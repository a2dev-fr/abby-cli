package cli

import (
	"context"
	"fmt"

	"github.com/a2dev-fr/abby-cli/internal/api"
	"github.com/a2dev-fr/abby-cli/internal/output"
	"github.com/spf13/cobra"
)

func newContactsCmd() *cobra.Command {
	contactsCmd := &cobra.Command{
		Use:     "contacts",
		Aliases: []string{"contact"},
		Short:   "Gérer les clients particuliers (Contacts)",
	}

	// 1. list
	var page, limit int
	var search string
	var archived bool
	listCmd := &cobra.Command{
		Use:   "list",
		Short: "Lister et rechercher des contacts particuliers (GET /contacts)",
		RunE: func(cmd *cobra.Command, args []string) error {
			client, printer, err := getClientAndPrinter()
			if err != nil {
				return err
			}
			filter := api.ListContactsFilter{
				Page:   page,
				Limit:  limit,
				Search: search,
			}
			if cmd.Flags().Changed("archived") {
				filter.Archived = &archived
			}
			res, err := client.ListContacts(context.Background(), filter)
			if err != nil {
				return err
			}
			return printer.PrintContacts(res)
		},
	}
	listCmd.Flags().IntVar(&page, "page", 1, "Numéro de page")
	listCmd.Flags().IntVar(&limit, "limit", 20, "Nombre d'éléments par page")
	listCmd.Flags().StringVarP(&search, "search", "s", "", "Recherche par nom, prénom ou email")
	listCmd.Flags().BoolVar(&archived, "archived", false, "Filtrer les contacts archivés")

	// 2. get
	getCmd := &cobra.Command{
		Use:   "get <contactId>",
		Short: "Récupérer les détails d'un contact (GET /contact/{id})",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			client, printer, err := getClientAndPrinter()
			if err != nil {
				return err
			}
			c, err := client.GetContact(context.Background(), args[0])
			if err != nil {
				return err
			}
			if printer.Format == output.FormatJSON {
				return printer.PrintJSON(c)
			}
			fmt.Printf("ID        : %s\n", c.ID)
			fmt.Printf("Nom       : %s %s\n", c.Firstname, c.Lastname)
			fmt.Printf("Emails    : %v\n", c.Emails)
			fmt.Printf("Téléphone : %s\n", c.Phone)
			if c.BillingAddress != nil {
				fmt.Printf("Adresse   : %s, %s %s (%s)\n",
					c.BillingAddress.Address, c.BillingAddress.ZipCode, c.BillingAddress.City, c.BillingAddress.Country)
			}
			return nil
		},
	}

	// 3. create
	var firstname, lastname, email, phone, jobTitle, notes string
	var address, city, zipCode, country string
	createCmd := &cobra.Command{
		Use:   "create",
		Short: "Créer un nouveau contact particulier (POST /contact)",
		RunE: func(cmd *cobra.Command, args []string) error {
			if firstname == "" || lastname == "" {
				return fmt.Errorf("--firstname et --lastname sont obligatoires")
			}
			client, printer, err := getClientAndPrinter()
			if err != nil {
				return err
			}
			dto := api.CreateContactDto{
				Firstname: firstname,
				Lastname:  lastname,
				Phone:     phone,
				JobTitle:  jobTitle,
				Notes:     notes,
			}
			if email != "" {
				dto.Emails = []string{email}
			}
			if address != "" || city != "" || zipCode != "" {
				if country == "" {
					country = "FR"
				}
				dto.BillingAddress = &api.ClientAddressDto{
					Address: address,
					City:    city,
					ZipCode: zipCode,
					Country: country,
				}
			}
			c, err := client.CreateContact(context.Background(), dto)
			if err != nil {
				return err
			}
			if printer.Format == output.FormatJSON {
				return printer.PrintJSON(c)
			}
			fmt.Printf("Contact créé avec succès ! ID: %s (%s %s)\n", c.ID, c.Firstname, c.Lastname)
			return nil
		},
	}
	createCmd.Flags().StringVar(&firstname, "firstname", "", "Prénom du contact (obligatoire)")
	createCmd.Flags().StringVar(&lastname, "lastname", "", "Nom de famille du contact (obligatoire)")
	createCmd.Flags().StringVar(&email, "email", "", "Adresse email du contact")
	createCmd.Flags().StringVar(&phone, "phone", "", "Numéro de téléphone")
	createCmd.Flags().StringVar(&jobTitle, "job-title", "", "Fonction / Poste")
	createCmd.Flags().StringVar(&notes, "notes", "", "Notes internes")
	createCmd.Flags().StringVar(&address, "address", "", "Rue (adresse de facturation)")
	createCmd.Flags().StringVar(&city, "city", "", "Ville")
	createCmd.Flags().StringVar(&zipCode, "zip", "", "Code postal")
	createCmd.Flags().StringVar(&country, "country", "FR", "Code pays ISO-2 (ex: FR, BE, CH)")

	// 4. update
	updateCmd := &cobra.Command{
		Use:   "update <contactId>",
		Short: "Mettre à jour un contact particulier (PUT /contact/{id})",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if firstname == "" || lastname == "" {
				return fmt.Errorf("--firstname et --lastname sont obligatoires")
			}
			client, printer, err := getClientAndPrinter()
			if err != nil {
				return err
			}
			dto := api.CreateContactDto{
				Firstname: firstname,
				Lastname:  lastname,
				Phone:     phone,
				JobTitle:  jobTitle,
				Notes:     notes,
			}
			if email != "" {
				dto.Emails = []string{email}
			}
			if address != "" || city != "" || zipCode != "" {
				if country == "" {
					country = "FR"
				}
				dto.BillingAddress = &api.ClientAddressDto{
					Address: address,
					City:    city,
					ZipCode: zipCode,
					Country: country,
				}
			}
			c, err := client.UpdateContact(context.Background(), args[0], dto)
			if err != nil {
				return err
			}
			if printer.Format == output.FormatJSON {
				return printer.PrintJSON(c)
			}
			fmt.Printf("Contact mis à jour avec succès ! ID: %s (%s %s)\n", c.ID, c.Firstname, c.Lastname)
			return nil
		},
	}
	updateCmd.Flags().StringVar(&firstname, "firstname", "", "Prénom du contact (obligatoire)")
	updateCmd.Flags().StringVar(&lastname, "lastname", "", "Nom de famille du contact (obligatoire)")
	updateCmd.Flags().StringVar(&email, "email", "", "Adresse email du contact")
	updateCmd.Flags().StringVar(&phone, "phone", "", "Numéro de téléphone")
	updateCmd.Flags().StringVar(&jobTitle, "job-title", "", "Fonction / Poste")
	updateCmd.Flags().StringVar(&notes, "notes", "", "Notes internes")
	updateCmd.Flags().StringVar(&address, "address", "", "Rue (adresse de facturation)")
	updateCmd.Flags().StringVar(&city, "city", "", "Ville")
	updateCmd.Flags().StringVar(&zipCode, "zip", "", "Code postal")
	updateCmd.Flags().StringVar(&country, "country", "FR", "Code pays ISO-2")

	// 5. delete
	deleteCmd := &cobra.Command{
		Use:   "delete <contactId>",
		Short: "Supprimer un contact particulier (DELETE /contact/{id})",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			client, printer, err := getClientAndPrinter()
			if err != nil {
				return err
			}
			if err := client.DeleteContact(context.Background(), args[0]); err != nil {
				return err
			}
			if printer.Format == output.FormatJSON {
				return printer.PrintJSON(map[string]any{"deleted": true, "id": args[0]})
			}
			fmt.Printf("Contact %s supprimé.\n", args[0])
			return nil
		},
	}

	contactsCmd.AddCommand(listCmd, getCmd, createCmd, updateCmd, deleteCmd)
	return contactsCmd
}
