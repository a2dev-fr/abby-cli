package cli

import (
	"context"
	"fmt"
	"os"

	"github.com/a2dev-fr/abby-cli/internal/api"
	"github.com/a2dev-fr/abby-cli/internal/output"
	"github.com/spf13/cobra"
)

func newAssetsCmd() *cobra.Command {
	assetsCmd := &cobra.Command{
		Use:     "assets",
		Aliases: []string{"asset", "avoirs", "avoir"},
		Short:   "Gérer les avoirs (Notes de crédit depuis une facture, un acompte ou vierges)",
	}

	// 1. list
	var page, limit int
	var search, customerID string
	var states []string
	var test bool
	listCmd := &cobra.Command{
		Use:   "list",
		Short: "Lister les avoirs (GET /v2/billings?type=asset)",
		RunE: func(cmd *cobra.Command, args []string) error {
			client, printer, err := getClientAndPrinter()
			if err != nil {
				return err
			}
			filter := api.ListBillingsFilter{
				Page:       page,
				Limit:      limit,
				Search:     search,
				Types:      []api.BillingType{api.BillingTypeAsset},
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
	listCmd.Flags().IntVar(&limit, "limit", 20, "Nombre d'avoirs par page")
	listCmd.Flags().StringVarP(&search, "search", "s", "", "Recherche textuelle")
	listCmd.Flags().StringSliceVar(&states, "state", nil, "Filtrer par état (draft, finalized, paid)")
	listCmd.Flags().StringVar(&customerID, "customer-id", "", "Filtrer par ID client")
	listCmd.Flags().BoolVar(&test, "test", false, "Afficher les avoirs de test")

	// 2. get
	getCmd := &cobra.Command{
		Use:   "get <assetId>",
		Short: "Récupérer un avoir par ID (GET /v2/billing/asset/{assetId})",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			client, printer, err := getClientAndPrinter()
			if err != nil {
				return err
			}
			doc, err := client.GetAsset(context.Background(), args[0])
			if err != nil {
				return err
			}
			return printer.PrintBillingDocument(doc)
		},
	}

	// 3. create (blank asset for customerId)
	var createLines lineFlags
	createCmd := &cobra.Command{
		Use:   "create <customerId>",
		Short: "Créer un avoir vierge en brouillon pour un client (POST /v2/billing/asset/{customerId})",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			client, printer, err := getClientAndPrinter()
			if err != nil {
				return err
			}
			ctx := context.Background()
			doc, err := client.CreateBlankAsset(ctx, args[0])
			if err != nil {
				return err
			}
			if createLines.hasInput() {
				dto, lErr := createLines.buildDto()
				if lErr != nil {
					return fmt.Errorf("avoir %s créé en brouillon, mais erreur sur les lignes: %w", doc.ID, lErr)
				}
				doc, err = client.UpdateBillingLines(ctx, doc.ID, *dto)
				if err != nil {
					return fmt.Errorf("avoir %s créé en brouillon, mais échec PATCH lines: %w", doc.ID, err)
				}
			}
			return printer.PrintBillingDocument(doc)
		},
	}
	bindLineFlags(createCmd, &createLines)

	// 4. create-from-invoice
	createFromInvoiceCmd := &cobra.Command{
		Use:   "create-from-invoice <invoiceId>",
		Short: "Créer un avoir en brouillon à partir d'une facture finalisée (POST /v2/billing/invoice/{invoiceId}/asset)",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			client, printer, err := getClientAndPrinter()
			if err != nil {
				return err
			}
			doc, err := client.CreateAssetFromInvoice(context.Background(), args[0])
			if err != nil {
				return err
			}
			return printer.PrintBillingDocument(doc)
		},
	}

	// 5. create-from-advance
	createFromAdvanceCmd := &cobra.Command{
		Use:   "create-from-advance <advanceId>",
		Short: "Créer un avoir en brouillon à partir d'une facture d'acompte (POST /v2/billing/advance/{advanceId}/asset)",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			client, printer, err := getClientAndPrinter()
			if err != nil {
				return err
			}
			doc, err := client.CreateAssetFromAdvance(context.Background(), args[0])
			if err != nil {
				return err
			}
			return printer.PrintBillingDocument(doc)
		},
	}

	// 6. set-lines
	var setLines lineFlags
	setLinesCmd := &cobra.Command{
		Use:   "set-lines <assetId>",
		Short: "Mettre à jour les lignes d'un avoir brouillon (PATCH /v2/billing/{billingId}/lines)",
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

	// 7. finalize
	finalizeCmd := &cobra.Command{
		Use:   "finalize <assetId>",
		Short: "Finaliser un avoir brouillon (PATCH /v2/billing/{billingId}/finalize)",
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

	// 8. download
	var outPDF string
	downloadCmd := &cobra.Command{
		Use:   "download <assetId>",
		Short: "Télécharger le PDF de l'avoir (GET /v2/billing/{billingId}/download)",
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
				dest = fmt.Sprintf("avoir-%s.pdf", args[0])
			}
			if err := os.WriteFile(dest, data, 0644); err != nil {
				return err
			}
			if printer.Format == output.FormatJSON {
				return printer.PrintJSON(map[string]any{"assetId": args[0], "file": dest, "bytes": len(data)})
			}
			fmt.Printf("Avoir téléchargé : %s (%d octets)\n", dest, len(data))
			return nil
		},
	}
	downloadCmd.Flags().StringVarP(&outPDF, "out", "f", "", "Chemin du fichier PDF")

	// 9. update-info
	var headerNote, footerNote, terms string
	updateInfoCmd := &cobra.Command{
		Use:   "update-info <assetId>",
		Short: "Mettre à jour les informations générales de l'avoir (PATCH /v2/billing/asset/{assetId}/general-informations)",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			client, printer, err := getClientAndPrinter()
			if err != nil {
				return err
			}
			doc, err := client.UpdateAssetGeneralInfo(context.Background(), args[0], api.UpdateGeneralInformationsDto{
				HeaderNote:                      headerNote,
				FooterNote:                      footerNote,
				GeneralTermsAndConditionsOfSale: terms,
			})
			if err != nil {
				return err
			}
			return printer.PrintBillingDocument(doc)
		},
	}
	updateInfoCmd.Flags().StringVar(&headerNote, "header-note", "", "Note d'en-tête")
	updateInfoCmd.Flags().StringVar(&footerNote, "footer-note", "", "Note de pied de page")
	updateInfoCmd.Flags().StringVar(&terms, "terms", "", "Conditions générales")

	// 10. update-timeline
	var emittedAtStr, paymentDelayStr string
	updateTimelineCmd := &cobra.Command{
		Use:   "update-timeline <assetId>",
		Short: "Mettre à jour les dates de l'avoir (PATCH /v2/billing/asset/{assetId}/timeline)",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			emittedSec, err := output.ParseDateOrUnixSeconds(emittedAtStr)
			if err != nil {
				return fmt.Errorf("--emitted-at: %w", err)
			}
			client, printer, err := getClientAndPrinter()
			if err != nil {
				return err
			}
			doc, err := client.UpdateAssetTimeline(context.Background(), args[0], api.UpdateTimelineDto{
				EmittedAt:    emittedSec,
				PaymentDelay: api.PaymentDelay(paymentDelayStr),
			})
			if err != nil {
				return err
			}
			return printer.PrintBillingDocument(doc)
		},
	}
	updateTimelineCmd.Flags().StringVar(&emittedAtStr, "emitted-at", "today", "Date d'émission (YYYY-MM-DD ou secondes Unix)")
	updateTimelineCmd.Flags().StringVar(&paymentDelayStr, "payment-delay", string(api.PaymentDelayAtReception), "Délai de remboursement / paiement")

	// 11. send-email
	var emails []string
	var subject, content string
	var inCopy bool
	sendEmailCmd := &cobra.Command{
		Use:   "send-email <assetId>",
		Short: "Envoyer l'avoir par email au client (POST /v2/billing/{billingId}/send-by-email)",
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
				return printer.PrintJSON(map[string]any{"sent": true, "assetId": args[0], "emails": emails})
			}
			fmt.Printf("Avoir %s envoyé par email à %v.\n", args[0], emails)
			return nil
		},
	}
	sendEmailCmd.Flags().StringSliceVar(&emails, "to", nil, "Destinataire(s) email")
	sendEmailCmd.Flags().StringVar(&subject, "subject", "", "Objet de l'email")
	sendEmailCmd.Flags().StringVar(&content, "content", "", "Corps de l'email")
	sendEmailCmd.Flags().BoolVar(&inCopy, "copy", false, "Recevoir une copie")

	assetsCmd.AddCommand(
		listCmd,
		getCmd,
		createCmd,
		createFromInvoiceCmd,
		createFromAdvanceCmd,
		setLinesCmd,
		updateInfoCmd,
		updateTimelineCmd,
		finalizeCmd,
		downloadCmd,
		sendEmailCmd,
	)
	return assetsCmd
}
