package resources_test

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/poyrazk/terraform-provider-thecloud/internal/provider"
)

func TestAccLoadBalancerResource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: provider.TestAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: providerConfig() + `
resource "thecloud_vpc" "lb_vpc" {
  name = "lb-test-vpc"
}

resource "thecloud_load_balancer" "test" {
  name   = "test-lb"
  vpc_id = thecloud_vpc.lb_vpc.id
  port   = 80
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("thecloud_load_balancer.test", "name", "test-lb"),
					resource.TestCheckResourceAttr("thecloud_load_balancer.test", "port", "80"),
					resource.TestCheckResourceAttrSet("thecloud_load_balancer.test", "id"),
				),
			},
		},
	})
}
