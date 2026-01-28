package resources_test

import (
	"fmt"
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/providerserver"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
	"github.com/poyrazk/terraform-provider-thecloud/internal/provider"
)

var testAccProtoV6ProviderFactories = map[string]func() (tfprotov6.ProviderServer, error){
	"thecloud": providerserver.NewProtocol6WithError(provider.New("test")()),
}

func providerConfig() string {
	endpoint := os.Getenv("THECLOUD_ENDPOINT")
	if endpoint == "" {
		endpoint = "http://localhost:8080"
	}
	apiKey := os.Getenv("THECLOUD_API_KEY")
	if apiKey == "" {
		apiKey = "test-key"
	}

	return fmt.Sprintf(`
provider "thecloud" {
  endpoint = %q
  api_key  = %q
}
`, endpoint, apiKey)
}

func testAccPreCheck(t *testing.T) {
	if os.Getenv("THECLOUD_API_KEY") == "" {
		t.Fatal("THECLOUD_API_KEY must be set for acceptance tests")
	}
	if os.Getenv("THECLOUD_ENDPOINT") == "" {
		t.Fatal("THECLOUD_ENDPOINT must be set for acceptance tests")
	}
}
