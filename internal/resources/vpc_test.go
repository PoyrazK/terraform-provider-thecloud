package resources_test

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/poyrazk/terraform-provider-thecloud/internal/provider"
)

func TestAccVpcResource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: provider.TestAccProtoV6ProviderFactories,
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
					resource.TestCheckResourceAttr("thecloud_vpc.test", "name", "test-vpc"),
					resource.TestCheckResourceAttr("thecloud_vpc.test", "cidr_block", "10.0.0.0/16"),
					resource.TestCheckResourceAttrSet("thecloud_vpc.test", "id"),
					resource.TestCheckResourceAttr("thecloud_vpc.test", "status", "available"),
				),
			},
			// ImportState testing
			{
				ResourceName:      "thecloud_vpc.test",
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
