package cli

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/a2dev-fr/abby-cli/internal/api"
	"github.com/a2dev-fr/abby-cli/internal/output"
	"github.com/spf13/cobra"
)

func buildPurchaseEntries(entriesJSON string, defaultCents int64, isPersonal bool, vatID, accountNumber int) ([]api.PurchaseEntryLineDto, int64, error) {
	if strings.TrimSpace(entriesJSON) != "" {
		var entries []api.PurchaseEntryLineDto
		if err := json.Unmarshal([]byte(entriesJSON), &entries); err != nil {
			return nil, 0, fmt.Errorf("--entries-json invalide: %w", err)
		}
		var sum int64
		for _, e := range entries {
			sum += e.Amount
		}
		return entries, sum, nil
	}
	entry := api.PurchaseEntryLineDto{
		Amount:     defaultCents,
		IsPersonal: isPersonal,
		VatID:      vatID,
	}
	if accountNumber > 0 {
		entry.AccountingAccountNumber = accountNumber
	}
	return []api.PurchaseEntryLineDto{entry}, defaultCents, nil
}

func newAccountingCmd() *cobra.Command {
	accCmd := &cobra.Command{
		Use:     "accounting",
		Aliases: []string{"compta", "books"},
		Short:   "Gérer la comptabilité (Livre des recettes, Livre des achats, Fournisseurs, Plan comptable)",
	}

	accCmd.AddCommand(
		newIncomeBookCmd(),
		newPurchaseRegisterCmd(),
		newProvidersCmd(),
		newAccountingAccountsCmd(),
		newTransactionCmd(),
	)
	return accCmd
}

func newIncomeBookCmd() *cobra.Command {
	incomesCmd := &cobra.Command{
		Use:     "incomes",
		Aliases: []string{"income", "recettes", "recette", "income-book"},
		Short:   "Gérer le Livre des recettes (/incomeBook)",
	}

	var clientName, reference, paidAt string
	var htCents, ttcCents, vatCents int64
	var htEuros, ttcEuros float64
	var vatID, productType, paymentMethod int
	var isSap, isTaxIncluded bool

	createCmd := &cobra.Command{
		Use:   "create",
		Short: "Ajouter une recette au Livre des recettes (POST /incomeBook)",
		RunE: func(cmd *cobra.Command, args []string) error {
			if clientName == "" {
				return fmt.Errorf("--client est obligatoire")
			}
			if htCents == 0 && htEuros != 0 {
				htCents = output.EurosToCents(htEuros)
			}
			if ttcCents == 0 && ttcEuros != 0 {
				ttcCents = output.EurosToCents(ttcEuros)
			}
			if ttcCents == 0 && htCents != 0 {
				ttcCents = htCents + vatCents
			}
			if htCents == 0 && ttcCents != 0 {
				htCents = ttcCents - vatCents
			}
			if htCents == 0 && ttcCents == 0 {
				return fmt.Errorf("spécifiez le montant via --ht-cents / --ttc-cents (ou --ht-euros / --ttc-euros)")
			}
			if paidAt == "" {
				paidAt = time.Now().Format("2006-01-02")
			}

			client, printer, err := getClientAndPrinter()
			if err != nil {
				return err
			}
			dto := api.CreateIncomeBookItemDto{
				Client:            clientName,
				Reference:         reference,
				PaidAt:            paidAt,
				PriceWithoutTax:   htCents,
				PriceTotalTax:     ttcCents,
				VatAmount:         vatCents,
				VatID:             vatID,
				ProductType:       productType,
				PaymentMethodUsed: map[string]any{"value": paymentMethod},
				IsSap:             isSap,
				IsTaxIncluded:     isTaxIncluded,
			}
			item, err := client.CreateIncomeBookItem(context.Background(), dto)
			if err != nil {
				return err
			}
			if printer.Format == output.FormatJSON {
				return printer.PrintJSON(item)
			}
			fmt.Printf("Recette ajoutée au Livre des recettes ! ID: %s (Client: %s, TTC: %s)\n",
				item.ID, clientName, output.FormatCents(item.PriceTotalTax, "EUR"))
			return nil
		},
	}
	createCmd.Flags().StringVar(&clientName, "client", "", "Nom du client (obligatoire)")
	createCmd.Flags().StringVar(&reference, "reference", "", "Référence de la pièce / facture")
	createCmd.Flags().StringVar(&paidAt, "paid-at", "", "Date d'encaissement YYYY-MM-DD [défaut: aujourd'hui]")
	createCmd.Flags().Int64Var(&htCents, "ht-cents", 0, "Montant HT en centimes")
	createCmd.Flags().Int64Var(&ttcCents, "ttc-cents", 0, "Montant TTC en centimes")
	createCmd.Flags().Int64Var(&vatCents, "vat-cents", 0, "Montant TVA en centimes")
	createCmd.Flags().Float64Var(&htEuros, "ht-euros", 0, "Montant HT en euros")
	createCmd.Flags().Float64Var(&ttcEuros, "ttc-euros", 0, "Montant TTC en euros")
	createCmd.Flags().IntVar(&vatID, "vat-id", 5, "ID taux de TVA (1=20%, 2=10%, 3=5.5%, 4=2.1%, 5=0% HT, 6=8.5%)")
	createCmd.Flags().IntVar(&productType, "product-type", 3, "Nature de la recette (1=Marchandises, 2=Services BIC, 3=Services BNC, 4=Produits fabriqués, 5=Débours)")
	createCmd.Flags().IntVar(&paymentMethod, "payment-method", 1, "Mode de règlement (1=Virement, 2=CB, 3=Espèces, 4=Chèque, 5=Prélèvement)")
	createCmd.Flags().BoolVar(&isSap, "sap", false, "Service à la personne")
	createCmd.Flags().BoolVar(&isTaxIncluded, "tax-included", false, "TVA incluse")

	deleteCmd := &cobra.Command{
		Use:   "delete <incomeId>",
		Short: "Supprimer une entrée du Livre des recettes (DELETE /incomeBook/{id})",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			client, printer, err := getClientAndPrinter()
			if err != nil {
				return err
			}
			if err := client.DeleteIncomeBookItem(context.Background(), args[0]); err != nil {
				return err
			}
			if printer.Format == output.FormatJSON {
				return printer.PrintJSON(map[string]any{"deleted": true, "id": args[0]})
			}
			fmt.Printf("Recette %s supprimée.\n", args[0])
			return nil
		},
	}

	incomesCmd.AddCommand(createCmd, deleteCmd)
	return incomesCmd
}

func newPurchaseRegisterCmd() *cobra.Command {
	purchasesCmd := &cobra.Command{
		Use:     "purchases",
		Aliases: []string{"purchase", "achats", "achat", "purchase-register"},
		Short:   "Gérer le Livre des achats (/v3/purchaseRegister & /v2/purchaseRegister)",
	}

	// 1. list
	var page, limit int
	var search, providerID, fromDate, toDate string
	listCmd := &cobra.Command{
		Use:   "list",
		Short: "Lister les écritures du Livre des achats (GET /v3/purchaseRegister/list)",
		RunE: func(cmd *cobra.Command, args []string) error {
			client, printer, err := getClientAndPrinter()
			if err != nil {
				return err
			}
			res, err := client.ListPurchaseRegisterV3(context.Background(), api.ListPurchaseRegisterFilter{
				Page:       page,
				Limit:      limit,
				Search:     search,
				ProviderID: providerID,
				FromDate:   fromDate,
				ToDate:     toDate,
			})
			if err != nil {
				return err
			}
			return printer.PrintPurchaseRegisterList(res)
		},
	}
	listCmd.Flags().IntVar(&page, "page", 1, "Numéro de page")
	listCmd.Flags().IntVar(&limit, "limit", 20, "Nombre d'éléments par page")
	listCmd.Flags().StringVarP(&search, "search", "s", "", "Recherche par libellé ou référence")
	listCmd.Flags().StringVar(&providerID, "provider-id", "", "Filtrer par ID fournisseur")
	listCmd.Flags().StringVar(&fromDate, "from", "", "Date de début (YYYY-MM-DD)")
	listCmd.Flags().StringVar(&toDate, "to", "", "Date de fin (YYYY-MM-DD)")

	// 2. get
	getCmd := &cobra.Command{
		Use:   "get <purchaseId>",
		Short: "Récupérer une écriture d'achat par ID (GET /v3/purchaseRegister/{id})",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			client, printer, err := getClientAndPrinter()
			if err != nil {
				return err
			}
			item, err := client.GetPurchaseRegisterV3(context.Background(), args[0])
			if err != nil {
				return err
			}
			return printer.PrintJSON(item)
		},
	}

	// 3. create
	var label, reference, thirdPartyID, valueDate, paymentMethodStr, entriesJSON string
	var amountCents int64
	var amountEuros float64
	var vatID, accountNumber int
	var isPersonal, useV2 bool
	createCmd := &cobra.Command{
		Use:   "create",
		Short: "Créer une dépense dans le Livre des achats (POST /v3/purchaseRegister ou /v2/purchaseRegister)",
		RunE: func(cmd *cobra.Command, args []string) error {
			if label == "" || thirdPartyID == "" {
				return fmt.Errorf("--label et --provider-id (identifiant fournisseur issu de `abby accounting providers list`) sont obligatoires")
			}
			cents := amountCents
			if cents == 0 && amountEuros != 0 {
				cents = output.EurosToCents(amountEuros)
			}
			if valueDate == "" {
				valueDate = time.Now().Format("2006-01-02")
			}

			entries, totalEntriesCents, err := buildPurchaseEntries(entriesJSON, cents, isPersonal, vatID, accountNumber)
			if err != nil {
				return err
			}
			if cents == 0 {
				cents = totalEntriesCents
			}
			if cents == 0 {
				return fmt.Errorf("spécifiez le montant TTC via --amount-cents, --amount-euros ou --entries-json")
			}

			var pm any = paymentMethodStr
			if useV2 {
				pm = 1 // Legacy integer default for /v2/purchaseRegister
			}

			dto := api.CreatePurchaseRegisterDto{
				ValueDate:         valueDate,
				PaymentMethodUsed: pm,
				Amount:            cents,
				ThirdPartyID:      thirdPartyID,
				Label:             label,
				Reference:         reference,
				Entries:           entries,
			}

			client, printer, err := getClientAndPrinter()
			if err != nil {
				return err
			}
			var item *api.PurchaseRegisterItem
			if useV2 {
				item, err = client.CreatePurchaseRegisterV2(context.Background(), dto)
			} else {
				item, err = client.CreatePurchaseRegisterV3(context.Background(), dto)
			}
			if err != nil {
				return err
			}
			if printer.Format == output.FormatJSON {
				return printer.PrintJSON(item)
			}
			fmt.Printf("Achat enregistré avec succès ! ID: %s (%s — %s)\n",
				item.ID, item.Label, output.FormatCents(item.Amount, "EUR"))
			return nil
		},
	}
	createCmd.Flags().StringVar(&label, "label", "", "Libellé de l'achat (obligatoire)")
	createCmd.Flags().StringVar(&thirdPartyID, "provider-id", "", "ID du fournisseur (obligatoire, voir `abby accounting providers`)")
	createCmd.Flags().StringVar(&reference, "reference", "", "Référence de la facture fournisseur")
	createCmd.Flags().StringVar(&valueDate, "date", "", "Date de l'achat YYYY-MM-DD [défaut: aujourd'hui]")
	createCmd.Flags().Int64Var(&amountCents, "amount-cents", 0, "Montant TTC en centimes")
	createCmd.Flags().Float64Var(&amountEuros, "amount-euros", 0, "Montant TTC en euros")
	createCmd.Flags().IntVar(&vatID, "vat-id", 1, "ID TVA (1=20%, 2=10%, 3=5.5%, 4=2.1%, 5=0% HT, 6=8.5% DOM, 7=UE, 8=Hors UE)")
	createCmd.Flags().IntVar(&accountNumber, "account-number", 0, "Numéro de compte comptable optionnel (ex: 606300, 622600)")
	createCmd.Flags().BoolVar(&isPersonal, "personal", false, "Marquer comme dépense personnelle")
	createCmd.Flags().StringVar(&entriesJSON, "entries-json", "", "Tableau JSON de ventilation multi-lignes ([]PurchaseEntryLineDto)")
	createCmd.Flags().StringVar(&paymentMethodStr, "payment-method", "BANK_CARD", "Mode de paiement (BANK_CARD, BANK_TRANSFER, CASH, CHECK, OTHER)")
	createCmd.Flags().BoolVar(&useV2, "v2", false, "Utiliser l'endpoint historique /v2/purchaseRegister au lieu de /v3/purchaseRegister")

	// 4. update
	updateCmd := &cobra.Command{
		Use:   "update <purchaseId>",
		Short: "Mettre à jour une écriture du Livre des achats (PUT /v3/purchaseRegister/{id})",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if label == "" || thirdPartyID == "" {
				return fmt.Errorf("--label et --provider-id sont obligatoires")
			}
			cents := amountCents
			if cents == 0 && amountEuros != 0 {
				cents = output.EurosToCents(amountEuros)
			}
			if valueDate == "" {
				valueDate = time.Now().Format("2006-01-02")
			}
			entries, totalEntriesCents, err := buildPurchaseEntries(entriesJSON, cents, isPersonal, vatID, accountNumber)
			if err != nil {
				return err
			}
			if cents == 0 {
				cents = totalEntriesCents
			}
			dto := api.CreatePurchaseRegisterDto{
				ValueDate:         valueDate,
				PaymentMethodUsed: paymentMethodStr,
				Amount:            cents,
				ThirdPartyID:      thirdPartyID,
				Label:             label,
				Reference:         reference,
				Entries:           entries,
			}
			client, printer, err := getClientAndPrinter()
			if err != nil {
				return err
			}
			item, err := client.UpdatePurchaseRegisterV3(context.Background(), args[0], dto)
			if err != nil {
				return err
			}
			return printer.PrintJSON(item)
		},
	}
	updateCmd.Flags().StringVar(&label, "label", "", "Libellé de l'achat (obligatoire)")
	updateCmd.Flags().StringVar(&thirdPartyID, "provider-id", "", "ID du fournisseur (obligatoire)")
	updateCmd.Flags().StringVar(&reference, "reference", "", "Référence")
	updateCmd.Flags().StringVar(&valueDate, "date", "", "Date YYYY-MM-DD")
	updateCmd.Flags().Int64Var(&amountCents, "amount-cents", 0, "Montant TTC en centimes")
	updateCmd.Flags().Float64Var(&amountEuros, "amount-euros", 0, "Montant TTC en euros")
	updateCmd.Flags().IntVar(&vatID, "vat-id", 1, "ID TVA (1..8)")
	updateCmd.Flags().IntVar(&accountNumber, "account-number", 0, "Numéro de compte comptable")
	updateCmd.Flags().BoolVar(&isPersonal, "personal", false, "Dépense personnelle")
	updateCmd.Flags().StringVar(&entriesJSON, "entries-json", "", "Tableau JSON de ventilation multi-lignes")
	updateCmd.Flags().StringVar(&paymentMethodStr, "payment-method", "BANK_CARD", "Mode de paiement")

	// 5. delete
	var deleteV2 bool
	deleteCmd := &cobra.Command{
		Use:   "delete <purchaseId>",
		Short: "Supprimer une écriture du Livre des achats (DELETE /v3/purchaseRegister/{id})",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			client, printer, err := getClientAndPrinter()
			if err != nil {
				return err
			}
			if deleteV2 {
				err = client.DeletePurchaseRegisterV2(context.Background(), args[0])
			} else {
				err = client.DeletePurchaseRegisterV3(context.Background(), args[0])
			}
			if err != nil {
				return err
			}
			if printer.Format == output.FormatJSON {
				return printer.PrintJSON(map[string]any{"deleted": true, "id": args[0]})
			}
			fmt.Printf("Achat %s supprimé.\n", args[0])
			return nil
		},
	}
	deleteCmd.Flags().BoolVar(&deleteV2, "v2", false, "Utiliser DELETE /v2/purchaseRegister/{id}")

	purchasesCmd.AddCommand(listCmd, getCmd, createCmd, updateCmd, deleteCmd)
	return purchasesCmd
}

func newProvidersCmd() *cobra.Command {
	providersCmd := &cobra.Command{
		Use:     "providers",
		Aliases: []string{"provider", "fournisseurs", "fournisseur"},
		Short:   "Gérer les fournisseurs pour le Livre des achats (/providers, /provider)",
	}

	var page, limit int
	var search string
	listCmd := &cobra.Command{
		Use:   "list",
		Short: "Lister les fournisseurs (GET /providers)",
		RunE: func(cmd *cobra.Command, args []string) error {
			client, printer, err := getClientAndPrinter()
			if err != nil {
				return err
			}
			res, err := client.ListProviders(context.Background(), page, limit, search)
			if err != nil {
				return err
			}
			return printer.PrintProviders(res)
		},
	}
	listCmd.Flags().IntVar(&page, "page", 1, "Numéro de page")
	listCmd.Flags().IntVar(&limit, "limit", 20, "Nombre d'éléments par page")
	listCmd.Flags().StringVarP(&search, "search", "s", "", "Recherche par nom")

	var name, siret, vatNumber, address, zip, city, country, phone string
	createCmd := &cobra.Command{
		Use:   "create",
		Short: "Créer un fournisseur (POST /provider)",
		RunE: func(cmd *cobra.Command, args []string) error {
			if name == "" {
				return fmt.Errorf("--name est obligatoire")
			}
			client, printer, err := getClientAndPrinter()
			if err != nil {
				return err
			}
			dto := api.ProviderDto{
				Name:      name,
				Siret:     siret,
				VatNumber: vatNumber,
				Address:   address,
				ZipCode:   zip,
				City:      city,
				Country:   country,
				Phone:     phone,
			}
			pr, err := client.CreateProvider(context.Background(), dto)
			if err != nil {
				return err
			}
			if printer.Format == output.FormatJSON {
				return printer.PrintJSON(pr)
			}
			fmt.Printf("Fournisseur créé avec succès ! ID: %s (%s)\n", pr.ID, pr.Name)
			return nil
		},
	}
	createCmd.Flags().StringVar(&name, "name", "", "Nom du fournisseur (obligatoire)")
	createCmd.Flags().StringVar(&siret, "siret", "", "SIRET du fournisseur")
	createCmd.Flags().StringVar(&vatNumber, "vat-number", "", "Numéro de TVA")
	createCmd.Flags().StringVar(&address, "address", "", "Adresse")
	createCmd.Flags().StringVar(&zip, "zip", "", "Code postal")
	createCmd.Flags().StringVar(&city, "city", "", "Ville")
	createCmd.Flags().StringVar(&country, "country", "FR", "Code pays")
	createCmd.Flags().StringVar(&phone, "phone", "", "Téléphone")

	updateProviderCmd := &cobra.Command{
		Use:   "update <providerId>",
		Short: "Mettre à jour un fournisseur (PUT /provider/{id})",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if name == "" {
				return fmt.Errorf("--name est obligatoire")
			}
			client, printer, err := getClientAndPrinter()
			if err != nil {
				return err
			}
			pr, err := client.UpdateProvider(context.Background(), args[0], api.ProviderDto{
				Name:      name,
				Siret:     siret,
				VatNumber: vatNumber,
				Address:   address,
				ZipCode:   zip,
				City:      city,
				Country:   country,
				Phone:     phone,
			})
			if err != nil {
				return err
			}
			return printer.PrintJSON(pr)
		},
	}
	updateProviderCmd.Flags().StringVar(&name, "name", "", "Nom du fournisseur (obligatoire)")
	updateProviderCmd.Flags().StringVar(&siret, "siret", "", "SIRET du fournisseur")
	updateProviderCmd.Flags().StringVar(&vatNumber, "vat-number", "", "Numéro de TVA")
	updateProviderCmd.Flags().StringVar(&address, "address", "", "Adresse")
	updateProviderCmd.Flags().StringVar(&zip, "zip", "", "Code postal")
	updateProviderCmd.Flags().StringVar(&city, "city", "", "Ville")
	updateProviderCmd.Flags().StringVar(&country, "country", "FR", "Code pays")
	updateProviderCmd.Flags().StringVar(&phone, "phone", "", "Téléphone")

	deleteCmd := &cobra.Command{
		Use:   "delete <providerId>",
		Short: "Supprimer un fournisseur (DELETE /provider/{id})",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			client, printer, err := getClientAndPrinter()
			if err != nil {
				return err
			}
			if err := client.DeleteProvider(context.Background(), args[0]); err != nil {
				return err
			}
			if printer.Format == output.FormatJSON {
				return printer.PrintJSON(map[string]any{"deleted": true, "id": args[0]})
			}
			fmt.Printf("Fournisseur %s supprimé.\n", args[0])
			return nil
		},
	}

	providersCmd.AddCommand(listCmd, createCmd, updateProviderCmd, deleteCmd)
	return providersCmd
}

func newAccountingAccountsCmd() *cobra.Command {
	var includeHidden bool
	cmd := &cobra.Command{
		Use:     "accounts",
		Aliases: []string{"plan-comptable"},
		Short:   "Lister les comptes du plan comptable de l'entreprise (GET /v2/accounting-accounts)",
		RunE: func(cmd *cobra.Command, args []string) error {
			client, printer, err := getClientAndPrinter()
			if err != nil {
				return err
			}
			accounts, err := client.ListAccountingAccounts(context.Background(), includeHidden)
			if err != nil {
				return err
			}
			return printer.PrintAccountingAccounts(accounts)
		},
	}
	cmd.Flags().BoolVar(&includeHidden, "all", false, "Inclure les comptes comptables masqués")
	return cmd
}

func newTransactionCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "transaction <transactionId>",
		Short: "Récupérer une transaction bancaire par ID (GET /v2/transactions/{id})",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			client, printer, err := getClientAndPrinter()
			if err != nil {
				return err
			}
			tx, err := client.GetTransaction(context.Background(), args[0])
			if err != nil {
				return err
			}
			return printer.PrintJSON(tx)
		},
	}
}
