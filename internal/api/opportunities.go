package api

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
)

// OpportunityDto represents a CRM opportunity / deal in Abby (/v2/opportunity).
type OpportunityDto struct {
	ID          string `json:"id,omitempty"`
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
	DueDate     string `json:"dueDate,omitempty"` // YYYY-MM-DD
	Amount      int64  `json:"amount"`            // Amount in cents
	CustomerID  string `json:"customerId"`
	CategoryID  string `json:"categoryId"`
	AfterRank   string `json:"afterRank,omitempty"`
}

// OpportunityCategoryDto represents a Kanban column / category for opportunities.
type OpportunityCategoryDto struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

// ListOpportunityCategories retrieves the CRM opportunity categories.
// Endpoint: GET /v2/opportunity/categories
func (c *Client) ListOpportunityCategories(ctx context.Context) (json.RawMessage, error) {
	var raw json.RawMessage
	if err := c.Do(ctx, http.MethodGet, "/v2/opportunity/categories", nil, nil, &raw); err != nil {
		return nil, err
	}
	return raw, nil
}

// GetOpportunity retrieves a CRM opportunity by ID.
// Endpoint: GET /v2/opportunity/{id}
func (c *Client) GetOpportunity(ctx context.Context, id string) (json.RawMessage, error) {
	var raw json.RawMessage
	path := fmt.Sprintf("/v2/opportunity/%s", url.PathEscape(id))
	if err := c.Do(ctx, http.MethodGet, path, nil, nil, &raw); err != nil {
		return nil, err
	}
	return raw, nil
}

// CreateOpportunity creates a new CRM opportunity.
// Endpoint: POST /v2/opportunity
func (c *Client) CreateOpportunity(ctx context.Context, dto OpportunityDto) (json.RawMessage, error) {
	var raw json.RawMessage
	if err := c.Do(ctx, http.MethodPost, "/v2/opportunity", nil, dto, &raw); err != nil {
		return nil, err
	}
	return raw, nil
}

// UpdateOpportunity updates an existing CRM opportunity.
// Endpoint: PUT /v2/opportunity/{id}
func (c *Client) UpdateOpportunity(ctx context.Context, id string, dto OpportunityDto) (json.RawMessage, error) {
	var raw json.RawMessage
	path := fmt.Sprintf("/v2/opportunity/%s", url.PathEscape(id))
	if err := c.Do(ctx, http.MethodPut, path, nil, dto, &raw); err != nil {
		return nil, err
	}
	return raw, nil
}

// DeleteOpportunity deletes a CRM opportunity.
// Endpoint: DELETE /v2/opportunity/{id}
func (c *Client) DeleteOpportunity(ctx context.Context, id string) error {
	path := fmt.Sprintf("/v2/opportunity/%s", url.PathEscape(id))
	return c.Do(ctx, http.MethodDelete, path, nil, nil, nil)
}
