package cli

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/a2dev-fr/abby-cli/internal/api"
	"github.com/a2dev-fr/abby-cli/internal/output"
	"github.com/spf13/cobra"
)

// lineFlags groups CLI flags for building or updating billing lines.
type lineFlags struct {
	designation    string
	description    string
	reference      string
	unitPriceCents int64
	unitPriceEuros float64
	quantity       float64
	quantityUnit   string
	productType    string
	vatCode        string
	linesJSON      string
	linesFile      string
}

func bindLineFlags(cmd *cobra.Command, lf *lineFlags) {
	cmd.Flags().StringVar(&lf.designation, "designation", "", "Désignation de la ligne (ex: 'Développement web')")
	cmd.Flags().StringVar(&lf.description, "description", "", "Description détaillée de la prestation")
	cmd.Flags().StringVar(&lf.reference, "reference", "", "Référence produit / SKU")
	cmd.Flags().Int64Var(&lf.unitPriceCents, "unit-price-cents", 0, "Prix unitaire HT en CENTIMES (ex: 50000 = 500.00 €)")
	cmd.Flags().Float64Var(&lf.unitPriceEuros, "unit-price-euros", 0, "Prix unitaire HT en EUROS (converti automatiquement en centimes)")
	cmd.Flags().Float64Var(&lf.quantity, "quantity", 1, "Quantité")
	cmd.Flags().StringVar(&lf.quantityUnit, "unit", string(api.ProductUnitUnit), "Unité (unit, hour, day, fixed_rate, month, kilometer)")
	cmd.Flags().StringVar(&lf.productType, "type", string(api.ProductTypeServiceDelivery), "Nature (service_delivery, sale_of_goods, commercial_or_craft_services, disbursement)")
	cmd.Flags().StringVar(&lf.vatCode, "vat", string(api.VatCodeFR2000), "Code TVA (FR_2000=20%, FR_1000=10%, FR_550=5.5%, FR_210=2.1%, FR_00HT=0%)")
	cmd.Flags().StringVar(&lf.linesJSON, "lines-json", "", "Tableau JSON complet de lignes (remplace les flags unitaires)")
	cmd.Flags().StringVar(&lf.linesFile, "lines-file", "", "Chemin vers un fichier JSON contenant les lignes (UpdateBillingLinesDto ou []UpdateLineDto)")
}

func (lf *lineFlags) hasInput() bool {
	return lf.designation != "" || lf.linesJSON != "" || lf.linesFile != ""
}

func (lf *lineFlags) buildDto() (*api.UpdateBillingLinesDto, error) {
	rawJSON := strings.TrimSpace(lf.linesJSON)
	if lf.linesFile != "" {
		data, err := os.ReadFile(lf.linesFile)
		if err != nil {
			return nil, fmt.Errorf("lecture de --lines-file impossible: %w", err)
		}
		rawJSON = strings.TrimSpace(string(data))
	}

	if rawJSON != "" {
		if strings.HasPrefix(rawJSON, "[") {
			var lines []api.UpdateLineDto
			if err := json.Unmarshal([]byte(rawJSON), &lines); err != nil {
				return nil, fmt.Errorf("JSON de lignes invalide: %w", err)
			}
			return &api.UpdateBillingLinesDto{Lines: lines}, nil
		}
		var dto api.UpdateBillingLinesDto
		if err := json.Unmarshal([]byte(rawJSON), &dto); err != nil {
			return nil, fmt.Errorf("JSON UpdateBillingLinesDto invalide: %w", err)
		}
		return &dto, nil
	}

	if lf.designation == "" {
		return nil, fmt.Errorf("spécifiez --designation (avec --unit-price-cents ou --unit-price-euros) ou --lines-json / --lines-file")
	}

	cents := lf.unitPriceCents
	if cents == 0 && lf.unitPriceEuros != 0 {
		cents = output.EurosToCents(lf.unitPriceEuros)
	}

	line := api.UpdateLineDto{
		Designation:  lf.designation,
		Description:  lf.description,
		Reference:    lf.reference,
		UnitPrice:    cents,
		Quantity:     lf.quantity,
		QuantityUnit: api.ProductUnit(lf.quantityUnit),
		Type:         api.ProductType(lf.productType),
		VatCode:       normalizeVatCode(lf.vatCode),
	}
	return &api.UpdateBillingLinesDto{Lines: []api.UpdateLineDto{line}}, nil
}

// normalizeVatCode accepts both API v2 codes (FR_2000) and shorthand codes (fr_20, 20, 0).
func normalizeVatCode(v string) api.VatCode {
	clean := strings.ToUpper(strings.TrimSpace(v))
	switch clean {
	case "FR_20", "20", "20%":
		return api.VatCodeFR2000
	case "FR_10", "10", "10%":
		return api.VatCodeFR1000
	case "FR_5_5", "5.5", "5,5", "5.5%":
		return api.VatCodeFR550
	case "FR_2_1", "2.1", "2,1", "2.1%":
		return api.VatCodeFR210
	case "FR_8_5", "8.5", "8.5%":
		return api.VatCodeFR850
	case "FR_0", "FR_NOT_APPLICABLE", "0", "0%":
		return api.VatCodeFR00HT
	default:
		return api.VatCode(clean)
	}
}

func newBillingsCmd() *cobra.Command {
	billingsCmd := &cobra.Command{
		Use:     "billings",
		Aliases: []string{"billing"},
		Short:   "Opérations transverses sur tous les documents de facturation (Factures, Devis, Avoirs, Acomptes)",
	}

	// 1. list
	var page, limit int
	var search, customerID, orgID, fromDate, toDate string
	var types, states []string
	var test, archived, late bool
	listCmd := &cobra.Command{
		Use:   "list",
		Short: "Lister les documents de facturation avec filtres (GET /v2/billings)",
		RunE: func(cmd *cobra.Command, args []string) error {
			client, printer, err := getClientAndPrinter()
			if err != nil {
				return err
			}
			filter := api.ListBillingsFilter{
				Page:           page,
				Limit:          limit,
				Search:         search,
				CustomerID:     customerID,
				OrganizationID: orgID,
				Test:           test,
				FromDate:       fromDate,
				ToDate:         toDate,
			}
			for _, t := range types {
				filter.Types = append(filter.Types, api.BillingType(t))
			}
			for _, s := range states {
				filter.States = append(filter.States, api.BillingState(s))
			}
			if cmd.Flags().Changed("archived") {
				filter.Archived = &archived
			}
			if cmd.Flags().Changed("late") {
				filter.Late = &late
			}
			res, err := client.ListBillings(context.Background(), filter)
			if err != nil {
				return err
			}
			return printer.PrintBillingList(res)
		},
	}
	listCmd.Flags().IntVar(&page, "page", 1, "Numéro de page")
	listCmd.Flags().IntVar(&limit, "limit", 20, "Nombre d'éléments par page")
	listCmd.Flags().StringVarP(&search, "search", "s", "", "Recherche textuelle (numéro, titre, client)")
	listCmd.Flags().StringSliceVar(&types, "type", nil, "Filtrer par type (invoice, estimate, asset, advance, purchase_order)")
	listCmd.Flags().StringSliceVar(&states, "state", nil, "Filtrer par état (draft, finalized, signed, refused, paid)")
	listCmd.Flags().StringVar(&customerID, "customer-id", "", "Filtrer par ID client (contactId ou organizationId)")
	listCmd.Flags().StringVar(&orgID, "organization-id", "", "Filtrer par ID organisation")
	listCmd.Flags().BoolVar(&test, "test", false, "Afficher les documents de test")
	listCmd.Flags().BoolVar(&archived, "archived", false, "Filtrer les documents archivés")
	listCmd.Flags().BoolVar(&late, "late", false, "Filtrer les documents en retard de paiement")
	listCmd.Flags().StringVar(&fromDate, "from", "", "Date de début (YYYY-MM-DD)")
	listCmd.Flags().StringVar(&toDate, "to", "", "Date de fin (YYYY-MM-DD)")

	// 2. get
	getCmd := &cobra.Command{
		Use:   "get <documentId>",
		Short: "Récupérer un document de facturation par ID (GET /v2/billing/{documentId})",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			client, printer, err := getClientAndPrinter()
			if err != nil {
				return err
			}
			doc, err := client.GetBillingByID(context.Background(), args[0])
			if err != nil {
				return err
			}
			return printer.PrintBillingDocument(doc)
		},
	}

	// 3. set-lines
	var lf lineFlags
	setLinesCmd := &cobra.Command{
		Use:   "set-lines <billingId>",
		Short: "Définir ou remplacer les lignes d'un document brouillon (PATCH /v2/billing/{billingId}/lines)",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			dto, err := lf.buildDto()
			if err != nil {
				return err
			}
			client, printer, err := getClientAndPrinter()
			if err != nil {
				return err
			}
			doc, err := client.UpdateBillingLines(context.Background(), args[0], *dto)
			if err != nil {
				return err
			}
			return printer.PrintBillingDocument(doc)
		},
	}
	bindLineFlags(setLinesCmd, &lf)

	// 4. finalize
	finalizeCmd := &cobra.Command{
		Use:   "finalize <billingId>",
		Short: "Finaliser un document brouillon (attribue le numéro définitif et le verrouille) (PATCH /v2/billing/{billingId}/finalize)",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			client, printer, err := getClientAndPrinter()
			if err != nil {
				return err
			}
			doc, err := client.FinalizeBilling(context.Background(), args[0])
			if err != nil {
				return err
			}
			return printer.PrintBillingDocument(doc)
		},
	}

	// 5. download
	var outPath string
	downloadCmd := &cobra.Command{
		Use:   "download <billingId>",
		Short: "Télécharger le PDF d'un document de facturation (GET /v2/billing/{billingId}/download)",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			client, printer, err := getClientAndPrinter()
			if err != nil {
				return err
			}
			data, _, err := client.DownloadBillingPDF(context.Background(), args[0])
			if err != nil {
				return err
			}
			dest := outPath
			if dest == "" {
				dest = fmt.Sprintf("abby-%s.pdf", args[0])
			}
			if err := os.WriteFile(dest, data, 0644); err != nil {
				return fmt.Errorf("écriture du fichier PDF impossible: %w", err)
			}
			if printer.Format == output.FormatJSON {
				return printer.PrintJSON(map[string]any{
					"billingId": args[0],
					"file":      dest,
					"bytes":     len(data),
				})
			}
			fmt.Printf("PDF téléchargé avec succès : %s (%d octets)\n", dest, len(data))
			return nil
		},
	}
	downloadCmd.Flags().StringVarP(&outPath, "out", "f", "", "Chemin du fichier PDF de sortie [défaut: abby-<id>.pdf]")

	// 6. send-email
	var emails []string
	var subject, content string
	var inCopy bool
	sendEmailCmd := &cobra.Command{
		Use:   "send-email <billingId>",
		Short: "Envoyer un document finalisé par email (POST /v2/billing/{billingId}/send-by-email)",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(emails) == 0 || subject == "" || content == "" {
				return fmt.Errorf("--to, --subject et --content sont obligatoires")
			}
			client, printer, err := getClientAndPrinter()
			if err != nil {
				return err
			}
			dto := api.SendBillingByEmailDto{
				Emails:  emails,
				Subject: subject,
				Content: content,
				InCopy:  inCopy,
			}
			if err := client.SendBillingByEmail(context.Background(), args[0], dto); err != nil {
				return err
			}
			if printer.Format == output.FormatJSON {
				return printer.PrintJSON(map[string]any{
					"sent":      true,
					"billingId": args[0],
					"emails":    emails,
				})
			}
			fmt.Printf("Document %s envoyé par email à %v.\n", args[0], emails)
			return nil
		},
	}
	sendEmailCmd.Flags().StringSliceVar(&emails, "to", nil, "Destinataire(s) email (séparés par virgule)")
	sendEmailCmd.Flags().StringVar(&subject, "subject", "", "Objet de l'email")
	sendEmailCmd.Flags().StringVar(&content, "content", "", "Corps de l'email (HTML supporté)")
	sendEmailCmd.Flags().BoolVar(&inCopy, "copy", false, "Recevoir une copie de l'email")

	// 7. set-title
	setTitleCmd := &cobra.Command{
		Use:   "set-title <documentId> <title>",
		Short: "Modifier le titre d'un document (PATCH /v2/billing/{documentId}/title)",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			client, printer, err := getClientAndPrinter()
			if err != nil {
				return err
			}
			doc, err := client.UpdateBillingTitle(context.Background(), args[0], args[1])
			if err != nil {
				return err
			}
			return printer.PrintBillingDocument(doc)
		},
	}

	// 8. archive / unarchive / delete
	archiveCmd := &cobra.Command{
		Use:   "archive <documentId>",
		Short: "Archiver un document de facturation (PATCH /v2/billing/{documentId}/archive)",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			client, printer, err := getClientAndPrinter()
			if err != nil {
				return err
			}
			if err := client.ArchiveBilling(context.Background(), args[0]); err != nil {
				return err
			}
			if printer.Format == output.FormatJSON {
				return printer.PrintJSON(map[string]any{"archived": true, "id": args[0]})
			}
			fmt.Printf("Document %s archivé.\n", args[0])
			return nil
		},
	}

	unarchiveCmd := &cobra.Command{
		Use:   "unarchive <documentId>",
		Short: "Désarchiver un document de facturation (PATCH /v2/billing/{documentId}/unarchive)",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			client, printer, err := getClientAndPrinter()
			if err != nil {
				return err
			}
			if err := client.UnarchiveBilling(context.Background(), args[0]); err != nil {
				return err
			}
			if printer.Format == output.FormatJSON {
				return printer.PrintJSON(map[string]any{"archived": false, "id": args[0]})
			}
			fmt.Printf("Document %s désarchivé.\n", args[0])
			return nil
		},
	}

	deleteCmd := &cobra.Command{
		Use:   "delete <documentId>",
		Short: "Supprimer définitivement un document brouillon (DELETE /v2/billing/{documentId})",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			client, printer, err := getClientAndPrinter()
			if err != nil {
				return err
			}
			if err := client.DeleteBillingDocument(context.Background(), args[0]); err != nil {
				return err
			}
			if printer.Format == output.FormatJSON {
				return printer.PrintJSON(map[string]any{"deleted": true, "id": args[0]})
			}
			fmt.Printf("Document brouillon %s supprimé.\n", args[0])
			return nil
		},
	}

	billingsCmd.AddCommand(
		listCmd,
		getCmd,
		setLinesCmd,
		finalizeCmd,
		downloadCmd,
		sendEmailCmd,
		setTitleCmd,
		archiveCmd,
		unarchiveCmd,
		deleteCmd,
	)
	return billingsCmd
}
