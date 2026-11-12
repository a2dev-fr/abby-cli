package api

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
)

// ============================================================================
// Livre des recettes (/incomeBook)
// ============================================================================

// CreateIncomeBookItemDto is the payload for POST /incomeBook.
// ProductType legacy enum: 1=Vente de marchandises, 2=Prestation de services (BIC), 3=Prestation de services (BNC), 4=Vente de produits fabriqués, 5=Débours.
// PaymentMethodLegacy enum: 1=Virement, 2=Carte bancaire, 3=Espèces, 4=Chèque, 5=Prélèvement, 6=PayPal, 7=Stripe, 8=CESU, 9=Autre.
type CreateIncomeBookItemDto struct {
	Client            string         `json:"client"`                      // Customer name or label
	Reference         string         `json:"reference,omitempty"`         // Invoice / receipt reference
	PaidAt            string         `json:"paidAt,omitempty"`            // ISO date or YYYY-MM-DD
	PriceWithoutTax   int64          `json:"priceWithoutTax"`             // Amount HT in cents
	PriceTotalTax     int64          `json:"priceTotalTax"`               // Amount TTC in cents
	VatAmount         int64          `json:"vatAmount,omitempty"`         // VAT amount in cents
	VatID             int            `json:"vatId,omitempty"`             // Legacy VAT ID (1=20%, 2=10%, 3=5.5%, 4=2.1%, 5=0% HT, 6=8.5%)
	ProductType       int            `json:"productType"`                 // Legacy product type (1..5)
	PaymentMethodUsed map[string]any `json:"paymentMethodUsed,omitempty"` // e.g. {"value": 1}
	IsSap             bool           `json:"isSap,omitempty"`
	IsTaxIncluded     bool           `json:"isTaxIncluded,omitempty"`
}

// IncomeBookItem represents an entry in the income book (livre des recettes).
type IncomeBookItem struct {
	ID              string `json:"_id"`
	Reference       string `json:"reference"`
	PriceWithoutTax int64  `json:"priceWithoutTax"`
	PriceTotalTax   int64  `json:"priceTotalTax"`
	VatAmount       int64  `json:"vatAmount"`
	VatID           int    `json:"vatId,omitempty"`
	ProductType     int    `json:"productType"`
}

// CreateIncomeBookItem adds a manual entry to the income book (livre des recettes).
// Endpoint: POST /incomeBook
func (c *Client) CreateIncomeBookItem(ctx context.Context, dto CreateIncomeBookItemDto) (*IncomeBookItem, error) {
	var out IncomeBookItem
	if err := c.Do(ctx, http.MethodPost, "/incomeBook", nil, dto, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// DeleteIncomeBookItem deletes an entry from the income book.
// Endpoint: DELETE /incomeBook/{id}
func (c *Client) DeleteIncomeBookItem(ctx context.Context, id string) error {
	path := fmt.Sprintf("/incomeBook/%s", url.PathEscape(id))
	return c.Do(ctx, http.MethodDelete, path, nil, nil, nil)
}

// ============================================================================
// Livre des achats (/v3/purchaseRegister & /v2/purchaseRegister)
// ============================================================================

// PurchaseEntryLineDto represents an annotated line inside a purchase register entry.
// VatID values: 1=20%, 2=10%, 3=5.5%, 4=2.1%, 5=0% HT, 6=8.5% DOM, 7=UE, 8=Hors UE.
type PurchaseEntryLineDto struct {
	Amount                  int64 `json:"amount"`                            // Amount TTC in cents allocated to this line
	IsPersonal              bool  `json:"isPersonal"`                        // True if personal expense
	VatID                   int   `json:"vatId,omitempty"`                   // 1..8
	AccountingAccountNumber int   `json:"accountingAccountNumber,omitempty"` // e.g. 606300, 622600
}

// CreatePurchaseRegisterDto is the payload for POST /v3/purchaseRegister and /v2/purchaseRegister.
type CreatePurchaseRegisterDto struct {
	ValueDate              string                 `json:"valueDate"`                        // ISO-8601 date string (e.g. "2026-11-12T00:00:00.000Z" or "2026-11-12")
	PaymentMethodUsed      any                    `json:"paymentMethodUsed"`                // Legacy int (1..9) or string ("BANK_TRANSFER", "BANK_CARD", "CASH", "CHECK", "OTHER")
	OtherPaymentMethodUsed string                 `json:"otherPaymentMethodUsed,omitempty"` // Used when PaymentMethodUsed is OTHER / 9
	Amount                 int64                  `json:"amount"`                           // Total TTC in cents (sum of entries[].amount)
	ThirdPartyID           string                 `json:"thirdPartyId"`                     // Provider (supplier) ID from /providers
	Label                  string                 `json:"label"`                            // Expense label
	Reference              string                 `json:"reference,omitempty"`              // Supplier invoice reference
	Entries                []PurchaseEntryLineDto `json:"entries"`                          // Breakdown lines
}

// PurchaseRegisterItem represents a purchase register entry returned by Abby.
type PurchaseRegisterItem struct {
	ID                string          `json:"id"`
	ValueDate         string          `json:"valueDate"`
	Label             string          `json:"label"`
	Reference         string          `json:"reference,omitempty"`
	Amount            int64           `json:"amount"`
	VatAmount         int64           `json:"vatAmount"`
	ThirdPartyID      string          `json:"thirdPartyId"`
	ThirdParty        *ProviderDto    `json:"thirdParty,omitempty"`
	PaymentMethodUsed any             `json:"paymentMethodUsed,omitempty"`
	Entries           json.RawMessage `json:"entries,omitempty"`
}

// ListPurchaseRegisterFilter defines query parameters for GET /v3/purchaseRegister/list.
type ListPurchaseRegisterFilter struct {
	Page       int
	Limit      int
	Search     string
	ProviderID string
	FromDate   string
	ToDate     string
	MinCents   *int64
	MaxCents   *int64
}

// ListPurchaseRegisterV3 lists entries from the purchase register (livre des achats).
// Endpoint: GET /v3/purchaseRegister/list
func (c *Client) ListPurchaseRegisterV3(ctx context.Context, filter ListPurchaseRegisterFilter) (*PaginatedResponse[PurchaseRegisterItem], error) {
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
	if filter.ProviderID != "" {
		q.Set("provider", filter.ProviderID)
	}
	if filter.FromDate != "" && filter.ToDate != "" {
		q.Add("range", filter.FromDate)
		q.Add("range", filter.ToDate)
	}
	if filter.MinCents != nil {
		q.Set("min", strconv.FormatInt(*filter.MinCents, 10))
	}
	if filter.MaxCents != nil {
		q.Set("max", strconv.FormatInt(*filter.MaxCents, 10))
	}

	var out PaginatedResponse[PurchaseRegisterItem]
	if err := c.Do(ctx, http.MethodGet, "/v3/purchaseRegister/list", q, nil, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// GetPurchaseRegisterV3 retrieves a purchase register entry by ID.
// Endpoint: GET /v3/purchaseRegister/{id}
func (c *Client) GetPurchaseRegisterV3(ctx context.Context, id string) (*PurchaseRegisterItem, error) {
	var out PurchaseRegisterItem
	path := fmt.Sprintf("/v3/purchaseRegister/%s", url.PathEscape(id))
	if err := c.Do(ctx, http.MethodGet, path, nil, nil, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// CreatePurchaseRegisterV3 creates a new entry in the purchase register (v3).
// Endpoint: POST /v3/purchaseRegister
func (c *Client) CreatePurchaseRegisterV3(ctx context.Context, dto CreatePurchaseRegisterDto) (*PurchaseRegisterItem, error) {
	var out PurchaseRegisterItem
	if err := c.Do(ctx, http.MethodPost, "/v3/purchaseRegister", nil, dto, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// CreatePurchaseRegisterV2 creates a new entry in the purchase register (v2 endpoint).
// Endpoint: POST /v2/purchaseRegister
func (c *Client) CreatePurchaseRegisterV2(ctx context.Context, dto CreatePurchaseRegisterDto) (*PurchaseRegisterItem, error) {
	var out PurchaseRegisterItem
	if err := c.Do(ctx, http.MethodPost, "/v2/purchaseRegister", nil, dto, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// UpdatePurchaseRegisterV3 updates a purchase register entry (v3).
// Endpoint: PUT /v3/purchaseRegister/{id}
func (c *Client) UpdatePurchaseRegisterV3(ctx context.Context, id string, dto CreatePurchaseRegisterDto) (*PurchaseRegisterItem, error) {
	var out PurchaseRegisterItem
	path := fmt.Sprintf("/v3/purchaseRegister/%s", url.PathEscape(id))
	if err := c.Do(ctx, http.MethodPut, path, nil, dto, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// DeletePurchaseRegisterV3 deletes a purchase register entry (v3).
// Endpoint: DELETE /v3/purchaseRegister/{id}
func (c *Client) DeletePurchaseRegisterV3(ctx context.Context, id string) error {
	path := fmt.Sprintf("/v3/purchaseRegister/%s", url.PathEscape(id))
	return c.Do(ctx, http.MethodDelete, path, nil, nil, nil)
}

// DeletePurchaseRegisterV2 deletes a purchase register entry (v2).
// Endpoint: DELETE /v2/purchaseRegister/{id}
func (c *Client) DeletePurchaseRegisterV2(ctx context.Context, id string) error {
	path := fmt.Sprintf("/v2/purchaseRegister/%s", url.PathEscape(id))
	return c.Do(ctx, http.MethodDelete, path, nil, nil, nil)
}

// ============================================================================
// Fournisseurs (Providers : /providers & /provider)
// ============================================================================

// ProviderDto represents a supplier (third-party of type PROVIDER) used in the purchase register.
type ProviderDto struct {
	ID             string `json:"id,omitempty"`
	Type           string `json:"type,omitempty"` // "provider"
	Name           string `json:"name"`
	CommercialName string `json:"commercialName,omitempty"`
	Siret          string `json:"siret,omitempty"`
	Siren          string `json:"siren,omitempty"`
	VatNumber      string `json:"vatNumber,omitempty"`
	Address        string `json:"address,omitempty"`
	ZipCode        string `json:"zipCode,omitempty"`
	City           string `json:"city,omitempty"`
	Country        string `json:"country,omitempty"`
	Phone          string `json:"phone,omitempty"`
	Website        string `json:"website,omitempty"`
	Notes          string `json:"notes,omitempty"`
}

// ListProviders returns a paginated list of suppliers (providers).
// Endpoint: GET /providers
func (c *Client) ListProviders(ctx context.Context, page, limit int, search string) (*PaginatedResponse[ProviderDto], error) {
	q := url.Values{}
	if page <= 0 {
		page = 1
	}
	if limit <= 0 {
		limit = 20
	}
	q.Set("page", strconv.Itoa(page))
	q.Set("limit", strconv.Itoa(limit))
	q.Set("type", "provider")
	if search != "" {
		q.Set("search", search)
	}

	var out PaginatedResponse[ProviderDto]
	if err := c.Do(ctx, http.MethodGet, "/providers", q, nil, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// CreateProvider creates a new supplier (provider) for purchase register entries.
// Endpoint: POST /provider
func (c *Client) CreateProvider(ctx context.Context, dto ProviderDto) (*ProviderDto, error) {
	var out ProviderDto
	if err := c.Do(ctx, http.MethodPost, "/provider", nil, dto, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// UpdateProvider updates an existing supplier (provider).
// Endpoint: PUT /provider/{id}
func (c *Client) UpdateProvider(ctx context.Context, id string, dto ProviderDto) (*ProviderDto, error) {
	var out ProviderDto
	path := fmt.Sprintf("/provider/%s", url.PathEscape(id))
	if err := c.Do(ctx, http.MethodPut, path, nil, dto, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// DeleteProvider deletes a supplier (provider).
// Endpoint: DELETE /provider/{id}
func (c *Client) DeleteProvider(ctx context.Context, id string) error {
	path := fmt.Sprintf("/provider/%s", url.PathEscape(id))
	return c.Do(ctx, http.MethodDelete, path, nil, nil, nil)
}

// ============================================================================
// Plan Comptable & Transactions (/v2/accounting-accounts, /v2/transactions/{id})
// ============================================================================

// AccountingAccountDto represents an account in the company's chart of accounts.
type AccountingAccountDto struct {
	ID          string `json:"id"`
	Number      int    `json:"number"`
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
	Category    int    `json:"category,omitempty"`
	IsPersonal  bool   `json:"isPersonal"`
	IsHidden    bool   `json:"isHidden"`
}

// ListAccountingAccounts retrieves the chart of accounts for the company.
// Endpoint: GET /v2/accounting-accounts
func (c *Client) ListAccountingAccounts(ctx context.Context, includeHidden bool) ([]AccountingAccountDto, error) {
	q := url.Values{}
	if !includeHidden {
		q.Set("isHidden", "false")
	}
	var out []AccountingAccountDto
	if err := c.Do(ctx, http.MethodGet, "/v2/accounting-accounts", q, nil, &out); err != nil {
		return nil, err
	}
	return out, nil
}

// GetTransaction retrieves a bank transaction by ID.
// Endpoint: GET /v2/transactions/{id}
func (c *Client) GetTransaction(ctx context.Context, id string) (json.RawMessage, error) {
	var raw json.RawMessage
	path := fmt.Sprintf("/v2/transactions/%s", url.PathEscape(id))
	if err := c.Do(ctx, http.MethodGet, path, nil, nil, &raw); err != nil {
		return nil, err
	}
	return raw, nil
}
