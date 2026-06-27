# Plane Terraform Provider API Mapping

Source of truth: live self-hosted OpenAPI schema fetched from
`https://plane.local.j-p.cloud/api/schema/` on 2026-06-27 (OpenAPI 3.0.3,
`title: The Plane REST API`, `version: 0.0.1`). All findings below are from the
**fetched schema**, not the public SaaS docs.

- **Auth:** `apiKey` in header, name `X-API-Key` (matches the brief).
- **Base path:** everything is under `/api/v1/`.
- **Pagination envelope** (cursor-based) on list endpoints:
  `results[]`, `next_cursor`, `prev_cursor`, `next_page_results`,
  `prev_page_results`, `total_count`, `count`, `total_pages`, `total_results`.
  List query params: `cursor`, `per_page`, `expand`, `fields`, `order_by`.
  Walk `next_cursor` while `next_page_results == true`.

## Scope reconciliation

The brief lists 6–7 resource areas. The live schema only supports **4** of them
as real Terraform resources. Three requested items do not exist in the
self-hosted v1 API:

| Brief item | Verdict | Evidence |
|---|---|---|
| 1. Workspace | No CRUD endpoint | `{slug}` is only ever a path parameter. There is no `/api/v1/workspaces/` collection and no `/api/v1/workspaces/{slug}/` detail endpoint. Only sub-collections exist (`/invitations/`, `/members/`, `/projects/`, `/assets/`, `/stickies/`). A workspace cannot be managed through this API. |
| 2. Project | Full CRUD | See below. |
| 3. Work Item States | Full CRUD | See below. |
| 4. Work Item Labels | Full CRUD | See below. |
| 5. Webhooks | Not exposed | Zero occurrences of `webhook` in the schema. |
| 6. Modules | Full CRUD | See below. |
| 7. General/SMTP settings | Not exposed | No `settings`, `smtp`, `email`, or `config` path or component schema. Self-hosted SMTP is instance-admin environment configuration (`EMAIL_HOST`, etc.), not a workspace-scoped API resource. |

Implemented managed resources: `plane_project`, `plane_state`, `plane_label`.
Planned but not yet implemented: `plane_module` (API exists, deferred).
Workspace CRUD, webhooks, and SMTP settings are out of scope for this API.

## Resource to endpoint mapping

All IDs are server-generated UUIDs (`id`, `readOnly`), so they are computed and
used for import. `slug` (workspace) and `project_id` are path parameters and
force replacement.

### plane_project
- Create: `POST   /api/v1/workspaces/{slug}/projects/` (body `ProjectCreateRequest`)
- Read:   `GET    /api/v1/workspaces/{slug}/projects/{pk}/`
- Update: `PATCH  /api/v1/workspaces/{slug}/projects/{pk}/` (body `PatchedProjectUpdateRequest`)
- Delete: `DELETE /api/v1/workspaces/{slug}/projects/{pk}/`
- Extra (out of scope, ignore): `POST/DELETE .../{project_id}/archive/`
- **Required (create):** `name` (<=255), `identifier` (<=12).
- Notable optional: `description`, `project_lead` (uuid, nullable),
  `default_assignee` (uuid, nullable), `network`, `emoji`/`icon_prop` (nullable),
  `module_view`/`cycle_view`/`issue_views_view`/`page_view`/`intake_view` (bool),
  `archive_in`/`close_in` (int), `timezone` (enum, **433 values**),
  `external_source`/`external_id`.
- **Mutability:** PATCH schema includes both `name` and `identifier`, so
  `identifier` is mutable per the schema. A live test is needed to confirm Plane
  honors an identifier change without breaking existing work-item keys.
- Import: `slug/project_id`.

### plane_state  (per project)
- Create: `POST   /api/v1/workspaces/{slug}/projects/{project_id}/states/` (body `StateRequest`)
- Read:   `GET    .../states/{state_id}/`
- Update: `PATCH  .../states/{state_id}/`
- Delete: `DELETE .../states/{state_id}/`
- **Required (create):** `name` (<=255), `color` (<=255).
- Optional: `description`, `sequence` (double), `group`
  (enum `GroupEnum`: backlog, unstarted, started, completed, cancelled, triage),
  `is_triage` (bool), `default` (bool), `external_source`/`external_id`.
- Computed: `id`, `slug`, `created_at`, `updated_at`, `created_by`,
  `updated_by`, `project`, `workspace`.
- Import: `slug/project_id/state_id`.

### plane_label  (per project)
- Create: `POST   .../{project_id}/labels/` (body `LabelCreateUpdateRequest`)
- Read/Update/Delete: `.../labels/{pk}/`
- **Required (create):** `name` (<=255).
- Optional: `color` (<=255), `description`, `parent` (uuid, nullable;
  self-reference for nested labels), `sort_order` (double),
  `external_source`/`external_id`.
- Import: `slug/project_id/label_id`.

### plane_module  (per project, not yet implemented)
- Create: `POST   .../{project_id}/modules/` (body `ModuleCreateRequest`)
- Read/Update/Delete: `.../modules/{pk}/`
- Extra (out of scope, ignore): `module-issues/*`, `archive/`, `archived-modules/*`.
- **Required (create):** `name` (<=255).
- Optional: `description`, `start_date`/`target_date` (date, nullable),
  `status` (enum `ModuleStatusEnum`: backlog, planned, in-progress, paused,
  completed, cancelled), `lead` (uuid, nullable), `members` (array of uuid),
  `external_source`/`external_id`.
- Computed: `id`, issue rollup counts (`total_issues`, `completed_issues`, etc.),
  timestamps, `project`, `workspace`.
- Import: `slug/project_id/module_id`.

## Ambiguities needing a live test (not resolvable from schema alone)
1. Project `identifier` mutability: schema allows PATCH; confirm Plane doesn't
   reject it or corrupt work-item keys.
2. State `default`: setting `default=true` on one state may implicitly unset it
   on another state (server-side side effect causing perpetual diff risk). Verify.
3. Module `members`: whether order is preserved / whether it's a set vs list
   (model as a set to be safe).
4. `external_source`/`external_id` appear on every resource and are likely for
   idempotent external-sync upserts; treat as optional, not managed by default.

## Codegen note
`terraform-plugin-codegen-openapi` would scaffold all 17 tags (Cycles, Pages,
Members, Stickies, Intake, etc.). Per the brief, only the 4 in-scope schemas
will be kept; everything else gets deleted rather than left half-wired.
