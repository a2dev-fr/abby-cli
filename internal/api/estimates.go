package api

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
)

// CreateEstimate creates a new draft estimate (devis) or purchase order (bon de commande) for a customerId.
// Endpoint: POST /v2/billing/estimate/{customerId}
func (c *Client) CreateEstimate(ctx context.Context, customerID string, estimateType string) (*BillingDocument, error) {
	var out BillingDocument
	path := fmt.Sprintf("/v2/billing/estimate/%s", url.PathEscape(customerID))
	dto := CreateEstimateDto{EstimateType: estimateType}
	if dto.EstimateType == "" {
		dto.EstimateType = "estimate"
	}
	if err := c.Do(ctx, http.MethodPost, path, nil, dto, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// UpdateEstimateGeneralInfo updates notes, terms, and payment methods on a draft estimate.
// Endpoint: PATCH /v2/billing/estimate/{estimateId}/general-informations
func (c *Client) UpdateEstimateGeneralInfo(ctx context.Context, estimateID string, dto UpdateGeneralInformationsDto) (*BillingDocument, error) {
	var out BillingDocument
	path := fmt.Sprintf("/v2/billing/estimate/%s/general-informations", url.PathEscape(estimateID))
	if err := c.Do(ctx, http.MethodPatch, path, nil, dto, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// UpdateEstimateTimeline updates emittedAt, expiredAt, paymentDelay, etc. (timestamps in seconds).
// Endpoint: PATCH /v2/billing/estimate/{estimateId}/timeline
func (c *Client) UpdateEstimateTimeline(ctx context.Context, estimateID string, dto UpdateTimelineDto) (*BillingDocument, error) {
	var out BillingDocument
	path := fmt.Sprintf("/v2/billing/estimate/%s/timeline", url.PathEscape(estimateID))
	if err := c.Do(ctx, http.MethodPatch, path, nil, dto, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// SignEstimate marks a finalized estimate as signed by the customer.
// Endpoint: PATCH /v2/billing/estimate/{estimateId}/sign
func (c *Client) SignEstimate(ctx context.Context, estimateID string) (*BillingDocument, error) {
	var out BillingDocument
	path := fmt.Sprintf("/v2/billing/estimate/%s/sign", url.PathEscape(estimateID))
	if err := c.Do(ctx, http.MethodPatch, path, nil, nil, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// UnsignEstimate reverts the signature on an estimate.
// Endpoint: PATCH /v2/billing/estimate/{estimateId}/unsign
func (c *Client) UnsignEstimate(ctx context.Context, estimateID string) (*BillingDocument, error) {
	var out BillingDocument
	path := fmt.Sprintf("/v2/billing/estimate/%s/unsign", url.PathEscape(estimateID))
	if err := c.Do(ctx, http.MethodPatch, path, nil, nil, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// RefuseEstimate marks a finalized estimate as refused.
// Endpoint: PATCH /v2/billing/estimate/{estimateId}/refuse
func (c *Client) RefuseEstimate(ctx context.Context, estimateID string) (*BillingDocument, error) {
	var out BillingDocument
	path := fmt.Sprintf("/v2/billing/estimate/%s/refuse", url.PathEscape(estimateID))
	if err := c.Do(ctx, http.MethodPatch, path, nil, nil, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// UnrefuseEstimate reverts the refusal on an estimate.
// Endpoint: PATCH /v2/billing/estimate/{estimateId}/unrefuse
func (c *Client) UnrefuseEstimate(ctx context.Context, estimateID string) (*BillingDocument, error) {
	var out BillingDocument
	path := fmt.Sprintf("/v2/billing/estimate/%s/unrefuse", url.PathEscape(estimateID))
	if err := c.Do(ctx, http.MethodPatch, path, nil, nil, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// CreateAdvanceFromEstimate creates a draft advance invoice (facture d'acompte) from a signed estimate.
// Endpoint: POST /v2/billing/estimate/{id}/advance
func (c *Client) CreateAdvanceFromEstimate(ctx context.Context, estimateID string, dto *CreateAdvanceDto) (*BillingDocument, error) {
	var out BillingDocument
	path := fmt.Sprintf("/v2/billing/estimate/%s/advance", url.PathEscape(estimateID))
	var body any
	if dto != nil && dto.Mode != "" {
		body = dto
	}
	if err := c.Do(ctx, http.MethodPost, path, nil, body, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// CreateFinalInvoiceFromEstimate creates the final balance invoice (facture de solde) from a signed estimate.
// Endpoint: POST /v2/billing/estimate/{id}/final-invoice
func (c *Client) CreateFinalInvoiceFromEstimate(ctx context.Context, estimateID string) (*BillingDocument, error) {
	var out BillingDocument
	path := fmt.Sprintf("/v2/billing/estimate/%s/final-invoice", url.PathEscape(estimateID))
	if err := c.Do(ctx, http.MethodPost, path, nil, nil, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// GetAdvance retrieves an advance invoice by ID.
// Endpoint: GET /v2/billing/advance/{advanceId}
func (c *Client) GetAdvance(ctx context.Context, advanceID string) (*BillingDocument, error) {
	var out BillingDocument
	path := fmt.Sprintf("/v2/billing/advance/%s", url.PathEscape(advanceID))
	if err := c.Do(ctx, http.MethodGet, path, nil, nil, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// UpdateAdvanceAmount updates the amount or percentage of a draft advance invoice.
// Endpoint: PATCH /v2/billing/advance/{advanceId}/amount
func (c *Client) UpdateAdvanceAmount(ctx context.Context, advanceID string, dto CreateAdvanceDto) (*BillingDocument, error) {
	var out BillingDocument
	path := fmt.Sprintf("/v2/billing/advance/%s/amount", url.PathEscape(advanceID))
	if err := c.Do(ctx, http.MethodPatch, path, nil, dto, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// MarkAdvanceAsPaid records payment(s) on a finalized advance invoice.
// Endpoint: POST /v2/accounting-billing/advance/{advanceId}/reconciliate
func (c *Client) MarkAdvanceAsPaid(ctx context.Context, advanceID string, dto ReconciliateInvoiceDto) error {
	path := fmt.Sprintf("/v2/accounting-billing/advance/%s/reconciliate", url.PathEscape(advanceID))
	return c.Do(ctx, http.MethodPost, path, nil, dto, nil)
}
