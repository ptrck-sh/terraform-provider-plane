package client

import (
	"context"
	"fmt"
	"net/http"
)

type Label struct {
	ID     string  `json:"id,omitempty"`
	Name   string  `json:"name,omitempty"`
	Color  string  `json:"color,omitempty"`
	Parent *string `json:"parent"`

	// Read-only.
	Project   string `json:"project,omitempty"`
	Workspace string `json:"workspace,omitempty"`
}

func labelsPath(slug, projectID string) string {
	return fmt.Sprintf("/api/v1/workspaces/%s/projects/%s/labels/", slug, projectID)
}

func labelPath(slug, projectID, id string) string {
	return fmt.Sprintf("/api/v1/workspaces/%s/projects/%s/labels/%s/", slug, projectID, id)
}

func (c *Client) CreateLabel(ctx context.Context, slug, projectID string, in Label) (*Label, error) {
	var out Label
	if err := c.do(ctx, http.MethodPost, labelsPath(slug, projectID), in, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

func (c *Client) GetLabel(ctx context.Context, slug, projectID, id string) (*Label, error) {
	var out Label
	if err := c.do(ctx, http.MethodGet, labelPath(slug, projectID, id), nil, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

func (c *Client) UpdateLabel(ctx context.Context, slug, projectID, id string, in Label) (*Label, error) {
	var out Label
	if err := c.do(ctx, http.MethodPatch, labelPath(slug, projectID, id), in, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

func (c *Client) DeleteLabel(ctx context.Context, slug, projectID, id string) error {
	return c.do(ctx, http.MethodDelete, labelPath(slug, projectID, id), nil, nil)
}

func (c *Client) ListLabels(ctx context.Context, slug, projectID string) ([]Label, error) {
	return listAll[Label](ctx, c, labelsPath(slug, projectID))
}
