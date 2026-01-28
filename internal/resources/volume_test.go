package resources_test

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

const volumeResourceName = "thecloud_volume.test"

func TestAccVolumeResource(t *testing.T) {
	rName := acctest.RandStringFromCharSet(10, acctest.CharSetAlphaNum)
	volumeName := fmt.Sprintf("test-volume-%s", rName)

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create and Read testing
			{
				Config: providerConfig() + fmt.Sprintf(`
resource "thecloud_volume" "test" {
  name    = "%s"
  size_gb = 10
}
`, volumeName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(volumeResourceName, "name", volumeName),
					resource.TestCheckResourceAttr(volumeResourceName, "size_gb", "10"),
					resource.TestCheckResourceAttrSet(volumeResourceName, "id"),
					resource.TestCheckResourceAttr(volumeResourceName, "status", "AVAILABLE"),
				),
			},
			// ImportState testing
			{
				ResourceName:      volumeResourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}
