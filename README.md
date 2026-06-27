# terraform-provider-plane

A Terraform/OpenTofu provider for a **self-hosted** [Plane](https://plane.so)
instance (`makeplane/plane`). It manages Plane resources through the REST API
(`X-API-Key` authentication).

> Built against the live OpenAPI schema of a self-hosted instance, not the SaaS
> docs. The endpoint/field mapping is documented in
> [`docs/api-mapping.md`](docs/api-mapping.md).

## Supported resources

The self-hosted v1 API only exposes a subset of Plane concepts as managed
resources. This provider deliberately implements exactly:

| Type | Kind | Notes |
|---|---|---|
| `plane_project` | resource | Full CRUD, per workspace |
| `plane_state` | resource | Work-item states, per project |
| `plane_label` | resource | Work-item labels, per project |
| `plane_workspace` | data source | Resolve a workspace by slug |

**Not supported, and why:** workspace CRUD, webhooks, and SMTP/email settings
have **no endpoint** in the self-hosted v1 API (SMTP is instance-level env-var
configuration). See `docs/api-mapping.md`.

## Usage

```hcl
terraform {
  required_providers {
    plane = {
      source = "ptrck-sh/plane"
    }
  }
}

provider "plane" {
  host    = "https://plane.example.com" # or PLANE_HOST
  api_key = var.plane_api_key           # or PLANE_API_KEY (sensitive)
}

resource "plane_project" "example" {
  workspace_slug = "my-workspace"
  name           = "Platform"
  identifier     = "PLAT"
}
```

### Pointing at a self-hosted instance

`host` is the base URL of your Plane deployment (e.g.
`https://plane.example.com`). The provider appends `/api/v1/...` paths itself.

### Generating an API key

In Plane, open **Profile settings**, then **Personal access tokens**. Create a token and
pass it as `api_key` (or export `PLANE_API_KEY`). Tokens are sent in the
`X-API-Key` header.

## Development

```sh
mise install          # Go, goreleaser, tfplugindocs, golangci-lint
mise run build        # go build
mise run test         # unit tests
mise run testacc      # acceptance tests (needs PLANE_HOST + PLANE_API_KEY)
mise run docs         # regenerate docs/ via tfplugindocs
```

Acceptance tests run against a real instance and skip when `PLANE_HOST` /
`PLANE_API_KEY` are unset.

## Releases & mirroring

Releases are cut by [GoReleaser](https://goreleaser.com) on Git tags
(`vX.Y.Z`), producing GPG-signed, registry-compatible artifacts.

This repository is canonical on GitLab and mirrored to GitHub. The mirror and
Terraform/OpenTofu Registry publishing are wired separately (see the `TODO` in
the infrastructure `gitlab/` IaC); this repo only produces the release
artifacts.

## License

See [`LICENSE`](LICENSE).
