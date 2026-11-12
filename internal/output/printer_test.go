package output

import (
	"bytes"
	"strings"
	"testing"

	"github.com/a2dev-fr/abby-cli/internal/api"
)

func TestEurosToCentsAndFormat(t *testing.T) {
	if got := EurosToCents(75.50); got != 7550 {
		t.Errorf("EurosToCents(75.50) = %d; attendu 7550", got)
	}
	if got := EurosToCents(0.99); got != 99 {
		t.Errorf("EurosToCents(0.99) = %d; attendu 99", got)
	}
	if got := FormatCents(150000, ""); got != "1500.00 EUR" {
		t.Errorf("FormatCents(150000, '') = %q; attendu '1500.00 EUR'", got)
	}
}

func TestParseDateOrUnixSeconds(t *testing.T) {
	// Date YYYY-MM-DD
	sec, err := ParseDateOrUnixSeconds("2026-05-01")
	if err != nil {
		t.Fatalf("erreur inattendue: %v", err)
	}
	if FormatUnixSeconds(sec) != "2026-05-01" {
		t.Errorf("attendu 2026-05-01, obtenu %s", FormatUnixSeconds(sec))
	}

	// RFC3339
	secRFC, err := ParseDateOrUnixSeconds("2026-05-01T10:00:00Z")
	if err != nil || secRFC <= 0 {
		t.Fatalf("RFC3339 err: %v", err)
	}

	// "today" / "now"
	secNow, err := ParseDateOrUnixSeconds("today")
	if err != nil || secNow <= 0 {
		t.Fatalf("today err: %v", err)
	}

	// Conversion automatique si quelqu'un passe des millisecondes par erreur
	msSec, err := ParseDateOrUnixSeconds("1777593600000")
	if err != nil {
		t.Fatalf("erreur inattendue: %v", err)
	}
	if msSec != 1777593600 {
		t.Errorf("attendu 1777593600 secondes, obtenu %d", msSec)
	}

	// Format invalide
	if _, err := ParseDateOrUnixSeconds("not-a-date"); err == nil {
		t.Error("attendait une erreur pour une date invalide")
	}

	if got := FormatUnixSeconds(0); got != "-" {
		t.Errorf("FormatUnixSeconds(0) = %q; attendu '-'", got)
	}
}

func TestPrinter_TableAndJSONOutputs(t *testing.T) {
	var buf bytes.Buffer
	p := NewPrinter("table")
	p.Out = &buf

	// 1. Contacts
	err := p.PrintContacts(&api.PaginatedResponse[api.ContactSummary]{
		Docs: []api.ContactSummary{
			{
				ID:        "c_1",
				Firstname: "Jean",
				Lastname:  "Dupont",
				Emails:    []string{"jean@dupont.fr"},
				BillingAddress: &api.ClientAddressDto{
					City: "Paris",
				},
			},
		},
		TotalDocs:  1,
		Page:       1,
		TotalPages: 1,
	})
	if err != nil || !strings.Contains(buf.String(), "Jean") || !strings.Contains(buf.String(), "Paris") {
		t.Errorf("PrintContacts table inattendu: %v / %s", err, buf.String())
	}

	// 2. Organizations
	buf.Reset()
	err = p.PrintOrganizations(&api.PaginatedResponse[api.OrganizationSummary]{
		Docs: []api.OrganizationSummary{
			{
				ID:    "org_1",
				Name:  "Dupont SARL",
				Siret: "12345678900012",
			},
		},
		TotalDocs: 1,
	})
	if err != nil || !strings.Contains(buf.String(), "Dupont SARL") {
		t.Errorf("PrintOrganizations table inattendu: %v / %s", err, buf.String())
	}

	// 3. Billing Document & List
	buf.Reset()
	doc := api.BillingDocument{
		ID:          "inv_1",
		Number:      "FA-2026-001",
		Title:       "Mission Go",
		Type:        api.BillingTypeInvoice,
		State:       api.BillingStateFinalized,
		EmittedAt:   1777593600,
		DueAt:       1780185600,
		Finalizable: true,
		Customer:    &api.BillingCustomerDto{ID: "org_1", Name: "Dupont SARL"},
		Lines: []api.BillingLineSummary{
			{Designation: "Dev Go", Quantity: 2, QuantityUnit: api.ProductUnitDay, UnitPrice: 50000, VatCode: api.VatCodeFR2000},
		},
		Total: &api.BillingTotalDto{WithoutTaxes: 100000, WithTaxes: 120000},
		FinalizeRequirements: []api.FinalizeRequirementDto{
			{Code: "MISSING_ADDR", Message: "Adresse requise", Field: "address"},
		},
	}
	if err := p.PrintBillingDocument(&doc); err != nil || !strings.Contains(buf.String(), "FA-2026-001") || !strings.Contains(buf.String(), "1200.00 EUR") {
		t.Errorf("PrintBillingDocument table inattendu: %v / %s", err, buf.String())
	}

	buf.Reset()
	if err := p.PrintBillingList(&api.PaginatedResponse[api.BillingDocument]{Docs: []api.BillingDocument{doc}, TotalDocs: 1}); err != nil || !strings.Contains(buf.String(), "FA-2026-001") {
		t.Errorf("PrintBillingList table inattendu: %v / %s", err, buf.String())
	}

	// 4. Products
	buf.Reset()
	prod := api.ProductDto{
		ID:             "prod_1",
		Designation:    "Forfait Audit",
		Reference:      "AUD-01",
		Type:           api.ProductTypeServiceDelivery,
		Unit:           api.ProductUnitFixedRate,
		UnitPrice:      80000,
		PriceWithTaxes: 96000,
		VatCode:        api.VatCodeFR2000,
	}
	if err := p.PrintProduct(&prod); err != nil || !strings.Contains(buf.String(), "800.00 EUR") {
		t.Errorf("PrintProduct table inattendu: %v / %s", err, buf.String())
	}
	buf.Reset()
	if err := p.PrintProducts(&api.PaginatedResponse[api.ProductDto]{Docs: []api.ProductDto{prod}, TotalDocs: 1}); err != nil || !strings.Contains(buf.String(), "Forfait Audit") {
		t.Errorf("PrintProducts table inattendu: %v / %s", err, buf.String())
	}

	// 5. Purchases, Providers, Accounts
	buf.Reset()
	if err := p.PrintPurchaseRegisterList(&api.PaginatedResponse[api.PurchaseRegisterItem]{
		Docs: []api.PurchaseRegisterItem{
			{ID: "pur_1", ValueDate: "2026-11-12T00:00:00Z", Label: "Serveur OVH", Amount: 5999, VatAmount: 1000, ThirdParty: &api.ProviderDto{Name: "OVH"}},
		},
	}); err != nil || !strings.Contains(buf.String(), "Serveur OVH") {
		t.Errorf("PrintPurchaseRegisterList inattendu: %v / %s", err, buf.String())
	}

	buf.Reset()
	if err := p.PrintProviders(&api.PaginatedResponse[api.ProviderDto]{
		Docs: []api.ProviderDto{{ID: "prov_1", Name: "OVHcloud", Siret: "42476141900045"}},
	}); err != nil || !strings.Contains(buf.String(), "OVHcloud") {
		t.Errorf("PrintProviders inattendu: %v / %s", err, buf.String())
	}

	buf.Reset()
	if err := p.PrintAccountingAccounts([]api.AccountingAccountDto{
		{ID: "acc_1", Number: 606300, Name: "Fournitures d'entretien"},
	}); err != nil || !strings.Contains(buf.String(), "606300") {
		t.Errorf("PrintAccountingAccounts inattendu: %v / %s", err, buf.String())
	}

	// 6. Mode JSON
	jsonPrinter := NewPrinter("json")
	buf.Reset()
	jsonPrinter.Out = &buf
	if err := jsonPrinter.PrintBillingDocument(&doc); err != nil || !strings.Contains(buf.String(), `"id": "inv_1"`) {
		t.Errorf("PrintBillingDocument JSON inattendu: %v / %s", err, buf.String())
	}
}
