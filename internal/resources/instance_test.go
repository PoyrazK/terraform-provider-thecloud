package resources_test

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccInstanceResource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create and Read testing
			{
				Config: providerConfig() + `
resource "thecloud_vpc" "inst_vpc" {
  name       = "inst-vpc"
  cidr_block = "10.0.0.0/16"
}

resource "thecloud_instance" "test" {
  name   = "test-instance"
  image  = "ubuntu-20.04"
  vpc_id = thecloud_vpc.inst_vpc.id
  ports  = "80:80"
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("thecloud_instance.test", "name", "test-instance"),
					resource.TestCheckResourceAttr("thecloud_instance.test", "image", "ubuntu-20.04"),
					resource.TestCheckResourceAttrSet("thecloud_instance.test", "vpc_id"),
					resource.TestCheckResourceAttrSet("thecloud_instance.test", "id"),
					resource.TestCheckResourceAttr("thecloud_instance.test", "status", "STARTING"),
				),
			},
			// ImportState testing
			{
				ResourceName:      "thecloud_instance.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}
