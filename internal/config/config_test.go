package config

import (
	"os"
	"strings"
	"testing"
)

func TestConfig_ValidateAndMaskedKey(t *testing.T) {
	cfg := &Config{
		APIKey:  "",
		BaseURL: DefaultBaseURL,
	}
	if err := cfg.Validate(); err == nil {
		t.Fatal("attendait une erreur pour une clé API vide")
	}
	if got := cfg.MaskedKey(); got != "<non configurée>" {
		t.Errorf("MaskedKey vide = %q", got)
	}

	cfg.APIKey = "short"
	if got := cfg.MaskedKey(); got != "****" {
		t.Errorf("MaskedKey court = %q; attendu ****", got)
	}

	cfg.APIKey = "suk-1234567890abcdef"
	if got := cfg.MaskedKey(); got != "suk-12...cdef" {
		t.Errorf("MaskedKey long = %q; attendu suk-12...cdef", got)
	}

	cfg.BaseURL = "http://api.app-abby.com"
	if err := cfg.Validate(); err == nil || !strings.Contains(err.Error(), "HTTPS") {
		t.Fatalf("attendait une erreur HTTPS, obtenu: %v", err)
	}

	cfg.BaseURL = "https://api.app-abby.com"
	if err := cfg.Validate(); err != nil {
		t.Fatalf("Validate() inattendu: %v", err)
	}
}

func TestConfig_EnvOverrides(t *testing.T) {
	t.Setenv(EnvAPIKey, "suk-env-key-9999")
	t.Setenv(EnvBaseURL, "https://api.abby.fr")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() err: %v", err)
	}
	if cfg.APIKey != "suk-env-key-9999" {
		t.Errorf("APIKey = %q; attendu suk-env-key-9999", cfg.APIKey)
	}
	if cfg.BaseURL != "https://api.abby.fr" {
		t.Errorf("BaseURL = %q; attendu https://api.abby.fr", cfg.BaseURL)
	}
}

func TestConfig_SaveAndLoad(t *testing.T) {
	tmpDir := t.TempDir()
	// Redirect AppData / XDG_CONFIG_HOME to temporary directory for isolated test
	t.Setenv("APPDATA", tmpDir)
	t.Setenv("XDG_CONFIG_HOME", tmpDir)
	_ = os.Unsetenv(EnvAPIKey)
	_ = os.Unsetenv(EnvBaseURL)

	original := &Config{
		APIKey:  "suk-saved-secret-key",
		BaseURL: "https://api.abby.fr",
		Timeout: 45,
	}
	if err := original.Save(); err != nil {
		t.Fatalf("Save() err: %v", err)
	}

	loaded, err := Load()
	if err != nil {
		t.Fatalf("Load() err: %v", err)
	}
	if loaded.APIKey != original.APIKey || loaded.BaseURL != original.BaseURL || loaded.Timeout != 45 {
		t.Errorf("config chargée différente de l'originale: %+v vs %+v", loaded, original)
	}
}
