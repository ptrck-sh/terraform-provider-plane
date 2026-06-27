# Claude Instructions

Follow `AGENTS.md` first. It contains the shared repository policy for human and automated contributors.

When working in this repository:

- This is a `terraform-plugin-framework` provider. Use the framework (not SDKv2) idioms: `resource.Resource`, `datasource.DataSource`, typed schema models with `tfsdk` tags.
- Treat `docs/api-mapping.md` as the source of truth for endpoints and field semantics. If the live schema and this doc disagree, re-fetch the schema and fix the doc.
- Fix `Required` / `Optional` / `Computed` / `PlanModifiers` (e.g. `RequiresReplace`) by hand against the schema — never trust codegen defaults.
- Path-parameter fields (`workspace_slug`, `project_id`) are `RequiresReplace`.
- Generated docs under `docs/` are owned by `tfplugindocs` — do not hand-edit them; edit schema descriptions and `examples/` instead.
- Report validation results clearly, including intentional alint warnings.
