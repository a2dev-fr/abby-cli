package cli

import (
	"context"
	"fmt"

	"github.com/a2dev-fr/abby-cli/internal/api"
	"github.com/a2dev-fr/abby-cli/internal/output"
	"github.com/spf13/cobra"
)

func newOpportunitiesCmd() *cobra.Command {
	oppCmd := &cobra.Command{
		Use:     "opportunities",
		Aliases: []string{"opportunity", "crm"},
		Short:   "Gérer les opportunités commerciales / CRM (/v2/opportunity)",
	}

	categoriesCmd := &cobra.Command{
		Use:   "categories",
		Short: "Lister les catégories / colonnes d'opportunités (GET /v2/opportunity/categories)",
		RunE: func(cmd *cobra.Command, args []string) error {
			client, printer, err := getClientAndPrinter()
			if err != nil {
				return err
			}
			raw, err := client.ListOpportunityCategories(context.Background())
			if err != nil {
				return err
			}
			return printer.PrintJSON(raw)
		},
	}

	getCmd := &cobra.Command{
		Use:   "get <opportunityId>",
		Short: "Récupérer une opportunité par ID (GET /v2/opportunity/{id})",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			client, printer, err := getClientAndPrinter()
			if err != nil {
				return err
			}
			raw, err := client.GetOpportunity(context.Background(), args[0])
			if err != nil {
				return err
			}
			return printer.PrintJSON(raw)
		},
	}

	var name, description, dueDate, customerID, categoryID string
	var amountCents int64
	var amountEuros float64
	createCmd := &cobra.Command{
		Use:   "create",
		Short: "Créer une opportunité commerciale (POST /v2/opportunity)",
		RunE: func(cmd *cobra.Command, args []string) error {
			if name == "" || customerID == "" || categoryID == "" {
				return fmt.Errorf("--name, --customer-id et --category-id sont obligatoires")
			}
			cents := amountCents
			if cents == 0 && amountEuros != 0 {
				cents = output.EurosToCents(amountEuros)
			}
			client, printer, err := getClientAndPrinter()
			if err != nil {
				return err
			}
			raw, err := client.CreateOpportunity(context.Background(), api.OpportunityDto{
				Name:        name,
				Description: description,
				DueDate:     dueDate,
				Amount:      cents,
				CustomerID:  customerID,
				CategoryID:  categoryID,
			})
			if err != nil {
				return err
			}
			return printer.PrintJSON(raw)
		},
	}
	createCmd.Flags().StringVar(&name, "name", "", "Titre de l'opportunité (obligatoire)")
	createCmd.Flags().StringVar(&description, "description", "", "Description")
	createCmd.Flags().StringVar(&dueDate, "due-date", "", "Date d'échéance YYYY-MM-DD")
	createCmd.Flags().Int64Var(&amountCents, "amount-cents", 0, "Montant estimé en centimes")
	createCmd.Flags().Float64Var(&amountEuros, "amount-euros", 0, "Montant estimé en euros")
	createCmd.Flags().StringVar(&customerID, "customer-id", "", "ID client (obligatoire)")
	createCmd.Flags().StringVar(&categoryID, "category-id", "", "ID catégorie / colonne (obligatoire)")

	updateCmd := &cobra.Command{
		Use:   "update <opportunityId>",
		Short: "Mettre à jour une opportunité commerciale (PUT /v2/opportunity/{id})",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			cents := amountCents
			if cents == 0 && amountEuros != 0 {
				cents = output.EurosToCents(amountEuros)
			}
			client, printer, err := getClientAndPrinter()
			if err != nil {
				return err
			}
			raw, err := client.UpdateOpportunity(context.Background(), args[0], api.OpportunityDto{
				Name:        name,
				Description: description,
				DueDate:     dueDate,
				Amount:      cents,
				CustomerID:  customerID,
				CategoryID:  categoryID,
			})
			if err != nil {
				return err
			}
			return printer.PrintJSON(raw)
		},
	}
	updateCmd.Flags().StringVar(&name, "name", "", "Titre de l'opportunité")
	updateCmd.Flags().StringVar(&description, "description", "", "Description")
	updateCmd.Flags().StringVar(&dueDate, "due-date", "", "Date d'échéance YYYY-MM-DD")
	updateCmd.Flags().Int64Var(&amountCents, "amount-cents", 0, "Montant estimé en centimes")
	updateCmd.Flags().Float64Var(&amountEuros, "amount-euros", 0, "Montant estimé en euros")
	updateCmd.Flags().StringVar(&customerID, "customer-id", "", "ID client")
	updateCmd.Flags().StringVar(&categoryID, "category-id", "", "ID catégorie")

	deleteCmd := &cobra.Command{
		Use:   "delete <opportunityId>",
		Short: "Supprimer une opportunité commerciale (DELETE /v2/opportunity/{id})",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			client, printer, err := getClientAndPrinter()
			if err != nil {
				return err
			}
			if err := client.DeleteOpportunity(context.Background(), args[0]); err != nil {
				return err
			}
			return printer.PrintJSON(map[string]any{"deleted": true, "id": args[0]})
		},
	}

	oppCmd.AddCommand(categoriesCmd, getCmd, createCmd, updateCmd, deleteCmd)
	return oppCmd
}
