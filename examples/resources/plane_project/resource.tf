resource "plane_project" "platform" {
  workspace_slug = "my-workspace"
  name           = "Platform"
  identifier     = "PLAT"
  description    = "Platform engineering work"

  cycle_view  = true
  module_view = true
}
