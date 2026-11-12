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

func newInvoicesCmd() *cobra.Command {
	invoicesCmd := &cobra.Command{
		Use:     "invoices",
		Aliases: []string{"invoice", "factures", "facture"},
		Short:   "Gérer le cycle de vie des factures (Création, Lignes, Finalisation, Paiement, Avoir)",
	}

	// 1. list
	var page, limit int
	var search, customerID, fromDate, toDate string
	var states []string
	var test, late bool
	listCmd := &cobra.Command{
		Use:   "list",
		Short: "Lister les factures (GET /v2/billings?type=invoice)",
		RunE: func(cmd *cobra.Command, args []string) error {
			client, printer, err := getClientAndPrinter()
			if err != nil {
				return err
			}
			filter := api.ListBillingsFilter{
				Page:       page,
				Limit:      limit,
				Search:     search,
				Types:      []api.BillingType{api.BillingTypeInvoice},
				CustomerID: customerID,
				Test:       test,
				FromDate:   fromDate,
				ToDate:     toDate,
			}
			for _, s := range states {
				filter.States = append(filter.States, api.BillingState(s))
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
	listCmd.Flags().IntVar(&limit, "limit", 20, "Nombre de factures par page")
	listCmd.Flags().StringVarP(&search, "search", "s", "", "Recherche par numéro, titre ou client")
	listCmd.Flags().StringSliceVar(&states, "state", nil, "Filtrer par état (draft, finalized, paid)")
	listCmd.Flags().StringVar(&customerID, "customer-id", "", "Filtrer par ID client (contactId ou organizationId)")
	listCmd.Flags().BoolVar(&test, "test", false, "Afficher les factures de test")
	listCmd.Flags().BoolVar(&late, "late", false, "Filtrer les factures en retard de paiement")
	listCmd.Flags().StringVar(&fromDate, "from", "", "Date de début (YYYY-MM-DD)")
	listCmd.Flags().StringVar(&toDate, "to", "", "Date de fin (YYYY-MM-DD)")

	// 2. get
	getCmd := &cobra.Command{
		Use:   "get <invoiceId>",
		Short: "Récupérer les détails d'une facture (GET /v2/billing/invoice/{invoiceId})",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			client, printer, err := getClientAndPrinter()
			if err != nil {
				return err
			}
			doc, err := client.GetInvoice(context.Background(), args[0])
			if err != nil {
				return err
			}
			return printer.PrintBillingDocument(doc)
		},
	}

	// 3. create (supports optional immediate line creation & title)
	var createLines lineFlags
	var createTitle string
	createCmd := &cobra.Command{
		Use:   "create <customerId>",
		Short: "Créer une nouvelle facture en brouillon (POST /v2/billing/invoice/{customerId})",
		Long: `Crée une facture en brouillon pour le client spécifié (contactId ou organizationId).
Si vous fournissez également des options de ligne (--designation et --unit-price-cents / --unit-price-euros, ou --lines-json),
les lignes sont automatiquement ajoutées au brouillon après sa création.`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			client, printer, err := getClientAndPrinter()
			if err != nil {
				return err
			}
			ctx := context.Background()
			doc, err := client.CreateInvoice(ctx, args[0])
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
					return fmt.Errorf("facture %s créée en brouillon, mais erreur sur les lignes: %w", doc.ID, lErr)
				}
				doc, err = client.UpdateBillingLines(ctx, doc.ID, *dto)
				if err != nil {
					return fmt.Errorf("facture %s créée en brouillon, mais échec PATCH lines: %w", doc.ID, err)
				}
			}
			return printer.PrintBillingDocument(doc)
		},
	}
	createCmd.Flags().StringVar(&createTitle, "title", "", "Titre optionnel de la facture")
	bindLineFlags(createCmd, &createLines)

	// 4. set-lines
	var setLines lineFlags
	setLinesCmd := &cobra.Command{
		Use:   "set-lines <invoiceId>",
		Short: "Mettre à jour les lignes d'une facture brouillon (PATCH /v2/billing/{billingId}/lines)",
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

	// 5. update-info
	var headerNote, footerNote, terms, mediator, otherMethod string
	var paymentMethods []string
	updateInfoCmd := &cobra.Command{
		Use:   "update-info <invoiceId>",
		Short: "Mettre à jour les informations générales (notes, CGV, moyens de paiement) (PATCH /v2/billing/invoice/{invoiceId}/general-informations)",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			client, printer, err := getClientAndPrinter()
			if err != nil {
				return err
			}
			dto := api.UpdateGeneralInformationsDto{
				HeaderNote:                      headerNote,
				FooterNote:                      footerNote,
				GeneralTermsAndConditionsOfSale: terms,
				MentionMediator:                 mediator,
				OtherPaymentMethod:              otherMethod,
			}
			for _, m := range paymentMethods {
				dto.PaymentMethods = append(dto.PaymentMethods, api.PaymentMethod(m))
			}
			doc, err := client.UpdateInvoiceGeneralInfo(context.Background(), args[0], dto)
			if err != nil {
				return err
			}
			return printer.PrintBillingDocument(doc)
		},
	}
	updateInfoCmd.Flags().StringVar(&headerNote, "header-note", "", "Note d'en-tête")
	updateInfoCmd.Flags().StringVar(&footerNote, "footer-note", "", "Note de pied de page")
	updateInfoCmd.Flags().StringVar(&terms, "terms", "", "Conditions générales de vente")
	updateInfoCmd.Flags().StringVar(&mediator, "mediator", "", "Mention médiateur")
	updateInfoCmd.Flags().StringSliceVar(&paymentMethods, "payment-methods", nil, "Moyens de paiement acceptés (transfer, credit_card, cheque, cash, direct_debit, other)")
	updateInfoCmd.Flags().StringVar(&otherMethod, "other-payment-method", "", "Précision si moyen de paiement = other")

	// 6. update-timeline
	var emittedAtStr, paymentDelayStr, customDueStr, deliveryStr string
	updateTimelineCmd := &cobra.Command{
		Use:   "update-timeline <invoiceId>",
		Short: "Mettre à jour les dates (émission, délai de paiement, livraison) en secondes Unix ou YYYY-MM-DD (PATCH /v2/billing/invoice/{invoiceId}/timeline)",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			emittedSec, err := output.ParseDateOrUnixSeconds(emittedAtStr)
			if err != nil {
				return fmt.Errorf("--emitted-at: %w", err)
			}
			dto := api.UpdateTimelineDto{
				EmittedAt:    emittedSec,
				PaymentDelay: api.PaymentDelay(paymentDelayStr),
			}
			if customDueStr != "" {
				dueSec, err := output.ParseDateOrUnixSeconds(customDueStr)
				if err != nil {
					return fmt.Errorf("--custom-due-date: %w", err)
				}
				dto.CustomDueDate = &dueSec
			}
			if deliveryStr != "" {
				delSec, err := output.ParseDateOrUnixSeconds(deliveryStr)
				if err != nil {
					return fmt.Errorf("--delivery-date: %w", err)
				}
				dto.DeliveryDate = &delSec
			}
			client, printer, err := getClientAndPrinter()
			if err != nil {
				return err
			}
			doc, err := client.UpdateInvoiceTimeline(context.Background(), args[0], dto)
			if err != nil {
				return err
			}
			return printer.PrintBillingDocument(doc)
		},
	}
	updateTimelineCmd.Flags().StringVar(&emittedAtStr, "emitted-at", "today", "Date d'émission (YYYY-MM-DD, 'today' ou timestamp Unix en secondes)")
	updateTimelineCmd.Flags().StringVar(&paymentDelayStr, "payment-delay", string(api.PaymentDelayThirtyDays), "Délai de paiement (at_reception, thirty_days, thirty_days_end_of_month, forty_five_days, sixty_days, other)")
	updateTimelineCmd.Flags().StringVar(&customDueStr, "custom-due-date", "", "Date d'échéance personnalisée si payment-delay=other (YYYY-MM-DD ou secondes Unix)")
	updateTimelineCmd.Flags().StringVar(&deliveryStr, "delivery-date", "", "Date de livraison (YYYY-MM-DD ou secondes Unix)")

	// 7. finalize
	finalizeCmd := &cobra.Command{
		Use:   "finalize <invoiceId>",
		Short: "Finaliser une facture brouillon (irréversible : attribue un numéro officiel et verrouille la facture) (PATCH /v2/billing/{billingId}/finalize)",
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

	// 8. mark-paid
	var paidAmountCents int64
	var paidAmountEuros float64
	var paidMethod, paidDate, paidTxID string
	markPaidCmd := &cobra.Command{
		Use:   "mark-paid <invoiceId>",
		Short: "Marquer une facture finalisée comme payée (POST /v2/accounting-billing/invoice/{invoiceId}/reconciliate)",
		Long: `Associe un règlement à une facture finalisée pour l'encaisser.
Si aucun montant (--amount-cents ou --amount-euros) n'est fourni, le CLI récupère automatiquement le montant TTC restant à payer sur la facture.`,
		Args: cobra.ExactArgs(1),
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
				inv, err := client.GetInvoice(ctx, args[0])
				if err != nil {
					return fmt.Errorf("impossible de récupérer le montant TTC de la facture: %w", err)
				}
				if inv.Total != nil {
					if inv.Total.RemainingWithTaxes > 0 {
						cents = inv.Total.RemainingWithTaxes
					} else {
						cents = inv.Total.EffectiveWithTaxes()
					}
				}
				if cents <= 0 {
					return fmt.Errorf("impossible de déterminer automatiquement le montant TTC à encaisser : précisez --amount-cents ou --amount-euros")
				}
			}
			recDate := paidDate
			if recDate == "" {
				recDate = time.Now().Format("2006-01-02")
			}
			dto := api.ReconciliateInvoiceDto{
				Payments: []api.PaymentDto{
					{
						Amount:        cents,
						ReceivedAt:    recDate,
						Method:        api.PaymentMethod(paidMethod),
						TransactionID: paidTxID,
					},
				},
			}
			if err := client.MarkInvoiceAsPaid(ctx, args[0], dto); err != nil {
				return err
			}
			if printer.Format == output.FormatJSON {
				return printer.PrintJSON(map[string]any{
					"invoiceId":   args[0],
					"state":       "paid",
					"amountCents": cents,
					"method":      paidMethod,
					"receivedAt":  recDate,
				})
			}
			fmt.Printf("Facture %s marquée comme payée (%s via %s le %s).\n",
				args[0], output.FormatCents(cents, "EUR"), paidMethod, recDate)
			return nil
		},
	}
	markPaidCmd.Flags().Int64Var(&paidAmountCents, "amount-cents", 0, "Montant encaissé TTC en centimes [défaut: total TTC restant de la facture]")
	markPaidCmd.Flags().Float64Var(&paidAmountEuros, "amount-euros", 0, "Montant encaissé TTC en euros")
	markPaidCmd.Flags().StringVar(&paidMethod, "method", string(api.PaymentMethodTransfer), "Moyen de paiement (transfer, credit_card, cheque, cash, direct_debit, stripe, paypal, other)")
	markPaidCmd.Flags().StringVar(&paidDate, "received-at", "", "Date de réception du paiement YYYY-MM-DD [défaut: aujourd'hui]")
	markPaidCmd.Flags().StringVar(&paidTxID, "transaction-id", "", "ID de transaction bancaire optionnel")

	// 9. mark-unpaid
	markUnpaidCmd := &cobra.Command{
		Use:   "mark-unpaid <invoiceId>",
		Short: "Annuler le statut payé d'une facture (POST /v2/accounting-billing/invoice/{invoiceId}/mark-as-unpaid)",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			client, printer, err := getClientAndPrinter()
			if err != nil {
				return err
			}
			if err := client.MarkInvoiceAsUnpaid(context.Background(), args[0]); err != nil {
				return err
			}
			if printer.Format == output.FormatJSON {
				return printer.PrintJSON(map[string]any{"invoiceId": args[0], "state": "finalized", "paid": false})
			}
			fmt.Printf("Facture %s marquée comme non payée.\n", args[0])
			return nil
		},
	}

	// 10. cancel (creates finalized credit note)
	var archiveAfterCancel bool
	cancelCmd := &cobra.Command{
		Use:   "cancel <invoiceId>",
		Short: "Annuler une facture finalisée en générant automatiquement un avoir finalisé (POST /v2/billing/invoice/{invoiceId}/cancel)",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			client, printer, err := getClientAndPrinter()
			if err != nil {
				return err
			}
			doc, err := client.CancelInvoice(context.Background(), args[0], archiveAfterCancel)
			if err != nil {
				return err
			}
			return printer.PrintBillingDocument(doc)
		},
	}
	cancelCmd.Flags().BoolVar(&archiveAfterCancel, "archive", false, "Archiver la facture et l'avoir après l'annulation")

	// 11. download
	var outPDF string
	downloadCmd := &cobra.Command{
		Use:   "download <invoiceId>",
		Short: "Télécharger le PDF de la facture (GET /v2/billing/{billingId}/download)",
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
				dest = fmt.Sprintf("facture-%s.pdf", args[0])
			}
			if err := os.WriteFile(dest, data, 0644); err != nil {
				return err
			}
			if printer.Format == output.FormatJSON {
				return printer.PrintJSON(map[string]any{"invoiceId": args[0], "file": dest, "bytes": len(data)})
			}
			fmt.Printf("Facture téléchargée : %s (%d octets)\n", dest, len(data))
			return nil
		},
	}
	downloadCmd.Flags().StringVarP(&outPDF, "out", "f", "", "Chemin du fichier PDF")

	// 12. send-email
	var emails []string
	var subject, content string
	var inCopy bool
	sendEmailCmd := &cobra.Command{
		Use:   "send-email <invoiceId>",
		Short: "Envoyer la facture par email au client (POST /v2/billing/{billingId}/send-by-email)",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(emails) == 0 || subject == "" || content == "" {
				return fmt.Errorf("--to, --subject et --content sont obligatoires")
			}
			client, printer, err := getClientAndPrinter()
			if err != nil {
				return err
			}
			err = client.SendBillingByEmail(context.Background(), args[0], api.SendBillingByEmailDto{
				Emails:  emails,
				Subject: subject,
				Content: content,
				InCopy:  inCopy,
			})
			if err != nil {
				return err
			}
			if printer.Format == output.FormatJSON {
				return printer.PrintJSON(map[string]any{"sent": true, "invoiceId": args[0], "emails": emails})
			}
			fmt.Printf("Facture %s envoyée par email à %v.\n", args[0], emails)
			return nil
		},
	}
	sendEmailCmd.Flags().StringSliceVar(&emails, "to", nil, "Destinataire(s) email")
	sendEmailCmd.Flags().StringVar(&subject, "subject", "", "Objet de l'email")
	sendEmailCmd.Flags().StringVar(&content, "content", "", "Corps de l'email")
	sendEmailCmd.Flags().BoolVar(&inCopy, "copy", false, "Recevoir une copie")

	// 13. set-locale
	setLocaleCmd := &cobra.Command{
		Use:   "set-locale <invoiceId> <locale>",
		Short: "Changer la langue de la facture (ex: fr, en) (PUT /v2/billing/invoice/{invoiceId}/locale)",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			client, printer, err := getClientAndPrinter()
			if err != nil {
				return err
			}
			doc, err := client.UpdateInvoiceLocale(context.Background(), args[0], args[1])
			if err != nil {
				return err
			}
			return printer.PrintBillingDocument(doc)
		},
	}

	// 14. set-currency
	setCurrencyCmd := &cobra.Command{
		Use:   "set-currency <invoiceId> <currencyCode>",
		Short: "Changer la devise de la facture (ex: EUR, USD, GBP, CHF) (PUT /v2/billing/invoice/{invoiceId}/currency)",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			client, printer, err := getClientAndPrinter()
			if err != nil {
				return err
			}
			doc, err := client.UpdateInvoiceCurrency(context.Background(), args[0], args[1])
			if err != nil {
				return err
			}
			return printer.PrintBillingDocument(doc)
		},
	}

	// 15. activate-online-payment
	activateOnlinePaymentCmd := &cobra.Command{
		Use:   "activate-online-payment <invoiceId>",
		Short: "Activer le paiement en ligne sur une facture finalisée (PATCH /v2/billing/invoice/{invoiceId}/activate-online-payment)",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			client, printer, err := getClientAndPrinter()
			if err != nil {
				return err
			}
			doc, err := client.ActivateInvoiceOnlinePayment(context.Background(), args[0])
			if err != nil {
				return err
			}
			return printer.PrintBillingDocument(doc)
		},
	}

	invoicesCmd.AddCommand(
		listCmd,
		getCmd,
		createCmd,
		setLinesCmd,
		updateInfoCmd,
		updateTimelineCmd,
		finalizeCmd,
		markPaidCmd,
		markUnpaidCmd,
		cancelCmd,
		downloadCmd,
		sendEmailCmd,
		setLocaleCmd,
		setCurrencyCmd,
		activateOnlinePaymentCmd,
	)
	return invoicesCmd
}
