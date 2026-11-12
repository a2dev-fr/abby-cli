package output

import (
	"encoding/json"
	"fmt"
	"io"
	"math"
	"os"
	"strconv"
	"strings"
	"text/tabwriter"
	"time"

	"github.com/a2dev-fr/abby-cli/internal/api"
)

// Format represents the CLI output format.
type Format string

const (
	FormatTable Format = "table"
	FormatJSON  Format = "json"
)

// Printer formats and prints CLI results.
type Printer struct {
	Format Format
	Out    io.Writer
}

// NewPrinter creates a Printer for the given format ("table" or "json").
func NewPrinter(format string) *Printer {
	f := FormatTable
	if strings.EqualFold(strings.TrimSpace(format), "json") {
		f = FormatJSON
	}
	return &Printer{
		Format: f,
		Out:    os.Stdout,
	}
}

// PrintJSON outputs any value as indented JSON.
func (p *Printer) PrintJSON(v any) error {
	enc := json.NewEncoder(p.Out)
	enc.SetIndent("", "  ")
	return enc.Encode(v)
}

// PrintContacts prints a list of contacts in table or JSON mode.
func (p *Printer) PrintContacts(res *api.PaginatedResponse[api.ContactSummary]) error {
	if p.Format == FormatJSON {
		return p.PrintJSON(res)
	}
	items := res.EffectiveItems()
	tw := tabwriter.NewWriter(p.Out, 0, 0, 2, ' ', 0)
	fmt.Fprintln(tw, "ID\tPRÉNOM\tNOM\tEMAIL\tTÉLÉPHONE\tVILLE")
	fmt.Fprintln(tw, "--\t------\t---\t-----\t---------\t-----")
	for _, c := range items {
		email := "-"
		if len(c.Emails) > 0 {
			email = c.Emails[0]
		}
		city := "-"
		if c.BillingAddress != nil && c.BillingAddress.City != "" {
			city = c.BillingAddress.City
		}
		fmt.Fprintf(tw, "%s\t%s\t%s\t%s\t%s\t%s\n",
			c.ID, c.Firstname, c.Lastname, email, emptyDash(c.Phone), city)
	}
	_ = tw.Flush()
	fmt.Fprintf(p.Out, "\nTotal: %d contact(s) (Page %d/%d)\n", res.EffectiveTotal(), maxInt(res.Page, 1), maxInt(res.TotalPages, 1))
	return nil
}

// PrintOrganizations prints a list of organizations in table or JSON mode.
func (p *Printer) PrintOrganizations(res *api.PaginatedResponse[api.OrganizationSummary]) error {
	if p.Format == FormatJSON {
		return p.PrintJSON(res)
	}
	items := res.EffectiveItems()
	tw := tabwriter.NewWriter(p.Out, 0, 0, 2, ' ', 0)
	fmt.Fprintln(tw, "ID\tNOM\tSIRET\tTVA\tEMAIL\tVILLE")
	fmt.Fprintln(tw, "--\t---\t-----\t---\t-----\t-----")
	for _, o := range items {
		email := "-"
		if len(o.Emails) > 0 {
			email = o.Emails[0]
		}
		city := "-"
		if o.BillingAddress != nil && o.BillingAddress.City != "" {
			city = o.BillingAddress.City
		}
		fmt.Fprintf(tw, "%s\t%s\t%s\t%s\t%s\t%s\n",
			o.ID, o.Name, emptyDash(o.Siret), emptyDash(o.VatNumber), email, city)
	}
	_ = tw.Flush()
	fmt.Fprintf(p.Out, "\nTotal: %d organisation(s) (Page %d/%d)\n", res.EffectiveTotal(), maxInt(res.Page, 1), maxInt(res.TotalPages, 1))
	return nil
}

// PrintBillingList prints a list of billing documents (invoices, estimates, assets, advances).
func (p *Printer) PrintBillingList(res *api.PaginatedResponse[api.BillingDocument]) error {
	if p.Format == FormatJSON {
		return p.PrintJSON(res)
	}
	items := res.EffectiveItems()
	tw := tabwriter.NewWriter(p.Out, 0, 0, 2, ' ', 0)
	fmt.Fprintln(tw, "ID\tNUMÉRO\tTYPE\tÉTAT\tCLIENT\tDATE ÉMISSION\tTOTAL HT\tTOTAL TTC")
	fmt.Fprintln(tw, "--\t------\t----\t----\t------\t-------------\t--------\t---------")
	for _, doc := range items {
		num := doc.Number
		if num == "" {
			num = "(brouillon)"
		}
		curr := doc.CurrencyCode
		if curr == "" {
			curr = "EUR"
		}
		fmt.Fprintf(tw, "%s\t%s\t%s\t%s\t%s\t%s\t%s\t%s\n",
			doc.ID,
			num,
			doc.Type,
			doc.State,
			doc.Customer.DisplayName(),
			FormatUnixSeconds(doc.EmittedAt),
			FormatCents(doc.Total.EffectiveWithoutTaxes(), curr),
			FormatCents(doc.Total.EffectiveWithTaxes(), curr),
		)
	}
	_ = tw.Flush()
	fmt.Fprintf(p.Out, "\nTotal: %d document(s) (Page %d/%d)\n", res.EffectiveTotal(), maxInt(res.Page, 1), maxInt(res.TotalPages, 1))
	return nil
}

// PrintBillingDocument prints a single billing document with its lines and totals.
func (p *Printer) PrintBillingDocument(doc *api.BillingDocument) error {
	if p.Format == FormatJSON {
		return p.PrintJSON(doc)
	}
	curr := doc.CurrencyCode
	if curr == "" {
		curr = "EUR"
	}
	num := doc.Number
	if num == "" {
		num = "(brouillon)"
	}

	fmt.Fprintf(p.Out, "Document ID      : %s\n", doc.ID)
	fmt.Fprintf(p.Out, "Numéro           : %s\n", num)
	fmt.Fprintf(p.Out, "Type / État      : %s / %s\n", doc.Type, doc.State)
	if doc.Title != "" {
		fmt.Fprintf(p.Out, "Titre            : %s\n", doc.Title)
	}
	fmt.Fprintf(p.Out, "Client           : %s (ID: %s)\n", doc.Customer.DisplayName(), emptyDash(safeCustomerID(doc.Customer)))
	fmt.Fprintf(p.Out, "Date d'émission  : %s (timestamp: %d)\n", FormatUnixSeconds(doc.EmittedAt), doc.EmittedAt)
	if doc.DueAt > 0 {
		fmt.Fprintf(p.Out, "Date d'échéance  : %s\n", FormatUnixSeconds(doc.DueAt))
	}
	if doc.ExpiredAt > 0 {
		fmt.Fprintf(p.Out, "Date expiration  : %s\n", FormatUnixSeconds(doc.ExpiredAt))
	}
	fmt.Fprintf(p.Out, "Finalisable      : %v (Éditable: %v)\n", doc.Finalizable, doc.IsEditable)

	if len(doc.FinalizeRequirements) > 0 {
		fmt.Fprintln(p.Out, "\nPrérequis manquants avant finalisation :")
		for _, req := range doc.FinalizeRequirements {
			fmt.Fprintf(p.Out, "  - [%s] %s (%s)\n", req.Code, req.Message, req.Field)
		}
	}

	if len(doc.Lines) > 0 {
		fmt.Fprintln(p.Out, "\nLignes de facturation :")
		tw := tabwriter.NewWriter(p.Out, 0, 0, 2, ' ', 0)
		fmt.Fprintln(tw, "  DÉSIGNATION\tQTÉ\tUNITÉ\tPRIX UNIT. HT\tTVA\tTYPE")
		fmt.Fprintln(tw, "  -----------\t---\t-----\t-------------\t---\t----")
		for _, l := range doc.Lines {
			fmt.Fprintf(tw, "  %s\t%.2f\t%s\t%s\t%s\t%s\n",
				l.Designation,
				l.Quantity,
				emptyDash(string(l.QuantityUnit)),
				FormatCents(l.UnitPrice, curr),
				emptyDash(string(l.VatCode)),
				emptyDash(string(l.Type)),
			)
		}
		_ = tw.Flush()
	}

	if doc.Total != nil {
		fmt.Fprintf(p.Out, "\nTotal HT  : %s (%d centimes)\n", FormatCents(doc.Total.EffectiveWithoutTaxes(), curr), doc.Total.EffectiveWithoutTaxes())
		fmt.Fprintf(p.Out, "Total TTC : %s (%d centimes)\n", FormatCents(doc.Total.EffectiveWithTaxes(), curr), doc.Total.EffectiveWithTaxes())
	}

	return nil
}

// PrintProducts prints a paginated list of catalog products in table or JSON mode.
func (p *Printer) PrintProducts(res *api.PaginatedResponse[api.ProductDto]) error {
	if p.Format == FormatJSON {
		return p.PrintJSON(res)
	}
	items := res.EffectiveItems()
	tw := tabwriter.NewWriter(p.Out, 0, 0, 2, ' ', 0)
	fmt.Fprintln(tw, "ID\tRÉFÉRENCE\tDÉSIGNATION\tTYPE\tUNITÉ\tPRIX HT\tTVA\tPRIX TTC")
	fmt.Fprintln(tw, "--\t---------\t-----------\t----\t-----\t-------\t---\t--------")
	for _, pr := range items {
		ht := pr.UnitPrice
		if ht == 0 && pr.PriceWithoutTaxes != 0 {
			ht = pr.PriceWithoutTaxes
		}
		fmt.Fprintf(tw, "%s\t%s\t%s\t%s\t%s\t%s\t%s\t%s\n",
			pr.ID,
			emptyDash(pr.Reference),
			pr.Designation,
			pr.Type,
			pr.Unit,
			FormatCents(ht, "EUR"),
			pr.VatCode,
			FormatCents(pr.PriceWithTaxes, "EUR"),
		)
	}
	_ = tw.Flush()
	fmt.Fprintf(p.Out, "\nTotal: %d produit(s) (Page %d/%d)\n", res.EffectiveTotal(), maxInt(res.Page, 1), maxInt(res.TotalPages, 1))
	return nil
}

// PrintProduct prints a single catalog product in table or JSON mode.
func (p *Printer) PrintProduct(pr *api.ProductDto) error {
	if p.Format == FormatJSON {
		return p.PrintJSON(pr)
	}
	fmt.Fprintf(p.Out, "Produit ID  : %s\n", pr.ID)
	fmt.Fprintf(p.Out, "Désignation : %s\n", pr.Designation)
	if pr.Reference != "" {
		fmt.Fprintf(p.Out, "Référence   : %s\n", pr.Reference)
	}
	if pr.Description != "" {
		fmt.Fprintf(p.Out, "Description : %s\n", pr.Description)
	}
	fmt.Fprintf(p.Out, "Type / Unité: %s / %s\n", pr.Type, pr.Unit)
	fmt.Fprintf(p.Out, "Code TVA    : %s\n", pr.VatCode)
	fmt.Fprintf(p.Out, "Prix HT     : %s (%d centimes)\n", FormatCents(pr.UnitPrice, "EUR"), pr.UnitPrice)
	fmt.Fprintf(p.Out, "Prix TTC    : %s (%d centimes)\n", FormatCents(pr.PriceWithTaxes, "EUR"), pr.PriceWithTaxes)
	return nil
}

// PrintPurchaseRegisterList prints a paginated list of purchase register entries (livre des achats).
func (p *Printer) PrintPurchaseRegisterList(res *api.PaginatedResponse[api.PurchaseRegisterItem]) error {
	if p.Format == FormatJSON {
		return p.PrintJSON(res)
	}
	items := res.EffectiveItems()
	tw := tabwriter.NewWriter(p.Out, 0, 0, 2, ' ', 0)
	fmt.Fprintln(tw, "ID\tDATE\tLIBELLÉ\tRÉFÉRENCE\tFOURNISSEUR\tMONTANT TTC\tTVA")
	fmt.Fprintln(tw, "--\t----\t-------\t---------\t-----------\t-----------\t---")
	for _, it := range items {
		providerName := it.ThirdPartyID
		if it.ThirdParty != nil && it.ThirdParty.Name != "" {
			providerName = it.ThirdParty.Name
		}
		dateStr := it.ValueDate
		if len(dateStr) >= 10 {
			dateStr = dateStr[:10]
		}
		fmt.Fprintf(tw, "%s\t%s\t%s\t%s\t%s\t%s\t%s\n",
			it.ID,
			emptyDash(dateStr),
			it.Label,
			emptyDash(it.Reference),
			emptyDash(providerName),
			FormatCents(it.Amount, "EUR"),
			FormatCents(it.VatAmount, "EUR"),
		)
	}
	_ = tw.Flush()
	fmt.Fprintf(p.Out, "\nTotal: %d achat(s) (Page %d/%d)\n", res.EffectiveTotal(), maxInt(res.Page, 1), maxInt(res.TotalPages, 1))
	return nil
}

// PrintProviders prints a paginated list of suppliers (providers).
func (p *Printer) PrintProviders(res *api.PaginatedResponse[api.ProviderDto]) error {
	if p.Format == FormatJSON {
		return p.PrintJSON(res)
	}
	items := res.EffectiveItems()
	tw := tabwriter.NewWriter(p.Out, 0, 0, 2, ' ', 0)
	fmt.Fprintln(tw, "ID\tNOM\tSIRET\tTVA\tVILLE\tTÉLÉPHONE")
	fmt.Fprintln(tw, "--\t---\t-----\t---\t-----\t---------")
	for _, pr := range items {
		fmt.Fprintf(tw, "%s\t%s\t%s\t%s\t%s\t%s\n",
			pr.ID,
			pr.Name,
			emptyDash(pr.Siret),
			emptyDash(pr.VatNumber),
			emptyDash(pr.City),
			emptyDash(pr.Phone),
		)
	}
	_ = tw.Flush()
	fmt.Fprintf(p.Out, "\nTotal: %d fournisseur(s) (Page %d/%d)\n", res.EffectiveTotal(), maxInt(res.Page, 1), maxInt(res.TotalPages, 1))
	return nil
}

// PrintAccountingAccounts prints the chart of accounts in table or JSON mode.
func (p *Printer) PrintAccountingAccounts(accounts []api.AccountingAccountDto) error {
	if p.Format == FormatJSON {
		return p.PrintJSON(accounts)
	}
	tw := tabwriter.NewWriter(p.Out, 0, 0, 2, ' ', 0)
	fmt.Fprintln(tw, "NUMÉRO\tINTITULÉ\tDESCRIPTION\tPERSONNEL\tID")
	fmt.Fprintln(tw, "------\t--------\t-----------\t---------\t--")
	for _, a := range accounts {
		fmt.Fprintf(tw, "%d\t%s\t%s\t%v\t%s\n",
			a.Number, a.Name, emptyDash(a.Description), a.IsPersonal, a.ID)
	}
	_ = tw.Flush()
	fmt.Fprintf(p.Out, "\nTotal: %d compte(s) comptable(s)\n", len(accounts))
	return nil
}

// FormatCents formats an integer amount in cents into a readable decimal currency string.
func FormatCents(cents int64, currency string) string {
	if currency == "" {
		currency = "EUR"
	}
	euros := float64(cents) / 100.0
	return fmt.Sprintf("%.2f %s", euros, currency)
}

// EurosToCents converts a float amount in euros to integer cents (rounded to nearest cent).
func EurosToCents(euros float64) int64 {
	return int64(math.Round(euros * 100.0))
}

// FormatUnixSeconds formats a Unix timestamp (in seconds) as YYYY-MM-DD.
func FormatUnixSeconds(ts int64) string {
	if ts <= 0 {
		return "-"
	}
	return time.Unix(ts, 0).Format("2006-01-02")
}

// ParseDateOrUnixSeconds parses either a YYYY-MM-DD date string or a Unix timestamp in seconds.
func ParseDateOrUnixSeconds(val string) (int64, error) {
	v := strings.TrimSpace(val)
	if v == "" || strings.EqualFold(v, "now") || strings.EqualFold(v, "today") {
		return time.Now().Unix(), nil
	}
	if sec, err := strconv.ParseInt(v, 10, 64); err == nil {
		// Guard against accidental milliseconds (> year 3000 in seconds)
		if sec > 32503680000 {
			return sec / 1000, nil
		}
		return sec, nil
	}
	if t, err := time.Parse("2006-01-02", v); err == nil {
		return t.Unix(), nil
	}
	if t, err := time.Parse(time.RFC3339, v); err == nil {
		return t.Unix(), nil
	}
	return 0, fmt.Errorf("format de date invalide %q : attendu YYYY-MM-DD ou timestamp Unix en secondes", val)
}

func emptyDash(s string) string {
	if strings.TrimSpace(s) == "" {
		return "-"
	}
	return s
}

func safeCustomerID(c *api.BillingCustomerDto) string {
	if c == nil {
		return ""
	}
	return c.ID
}

func maxInt(a, b int) int {
	if a > b {
		return a
	}
	return b
}
