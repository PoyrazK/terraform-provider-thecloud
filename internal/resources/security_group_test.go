package resources_test

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

const sgResourceName = "thecloud_security_group.test"

func TestAccSecurityGroupResource(t *testing.T) {
	rName := acctest.RandStringFromCharSet(10, acctest.CharSetAlphaNum)
	vpcName := fmt.Sprintf("sg-test-vpc-%s", rName)
	sgName := fmt.Sprintf("test-sg-%s", rName)

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create and Read testing
			{
				Config: providerConfig() + fmt.Sprintf(`
resource "thecloud_vpc" "sg_vpc" {
  name       = "%s"
  cidr_block = "10.0.0.0/16"
}

resource "thecloud_security_group" "test" {
  name        = "%s"
  vpc_id      = thecloud_vpc.sg_vpc.id
  description = "test security group"
}
`, vpcName, sgName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(sgResourceName, "name", sgName),
					resource.TestCheckResourceAttr(sgResourceName, "description", "test security group"),
					resource.TestCheckResourceAttrSet(sgResourceName, "vpc_id"),
					resource.TestCheckResourceAttrSet(sgResourceName, "id"),
				),
			},
			// ImportState testing
			{
				ResourceName:      sgResourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}
