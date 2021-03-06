package ibm

import (
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/terraform"

	"github.com/IBM-Cloud/bluemix-go/models"
)

func TestAccIBMCertificateManagerOrder_Basic(t *testing.T) {
	var conf models.CertificateInfo
	name1 := fmt.Sprintf("tf-acc-test1-%s", acctest.RandStringFromCharSet(10, acctest.CharSetAlphaNum))
	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckIBMCertificateManagerOrderDestroy,
		Steps: []resource.TestStep{
			resource.TestStep{
				Config: testAccCheckIBMCertificateManagerOrder_basic(name1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckIBMCMOrderExists("ibm_certificate_manager_order.cert", conf),
					resource.TestCheckResourceAttr("ibm_certificate_manager_order.cert", "name", name1),
				),
			},
			resource.TestStep{
				ResourceName:      "ibm_certificate_manager_order.cert",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func testAccCheckIBMCertificateManagerOrderDestroy(s *terraform.State) error {
	time.Sleep(100 * time.Second)
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "ibm_certificate_manager_order" {
			continue
		}
		certID := rs.Primary.ID
		cmClient, err := testAccProvider.Meta().(ClientSession).CertificateManagerAPI()
		if err != nil {
			return err
		}
		certAPI := cmClient.Certificate()
		_, err = certAPI.GetCertData(certID)

		if err != nil && !strings.Contains(err.Error(), "404") {
			return fmt.Errorf("Error checking if instance (%s) has been destroyed: %s", rs.Primary.ID, err)
		}
	}
	return nil

}

func testAccCheckIBMCertificateManagerOrder_basic(name1 string) string {
	return fmt.Sprintf(`
	resource "ibm_certificate_manager_order" "cert" {
		certificate_manager_instance_id = ibm_resource_instance.cm.id
		name                            = "%s"
		description                     = "test description"
		domains                         = ["example.com"]
		rotate_keys                     = false
		domain_validation_method        = "dns-01"
		dns_provider_instance_crn       = ibm_cis.instance.id
	  }
	  
	  `, name1)
}

func testAccCheckIBMCMOrderExists(n string, obj models.CertificateInfo) resource.TestCheckFunc {

	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		cmClient, err := testAccProvider.Meta().(ClientSession).CertificateManagerAPI()
		if err != nil {
			return err
		}
		certID := rs.Primary.ID

		crt, err := cmClient.Certificate().GetMetaData(certID)
		if err != nil {
			return err
		}

		obj = crt
		return nil
	}
}
