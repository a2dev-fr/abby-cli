package api

import (
	"context"
	"encoding/json"
	"net/http"
)

// CompanyMe represents the key fields returned by GET /v2/company/me.
type CompanyMe struct {
	Company struct {
		ID             string `json:"id"`
		Name           string `json:"name"`
		CommercialName string `json:"commercialName,omitempty"`
		Siret          string `json:"siret,omitempty"`
		VatNumber      string `json:"vatNumber,omitempty"`
		Email          string `json:"email,omitempty"`
	} `json:"company"`
	User struct {
		ID        string `json:"id"`
		Firstname string `json:"firstname"`
		Lastname  string `json:"lastname"`
		Email     string `json:"email"`
	} `json:"user"`
	TestMode bool `json:"testMode,omitempty"`
}

// GetMe retrieves the current authenticated company, user and preferences.
// Endpoint: GET /v2/company/me
func (c *Client) GetMe(ctx context.Context) (*CompanyMe, json.RawMessage, error) {
	var raw json.RawMessage
	if err := c.Do(ctx, http.MethodGet, "/v2/company/me", nil, nil, &raw); err != nil {
		return nil, nil, err
	}
	var me CompanyMe
	_ = json.Unmarshal(raw, &me)
	return &me, raw, nil
}
