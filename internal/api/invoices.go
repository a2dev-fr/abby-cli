package api

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
)

// GetInvoice retrieves an invoice by its ID.
// Endpoint: GET /v2/billing/invoice/{invoiceId}
func (c *Client) GetInvoice(ctx context.Context, invoiceID string) (*BillingDocument, error) {
	var out BillingDocument
	path := fmt.Sprintf("/v2/billing/invoice/%s", url.PathEscape(invoiceID))
	if err := c.Do(ctx, http.MethodGet, path, nil, nil, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// CreateInvoice creates a new draft invoice for a customerId (contactId or organizationId).
// Endpoint: POST /v2/billing/invoice/{customerId}
func (c *Client) CreateInvoice(ctx context.Context, customerID string) (*BillingDocument, error) {
	var out BillingDocument
	path := fmt.Sprintf("/v2/billing/invoice/%s", url.PathEscape(customerID))
	if err := c.Do(ctx, http.MethodPost, path, nil, nil, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// UpdateInvoiceGeneralInfo updates the header/footer notes, terms, and payment methods of a draft invoice.
// Endpoint: PATCH /v2/billing/invoice/{invoiceId}/general-informations
func (c *Client) UpdateInvoiceGeneralInfo(ctx context.Context, invoiceID string, dto UpdateGeneralInformationsDto) (*BillingDocument, error) {
	var out BillingDocument
	path := fmt.Sprintf("/v2/billing/invoice/%s/general-informations", url.PathEscape(invoiceID))
	if err := c.Do(ctx, http.MethodPatch, path, nil, dto, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// UpdateInvoiceTimeline updates emittedAt, paymentDelay, customDueDate, and deliveryDate (timestamps in seconds).
// Endpoint: PATCH /v2/billing/invoice/{invoiceId}/timeline
func (c *Client) UpdateInvoiceTimeline(ctx context.Context, invoiceID string, dto UpdateTimelineDto) (*BillingDocument, error) {
	var out BillingDocument
	path := fmt.Sprintf("/v2/billing/invoice/%s/timeline", url.PathEscape(invoiceID))
	if err := c.Do(ctx, http.MethodPatch, path, nil, dto, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// UpdateInvoiceLocale updates the document language of an invoice (e.g. "fr", "en").
// Endpoint: PUT /v2/billing/invoice/{invoiceId}/locale
func (c *Client) UpdateInvoiceLocale(ctx context.Context, invoiceID, locale string) (*BillingDocument, error) {
	var out BillingDocument
	path := fmt.Sprintf("/v2/billing/invoice/%s/locale", url.PathEscape(invoiceID))
	if err := c.Do(ctx, http.MethodPut, path, nil, UpdateLocaleDto{Locale: locale}, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// UpdateInvoiceCurrency updates the currency of an invoice (e.g. "EUR", "USD").
// Endpoint: PUT /v2/billing/invoice/{invoiceId}/currency
func (c *Client) UpdateInvoiceCurrency(ctx context.Context, invoiceID, currencyCode string) (*BillingDocument, error) {
	var out BillingDocument
	path := fmt.Sprintf("/v2/billing/invoice/%s/currency", url.PathEscape(invoiceID))
	if err := c.Do(ctx, http.MethodPut, path, nil, UpdateCurrencyDto{CurrencyCode: currencyCode}, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// ActivateInvoiceOnlinePayment activates online payment on a finalized invoice.
// Endpoint: PATCH /v2/billing/invoice/{invoiceId}/activate-online-payment
func (c *Client) ActivateInvoiceOnlinePayment(ctx context.Context, invoiceID string) (*BillingDocument, error) {
	var out BillingDocument
	path := fmt.Sprintf("/v2/billing/invoice/%s/activate-online-payment", url.PathEscape(invoiceID))
	if err := c.Do(ctx, http.MethodPatch, path, nil, nil, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// CreateAssetFromInvoice creates a draft credit note (avoir / asset) from a finalized invoice.
// Endpoint: POST /v2/billing/invoice/{invoiceId}/asset
func (c *Client) CreateAssetFromInvoice(ctx context.Context, invoiceID string) (*BillingDocument, error) {
	var out BillingDocument
	path := fmt.Sprintf("/v2/billing/invoice/%s/asset", url.PathEscape(invoiceID))
	if err := c.Do(ctx, http.MethodPost, path, nil, nil, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// CancelInvoice cancels a finalized invoice by automatically creating and finalizing a credit note (avoir).
// Endpoint: POST /v2/billing/invoice/{invoiceId}/cancel
func (c *Client) CancelInvoice(ctx context.Context, invoiceID string, archiveAfter bool) (*BillingDocument, error) {
	var out BillingDocument
	path := fmt.Sprintf("/v2/billing/invoice/%s/cancel", url.PathEscape(invoiceID))
	dto := CancelBillingDto{WantArchiveAfterCancellation: archiveAfter}
	if err := c.Do(ctx, http.MethodPost, path, nil, dto, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// MarkInvoiceAsPaid records payment(s) on a finalized invoice and marks it as paid.
// Note: The sum of payments (in cents) must match the invoice remaining amount TTC.
// Endpoint: POST /v2/accounting-billing/invoice/{invoiceId}/reconciliate
func (c *Client) MarkInvoiceAsPaid(ctx context.Context, invoiceID string, dto ReconciliateInvoiceDto) error {
	path := fmt.Sprintf("/v2/accounting-billing/invoice/%s/reconciliate", url.PathEscape(invoiceID))
	return c.Do(ctx, http.MethodPost, path, nil, dto, nil)
}

// MarkInvoiceAsUnpaid removes payments from a paid invoice and reverts it to finalized/unpaid.
// Endpoint: POST /v2/accounting-billing/invoice/{invoiceId}/mark-as-unpaid
func (c *Client) MarkInvoiceAsUnpaid(ctx context.Context, invoiceID string) error {
	path := fmt.Sprintf("/v2/accounting-billing/invoice/%s/mark-as-unpaid", url.PathEscape(invoiceID))
	return c.Do(ctx, http.MethodPost, path, nil, nil, nil)
}
