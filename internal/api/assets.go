package api

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
)

// GetAsset retrieves a credit note (avoir / asset) by ID.
// Endpoint: GET /v2/billing/asset/{assetId}
func (c *Client) GetAsset(ctx context.Context, assetID string) (*BillingDocument, error) {
	var out BillingDocument
	path := fmt.Sprintf("/v2/billing/asset/%s", url.PathEscape(assetID))
	if err := c.Do(ctx, http.MethodGet, path, nil, nil, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// CreateBlankAsset creates a new blank draft credit note (avoir vierge) for a customerId.
// Endpoint: POST /v2/billing/asset/{customerId}
func (c *Client) CreateBlankAsset(ctx context.Context, customerID string) (*BillingDocument, error) {
	var out BillingDocument
	path := fmt.Sprintf("/v2/billing/asset/%s", url.PathEscape(customerID))
	if err := c.Do(ctx, http.MethodPost, path, nil, nil, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// CreateAssetFromAdvance creates a draft credit note (avoir) from a finalized advance invoice.
// Endpoint: POST /v2/billing/advance/{advanceId}/asset
func (c *Client) CreateAssetFromAdvance(ctx context.Context, advanceID string) (*BillingDocument, error) {
	var out BillingDocument
	path := fmt.Sprintf("/v2/billing/advance/%s/asset", url.PathEscape(advanceID))
	if err := c.Do(ctx, http.MethodPost, path, nil, nil, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// UpdateAssetGeneralInfo updates notes, terms, and payment methods on a draft credit note.
// Endpoint: PATCH /v2/billing/asset/{assetId}/general-informations
func (c *Client) UpdateAssetGeneralInfo(ctx context.Context, assetID string, dto UpdateGeneralInformationsDto) (*BillingDocument, error) {
	var out BillingDocument
	path := fmt.Sprintf("/v2/billing/asset/%s/general-informations", url.PathEscape(assetID))
	if err := c.Do(ctx, http.MethodPatch, path, nil, dto, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// UpdateAssetTimeline updates emittedAt, paymentDelay, etc. on a draft credit note (timestamps in seconds).
// Endpoint: PATCH /v2/billing/asset/{assetId}/timeline
func (c *Client) UpdateAssetTimeline(ctx context.Context, assetID string, dto UpdateTimelineDto) (*BillingDocument, error) {
	var out BillingDocument
	path := fmt.Sprintf("/v2/billing/asset/%s/timeline", url.PathEscape(assetID))
	if err := c.Do(ctx, http.MethodPatch, path, nil, dto, &out); err != nil {
		return nil, err
	}
	return &out, nil
}
