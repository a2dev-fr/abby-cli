package api

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

// Client is the HTTP client for the Abby REST API.
type Client struct {
	baseURL    string
	apiKey     string
	httpClient *http.Client
	userAgent  string
}

// Option configures a Client.
type Option func(*Client)

// WithTimeout sets the HTTP client timeout.
func WithTimeout(d time.Duration) Option {
	return func(c *Client) {
		c.httpClient.Timeout = d
	}
}

// WithHTTPClient overrides the underlying http.Client (useful for unit tests).
func WithHTTPClient(hc *http.Client) Option {
	return func(c *Client) {
		c.httpClient = hc
	}
}

// NewClient creates a new Abby API client.
func NewClient(baseURL, apiKey string, opts ...Option) (*Client, error) {
	cleanURL := strings.TrimRight(strings.TrimSpace(baseURL), "/")
	if cleanURL == "" {
		cleanURL = "https://api.app-abby.com"
	}
	if !strings.HasPrefix(strings.ToLower(cleanURL), "https://") {
		return nil, fmt.Errorf("sécurité Abby: l'URL de base doit utiliser HTTPS (%s)", cleanURL)
	}
	cleanKey := strings.TrimSpace(apiKey)
	if cleanKey == "" {
		return nil, fmt.Errorf("clé API Abby manquante")
	}

	c := &Client{
		baseURL: cleanURL,
		apiKey:  cleanKey,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
		userAgent: "abby-cli-go/1.0",
	}

	for _, opt := range opts {
		opt(c)
	}

	return c, nil
}

// Do executes a JSON API request and decodes the response into out (if non-nil).
// Note: If body is nil, Content-Type is omitted to comply with Abby API strict body parser.
func (c *Client) Do(ctx context.Context, method, path string, query url.Values, body any, out any) error {
	reqURL := c.baseURL + "/" + strings.TrimLeft(path, "/")
	if len(query) > 0 {
		reqURL += "?" + query.Encode()
	}

	var bodyReader io.Reader
	hasBody := body != nil
	if hasBody {
		payload, err := json.Marshal(body)
		if err != nil {
			return fmt.Errorf("erreur de sérialisation JSON: %w", err)
		}
		bodyReader = bytes.NewReader(payload)
	}

	req, err := http.NewRequestWithContext(ctx, method, reqURL, bodyReader)
	if err != nil {
		return fmt.Errorf("création de la requête HTTP impossible: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+c.apiKey)
	req.Header.Set("Accept", "application/json")
	req.Header.Set("User-Agent", c.userAgent)
	if hasBody {
		req.Header.Set("Content-Type", "application/json")
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("échec de l'appel HTTP [%s %s]: %w", method, reqURL, err)
	}
	defer resp.Body.Close()

	respBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("lecture de la réponse HTTP impossible: %w", err)
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return ParseAPIError(resp.StatusCode, method, path, respBytes)
	}

	if out != nil && len(respBytes) > 0 {
		if rawOut, ok := out.(*json.RawMessage); ok {
			*rawOut = append((*rawOut)[0:0], respBytes...)
			return nil
		}
		if err := json.Unmarshal(respBytes, out); err != nil {
			return fmt.Errorf("décodage JSON de la réponse impossible: %w ( brut: %s )", err, truncate(string(respBytes), 200))
		}
	}

	return nil
}

// Download fetches binary content (e.g., PDF documents from /v2/billing/{id}/download).
func (c *Client) Download(ctx context.Context, path string) ([]byte, string, error) {
	reqURL := c.baseURL + "/" + strings.TrimLeft(path, "/")
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, reqURL, nil)
	if err != nil {
		return nil, "", err
	}

	req.Header.Set("Authorization", "Bearer "+c.apiKey)
	req.Header.Set("User-Agent", c.userAgent)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, "", fmt.Errorf("échec du téléchargement [%s]: %w", reqURL, err)
	}
	defer resp.Body.Close()

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, "", err
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, "", ParseAPIError(resp.StatusCode, http.MethodGet, path, data)
	}

	contentDisp := resp.Header.Get("Content-Disposition")
	return data, contentDisp, nil
}

func truncate(s string, max int) string {
	if len(s) <= max {
		return s
	}
	return s[:max] + "..."
}
