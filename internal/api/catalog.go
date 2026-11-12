package api

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
)

// ProductDto represents a product in the Abby catalog (/v2/catalog).
type ProductDto struct {
	ID                      string      `json:"id,omitempty"`
	Type                    ProductType `json:"type"`
	Unit                    ProductUnit `json:"unit"`
	Designation             string      `json:"designation"`
	Description             string      `json:"description,omitempty"`
	Reference               string      `json:"reference,omitempty"`
	VatCode                 VatCode     `json:"vatCode"`
	AmountInCents           *int64      `json:"amountInCents,omitempty"` // Used in POST/PUT /v2/catalog/product
	UnitPrice               int64       `json:"unitPrice,omitempty"`     // Returned in ReadProductDto (HT in cents)
	PriceWithoutTaxes       int64       `json:"priceWithoutTaxes,omitempty"`
	VatAmount               int64       `json:"vatAmount,omitempty"`
	PriceWithTaxes          int64       `json:"priceWithTaxes,omitempty"`
	PurchasePrice           *int64      `json:"purchasePrice,omitempty"`
	ProductMargin           *float64    `json:"productMargin,omitempty"`
	LockUnitPrice           *bool       `json:"lockUnitPrice,omitempty"`
	TaxesIncluded           bool        `json:"taxesIncluded"`
	IsDeliveryOfGood        bool        `json:"isDeliveryOfGood"`
	PersonalServiceActivity string      `json:"personalServiceActivity,omitempty"`
	Stock                   int         `json:"stock,omitempty"`
}

// ListProductsFilter defines query parameters for GET /v2/catalog.
type ListProductsFilter struct {
	Page          int
	Limit         int
	Search        string
	SapCompatible *bool
}

// ListProducts retrieves a paginated list of products from the catalog.
// Endpoint: GET /v2/catalog
func (c *Client) ListProducts(ctx context.Context, filter ListProductsFilter) (*PaginatedResponse[ProductDto], error) {
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
	if filter.SapCompatible != nil {
		q.Set("sapCompatible", strconv.FormatBool(*filter.SapCompatible))
	}

	var out PaginatedResponse[ProductDto]
	if err := c.Do(ctx, http.MethodGet, "/v2/catalog", q, nil, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// GetProduct retrieves a single catalog product by ID.
// Endpoint: GET /v2/catalog/product/{productId}
func (c *Client) GetProduct(ctx context.Context, productID string) (*ProductDto, error) {
	var out ProductDto
	path := fmt.Sprintf("/v2/catalog/product/%s", url.PathEscape(productID))
	if err := c.Do(ctx, http.MethodGet, path, nil, nil, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// CreateProduct creates a new product in the catalog.
// Endpoint: POST /v2/catalog/product
func (c *Client) CreateProduct(ctx context.Context, dto ProductDto) (*ProductDto, error) {
	var out ProductDto
	if err := c.Do(ctx, http.MethodPost, "/v2/catalog/product", nil, dto, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// UpdateProduct updates an existing product in the catalog.
// Endpoint: PUT /v2/catalog/product/{productId}
func (c *Client) UpdateProduct(ctx context.Context, productID string, dto ProductDto) (*ProductDto, error) {
	var out ProductDto
	path := fmt.Sprintf("/v2/catalog/product/%s", url.PathEscape(productID))
	if err := c.Do(ctx, http.MethodPut, path, nil, dto, &out); err != nil {
		return nil, err
	}
	return &out, nil
}
