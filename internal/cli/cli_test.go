package cli

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/a2dev-fr/abby-cli/internal/api"
)

func TestNormalizeVatCode(t *testing.T) {
	cases := map[string]api.VatCode{
		"fr_20":             api.VatCodeFR2000,
		"20%":               api.VatCodeFR2000,
		"FR_2000":           api.VatCodeFR2000,
		"10":                api.VatCodeFR1000,
		"5.5":               api.VatCodeFR550,
		"2,1":               api.VatCodeFR210,
		"8.5%":              api.VatCodeFR850,
		"0":                 api.VatCodeFR00HT,
		"fr_not_applicable": api.VatCodeFR00HT,
		"FR_00UE":           api.VatCodeFR00UE,
	}

	for in, want := range cases {
		if got := normalizeVatCode(in); got != want {
			t.Errorf("normalizeVatCode(%q) = %q; attendu %q", in, got, want)
		}
	}
}

func TestLineFlags_BuildDto(t *testing.T) {
	// 1. Flag unitaire en euros -> conversion en centimes
	lf := &lineFlags{
		designation:    "Prestation Go",
		unitPriceEuros: 500.50,
		quantity:       2,
		quantityUnit:   "day",
		productType:    "service_delivery",
		vatCode:        "20%",
	}
	dto, err := lf.buildDto()
	if err != nil {
		t.Fatalf("buildDto err: %v", err)
	}
	if len(dto.Lines) != 1 || dto.Lines[0].UnitPrice != 50050 || dto.Lines[0].VatCode != api.VatCodeFR2000 {
		t.Errorf("ligne générée inattendue: %+v", dto.Lines)
	}

	// 2. Tableau JSON via --lines-json
	lfJSON := &lineFlags{
		linesJSON: `[{"designation":"Ligne 1","unitPrice":10000,"quantity":1,"quantityUnit":"unit","type":"service_delivery","vatCode":"FR_2000"}]`,
	}
	dtoJSON, err := lfJSON.buildDto()
	if err != nil || len(dtoJSON.Lines) != 1 || dtoJSON.Lines[0].UnitPrice != 10000 {
		t.Fatalf("buildDto --lines-json échec: %v / %+v", err, dtoJSON)
	}

	// 3. Fichier JSON via --lines-file
	tmpFile := filepath.Join(t.TempDir(), "lines.json")
	_ = os.WriteFile(tmpFile, []byte(`{"lines":[{"designation":"Depuis fichier","unitPrice":25000,"quantity":1,"quantityUnit":"fixed_rate","type":"service_delivery","vatCode":"FR_2000"}]}`), 0600)
	lfFile := &lineFlags{linesFile: tmpFile}
	dtoFile, err := lfFile.buildDto()
	if err != nil || len(dtoFile.Lines) != 1 || dtoFile.Lines[0].Designation != "Depuis fichier" {
		t.Fatalf("buildDto --lines-file échec: %v / %+v", err, dtoFile)
	}

	// 4. Erreur si aucune entrée
	lfEmpty := &lineFlags{}
	if _, err := lfEmpty.buildDto(); err == nil {
		t.Error("attendait une erreur lorsque ni --designation ni --lines-json ne sont fournis")
	}
}

func TestRootCmd_AuthLifecycle(t *testing.T) {
	tmpDir := t.TempDir()
	t.Setenv("APPDATA", tmpDir)
	t.Setenv("XDG_CONFIG_HOME", tmpDir)
	_ = os.Unsetenv("ABBY_API_KEY")

	root := NewRootCmd()
	root.SetArgs([]string{"auth", "set-key", "suk-test-cli-123456", "--set-base-url", "https://api.abby.fr"})
	if err := root.Execute(); err != nil {
		t.Fatalf("auth set-key err: %v", err)
	}

	rootStatus := NewRootCmd()
	rootStatus.SetArgs([]string{"auth", "status", "-o", "json"})
	if err := rootStatus.Execute(); err != nil {
		t.Fatalf("auth status err: %v", err)
	}

	rootClear := NewRootCmd()
	rootClear.SetArgs([]string{"auth", "clear"})
	if err := rootClear.Execute(); err != nil {
		t.Fatalf("auth clear err: %v", err)
	}
}

func TestBuildPurchaseEntries(t *testing.T) {
	// 1. Ventilation multi-taux JSON
	entries, total, err := buildPurchaseEntries(`[{"amount":12000,"vatId":1,"accountingAccountNumber":606300},{"amount":5500,"vatId":2,"accountingAccountNumber":625600}]`, 0, false, 0, 0)
	if err != nil {
		t.Fatalf("buildPurchaseEntries JSON err: %v", err)
	}
	if len(entries) != 2 || total != 17500 || entries[0].Amount != 12000 || entries[1].AccountingAccountNumber != 625600 {
		t.Errorf("entries JSON inattendues: %+v (total=%d)", entries, total)
	}

	// 2. Flag unitaire en centimes
	single, totalSingle, err := buildPurchaseEntries("", 4999, false, 1, 606300)
	if err != nil {
		t.Fatalf("buildPurchaseEntries unitaire err: %v", err)
	}
	if len(single) != 1 || totalSingle != 4999 || single[0].Amount != 4999 || single[0].VatID != 1 {
		t.Errorf("entry unitaire inattendue: %+v (total=%d)", single, totalSingle)
	}
}
