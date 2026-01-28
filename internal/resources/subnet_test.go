package resources_test

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

const subnetResourceName = "thecloud_subnet.test"

func TestAccSubnetResource(t *testing.T) {
	rName := acctest.RandStringFromCharSet(10, acctest.CharSetAlphaNum)
	vpcName := fmt.Sprintf("subnet-test-vpc-%s", rName)
	subnetName := fmt.Sprintf("test-subnet-%s", rName)

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create and Read testing
			{
				Config: providerConfig() + fmt.Sprintf(`
resource "thecloud_vpc" "subnet_vpc" {
  name       = "%s"
  cidr_block = "10.0.0.0/16"
}

resource "thecloud_subnet" "test" {
  vpc_id            = thecloud_vpc.subnet_vpc.id
  name              = "%s"
  cidr_block        = "10.0.1.0/24"
  availability_zone = "us-east-1a"
}
`, vpcName, subnetName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(subnetResourceName, "name", subnetName),
					resource.TestCheckResourceAttr(subnetResourceName, "cidr_block", "10.0.1.0/24"),
					resource.TestCheckResourceAttr(subnetResourceName, "availability_zone", "us-east-1a"),
					resource.TestCheckResourceAttrSet(subnetResourceName, "id"),
					resource.TestCheckResourceAttrSet(subnetResourceName, "vpc_id"),
				),
			},
			// ImportState testing
			{
				ResourceName:      subnetResourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}
