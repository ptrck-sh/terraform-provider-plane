package client

import (
	"context"
	"fmt"
	"net/http"
)

// Project maps the in-scope fields of a Plane project. Fields beyond the
// provider's schema (rollup counts, emoji, external_* sync keys, etc.) are
// intentionally omitted. See docs/api-mapping.md.
type Project struct {
	ID          string  `json:"id,omitempty"`
	Name        string  `json:"name,omitempty"`
	Identifier  string  `json:"identifier,omitempty"`
	Description string  `json:"description"`
	ProjectLead *string `json:"project_lead"`
	// DefaultAssignee is the default work-item assignee (user UUID).
	DefaultAssignee *string `json:"default_assignee"`
	ModuleView      *bool   `json:"module_view,omitempty"`
	CycleView       *bool   `json:"cycle_view,omitempty"`
	IssueViewsView  *bool   `json:"issue_views_view,omitempty"`
	PageView        *bool   `json:"page_view,omitempty"`
	IntakeView      *bool   `json:"intake_view,omitempty"`
	ArchiveIn       *int64  `json:"archive_in,omitempty"`
	CloseIn         *int64  `json:"close_in,omitempty"`
	Timezone        string  `json:"timezone,omitempty"`

	// Read-only.
	Workspace string `json:"workspace,omitempty"`
	CreatedAt string `json:"created_at,omitempty"`
	UpdatedAt string `json:"updated_at,omitempty"`
}

func projectsPath(slug string) string {
	return fmt.Sprintf("/api/v1/workspaces/%s/projects/", slug)
}

func projectPath(slug, id string) string {
	return fmt.Sprintf("/api/v1/workspaces/%s/projects/%s/", slug, id)
}

// CreateProject creates a project in the given workspace.
func (c *Client) CreateProject(ctx context.Context, slug string, in Project) (*Project, error) {
	var out Project
	if err := c.do(ctx, http.MethodPost, projectsPath(slug), in, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// GetProject fetches a project by ID. A 404 is returned as an *APIError.
func (c *Client) GetProject(ctx context.Context, slug, id string) (*Project, error) {
	var out Project
	if err := c.do(ctx, http.MethodGet, projectPath(slug, id), nil, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// UpdateProject patches a project. Only set fields are sent.
func (c *Client) UpdateProject(ctx context.Context, slug, id string, in Project) (*Project, error) {
	var out Project
	if err := c.do(ctx, http.MethodPatch, projectPath(slug, id), in, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// DeleteProject deletes a project by ID.
func (c *Client) DeleteProject(ctx context.Context, slug, id string) error {
	return c.do(ctx, http.MethodDelete, projectPath(slug, id), nil, nil)
}

// ListProjects returns every project in a workspace, walking pagination.
func (c *Client) ListProjects(ctx context.Context, slug string) ([]Project, error) {
	return listAll[Project](ctx, c, projectsPath(slug))
}
