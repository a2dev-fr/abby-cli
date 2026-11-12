package cli

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/a2dev-fr/abby-cli/internal/api"
	"github.com/a2dev-fr/abby-cli/internal/output"
	"github.com/spf13/cobra"
)

func newAdvancesCmd() *cobra.Command {
	advancesCmd := &cobra.Command{
		Use:     "advances",
		Aliases: []string{"advance", "acomptes", "acompte"},
		Short:   "Gérer les factures d'acompte (/v2/billing/advance/*)",
	}

	// 1. list
	var page, limit int
	var search, customerID string
	var states []string
	var test bool
	listCmd := &cobra.Command{
		Use:   "list",
		Short: "Lister les factures d'acompte (GET /v2/billings?type=advance)",
		RunE: func(cmd *cobra.Command, args []string) error {
			client, printer, err := getClientAndPrinter()
			if err != nil {
				return err
			}
			filter := api.ListBillingsFilter{
				Page:       page,
				Limit:      limit,
				Search:     search,
				Types:      []api.BillingType{api.BillingTypeAdvance},
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
	listCmd.Flags().IntVar(&limit, "limit", 20, "Nombre d'acomptes par page")
	listCmd.Flags().StringVarP(&search, "search", "s", "", "Recherche textuelle")
	listCmd.Flags().StringSliceVar(&states, "state", nil, "Filtrer par état (draft, finalized, paid)")
	listCmd.Flags().StringVar(&customerID, "customer-id", "", "Filtrer par ID client")
	listCmd.Flags().BoolVar(&test, "test", false, "Afficher les acomptes de test")

	// 2. get
	getCmd := &cobra.Command{
		Use:   "get <advanceId>",
		Short: "Récupérer une facture d'acompte par ID (GET /v2/billing/advance/{advanceId})",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			client, printer, err := getClientAndPrinter()
			if err != nil {
				return err
			}
			doc, err := client.GetAdvance(context.Background(), args[0])
			if err != nil {
				return err
			}
			return printer.PrintBillingDocument(doc)
		},
	}

	// 3. update-amount
	var percentage float64
	var amountCents int64
	var amountEuros float64
	updateAmountCmd := &cobra.Command{
		Use:   "update-amount <advanceId>",
		Short: "Modifier le montant ou le pourcentage d'un acompte brouillon (PATCH /v2/billing/advance/{advanceId}/amount)",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			var dto api.CreateAdvanceDto
			if percentage > 0 {
				dto = api.CreateAdvanceDto{
					Mode:       "PERCENTAGE",
					Percentage: percentage,
				}
			} else if amountCents > 0 || amountEuros > 0 {
				cents := amountCents
				if cents == 0 {
					cents = output.EurosToCents(amountEuros)
				}
				dto = api.CreateAdvanceDto{
					Mode:   "AMOUNT",
					Amount: &cents,
				}
			} else {
				return fmt.Errorf("spécifiez --percentage ou --amount-cents / --amount-euros")
			}
			client, printer, err := getClientAndPrinter()
			if err != nil {
				return err
			}
			doc, err := client.UpdateAdvanceAmount(context.Background(), args[0], dto)
			if err != nil {
				return err
			}
			return printer.PrintBillingDocument(doc)
		},
	}
	updateAmountCmd.Flags().Float64Var(&percentage, "percentage", 0, "Pourcentage du devis (ex: 30)")
	updateAmountCmd.Flags().Int64Var(&amountCents, "amount-cents", 0, "Montant HT en centimes")
	updateAmountCmd.Flags().Float64Var(&amountEuros, "amount-euros", 0, "Montant HT en euros")

	// 4. finalize
	finalizeCmd := &cobra.Command{
		Use:   "finalize <advanceId>",
		Short: "Finaliser une facture d'acompte brouillon (PATCH /v2/billing/{billingId}/finalize)",
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

	// 5. mark-paid
	var paidAmountCents int64
	var paidAmountEuros float64
	var paidMethod, paidDate string
	markPaidCmd := &cobra.Command{
		Use:   "mark-paid <advanceId>",
		Short: "Marquer un acompte finalisé comme payé (POST /v2/accounting-billing/advance/{advanceId}/reconciliate)",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			client, printer, err := getClientAndPrinter()
			if err != nil {
				return err
			}
			ctx := context.Background()
			cents := paidAmountCents
			if cents == 0 && paidAmountEuros != 0 {
				cents = output.EurosToCents(paidAmountEuros)
			}
			if cents == 0 {
				adv, err := client.GetAdvance(ctx, args[0])
				if err != nil {
					return fmt.Errorf("impossible de récupérer l'acompte: %w", err)
				}
				if adv.Total != nil {
					if adv.Total.RemainingWithTaxes > 0 {
						cents = adv.Total.RemainingWithTaxes
					} else {
						cents = adv.Total.EffectiveWithTaxes()
					}
				}
				if cents <= 0 {
					return fmt.Errorf("précisez --amount-cents ou --amount-euros")
				}
			}
			recDate := paidDate
			if recDate == "" {
				recDate = time.Now().Format("2006-01-02")
			}
			err = client.MarkAdvanceAsPaid(ctx, args[0], api.ReconciliateInvoiceDto{
				Payments: []api.PaymentDto{
					{
						Amount:     cents,
						ReceivedAt: recDate,
						Method:     api.PaymentMethod(paidMethod),
					},
				},
			})
			if err != nil {
				return err
			}
			if printer.Format == output.FormatJSON {
				return printer.PrintJSON(map[string]any{"advanceId": args[0], "state": "paid", "amountCents": cents})
			}
			fmt.Printf("Acompte %s marqué comme payé (%s).\n", args[0], output.FormatCents(cents, "EUR"))
			return nil
		},
	}
	markPaidCmd.Flags().Int64Var(&paidAmountCents, "amount-cents", 0, "Montant TTC en centimes")
	markPaidCmd.Flags().Float64Var(&paidAmountEuros, "amount-euros", 0, "Montant TTC en euros")
	markPaidCmd.Flags().StringVar(&paidMethod, "method", string(api.PaymentMethodTransfer), "Moyen de paiement")
	markPaidCmd.Flags().StringVar(&paidDate, "received-at", "", "Date YYYY-MM-DD")

	// 6. cancel
	var archiveAfter bool
	cancelCmd := &cobra.Command{
		Use:   "cancel <advanceId>",
		Short: "Annuler un acompte finalisé en créant automatiquement un avoir (POST /v2/billing/advance/{advanceId}/cancel)",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			client, printer, err := getClientAndPrinter()
			if err != nil {
				return err
			}
			doc, err := client.CancelAdvance(context.Background(), args[0], archiveAfter)
			if err != nil {
				return err
			}
			return printer.PrintBillingDocument(doc)
		},
	}
	cancelCmd.Flags().BoolVar(&archiveAfter, "archive", false, "Archiver l'acompte et l'avoir après l'annulation")

	// 7. download
	var outPDF string
	downloadCmd := &cobra.Command{
		Use:   "download <advanceId>",
		Short: "Télécharger le PDF de l'acompte (GET /v2/billing/{billingId}/download)",
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
				dest = fmt.Sprintf("acompte-%s.pdf", args[0])
			}
			if err := os.WriteFile(dest, data, 0644); err != nil {
				return err
			}
			if printer.Format == output.FormatJSON {
				return printer.PrintJSON(map[string]any{"advanceId": args[0], "file": dest, "bytes": len(data)})
			}
			fmt.Printf("Acompte téléchargé : %s (%d octets)\n", dest, len(data))
			return nil
		},
	}
	downloadCmd.Flags().StringVarP(&outPDF, "out", "f", "", "Chemin du fichier PDF")

	advancesCmd.AddCommand(
		listCmd,
		getCmd,
		updateAmountCmd,
		finalizeCmd,
		markPaidCmd,
		cancelCmd,
		downloadCmd,
	)
	return advancesCmd
}
