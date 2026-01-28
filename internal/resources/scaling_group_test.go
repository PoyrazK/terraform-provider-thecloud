package resources_test

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

const scalingGroupResourceName = "thecloud_scaling_group.test"

func TestAccScalingGroupResource(t *testing.T) {
	rName := acctest.RandStringFromCharSet(10, acctest.CharSetAlphaNum)
	vpcName := fmt.Sprintf("sg-test-vpc-%s", rName)
	asgName := fmt.Sprintf("test-asg-%s", rName)

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create and Read testing
			{
				Config: providerConfig() + fmt.Sprintf(`
resource "thecloud_vpc" "asg_vpc" {
  name       = "%s"
  cidr_block = "10.0.0.0/16"
}

resource "thecloud_scaling_group" "test" {
  name          = "%s"
  vpc_id        = thecloud_vpc.asg_vpc.id
  image         = "ubuntu-20.04"
  min_instances = 1
  max_instances = 3
  desired_count = 2
}
`, vpcName, asgName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(scalingGroupResourceName, "name", asgName),
					resource.TestCheckResourceAttr(scalingGroupResourceName, "image", "ubuntu-20.04"),
					resource.TestCheckResourceAttr(scalingGroupResourceName, "min_instances", "1"),
					resource.TestCheckResourceAttr(scalingGroupResourceName, "max_instances", "3"),
					resource.TestCheckResourceAttr(scalingGroupResourceName, "desired_count", "2"),
					resource.TestCheckResourceAttrSet(scalingGroupResourceName, "id"),
					resource.TestCheckResourceAttrSet(scalingGroupResourceName, "status"),
				),
			},
			// ImportState testing
			{
				ResourceName:      scalingGroupResourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}
