package cli

import (
	"context"
	"fmt"
	"os"

	"github.com/a2dev-fr/abby-cli/internal/api"
	"github.com/a2dev-fr/abby-cli/internal/output"
	"github.com/spf13/cobra"
)

func newEstimatesCmd() *cobra.Command {
	estimatesCmd := &cobra.Command{
		Use:     "estimates",
		Aliases: []string{"estimate", "devis"},
		Short:   "Gérer le cycle de vie des devis et bons de commande (Création, Signature, Acomptes, Facture de solde)",
	}

	// 1. list
	var page, limit int
	var search, customerID string
	var states []string
	var test, purchaseOrders bool
	listCmd := &cobra.Command{
		Use:   "list",
		Short: "Lister les devis ou bons de commande (GET /v2/billings?type=estimate)",
		RunE: func(cmd *cobra.Command, args []string) error {
			client, printer, err := getClientAndPrinter()
			if err != nil {
				return err
			}
			docType := api.BillingTypeEstimate
			if purchaseOrders {
				docType = api.BillingTypePurchaseOrder
			}
			filter := api.ListBillingsFilter{
				Page:       page,
				Limit:      limit,
				Search:     search,
				Types:      []api.BillingType{docType},
				CustomerID: customerID,
				Test:       test,
			}
			for _, s := range states {
				filter.States = append(filter.States, api.BillingState(s))
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
	listCmd.Flags().StringVarP(&search, "search", "s", "", "Recherche textuelle")
	listCmd.Flags().StringSliceVar(&states, "state", nil, "Filtrer par état (draft, finalized, signed, refused, paid)")
	listCmd.Flags().StringVar(&customerID, "customer-id", "", "Filtrer par ID client")
	listCmd.Flags().BoolVar(&purchaseOrders, "purchase-order", false, "Lister les bons de commande (purchase_order) au lieu des devis")
	listCmd.Flags().BoolVar(&test, "test", false, "Afficher les devis de test")

	// 2. get
	getCmd := &cobra.Command{
		Use:   "get <estimateId>",
		Short: "Récupérer un devis par ID (GET /v2/billing/{documentId})",
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

	// 3. create
	var createLines lineFlags
	var createTitle string
	var isPurchaseOrder bool
	createCmd := &cobra.Command{
		Use:   "create <customerId>",
		Short: "Créer un nouveau devis (ou bon de commande) en brouillon (POST /v2/billing/estimate/{customerId})",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			client, printer, err := getClientAndPrinter()
			if err != nil {
				return err
			}
			ctx := context.Background()
			estType := "estimate"
			if isPurchaseOrder {
				estType = "purchase_order"
			}
			doc, err := client.CreateEstimate(ctx, args[0], estType)
			if err != nil {
				return err
			}
			if createTitle != "" {
				if updated, tErr := client.UpdateBillingTitle(ctx, doc.ID, createTitle); tErr == nil {
					doc = updated
				}
			}
			if createLines.hasInput() {
				dto, lErr := createLines.buildDto()
				if lErr != nil {
					return fmt.Errorf("devis %s créé en brouillon, mais erreur sur les lignes: %w", doc.ID, lErr)
				}
				doc, err = client.UpdateBillingLines(ctx, doc.ID, *dto)
				if err != nil {
					return fmt.Errorf("devis %s créé en brouillon, mais échec PATCH lines: %w", doc.ID, err)
				}
			}
			return printer.PrintBillingDocument(doc)
		},
	}
	createCmd.Flags().StringVar(&createTitle, "title", "", "Titre du devis")
	createCmd.Flags().BoolVar(&isPurchaseOrder, "purchase-order", false, "Créer un bon de commande (purchase_order) au lieu d'un devis")
	bindLineFlags(createCmd, &createLines)

	// 4. set-lines
	var setLines lineFlags
	setLinesCmd := &cobra.Command{
		Use:   "set-lines <estimateId>",
		Short: "Mettre à jour les lignes d'un devis brouillon (PATCH /v2/billing/{billingId}/lines)",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			dto, err := setLines.buildDto()
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
	bindLineFlags(setLinesCmd, &setLines)

	// 5. update-timeline
	var emittedAtStr, expiredAtStr, paymentDelayStr string
	updateTimelineCmd := &cobra.Command{
		Use:   "update-timeline <estimateId>",
		Short: "Mettre à jour les dates du devis (émission, expiration, délai de paiement) (PATCH /v2/billing/estimate/{estimateId}/timeline)",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			emittedSec, err := output.ParseDateOrUnixSeconds(emittedAtStr)
			if err != nil {
				return fmt.Errorf("--emitted-at: %w", err)
			}
			expiredSec, err := output.ParseDateOrUnixSeconds(expiredAtStr)
			if err != nil {
				return fmt.Errorf("--expired-at: %w", err)
			}
			dto := api.UpdateTimelineDto{
				EmittedAt:    emittedSec,
				ExpiredAt:    &expiredSec,
				PaymentDelay: api.PaymentDelay(paymentDelayStr),
			}
			client, printer, err := getClientAndPrinter()
			if err != nil {
				return err
			}
			doc, err := client.UpdateEstimateTimeline(context.Background(), args[0], dto)
			if err != nil {
				return err
			}
			return printer.PrintBillingDocument(doc)
		},
	}
	updateTimelineCmd.Flags().StringVar(&emittedAtStr, "emitted-at", "today", "Date d'émission (YYYY-MM-DD ou secondes Unix)")
	updateTimelineCmd.Flags().StringVar(&expiredAtStr, "expired-at", "", "Date d'expiration du devis (YYYY-MM-DD ou secondes Unix) (obligatoire)")
	updateTimelineCmd.Flags().StringVar(&paymentDelayStr, "payment-delay", string(api.PaymentDelayThirtyDays), "Délai de paiement")

	// 6. finalize
	finalizeCmd := &cobra.Command{
		Use:   "finalize <estimateId>",
		Short: "Finaliser un devis brouillon (PATCH /v2/billing/{billingId}/finalize)",
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

	// 7. sign / unsign / refuse / unrefuse
	signCmd := &cobra.Command{
		Use:   "sign <estimateId>",
		Short: "Marquer un devis finalisé comme signé (PATCH /v2/billing/estimate/{estimateId}/sign)",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			client, printer, err := getClientAndPrinter()
			if err != nil {
				return err
			}
			doc, err := client.SignEstimate(context.Background(), args[0])
			if err != nil {
				return err
			}
			return printer.PrintBillingDocument(doc)
		},
	}

	unsignCmd := &cobra.Command{
		Use:   "unsign <estimateId>",
		Short: "Annuler la signature d'un devis (PATCH /v2/billing/estimate/{estimateId}/unsign)",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			client, printer, err := getClientAndPrinter()
			if err != nil {
				return err
			}
			doc, err := client.UnsignEstimate(context.Background(), args[0])
			if err != nil {
				return err
			}
			return printer.PrintBillingDocument(doc)
		},
	}

	refuseCmd := &cobra.Command{
		Use:   "refuse <estimateId>",
		Short: "Marquer un devis finalisé comme refusé (PATCH /v2/billing/estimate/{estimateId}/refuse)",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			client, printer, err := getClientAndPrinter()
			if err != nil {
				return err
			}
			doc, err := client.RefuseEstimate(context.Background(), args[0])
			if err != nil {
				return err
			}
			return printer.PrintBillingDocument(doc)
		},
	}

	unrefuseCmd := &cobra.Command{
		Use:   "unrefuse <estimateId>",
		Short: "Annuler le refus d'un devis (PATCH /v2/billing/estimate/{estimateId}/unrefuse)",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			client, printer, err := getClientAndPrinter()
			if err != nil {
				return err
			}
			doc, err := client.UnrefuseEstimate(context.Background(), args[0])
			if err != nil {
				return err
			}
			return printer.PrintBillingDocument(doc)
		},
	}

	// 8. create-advance
	var advPercentage float64
	var advAmountCents int64
	var advAmountEuros float64
	createAdvanceCmd := &cobra.Command{
		Use:   "create-advance <estimateId>",
		Short: "Créer une facture d'acompte (brouillon) depuis un devis signé (POST /v2/billing/estimate/{id}/advance)",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			client, printer, err := getClientAndPrinter()
			if err != nil {
				return err
			}
			var dto *api.CreateAdvanceDto
			if advPercentage > 0 {
				dto = &api.CreateAdvanceDto{
					Mode:       "PERCENTAGE",
					Percentage: advPercentage,
				}
			} else if advAmountCents > 0 || advAmountEuros > 0 {
				cents := advAmountCents
				if cents == 0 {
					cents = output.EurosToCents(advAmountEuros)
				}
				dto = &api.CreateAdvanceDto{
					Mode:   "AMOUNT",
					Amount: &cents,
				}
			}
			doc, err := client.CreateAdvanceFromEstimate(context.Background(), args[0], dto)
			if err != nil {
				return err
			}
			return printer.PrintBillingDocument(doc)
		},
	}
	createAdvanceCmd.Flags().Float64Var(&advPercentage, "percentage", 0, "Pourcentage d'acompte (ex: 30 pour 30%) [défaut API sans body: 30%]")
	createAdvanceCmd.Flags().Int64Var(&advAmountCents, "amount-cents", 0, "Montant fixe d'acompte HT en centimes")
	createAdvanceCmd.Flags().Float64Var(&advAmountEuros, "amount-euros", 0, "Montant fixe d'acompte HT en euros")

	// 9. create-final-invoice
	createFinalInvoiceCmd := &cobra.Command{
		Use:   "create-final-invoice <estimateId>",
		Short: "Créer la facture finale (de solde) depuis un devis signé (POST /v2/billing/estimate/{id}/final-invoice)",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			client, printer, err := getClientAndPrinter()
			if err != nil {
				return err
			}
			doc, err := client.CreateFinalInvoiceFromEstimate(context.Background(), args[0])
			if err != nil {
				return err
			}
			return printer.PrintBillingDocument(doc)
		},
	}

	// 10. download
	var outPDF string
	downloadCmd := &cobra.Command{
		Use:   "download <estimateId>",
		Short: "Télécharger le PDF du devis (GET /v2/billing/{billingId}/download)",
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
			dest := outPDF
			if dest == "" {
				dest = fmt.Sprintf("devis-%s.pdf", args[0])
			}
			if err := os.WriteFile(dest, data, 0644); err != nil {
				return err
			}
			if printer.Format == output.FormatJSON {
				return printer.PrintJSON(map[string]any{"estimateId": args[0], "file": dest, "bytes": len(data)})
			}
			fmt.Printf("Devis téléchargé : %s (%d octets)\n", dest, len(data))
			return nil
		},
	}
	downloadCmd.Flags().StringVarP(&outPDF, "out", "f", "", "Chemin du fichier PDF")

	estimatesCmd.AddCommand(
		listCmd,
		getCmd,
		createCmd,
		setLinesCmd,
		updateTimelineCmd,
		finalizeCmd,
		signCmd,
		unsignCmd,
		refuseCmd,
		unrefuseCmd,
		createAdvanceCmd,
		createFinalInvoiceCmd,
		downloadCmd,
	)
	return estimatesCmd
}
