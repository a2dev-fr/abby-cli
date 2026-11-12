package api

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
)

// ListContactsFilter defines query parameters for GET /contacts.
type ListContactsFilter struct {
	Page     int
	Limit    int
	Search   string
	Archived *bool
}

// ListContacts returns a paginated list of individual customers (contacts).
// Endpoint: GET /contacts
func (c *Client) ListContacts(ctx context.Context, filter ListContactsFilter) (*PaginatedResponse[ContactSummary], error) {
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

	var out PaginatedResponse[ContactSummary]
	if err := c.Do(ctx, http.MethodGet, "/contacts", q, nil, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// GetContact retrieves a contact by ID.
// Endpoint: GET /contact/{id}
func (c *Client) GetContact(ctx context.Context, id string) (*ContactSummary, error) {
	var out ContactSummary
	path := fmt.Sprintf("/contact/%s", url.PathEscape(id))
	if err := c.Do(ctx, http.MethodGet, path, nil, nil, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// CreateContact creates a new contact.
// Endpoint: POST /contact
func (c *Client) CreateContact(ctx context.Context, dto CreateContactDto) (*ContactSummary, error) {
	var out ContactSummary
	if err := c.Do(ctx, http.MethodPost, "/contact", nil, dto, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// UpdateContact updates an existing contact.
// Endpoint: PUT /contact/{id}
func (c *Client) UpdateContact(ctx context.Context, id string, dto CreateContactDto) (*ContactSummary, error) {
	var out ContactSummary
	path := fmt.Sprintf("/contact/%s", url.PathEscape(id))
	if err := c.Do(ctx, http.MethodPut, path, nil, dto, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// DeleteContact deletes a contact by ID.
// Endpoint: DELETE /contact/{id}
func (c *Client) DeleteContact(ctx context.Context, id string) error {
	path := fmt.Sprintf("/contact/%s", url.PathEscape(id))
	return c.Do(ctx, http.MethodDelete, path, nil, nil, nil)
}
