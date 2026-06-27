package provider

import (
	"fmt"
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

// TestAccProjectResource exercises create, update, import, and destroy against a
// real Plane instance. It needs PLANE_WORKSPACE_SLUG in addition to the
// provider env vars and skips when any are unset.
func TestAccProjectResource(t *testing.T) {
	slug := os.Getenv("PLANE_WORKSPACE_SLUG")
	if slug == "" {
		t.Skip("PLANE_WORKSPACE_SLUG not set; skipping acceptance test")
	}

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccProjectConfig(slug, "Acc Test Project", "ACC", "first"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("plane_project.test", "name", "Acc Test Project"),
					resource.TestCheckResourceAttr("plane_project.test", "identifier", "ACC"),
					resource.TestCheckResourceAttr("plane_project.test", "description", "first"),
					resource.TestCheckResourceAttrSet("plane_project.test", "id"),
					resource.TestCheckResourceAttrSet("plane_project.test", "workspace"),
				),
			},
			{
				Config: testAccProjectConfig(slug, "Acc Test Project Renamed", "ACC", "second"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("plane_project.test", "name", "Acc Test Project Renamed"),
					resource.TestCheckResourceAttr("plane_project.test", "description", "second"),
				),
			},
			{
				ResourceName:      "plane_project.test",
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateIdFunc: func(s *terraform.State) (string, error) {
					rs, ok := s.RootModule().Resources["plane_project.test"]
					if !ok {
						return "", fmt.Errorf("plane_project.test not found in state")
					}
					return fmt.Sprintf("%s/%s", slug, rs.Primary.Attributes["id"]), nil
				},
			},
		},
	})
}

func testAccProjectConfig(slug, name, identifier, description string) string {
	return fmt.Sprintf(`
resource "plane_project" "test" {
  workspace_slug = %[1]q
  name           = %[2]q
  identifier     = %[3]q
  description    = %[4]q
}
`, slug, name, identifier, description)
}
