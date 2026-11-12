package cli

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/a2dev-fr/abby-cli/internal/config"
	"github.com/a2dev-fr/abby-cli/internal/output"
	"github.com/spf13/cobra"
)

func newAuthCmd() *cobra.Command {
	authCmd := &cobra.Command{
		Use:   "auth",
		Short: "Gérer l'authentification par clé API Abby (Bearer suk-*****)",
	}

	var baseURL string
	setKeyCmd := &cobra.Command{
		Use:   "set-key <api-key>",
		Short: "Enregistrer la clé API secrète Abby dans la configuration locale",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			key := strings.TrimSpace(args[0])
			cfg, err := config.Load()
			if err != nil {
				return err
			}
			cfg.APIKey = key
			if baseURL != "" {
				cfg.BaseURL = strings.TrimSpace(baseURL)
			}
			if err := cfg.Validate(); err != nil {
				return err
			}
			if err := cfg.Save(); err != nil {
				return err
			}
			path, _ := config.ConfigFilePath()
			fmt.Printf("Clé API enregistrée avec succès (%s) dans %s\n", cfg.MaskedKey(), path)
			return nil
		},
	}
	setKeyCmd.Flags().StringVar(&baseURL, "set-base-url", "", "Définir également l'URL de base (ex: https://api.app-abby.com ou https://api.abby.fr)")

	statusCmd := &cobra.Command{
		Use:   "status",
		Short: "Afficher l'état de la configuration locale et la source de la clé API",
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := resolveConfig()
			if err != nil {
				return err
			}
			path, _ := config.ConfigFilePath()
			source := "fichier de configuration"
			if flagAPIKey != "" {
				source = "flag --api-key"
			} else if os.Getenv(config.EnvAPIKey) != "" {
				source = "variable d'environnement " + config.EnvAPIKey
			}
			p := output.NewPrinter(flagOutput)
			if p.Format == output.FormatJSON {
				return p.PrintJSON(map[string]any{
					"configured":  cfg.APIKey != "",
					"masked_key":  cfg.MaskedKey(),
					"source":      source,
					"base_url":    cfg.BaseURL,
					"config_path": path,
				})
			}
			fmt.Printf("Fichier config : %s\n", path)
			fmt.Printf("Base URL       : %s\n", cfg.BaseURL)
			fmt.Printf("Clé API        : %s (source: %s)\n", cfg.MaskedKey(), source)
			return nil
		},
	}

	whoamiCmd := &cobra.Command{
		Use:   "whoami",
		Short: "Vérifier la clé API en récupérant les informations du compte Abby (GET /v2/company/me)",
		RunE: func(cmd *cobra.Command, args []string) error {
			client, printer, err := getClientAndPrinter()
			if err != nil {
				return err
			}
			me, raw, err := client.GetMe(context.Background())
			if err != nil {
				return err
			}
			if printer.Format == output.FormatJSON {
				return printer.PrintJSON(raw)
			}
			fmt.Printf("Connexion Abby réussie !\n")
			fmt.Printf("Utilisateur : %s %s (%s)\n", me.User.Firstname, me.User.Lastname, me.User.Email)
			fmt.Printf("Entreprise  : %s (ID: %s)\n", me.Company.Name, me.Company.ID)
			if me.Company.Siret != "" {
				fmt.Printf("SIRET       : %s\n", me.Company.Siret)
			}
			if me.Company.VatNumber != "" {
				fmt.Printf("TVA Intra.  : %s\n", me.Company.VatNumber)
			}
			return nil
		},
	}

	clearCmd := &cobra.Command{
		Use:   "clear",
		Short: "Supprimer la clé API enregistrée localement",
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := config.Load()
			if err != nil {
				return err
			}
			cfg.APIKey = ""
			if err := cfg.Save(); err != nil {
				return err
			}
			fmt.Println("Clé API supprimée de la configuration locale.")
			return nil
		},
	}

	authCmd.AddCommand(setKeyCmd, statusCmd, whoamiCmd, clearCmd)
	return authCmd
}
