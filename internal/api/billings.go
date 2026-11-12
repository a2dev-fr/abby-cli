package api

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
)

// ListBillingsFilter defines query parameters for GET /v2/billings.
type ListBillingsFilter struct {
	Page           int
	Limit          int
	Search         string
	Types          []BillingType
	States         []BillingState
	CustomerID     string
	OrganizationID string
	Test           bool
	Archived       *bool
	Late           *bool
	FromDate       string // YYYY-MM-DD
	ToDate         string // YYYY-MM-DD
}

// ListBillings lists billing documents across types (invoice, estimate, asset, advance, purchase_order).
// Endpoint: GET /v2/billings
func (c *Client) ListBillings(ctx context.Context, filter ListBillingsFilter) (*PaginatedResponse[BillingDocument], error) {
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
	q.Set("test", strconv.FormatBool(filter.Test))

	if filter.Search != "" {
		q.Set("search", filter.Search)
	}
	for _, t := range filter.Types {
		q.Add("type", string(t))
	}
	for _, s := range filter.States {
		q.Add("state", string(s))
	}
	if filter.CustomerID != "" {
		q.Set("customerId", filter.CustomerID)
	}
	if filter.OrganizationID != "" {
		q.Set("organizationId", filter.OrganizationID)
	}
	if filter.Archived != nil {
		q.Set("archived", strconv.FormatBool(*filter.Archived))
	}
	if filter.Late != nil {
		q.Set("late", strconv.FormatBool(*filter.Late))
	}
	if filter.FromDate != "" && filter.ToDate != "" {
		q.Add("range", filter.FromDate)
		q.Add("range", filter.ToDate)
	}

	var out PaginatedResponse[BillingDocument]
	if err := c.Do(ctx, http.MethodGet, "/v2/billings", q, nil, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// GetBillingByID retrieves any billing document by its ID.
// Endpoint: GET /v2/billing/{documentId}
func (c *Client) GetBillingByID(ctx context.Context, documentID string) (*BillingDocument, error) {
	var out BillingDocument
	path := fmt.Sprintf("/v2/billing/%s", url.PathEscape(documentID))
	if err := c.Do(ctx, http.MethodGet, path, nil, nil, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// DeleteBillingDocument permanently deletes a draft billing document.
// Endpoint: DELETE /v2/billing/{documentId}
func (c *Client) DeleteBillingDocument(ctx context.Context, documentID string) error {
	path := fmt.Sprintf("/v2/billing/%s", url.PathEscape(documentID))
	return c.Do(ctx, http.MethodDelete, path, nil, nil, nil)
}

// UpdateBillingLines replaces the billing lines on a draft document (invoice, estimate, or asset).
// Endpoint: PATCH /v2/billing/{billingId}/lines
func (c *Client) UpdateBillingLines(ctx context.Context, billingID string, dto UpdateBillingLinesDto) (*BillingDocument, error) {
	var out BillingDocument
	path := fmt.Sprintf("/v2/billing/%s/lines", url.PathEscape(billingID))
	if err := c.Do(ctx, http.MethodPatch, path, nil, dto, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// FinalizeBilling finalizes a draft billing document (assigns official document number and locks it).
// Endpoint: PATCH /v2/billing/{billingId}/finalize
func (c *Client) FinalizeBilling(ctx context.Context, billingID string) (*BillingDocument, error) {
	var out BillingDocument
	path := fmt.Sprintf("/v2/billing/%s/finalize", url.PathEscape(billingID))
	if err := c.Do(ctx, http.MethodPatch, path, nil, nil, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// DownloadBillingPDF downloads the PDF binary of a billing document.
// Endpoint: GET /v2/billing/{billingId}/download
func (c *Client) DownloadBillingPDF(ctx context.Context, billingID string) ([]byte, string, error) {
	path := fmt.Sprintf("/v2/billing/%s/download", url.PathEscape(billingID))
	return c.Download(ctx, path)
}

// SendBillingByEmail sends a finalized billing document by email.
// Endpoint: POST /v2/billing/{billingId}/send-by-email
func (c *Client) SendBillingByEmail(ctx context.Context, billingID string, dto SendBillingByEmailDto) error {
	path := fmt.Sprintf("/v2/billing/%s/send-by-email", url.PathEscape(billingID))
	return c.Do(ctx, http.MethodPost, path, nil, dto, nil)
}

// ArchiveBilling archives a billing document.
// Endpoint: PATCH /v2/billing/{documentId}/archive
func (c *Client) ArchiveBilling(ctx context.Context, documentID string) error {
	path := fmt.Sprintf("/v2/billing/%s/archive", url.PathEscape(documentID))
	return c.Do(ctx, http.MethodPatch, path, nil, nil, nil)
}

// UnarchiveBilling unarchives a billing document.
// Endpoint: PATCH /v2/billing/{documentId}/unarchive
func (c *Client) UnarchiveBilling(ctx context.Context, documentID string) error {
	path := fmt.Sprintf("/v2/billing/%s/unarchive", url.PathEscape(documentID))
	return c.Do(ctx, http.MethodPatch, path, nil, nil, nil)
}

// UpdateBillingTitle updates the title of a billing document.
// Endpoint: PATCH /v2/billing/{documentId}/title
func (c *Client) UpdateBillingTitle(ctx context.Context, documentID string, title string) (*BillingDocument, error) {
	var out BillingDocument
	path := fmt.Sprintf("/v2/billing/%s/title", url.PathEscape(documentID))
	t := title
	if err := c.Do(ctx, http.MethodPatch, path, nil, UpdateTitleDto{Title: &t}, &out); err != nil {
		return nil, err
	}
	return &out, nil
}
