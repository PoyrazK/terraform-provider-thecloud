package resources_test

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

const lbTargetResourceName = "thecloud_load_balancer_target.test"

func TestAccLoadBalancerTargetResource(t *testing.T) {
	rName := acctest.RandStringFromCharSet(10, acctest.CharSetAlphaNum)
	vpcName := fmt.Sprintf("lbt-test-vpc-%s", rName)
	lbName := fmt.Sprintf("test-lb-%s", rName)
	instanceName := fmt.Sprintf("test-inst-%s", rName)

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create and Read testing
			{
				Config: providerConfig() + fmt.Sprintf(`
resource "thecloud_vpc" "lbt_vpc" {
  name       = "%s"
  cidr_block = "10.0.0.0/16"
}

resource "thecloud_load_balancer" "lbt_lb" {
  name   = "%s"
  vpc_id = thecloud_vpc.lbt_vpc.id
  port   = 80
}

resource "thecloud_instance" "lbt_inst" {
  name   = "%s"
  image  = "ubuntu-20.04"
  vpc_id = thecloud_vpc.lbt_vpc.id
}

resource "thecloud_load_balancer_target" "test" {
  load_balancer_id = thecloud_load_balancer.lbt_lb.id
  instance_id      = thecloud_instance.lbt_inst.id
  port             = 80
  weight           = 100
}
`, vpcName, lbName, instanceName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet(lbTargetResourceName, "id"),
					resource.TestCheckResourceAttr(lbTargetResourceName, "port", "80"),
					resource.TestCheckResourceAttr(lbTargetResourceName, "weight", "100"),
				),
			},
		},
	})
}
