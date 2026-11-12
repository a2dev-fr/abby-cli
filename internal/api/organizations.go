package api

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
)

// ListOrganizationsFilter defines query parameters for GET /organizations.
type ListOrganizationsFilter struct {
	Page     int
	Limit    int
	Search   string
	Archived *bool
}

// ListOrganizations returns a paginated list of professional customers (organizations).
// Endpoint: GET /organizations
func (c *Client) ListOrganizations(ctx context.Context, filter ListOrganizationsFilter) (*PaginatedResponse[OrganizationSummary], error) {
	q := url.Values{}
	if filter.Page > 0 {
		q.Set("page", strconv.Itoa(filter.Page))
	} else {
		q.Set("page", "1")
	}
	if filter.Limit > 0 {
		q.Set("limit", strconv.Itoa(filter.Limit))
	} else {
		q.Set("limit", "20")
	}
	if filter.Search != "" {
		q.Set("search", filter.Search)
	}
	if filter.Archived != nil {
		q.Set("archived", strconv.FormatBool(*filter.Archived))
	}

	var out PaginatedResponse[OrganizationSummary]
	if err := c.Do(ctx, http.MethodGet, "/organizations", q, nil, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// GetOrganization retrieves an organization by ID.
// Endpoint: GET /organization/{id}
func (c *Client) GetOrganization(ctx context.Context, id string) (*OrganizationSummary, error) {
	var out OrganizationSummary
	path := fmt.Sprintf("/organization/%s", url.PathEscape(id))
	if err := c.Do(ctx, http.MethodGet, path, nil, nil, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// CreateOrganization creates a new professional customer (organization).
// Endpoint: POST /organization
func (c *Client) CreateOrganization(ctx context.Context, dto CreateOrganizationDto) (*OrganizationSummary, error) {
	var out OrganizationSummary
	if err := c.Do(ctx, http.MethodPost, "/organization", nil, dto, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// UpdateOrganization updates an existing organization.
// Endpoint: PUT /organization/{id}
func (c *Client) UpdateOrganization(ctx context.Context, id string, dto CreateOrganizationDto) (*OrganizationSummary, error) {
	var out OrganizationSummary
	path := fmt.Sprintf("/organization/%s", url.PathEscape(id))
	if err := c.Do(ctx, http.MethodPut, path, nil, dto, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// DeleteOrganization deletes an organization by ID.
// Endpoint: DELETE /organization/{organizationId}
func (c *Client) DeleteOrganization(ctx context.Context, organizationID string) error {
	path := fmt.Sprintf("/organization/%s", url.PathEscape(organizationID))
	return c.Do(ctx, http.MethodDelete, path, nil, nil, nil)
}

// ListOrganizationContacts retrieves contacts associated with an organization.
// Endpoint: GET /organization/{id}/contacts
func (c *Client) ListOrganizationContacts(ctx context.Context, organizationID string) ([]ContactSummary, error) {
	var out []ContactSummary
	path := fmt.Sprintf("/organization/%s/contacts", url.PathEscape(organizationID))
	if err := c.Do(ctx, http.MethodGet, path, nil, nil, &out); err != nil {
		return nil, err
	}
	return out, nil
}

// CreateOrganizationContact creates a new contact inside an organization.
// Endpoint: POST /organization/{id}/contact
func (c *Client) CreateOrganizationContact(ctx context.Context, organizationID string, dto CreateContactDto) (*ContactSummary, error) {
	var out ContactSummary
	path := fmt.Sprintf("/organization/%s/contact", url.PathEscape(organizationID))
	if err := c.Do(ctx, http.MethodPost, path, nil, dto, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// SetOrganizationDefaultContact sets the default contact of an organization.
// Endpoint: PATCH /organization/{organizationId}/default-contact
func (c *Client) SetOrganizationDefaultContact(ctx context.Context, organizationID, contactID string) error {
	path := fmt.Sprintf("/organization/%s/default-contact", url.PathEscape(organizationID))
	body := map[string]string{"contactId": contactID}
	return c.Do(ctx, http.MethodPatch, path, nil, body, nil)
}
