package output

import "testing"

func TestEurosToCentsAndFormat(t *testing.T) {
	if got := EurosToCents(75.50); got != 7550 {
		t.Errorf("EurosToCents(75.50) = %d; attendu 7550", got)
	}
	if got := EurosToCents(0.99); got != 99 {
		t.Errorf("EurosToCents(0.99) = %d; attendu 99", got)
	}
	if got := FormatCents(150000, "EUR"); got != "1500.00 EUR" {
		t.Errorf("FormatCents(150000, EUR) = %q; attendu '1500.00 EUR'", got)
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

	// Conversion automatique si quelqu'un passe des millisecondes par erreur
	msSec, err := ParseDateOrUnixSeconds("1777593600000")
	if err != nil {
		t.Fatalf("erreur inattendue: %v", err)
	}
	if msSec != 1777593600 {
		t.Errorf("attendu 1777593600 secondes, obtenu %d", msSec)
	}
}
