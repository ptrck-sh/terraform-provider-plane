package provider

import (
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/providerserver"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
)

// testAccProtoV6ProviderFactories wires the provider under test for acceptance
// tests run through terraform-plugin-testing.
var testAccProtoV6ProviderFactories = map[string]func() (tfprotov6.ProviderServer, error){
	"plane": providerserver.NewProtocol6WithError(New("test")()),
}

// testAccPreCheck skips acceptance tests gracefully unless a live Plane instance
// is configured. These tests never run against a mock.
func testAccPreCheck(t *testing.T) {
	t.Helper()
	for _, env := range []string{"PLANE_HOST", "PLANE_API_KEY"} {
		if os.Getenv(env) == "" {
			t.Skipf("%s not set; skipping acceptance test (set TF_ACC + PLANE_HOST + PLANE_API_KEY to run)", env)
		}
	}
}
