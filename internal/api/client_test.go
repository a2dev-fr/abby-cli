package api

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestNewClient_EnforcesHTTPS(t *testing.T) {
	_, err := NewClient("http://api.app-abby.com", "suk-test123456")
	if err == nil {
		t.Fatal("attendait une erreur pour une URL HTTP non sécurisée, mais nil reçu")
	}
	if !strings.Contains(err.Error(), "HTTPS") {
		t.Fatalf("message d'erreur inattendu: %v", err)
	}
}

func TestClient_AuthHeaderAndNoBodyContentType(t *testing.T) {
	var capturedAuth, capturedContentType string
	srv := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		capturedAuth = r.Header.Get("Authorization")
		capturedContentType = r.Header.Get("Content-Type")
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(BillingDocument{
			ID:     "inv_123",
			Number: "FA-2026-001",
			Type:   BillingTypeInvoice,
			State:  BillingStateFinalized,
		})
	}))
	defer srv.Close()

	client, err := NewClient(srv.URL, "suk-secret-token", WithHTTPClient(srv.Client()))
	if err != nil {
		t.Fatalf("NewClient err: %v", err)
	}

	doc, err := client.FinalizeBilling(context.Background(), "inv_123")
	if err != nil {
		t.Fatalf("FinalizeBilling err: %v", err)
	}

	if capturedAuth != "Bearer suk-secret-token" {
		t.Errorf("Authorization attendu 'Bearer suk-secret-token', obtenu %q", capturedAuth)
	}
	if capturedContentType != "" {
		t.Errorf("Content-Type devait être vide sans body, obtenu %q", capturedContentType)
	}
	if doc.Number != "FA-2026-001" || doc.State != BillingStateFinalized {
		t.Errorf("document décodé inattendu: %+v", doc)
	}
}

func TestClient_StructuredAPIError(t *testing.T) {
	srv := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte(`{
			"type": "rsc_error",
			"code": "validation_error",
			"message": "unitPrice must be an integer in cents",
			"param": "lines[0].unitPrice"
		}`))
	}))
	defer srv.Close()

	client, err := NewClient(srv.URL, "suk-secret-token", WithHTTPClient(srv.Client()))
	if err != nil {
		t.Fatalf("NewClient err: %v", err)
	}

	_, err = client.GetInvoice(context.Background(), "inv_bad")
	if err == nil {
		t.Fatal("attendait une erreur APIError, obtenu nil")
	}

	var apiErr *APIError
	if !errors.As(err, &apiErr) {
		t.Fatalf("attendait *APIError, obtenu %T: %v", err, err)
	}
	if apiErr.StatusCode != 400 || apiErr.Code != "validation_error" || apiErr.Param != "lines[0].unitPrice" {
		t.Errorf("champs APIError inattendus: %+v", apiErr)
	}
}
