package resources_test

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

const vpcResourceName = "thecloud_vpc.test"

func TestAccVpcResource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create and Read testing
			{
				Config: providerConfig() + `
resource "thecloud_vpc" "test" {
  name       = "test-vpc"
  cidr_block = "10.0.0.0/16"
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(vpcResourceName, "name", "test-vpc"),
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

func providerConfig() string {
	return `
provider "thecloud" {
  endpoint = "http://localhost:8080"
  api_key  = "test-key"
}
`
}
