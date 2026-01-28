package resources_test

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

const snapshotResourceName = "thecloud_snapshot.test"

func TestAccSnapshotResource(t *testing.T) {
	rName := acctest.RandStringFromCharSet(10, acctest.CharSetAlphaNum)
	volName := fmt.Sprintf("test-vol-%s", rName)
	snapshotDesc := fmt.Sprintf("test-snapshot-%s", rName)

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create and Read testing
			{
				Config: providerConfig() + fmt.Sprintf(`
resource "thecloud_volume" "snapshot_vol" {
  name    = "%s"
  size_gb = 10
}

resource "thecloud_snapshot" "test" {
  volume_id   = thecloud_volume.snapshot_vol.id
  description = "%s"
}
`, volName, snapshotDesc),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(snapshotResourceName, "description", snapshotDesc),
					resource.TestCheckResourceAttrSet(snapshotResourceName, "id"),
					resource.TestCheckResourceAttrSet(snapshotResourceName, "volume_id"),
					resource.TestCheckResourceAttrSet(snapshotResourceName, "status"),
				),
			},
			// ImportState testing
			{
				ResourceName:      snapshotResourceName,
				ImportState:       true,
				ImportStateVerify: true,
				// status might transition from CREATING to AVAILABLE
				ImportStateVerifyIgnore: []string{"status"},
			},
		},
	})
}
