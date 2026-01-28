package resources_test

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

const instanceResourceName = "thecloud_instance.test"

func TestAccInstanceResource(t *testing.T) {
	rName := acctest.RandStringFromCharSet(10, acctest.CharSetAlphaNum)
	vpcName := fmt.Sprintf("inst-vpc-%s", rName)
	instanceName := fmt.Sprintf("test-instance-%s", rName)

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create and Read testing
			{
				Config: providerConfig() + fmt.Sprintf(`
resource "thecloud_vpc" "inst_vpc" {
  name       = "%s"
  cidr_block = "10.0.0.0/16"
}

resource "thecloud_instance" "test" {
  name   = "%s"
  image  = "ubuntu-20.04"
  vpc_id = thecloud_vpc.inst_vpc.id
  ports  = "80:80"
}
`, vpcName, instanceName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(instanceResourceName, "name", instanceName),
					resource.TestCheckResourceAttr(instanceResourceName, "image", "ubuntu-20.04"),
					resource.TestCheckResourceAttrSet(instanceResourceName, "vpc_id"),
					resource.TestCheckResourceAttrSet(instanceResourceName, "id"),
					resource.TestCheckResourceAttr(instanceResourceName, "status", "STARTING"),
				),
			},
			// ImportState testing
			{
				ResourceName:      instanceResourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}
