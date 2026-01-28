package resources_test

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

const secretResourceName = "thecloud_secret.test"

func TestAccSecretResource(t *testing.T) {
	rName := acctest.RandStringFromCharSet(10, acctest.CharSetAlphaNum)
	secretName := fmt.Sprintf("test-secret-%s", rName)
	secretValue := "super-secret-value"

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create and Read testing
			{
				Config: providerConfig() + fmt.Sprintf(`
resource "thecloud_secret" "test" {
  name        = "%s"
  value       = "%s"
  description = "test secret"
}
`, secretName, secretValue),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(secretResourceName, "name", secretName),
					resource.TestCheckResourceAttr(secretResourceName, "value", secretValue),
					resource.TestCheckResourceAttr(secretResourceName, "description", "test secret"),
					resource.TestCheckResourceAttrSet(secretResourceName, "id"),
				),
			},
			// ImportState testing
			{
				ResourceName:      secretResourceName,
				ImportState:       true,
				ImportStateVerify: true,
				// Ignore value since it's not returned by Read
				ImportStateVerifyIgnore: []string{"value"},
			},
		},
	})
}
