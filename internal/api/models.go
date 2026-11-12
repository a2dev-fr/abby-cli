package api

import "encoding/json"

// ============================================================================
// Enums & Constants (Abby Public API v2)
// ============================================================================

// BillingType represents the document type in Abby.
type BillingType string

const (
	BillingTypeEstimate      BillingType = "estimate"
	BillingTypeInvoice       BillingType = "invoice"
	BillingTypeAsset         BillingType = "asset"
	BillingTypeAdvance       BillingType = "advance"
	BillingTypePurchaseOrder BillingType = "purchase_order"
)

// BillingState represents the lifecycle status of a billing document.
type BillingState string

const (
	BillingStateDraft     BillingState = "draft"
	BillingStateFinalized BillingState = "finalized"
	BillingStateSigned    BillingState = "signed"
	BillingStateRefused   BillingState = "refused"
	BillingStatePaid      BillingState = "paid"
)

// VatCode represents the VAT code used on billing lines in Abby API v2.
type VatCode string

const (
	VatCodeFR2000 VatCode = "FR_2000" // 20% (Taux normal)
	VatCodeFR1000 VatCode = "FR_1000" // 10% (Taux intermédiaire)
	VatCodeFR850  VatCode = "FR_850"  // 8.5% (Taux DOM)
	VatCodeFR550  VatCode = "FR_550"  // 5.5% (Taux réduit)
	VatCodeFR210  VatCode = "FR_210"  // 2.1% (Taux super-réduit)
	VatCodeFR00HT VatCode = "FR_00HT" // 0% (Exonéré / Franchise en base)
	VatCodeFR00UE VatCode = "FR_00UE" // 0% (Intracommunautaire UE)
	VatCodeFR0HUE VatCode = "FR_0HUE" // 0% (Hors UE)
)

// ProductType represents the nature of a billing line.
type ProductType string

const (
	ProductTypeSaleOfGoods             ProductType = "sale_of_goods"
	ProductTypeServiceDelivery         ProductType = "service_delivery"
	ProductTypeCommercialOrCraft       ProductType = "commercial_or_craft_services"
	ProductTypeSaleOfManufacturedGoods ProductType = "sale_of_manufactured_goods"
	ProductTypeDisbursement            ProductType = "disbursement"
)

// ProductUnit represents the unit of measurement for a line.
type ProductUnit string

const (
	ProductUnitUnit      ProductUnit = "unit"
	ProductUnitHour      ProductUnit = "hour"
	ProductUnitDay       ProductUnit = "day"
	ProductUnitFixedRate ProductUnit = "fixed_rate"
	ProductUnitMonth     ProductUnit = "month"
	ProductUnitKilometer ProductUnit = "kilometer"
)

// PaymentDelay represents the payment terms on a billing document.
type PaymentDelay string

const (
	PaymentDelayAtReception          PaymentDelay = "at_reception"
	PaymentDelayEndOfMonth           PaymentDelay = "end_of_month"
	PaymentDelaySevenDays            PaymentDelay = "seven_days"
	PaymentDelayTenDays              PaymentDelay = "ten_days"
	PaymentDelayFifteenDays          PaymentDelay = "fifteen_days"
	PaymentDelayThirtyDays           PaymentDelay = "thirty_days"
	PaymentDelayThirtyDaysEndOfMonth PaymentDelay = "thirty_days_end_of_month"
	PaymentDelayFortyFiveDays        PaymentDelay = "forty_five_days"
	PaymentDelaySixtyDays            PaymentDelay = "sixty_days"
	PaymentDelayOther                PaymentDelay = "other"
)

// PaymentMethod represents the payment method used for invoices/advances.
type PaymentMethod string

const (
	PaymentMethodTransfer    PaymentMethod = "transfer"
	PaymentMethodDirectDebit PaymentMethod = "direct_debit"
	PaymentMethodCreditCard  PaymentMethod = "credit_card"
	PaymentMethodCheque      PaymentMethod = "cheque"
	PaymentMethodCESU        PaymentMethod = "universal_employment_service_cheque"
	PaymentMethodCash        PaymentMethod = "cash"
	PaymentMethodPaypal      PaymentMethod = "paypal"
	PaymentMethodStripe      PaymentMethod = "stripe"
	PaymentMethodOther       PaymentMethod = "other"
)

// ============================================================================
// Clients: Contacts & Organizations
// ============================================================================

// ClientAddressDto represents a postal address for a contact or organization.
type ClientAddressDto struct {
	Address    string `json:"address"`
	Complement string `json:"complement,omitempty"`
	City       string `json:"city"`
	ZipCode    string `json:"zipCode"`
	State      string `json:"state,omitempty"`
	Country    string `json:"country"` // ISO 2-letter code, e.g. "FR"
}

// BankInformationsDto represents IBAN/BIC details for a client.
type BankInformationsDto struct {
	IBAN       string `json:"iban,omitempty"`
	BIC        string `json:"bic,omitempty"`
	HolderName string `json:"holderName,omitempty"`
}

// CreateContactDto is the request payload for POST /contact and PUT /contact/{id}.
type CreateContactDto struct {
	Firstname       string               `json:"firstname"`
	Lastname        string               `json:"lastname"`
	Phone           string               `json:"phone,omitempty"`
	JobTitle        string               `json:"jobTitle,omitempty"`
	Emails          []string             `json:"emails,omitempty"`
	Notes           string               `json:"notes,omitempty"`
	BillingAddress  *ClientAddressDto    `json:"billingAddress,omitempty"`
	DeliveryAddress *ClientAddressDto    `json:"deliveryAddress,omitempty"`
	BankInfo        *BankInformationsDto `json:"bankInformations,omitempty"`
}

// ContactSummary represents a contact returned by the API.
type ContactSummary struct {
	ID              string            `json:"id"`
	Firstname       string            `json:"firstname"`
	Lastname        string            `json:"lastname"`
	Emails          []string          `json:"emails,omitempty"`
	Phone           string            `json:"phone,omitempty"`
	JobTitle        string            `json:"jobTitle,omitempty"`
	Notes           string            `json:"notes,omitempty"`
	BillingAddress  *ClientAddressDto `json:"billingAddress,omitempty"`
	DeliveryAddress *ClientAddressDto `json:"deliveryAddress,omitempty"`
	ArchivedAt      *int64            `json:"archivedAt,omitempty"`
	CreatedAt       any               `json:"createdAt,omitempty"`
	Raw             json.RawMessage   `json:"-"`
}

// CreateOrganizationDto is the request payload for POST /organization and PUT /organization/{id}.
type CreateOrganizationDto struct {
	Name            string               `json:"name"`
	CommercialName  string               `json:"commercialName,omitempty"`
	Emails          []string             `json:"emails,omitempty"`
	Siret           string               `json:"siret,omitempty"`
	VatNumber       string               `json:"vatNumber,omitempty"`
	BillingAddress  *ClientAddressDto    `json:"billingAddress,omitempty"`
	DeliveryAddress *ClientAddressDto    `json:"deliveryAddress,omitempty"`
	Notes           string               `json:"notes,omitempty"`
	BankInfo        *BankInformationsDto `json:"bankInformations,omitempty"`
}

// OrganizationSummary represents an organization returned by the API.
type OrganizationSummary struct {
	ID              string            `json:"id"`
	Name            string            `json:"name"`
	CommercialName  string            `json:"commercialName,omitempty"`
	Siret           string            `json:"siret,omitempty"`
	VatNumber       string            `json:"vatNumber,omitempty"`
	Emails          []string          `json:"emails,omitempty"`
	BillingAddress  *ClientAddressDto `json:"billingAddress,omitempty"`
	DeliveryAddress *ClientAddressDto `json:"deliveryAddress,omitempty"`
	DefaultContact  *ContactSummary   `json:"defaultContact,omitempty"`
}

// PaginatedResponse represents a standard Abby paginated list.
type PaginatedResponse[T any] struct {
	Docs          []T   `json:"docs"`
	TotalDocs     int   `json:"totalDocs"`
	Limit         int   `json:"limit"`
	Page          int   `json:"page"`
	TotalPages    int   `json:"totalPages"`
	HasNextPage   bool  `json:"hasNextPage"`
	HasPrevPage   bool  `json:"hasPrevPage"`
	NextPage      *int  `json:"nextPage,omitempty"`
	PrevPage      *int  `json:"prevPage,omitempty"`
	PagingCounter int   `json:"pagingCounter,omitempty"`
	Items         []T   `json:"items,omitempty"` // Fallback if endpoint uses items instead of docs
	Total         int   `json:"total,omitempty"`
}

// EffectiveItems returns Docs if non-empty, otherwise Items.
func (p *PaginatedResponse[T]) EffectiveItems() []T {
	if len(p.Docs) > 0 {
		return p.Docs
	}
	return p.Items
}

// EffectiveTotal returns TotalDocs if non-zero, otherwise Total.
func (p *PaginatedResponse[T]) EffectiveTotal() int {
	if p.TotalDocs > 0 {
		return p.TotalDocs
	}
	return p.Total
}

// ============================================================================
// Billing Documents (Invoices, Estimates, Assets, Advances)
// ============================================================================

// DiscountDto represents a discount in percentage or fixed amount (in cents / basis points).
type DiscountDto struct {
	Mode   string `json:"mode"`   // "percentage" | "amount" (or uppercase)
	Amount int64  `json:"amount"` // In cents or basis points (e.g. 1000 for 10% or 10.00€)
}

// UpdateLineDto represents a line item when updating a billing document via PATCH /v2/billing/{billingId}/lines.
type UpdateLineDto struct {
	ProductID        string       `json:"productId,omitempty"`
	Designation      string       `json:"designation"`
	Description      string       `json:"description,omitempty"`
	Reference        string       `json:"reference,omitempty"`
	UnitPrice        int64        `json:"unitPrice"` // Price per unit in CENTS (e.g., 15000 = 150.00 €)
	Quantity         float64      `json:"quantity"`
	QuantityUnit     ProductUnit  `json:"quantityUnit"`
	Type             ProductType  `json:"type"`
	VatCode          VatCode      `json:"vatCode,omitempty"`
	Discount         *DiscountDto `json:"discount,omitempty"`
	IsDeliveryOfGood bool         `json:"isDeliveryOfGoods,omitempty"`
	IsTaxIncluded    bool         `json:"isTaxIncluded,omitempty"`
}

// UpdateBillingLinesDto is the request payload for PATCH /v2/billing/{billingId}/lines.
type UpdateBillingLinesDto struct {
	Lines    []UpdateLineDto `json:"lines"`
	Discount *DiscountDto    `json:"discount,omitempty"`
}

// UpdateGeneralInformationsDto is the payload for PATCH /v2/billing/{type}/{id}/general-informations.
type UpdateGeneralInformationsDto struct {
	HeaderNote                      string          `json:"headerNote,omitempty"`
	FooterNote                      string          `json:"footerNote,omitempty"`
	GeneralTermsAndConditionsOfSale string          `json:"generalTermsAndConditionsOfSale,omitempty"`
	MentionMediator                 string          `json:"mentionMediator,omitempty"`
	PaymentMethods                  []PaymentMethod `json:"paymentMethods,omitempty"`
	OtherPaymentMethod              string          `json:"otherPaymentMethod,omitempty"`
}

// UpdateTimelineDto is the payload for PATCH /v2/billing/{type}/{id}/timeline.
// All timestamps MUST be in Unix seconds (NOT milliseconds).
type UpdateTimelineDto struct {
	EmittedAt     int64        `json:"emittedAt"`               // Unix timestamp in seconds
	PaymentDelay  PaymentDelay `json:"paymentDelay"`            // e.g. "thirty_days", "at_reception", "other"
	CustomDueDate *int64       `json:"customDueDate,omitempty"` // Unix timestamp in seconds when PaymentDelay == "other"
	DeliveryDate  *int64       `json:"deliveryDate,omitempty"`  // Unix timestamp in seconds
	ExpiredAt     *int64       `json:"expiredAt,omitempty"`     // Required for estimates: expiration Unix timestamp in seconds
}

// CreateEstimateDto is the payload for POST /v2/billing/estimate/{customerId}.
type CreateEstimateDto struct {
	EstimateType string `json:"estimateType,omitempty"` // "estimate" (default) or "purchase_order"
}

// CreateAdvanceDto is the payload for POST /v2/billing/estimate/{id}/advance or PATCH /v2/billing/advance/{advanceId}/amount.
type CreateAdvanceDto struct {
	Mode       string  `json:"mode,omitempty"`       // "AMOUNT" or "PERCENTAGE"
	Amount     *int64  `json:"amount,omitempty"`     // In cents HT when mode == "AMOUNT"
	Percentage float64 `json:"percentage,omitempty"` // 0 < percentage <= 100 when mode == "PERCENTAGE"
}

// PaymentDto represents a payment entry when marking an invoice or advance as paid.
type PaymentDto struct {
	Amount        int64         `json:"amount"`                  // Amount in cents TTC
	ReceivedAt    string        `json:"receivedAt"`              // ISO 8601 date string or YYYY-MM-DD
	Method        PaymentMethod `json:"method"`                  // e.g. "transfer", "credit_card", "cash"
	OtherMethod   string        `json:"otherMethod,omitempty"`   // Used when Method == "other"
	TransactionID string        `json:"transactionId,omitempty"` // Optional bank transaction ID
}

// ReconciliateInvoiceDto is the payload for POST /v2/accounting-billing/invoice/{invoiceId}/reconciliate.
type ReconciliateInvoiceDto struct {
	Payments []PaymentDto `json:"payments"`
}

// CancelBillingDto is the payload for POST /v2/billing/invoice/{invoiceId}/cancel.
type CancelBillingDto struct {
	WantArchiveAfterCancellation bool `json:"wantArchiveAfterCancellation"`
}

// SendBillingByEmailDto is the payload for POST /v2/billing/{billingId}/send-by-email.
type SendBillingByEmailDto struct {
	Emails  []string `json:"emails"`
	Subject string   `json:"subject"`
	Content string   `json:"content"`
	InCopy  bool     `json:"inCopy,omitempty"`
}

// UpdateTitleDto is the payload for PATCH /v2/billing/{documentId}/title.
type UpdateTitleDto struct {
	Title *string `json:"title"`
}

// UpdateLocaleDto is the payload for PUT /v2/billing/{type}/{id}/locale.
type UpdateLocaleDto struct {
	Locale string `json:"locale"` // e.g., "fr", "en"
}

// UpdateCurrencyDto is the payload for PUT /v2/billing/{type}/{id}/currency.
type UpdateCurrencyDto struct {
	CurrencyCode string `json:"currencyCode"` // e.g., "EUR", "USD"
}

// BillingTotalDto holds the monetary totals of a billing document (in cents).
type BillingTotalDto struct {
	WithoutTaxes          int64 `json:"withoutTaxes"`
	WithTaxes             int64 `json:"withTaxes"`
	Tax                   int64 `json:"tax"`
	RemainingWithTaxes    int64 `json:"remainingWithTaxes,omitempty"`
	RemainingWithoutTaxes int64 `json:"remainingWithoutTaxes,omitempty"`
	TotalWithoutTaxes     int64 `json:"totalWithoutTaxes,omitempty"`
	TotalWithTaxes        int64 `json:"totalWithTaxes,omitempty"`
	TotalTax              int64 `json:"totalTax,omitempty"`
}

// EffectiveWithoutTaxes returns the HT total in cents across v2 schema variations.
func (t *BillingTotalDto) EffectiveWithoutTaxes() int64 {
	if t == nil {
		return 0
	}
	if t.WithoutTaxes != 0 {
		return t.WithoutTaxes
	}
	return t.TotalWithoutTaxes
}

// EffectiveWithTaxes returns the TTC total in cents across v2 schema variations.
func (t *BillingTotalDto) EffectiveWithTaxes() int64 {
	if t == nil {
		return 0
	}
	if t.WithTaxes != 0 {
		return t.WithTaxes
	}
	return t.TotalWithTaxes
}

// BillingCustomerDto represents the customer embedded in a billing document.
type BillingCustomerDto struct {
	ID             string `json:"id"`
	Firstname      string `json:"firstname,omitempty"`
	Lastname       string `json:"lastname,omitempty"`
	CommercialName string `json:"commercialName,omitempty"`
	Name           string `json:"name,omitempty"`
	Siret          string `json:"siret,omitempty"`
}

// DisplayName returns a human-readable customer name.
func (c *BillingCustomerDto) DisplayName() string {
	if c == nil {
		return "-"
	}
	if c.Name != "" {
		return c.Name
	}
	if c.CommercialName != "" {
		return c.CommercialName
	}
	full := c.Firstname + " " + c.Lastname
	if full != " " {
		return full
	}
	return c.ID
}

// BillingLineSummary represents a read billing line from Abby API.
type BillingLineSummary struct {
	ID                string      `json:"id,omitempty"`
	Designation       string      `json:"designation"`
	Description       string      `json:"description,omitempty"`
	Reference         string      `json:"reference,omitempty"`
	Quantity          float64     `json:"quantity"`
	QuantityUnit      ProductUnit `json:"quantityUnit,omitempty"`
	UnitPrice         int64       `json:"unitPrice"`
	TotalWithoutTaxes int64       `json:"totalWithoutTaxes,omitempty"`
	TotalWithTaxes    int64       `json:"totalWithTaxes,omitempty"`
	VatCode           VatCode     `json:"vatCode,omitempty"`
	Type              ProductType `json:"type,omitempty"`
}

// FinalizeRequirementDto indicates missing requirements before a document can be finalized.
type FinalizeRequirementDto struct {
	Field   string `json:"field,omitempty"`
	Message string `json:"message,omitempty"`
	Code    string `json:"code,omitempty"`
}

// BillingDocument represents an Abby billing document (Invoice, Estimate, Asset, Advance).
type BillingDocument struct {
	ID                   string                   `json:"id"`
	Number               string                   `json:"number,omitempty"`
	Title                string                   `json:"title,omitempty"`
	Type                 BillingType              `json:"type"`
	State                BillingState             `json:"state"`
	Customer             *BillingCustomerDto      `json:"customer,omitempty"`
	Locale               string                   `json:"locale,omitempty"`
	CurrencyCode         string                   `json:"currencyCode,omitempty"`
	EmittedAt            int64                    `json:"emittedAt"`
	DueAt                int64                    `json:"dueAt,omitempty"`
	ExpiredAt            int64                    `json:"expiredAt,omitempty"`
	FinalizedAt          int64                    `json:"finalizedAt,omitempty"`
	ArchivedAt           int64                    `json:"archivedAt,omitempty"`
	PaymentDelay         PaymentDelay             `json:"paymentDelay,omitempty"`
	Finalizable          bool                     `json:"finalizable"`
	IsEditable           bool                     `json:"isEditable"`
	IsDeletable          bool                     `json:"isDeletable"`
	Locked               bool                     `json:"locked"`
	Test                 bool                     `json:"test"`
	Lines                []BillingLineSummary     `json:"lines,omitempty"`
	Total                *BillingTotalDto         `json:"total,omitempty"`
	FinalizeRequirements []FinalizeRequirementDto `json:"finalizeRequirements,omitempty"`
}
