package resources_test

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

const vpcResourceName = "thecloud_vpc.test"

func TestAccVpcResource(t *testing.T) {
	rName := acctest.RandStringFromCharSet(10, acctest.CharSetAlphaNum)
	vpcName := fmt.Sprintf("test-vpc-%s", rName)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create and Read testing
			{
				Config: providerConfig() + fmt.Sprintf(`
resource "thecloud_vpc" "test" {
  name       = "%s"
  cidr_block = "10.0.0.0/16"
}
`, vpcName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(vpcResourceName, "name", vpcName),
					resource.TestCheckResourceAttr(vpcResourceName, "cidr_block", "10.0.0.0/16"),
					resource.TestCheckResourceAttrSet(vpcResourceName, "id"),
					resource.TestCheckResourceAttr(vpcResourceName, "status", "active"),
				),
			},
			// ImportState testing
			{
				ResourceName:      vpcResourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}
