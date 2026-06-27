package client

import (
	"context"
	"fmt"
	"net/http"
)

type State struct {
	ID      string `json:"id,omitempty"`
	Name    string `json:"name,omitempty"`
	Color   string `json:"color,omitempty"`
	Group   string `json:"group,omitempty"`
	Default bool   `json:"default"`

	// Read-only.
	Project   string `json:"project,omitempty"`
	Workspace string `json:"workspace,omitempty"`
}

func statesPath(slug, projectID string) string {
	return fmt.Sprintf("/api/v1/workspaces/%s/projects/%s/states/", slug, projectID)
}

func statePath(slug, projectID, id string) string {
	return fmt.Sprintf("/api/v1/workspaces/%s/projects/%s/states/%s/", slug, projectID, id)
}

func (c *Client) CreateState(ctx context.Context, slug, projectID string, in State) (*State, error) {
	var out State
	if err := c.do(ctx, http.MethodPost, statesPath(slug, projectID), in, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

func (c *Client) GetState(ctx context.Context, slug, projectID, id string) (*State, error) {
	var out State
	if err := c.do(ctx, http.MethodGet, statePath(slug, projectID, id), nil, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

func (c *Client) UpdateState(ctx context.Context, slug, projectID, id string, in State) (*State, error) {
	var out State
	if err := c.do(ctx, http.MethodPatch, statePath(slug, projectID, id), in, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

func (c *Client) DeleteState(ctx context.Context, slug, projectID, id string) error {
	return c.do(ctx, http.MethodDelete, statePath(slug, projectID, id), nil, nil)
}

func (c *Client) ListStates(ctx context.Context, slug, projectID string) ([]State, error) {
	return listAll[State](ctx, c, statesPath(slug, projectID))
}
