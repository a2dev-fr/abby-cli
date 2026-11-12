package cli

import (
	"fmt"
	"os"
	"time"

	"github.com/a2dev-fr/abby-cli/internal/api"
	"github.com/a2dev-fr/abby-cli/internal/config"
	"github.com/a2dev-fr/abby-cli/internal/output"
	"github.com/spf13/cobra"
)

var (
	flagAPIKey  string
	flagBaseURL string
	flagOutput  string
	flagTimeout int
)

// NewRootCmd constructs the root Cobra command and registers all subcommands.
func NewRootCmd() *cobra.Command {
	rootCmd := &cobra.Command{
		Use:   "abby",
		Short: "CLI officiel en Go pour piloter l'API REST Abby (Facturation & Clients)",
		Long: `abby-cli permet de piloter la plateforme Abby (https://docs.abby.fr) depuis le terminal ou via un agent IA.

Domaines couverts :
  - auth          : Configuration de la clé API (Bearer suk-***) et vérification du compte (/v2/company/me)
  - contacts      : Gestion des clients particuliers (/contacts, /contact/{id})
  - orgs          : Gestion des clients professionnels / entreprises (/organizations, /organization/{id})
  - invoices      : Cycle de vie complet des factures (/v2/billing/invoice/*, lignes, finalisation, paiement, avoir)
  - estimates     : Cycle de vie des devis et bons de commande (/v2/billing/estimate/*, signature, acomptes, facture de solde)
  - advances      : Gestion dédiée des factures d'acompte (/v2/billing/advance/*, montant, encaissement, annulation)
  - assets        : Gestion des avoirs (/v2/billing/asset/*)
  - billings      : Opérations transverses sur tous les documents de facturation (/v2/billings, PDF, email, archivage)
  - catalog       : Catalogue de produits et services (/v2/catalog, /v2/catalog/product/*)
  - accounting    : Comptabilité (Livre des recettes /incomeBook, Livre des achats /v3/purchaseRegister, Fournisseurs, Plan comptable)
  - opportunities : CRM et opportunités commerciales (/v2/opportunity/*)
  - raw           : Appel HTTP générique authentifié sur n'importe quel endpoint de l'API Abby`,
		SilenceUsage:  true,
		SilenceErrors: true,
	}

	rootCmd.PersistentFlags().StringVar(&flagAPIKey, "api-key", "", "Clé API secrète Abby (suk-***) [défaut: env ABBY_API_KEY ou fichier config]")
	rootCmd.PersistentFlags().StringVar(&flagBaseURL, "base-url", "", "URL de base de l'API Abby [défaut: https://api.app-abby.com]")
	rootCmd.PersistentFlags().StringVarP(&flagOutput, "output", "o", "table", "Format de sortie: 'table' (humain) ou 'json' (agent IA / scripts)")
	rootCmd.PersistentFlags().IntVar(&flagTimeout, "timeout", 0, "Timeout HTTP en secondes [défaut: 30]")

	rootCmd.AddCommand(
		newAuthCmd(),
		newContactsCmd(),
		newOrganizationsCmd(),
		newInvoicesCmd(),
		newEstimatesCmd(),
		newAdvancesCmd(),
		newAssetsCmd(),
		newBillingsCmd(),
		newCatalogCmd(),
		newAccountingCmd(),
		newOpportunitiesCmd(),
		newRawCmd(),
	)

	return rootCmd
}

// Execute runs the root command and exits with code 1 on error.
func Execute() {
	if err := NewRootCmd().Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Erreur: %v\n", err)
		os.Exit(1)
	}
}

func resolveConfig() (*config.Config, error) {
	cfg, err := config.Load()
	if err != nil {
		return nil, err
	}
	if flagAPIKey != "" {
		cfg.APIKey = flagAPIKey
	}
	if flagBaseURL != "" {
		cfg.BaseURL = flagBaseURL
	}
	if flagTimeout > 0 {
		cfg.Timeout = flagTimeout
	}
	return cfg, nil
}

func getClientAndPrinter() (*api.Client, *output.Printer, error) {
	cfg, err := resolveConfig()
	if err != nil {
		return nil, nil, err
	}
	if err := cfg.Validate(); err != nil {
		return nil, nil, err
	}
	client, err := api.NewClient(
		cfg.BaseURL,
		cfg.APIKey,
		api.WithTimeout(time.Duration(cfg.Timeout)*time.Second),
	)
	if err != nil {
		return nil, nil, err
	}
	printer := output.NewPrinter(flagOutput)
	return client, printer, nil
}
