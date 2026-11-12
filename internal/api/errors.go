package api

import (
	"encoding/json"
	"fmt"
)

// APIError represents the structured error returned by the Abby REST API.
// Reference: https://docs.abby.fr/api/erreurs
type APIError struct {
	StatusCode int    `json:"status_code"`
	Method     string `json:"method"`
	URL        string `json:"url"`
	Type       string `json:"type,omitempty"`    // "api_error" | "authentication_error" | "rsc_error"
	Code       string `json:"code,omitempty"`    // "internal_server_error" | "token_failure" | "validation_error" | "not_found" | "bad_request" | "forbidden"
	Message    string `json:"message,omitempty"` // Detailed error message
	Param      string `json:"param,omitempty"`   // Offending parameter if applicable
	RawBody    string `json:"raw_body,omitempty"`
}

func (e *APIError) Error() string {
	msg := e.Message
	if msg == "" {
		msg = e.RawBody
	}
	if e.Code != "" && e.Param != "" {
		return fmt.Sprintf("Abby API Error [%d %s %s] (%s / code=%s, param=%s): %s",
			e.StatusCode, e.Method, e.URL, e.Type, e.Code, e.Param, msg)
	}
	if e.Code != "" {
		return fmt.Sprintf("Abby API Error [%d %s %s] (%s / code=%s): %s",
			e.StatusCode, e.Method, e.URL, e.Type, e.Code, msg)
	}
	return fmt.Sprintf("Abby API Error [%d %s %s]: %s", e.StatusCode, e.Method, e.URL, msg)
}

// ParseAPIError parses an error response body from Abby API.
func ParseAPIError(statusCode int, method, reqURL string, body []byte) error {
	apiErr := &APIError{
		StatusCode: statusCode,
		Method:     method,
		URL:        reqURL,
		RawBody:    string(body),
	}

	if len(body) > 0 {
		_ = json.Unmarshal(body, apiErr)
	}

	if apiErr.Message == "" {
		switch statusCode {
		case 400:
			apiErr.Message = "Requête invalide (vérifiez les paramètres, montants en centimes ou timestamps en secondes)"
		case 401:
			apiErr.Message = "Clé API invalide ou expirée (format attendu: Bearer suk-*****)"
		case 403:
			apiErr.Message = "Accès refusé : la clé API n'a pas les permissions requises"
		case 404:
			apiErr.Message = "Ressource introuvable (vérifiez l'identifiant)"
		case 429:
			apiErr.Message = "Trop de requêtes envoyées à l'API Abby (Rate Limit atteint)"
		default:
			if statusCode >= 500 {
				apiErr.Message = "Erreur interne côté serveur Abby"
			}
		}
	}

	return apiErr
}
