package resources_test

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

const databaseResourceName = "thecloud_database.test"

func TestAccDatabaseResource(t *testing.T) {
	rName := acctest.RandStringFromCharSet(10, acctest.CharSetAlphaNum)
	dbName := fmt.Sprintf("test-db-%s", rName)

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create and Read testing
			{
				Config: providerConfig() + fmt.Sprintf(`
resource "thecloud_database" "test" {
  name    = "%s"
  engine  = "postgres"
  version = "14"
}
`, dbName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(databaseResourceName, "name", dbName),
					resource.TestCheckResourceAttr(databaseResourceName, "engine", "postgres"),
					resource.TestCheckResourceAttr(databaseResourceName, "version", "14"),
					resource.TestCheckResourceAttrSet(databaseResourceName, "id"),
					resource.TestCheckResourceAttrSet(databaseResourceName, "status"),
				),
			},
			// ImportState testing
			{
				ResourceName:      databaseResourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}
