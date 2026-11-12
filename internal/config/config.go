package config

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

const (
	// DefaultBaseURL is the official Abby public API endpoint.
	// Note: https://api.app-abby.com and https://api.abby.fr are both supported.
	DefaultBaseURL = "https://api.app-abby.com"
	DefaultTimeout = 30 * time.Second
	EnvAPIKey      = "ABBY_API_KEY"
	EnvBaseURL     = "ABBY_BASE_URL"
)

// Config holds the runtime configuration for the Abby CLI.
type Config struct {
	APIKey  string `json:"api_key"`
	BaseURL string `json:"base_url"`
	Timeout int    `json:"timeout_seconds,omitempty"`
}

// ConfigFilePath returns the path to ~/.config/abby/config.json (or OS equivalent).
func ConfigFilePath() (string, error) {
	dir, err := os.UserConfigDir()
	if err != nil {
		home, herr := os.UserHomeDir()
		if herr != nil {
			return "", fmt.Errorf("impossible de déterminer le dossier de configuration: %w", err)
		}
		dir = filepath.Join(home, ".config")
	}
	return filepath.Join(dir, "abby-cli", "config.json"), nil
}

// Load reads configuration from disk and overrides with environment variables.
func Load() (*Config, error) {
	cfg := &Config{
		BaseURL: DefaultBaseURL,
		Timeout: int(DefaultTimeout.Seconds()),
	}

	path, err := ConfigFilePath()
	if err == nil {
		if data, readErr := os.ReadFile(path); readErr == nil {
			_ = json.Unmarshal(data, cfg)
		}
	}

	if envKey := strings.TrimSpace(os.Getenv(EnvAPIKey)); envKey != "" {
		cfg.APIKey = envKey
	}
	if envURL := strings.TrimSpace(os.Getenv(EnvBaseURL)); envURL != "" {
		cfg.BaseURL = envURL
	}
	if cfg.BaseURL == "" {
		cfg.BaseURL = DefaultBaseURL
	}
	if cfg.Timeout <= 0 {
		cfg.Timeout = int(DefaultTimeout.Seconds())
	}

	return cfg, nil
}

// Save persists the configuration to disk with restricted permissions (0600).
func (c *Config) Save() error {
	path, err := ConfigFilePath()
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(path), 0700); err != nil {
		return fmt.Errorf("création du dossier config impossible: %w", err)
	}
	data, err := json.MarshalIndent(c, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0600)
}

// Validate checks that an API key is present and enforces HTTPS.
func (c *Config) Validate() error {
	if strings.TrimSpace(c.APIKey) == "" {
		return errors.New("clé API Abby manquante : utilisez `abby auth set-key <suk-...>` ou définissez la variable d'environnement ABBY_API_KEY")
	}
	if !strings.HasPrefix(strings.ToLower(c.BaseURL), "https://") {
		return errors.New("l'API Abby exige impérativement HTTPS (les appels HTTP sont rejetés)")
	}
	return nil
}

// MaskedKey returns an obfuscated representation of the API key for safe display.
func (c *Config) MaskedKey() string {
	k := strings.TrimSpace(c.APIKey)
	if k == "" {
		return "<non configurée>"
	}
	if len(k) <= 8 {
		return "****"
	}
	return k[:6] + "..." + k[len(k)-4:]
}
