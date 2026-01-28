package resources_test

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

const apiKeyResourceName = "thecloud_api_key.test"

func TestAccApiKeyResource(t *testing.T) {
	rName := acctest.RandStringFromCharSet(10, acctest.CharSetAlphaNum)
	keyName := fmt.Sprintf("test-key-%s", rName)

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create and Read testing
			{
				Config: providerConfig() + fmt.Sprintf(`
resource "thecloud_api_key" "test" {
  name = "%s"
}
`, keyName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(apiKeyResourceName, "name", keyName),
					resource.TestCheckResourceAttrSet(apiKeyResourceName, "id"),
					resource.TestCheckResourceAttrSet(apiKeyResourceName, "key"),
					resource.TestCheckResourceAttrSet(apiKeyResourceName, "created_at"),
				),
			},
			// ImportState testing
			{
				ResourceName:      apiKeyResourceName,
				ImportState:       true,
				ImportStateVerify: true,
				// Key is not returned by Read (List)
				// created_at has precision mismatch
				ImportStateVerifyIgnore: []string{"key", "created_at"},
			},
		},
	})
}
