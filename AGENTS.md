# Agent Instructions

This repository follows the homelab project baseline. Treat these files as durable project policy, not as generated boilerplate.

## What this is

A Terraform/OpenTofu provider for a self-hosted [Plane](https://plane.so) instance, built with `terraform-plugin-framework`. It wraps the Plane REST API (`X-API-Key` auth). The authoritative API mapping lives in `docs/api-mapping.md` — derived from the live OpenAPI schema, not the public SaaS docs.

## Working Rules

- Keep changes small and scoped. Prefer existing CI, pre-commit, Renovate, and alint conventions over new structure.
- Preserve `README.md`, `AGENTS.md`, `CLAUDE.md`, `.gitlab-ci.yml`, `.pre-commit-config.yaml`, `.yamllint.yaml`, `.alint.yml`, `.gitignore`, `.editorconfig`, and `mise.toml` unless there is a documented exception.
- Use tagged `ci-templates` components and tagged `project-template-default` alint policy.
- The API client (`internal/client/`) is hand-written and intentionally small. Match field requiredness, nullability, and `readOnly` markers to the OpenAPI schema, not to prose docs.
- **No polling loops.** Pagination may walk `next_cursor` to completion within a single Read. Anything async on Plane's side must be surfaced, never papered over with a retry-until-ready loop.
- Do not commit secrets, `.env` files, `mise.local.toml`, Terraform/OpenTofu state, or `dist/`.

## Scope

In-scope resources (the self-hosted v1 API only supports these): `plane_project`, `plane_state`, `plane_label`, `plane_module`, plus the `plane_workspace` data source. Workspace CRUD, webhooks, and SMTP/settings are **not** exposed by the API — see `docs/api-mapping.md` for the evidence. Do not scaffold out-of-scope resources.

## Validation

```sh
gofmt -l .          # must be empty
go vet ./...
go test ./... -race -cover
pre-commit run --all-files
alint check
```

Acceptance tests (`TestAcc*`) hit a real Plane instance, gated on `TF_ACC=1` and `PLANE_HOST`/`PLANE_API_KEY`. They skip gracefully when unset and must never run unauthenticated in default CI.
