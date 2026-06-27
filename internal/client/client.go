// Package client is a small hand-written client for the Plane REST API.
// It covers exactly the endpoints this provider manages; see docs/api-mapping.md.
package client

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
)

const userAgent = "terraform-provider-plane"

// Client talks to a self-hosted Plane instance using X-API-Key auth.
type Client struct {
	host       string
	apiKey     string
	httpClient *http.Client
}

// New returns a Client. host is the instance base URL (no trailing /api/v1).
// If httpClient is nil, http.DefaultClient is used.
func New(host, apiKey string, httpClient *http.Client) *Client {
	if httpClient == nil {
		httpClient = http.DefaultClient
	}
	return &Client{
		host:       strings.TrimRight(host, "/"),
		apiKey:     apiKey,
		httpClient: httpClient,
	}
}

// APIError is returned for any non-2xx response.
type APIError struct {
	StatusCode int
	Method     string
	Path       string
	Body       string
}

func (e *APIError) Error() string {
	return fmt.Sprintf("plane API %s %s: status %d: %s", e.Method, e.Path, e.StatusCode, e.Body)
}

// NotFound reports whether the error is a 404, used to drop resources from state.
func (e *APIError) NotFound() bool { return e.StatusCode == http.StatusNotFound }

// IsNotFound reports whether err is an *APIError with a 404 status.
func IsNotFound(err error) bool {
	var apiErr *APIError
	return errors.As(err, &apiErr) && apiErr.NotFound()
}

// do performs an API request. path is relative to the host (e.g. "/api/v1/...").
// body is JSON-encoded when non-nil; out is JSON-decoded when non-nil and a
// response body is present.
func (c *Client) do(ctx context.Context, method, path string, body, out any) error {
	var reader io.Reader
	if body != nil {
		b, err := json.Marshal(body)
		if err != nil {
			return fmt.Errorf("encoding request body: %w", err)
		}
		reader = bytes.NewReader(b)
	}

	req, err := http.NewRequestWithContext(ctx, method, c.host+path, reader)
	if err != nil {
		return fmt.Errorf("building request: %w", err)
	}
	req.Header.Set("X-API-Key", c.apiKey)
	req.Header.Set("Accept", "application/json")
	req.Header.Set("User-Agent", userAgent)
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("performing request: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("reading response body: %w", err)
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return &APIError{
			StatusCode: resp.StatusCode,
			Method:     method,
			Path:       path,
			Body:       strings.TrimSpace(string(respBody)),
		}
	}

	if out != nil && len(respBody) > 0 {
		if err := json.Unmarshal(respBody, out); err != nil {
			return fmt.Errorf("decoding response body: %w", err)
		}
	}
	return nil
}

// page is the cursor-pagination envelope Plane returns on list endpoints.
type page[T any] struct {
	Results         []T    `json:"results"`
	NextCursor      string `json:"next_cursor"`
	NextPageResults bool   `json:"next_page_results"`
}

// listAll walks cursor pagination to completion, following next_cursor while
// next_page_results is true. It does not poll or retry.
func listAll[T any](ctx context.Context, c *Client, basePath string) ([]T, error) {
	var all []T
	cursor := ""
	for {
		path := basePath
		if cursor != "" {
			sep := "?"
			if strings.Contains(path, "?") {
				sep = "&"
			}
			path = path + sep + "cursor=" + url.QueryEscape(cursor)
		}

		var pg page[T]
		if err := c.do(ctx, http.MethodGet, path, nil, &pg); err != nil {
			return nil, err
		}
		all = append(all, pg.Results...)

		if !pg.NextPageResults || pg.NextCursor == "" {
			return all, nil
		}
		cursor = pg.NextCursor
	}
}
