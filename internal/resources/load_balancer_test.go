package resources_test

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

const lbResourceName = "thecloud_load_balancer.test"

func TestAccLoadBalancerResource(t *testing.T) {
	rName := acctest.RandStringFromCharSet(10, acctest.CharSetAlphaNum)
	vpcName := fmt.Sprintf("lb-test-vpc-%s", rName)
	lbName := fmt.Sprintf("test-lb-%s", rName)

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: providerConfig() + fmt.Sprintf(`
resource "thecloud_vpc" "lb_vpc" {
  name       = "%s"
  cidr_block = "10.0.0.0/16"
}

resource "thecloud_load_balancer" "test" {
  name      = "%s"
  vpc_id    = thecloud_vpc.lb_vpc.id
  port      = 80
  algorithm = "round-robin"
}
`, vpcName, lbName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(lbResourceName, "name", lbName),
					resource.TestCheckResourceAttr(lbResourceName, "port", "80"),
					resource.TestCheckResourceAttrSet(lbResourceName, "id"),
				),
			},
		},
	})
}
