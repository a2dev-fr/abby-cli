package cli

import (
	"context"
	"fmt"

	"github.com/a2dev-fr/abby-cli/internal/api"
	"github.com/a2dev-fr/abby-cli/internal/output"
	"github.com/spf13/cobra"
)

func newOrganizationsCmd() *cobra.Command {
	orgsCmd := &cobra.Command{
		Use:     "orgs",
		Aliases: []string{"organizations", "organization", "org"},
		Short:   "Gérer les clients professionnels / entreprises (Organizations)",
	}

	// 1. list
	var page, limit int
	var search string
	var archived bool
	listCmd := &cobra.Command{
		Use:   "list",
		Short: "Lister et rechercher les entreprises clientes (GET /organizations)",
		RunE: func(cmd *cobra.Command, args []string) error {
			client, printer, err := getClientAndPrinter()
			if err != nil {
				return err
			}
			filter := api.ListOrganizationsFilter{
				Page:   page,
				Limit:  limit,
				Search: search,
			}
			if cmd.Flags().Changed("archived") {
				filter.Archived = &archived
			}
			res, err := client.ListOrganizations(context.Background(), filter)
			if err != nil {
				return err
			}
			return printer.PrintOrganizations(res)
		},
	}
	listCmd.Flags().IntVar(&page, "page", 1, "Numéro de page")
	listCmd.Flags().IntVar(&limit, "limit", 20, "Nombre d'éléments par page")
	listCmd.Flags().StringVarP(&search, "search", "s", "", "Recherche par raison sociale, SIRET ou email")
	listCmd.Flags().BoolVar(&archived, "archived", false, "Filtrer les organisations archivées")

	// 2. get
	getCmd := &cobra.Command{
		Use:   "get <organizationId>",
		Short: "Récupérer une organisation par ID (GET /organization/{id})",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			client, printer, err := getClientAndPrinter()
			if err != nil {
				return err
			}
			org, err := client.GetOrganization(context.Background(), args[0])
			if err != nil {
				return err
			}
			if printer.Format == output.FormatJSON {
				return printer.PrintJSON(org)
			}
			fmt.Printf("ID              : %s\n", org.ID)
			fmt.Printf("Raison sociale  : %s\n", org.Name)
			if org.CommercialName != "" {
				fmt.Printf("Nom commercial  : %s\n", org.CommercialName)
			}
			fmt.Printf("SIRET           : %s\n", org.Siret)
			fmt.Printf("TVA Intra.      : %s\n", org.VatNumber)
			fmt.Printf("Emails          : %v\n", org.Emails)
			if org.BillingAddress != nil {
				fmt.Printf("Adresse         : %s, %s %s (%s)\n",
					org.BillingAddress.Address, org.BillingAddress.ZipCode, org.BillingAddress.City, org.BillingAddress.Country)
			}
			return nil
		},
	}

	// 3. create
	var name, commercialName, email, siret, vatNumber, notes string
	var address, city, zipCode, country string
	createCmd := &cobra.Command{
		Use:   "create",
		Short: "Créer une nouvelle entreprise cliente (POST /organization)",
		RunE: func(cmd *cobra.Command, args []string) error {
			if name == "" {
				return fmt.Errorf("--name (raison sociale) est obligatoire")
			}
			client, printer, err := getClientAndPrinter()
			if err != nil {
				return err
			}
			dto := api.CreateOrganizationDto{
				Name:           name,
				CommercialName: commercialName,
				Siret:          siret,
				VatNumber:      vatNumber,
				Notes:          notes,
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
			org, err := client.CreateOrganization(context.Background(), dto)
			if err != nil {
				return err
			}
			if printer.Format == output.FormatJSON {
				return printer.PrintJSON(org)
			}
			fmt.Printf("Organisation créée avec succès ! ID: %s (%s)\n", org.ID, org.Name)
			return nil
		},
	}
	createCmd.Flags().StringVar(&name, "name", "", "Raison sociale de l'entreprise (obligatoire)")
	createCmd.Flags().StringVar(&commercialName, "commercial-name", "", "Nom commercial")
	createCmd.Flags().StringVar(&email, "email", "", "Email de facturation")
	createCmd.Flags().StringVar(&siret, "siret", "", "Numéro SIRET (14 chiffres)")
	createCmd.Flags().StringVar(&vatNumber, "vat-number", "", "Numéro de TVA intracommunautaire")
	createCmd.Flags().StringVar(&notes, "notes", "", "Notes internes")
	createCmd.Flags().StringVar(&address, "address", "", "Rue (adresse de facturation)")
	createCmd.Flags().StringVar(&city, "city", "", "Ville")
	createCmd.Flags().StringVar(&zipCode, "zip", "", "Code postal")
	createCmd.Flags().StringVar(&country, "country", "FR", "Code pays ISO-2")

	// 4. update
	updateCmd := &cobra.Command{
		Use:   "update <organizationId>",
		Short: "Mettre à jour une entreprise cliente (PUT /organization/{id})",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if name == "" {
				return fmt.Errorf("--name (raison sociale) est obligatoire")
			}
			client, printer, err := getClientAndPrinter()
			if err != nil {
				return err
			}
			dto := api.CreateOrganizationDto{
				Name:           name,
				CommercialName: commercialName,
				Siret:          siret,
				VatNumber:      vatNumber,
				Notes:          notes,
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
			org, err := client.UpdateOrganization(context.Background(), args[0], dto)
			if err != nil {
				return err
			}
			if printer.Format == output.FormatJSON {
				return printer.PrintJSON(org)
			}
			fmt.Printf("Organisation mise à jour avec succès ! ID: %s (%s)\n", org.ID, org.Name)
			return nil
		},
	}
	updateCmd.Flags().StringVar(&name, "name", "", "Raison sociale de l'entreprise (obligatoire)")
	updateCmd.Flags().StringVar(&commercialName, "commercial-name", "", "Nom commercial")
	updateCmd.Flags().StringVar(&email, "email", "", "Email de facturation")
	updateCmd.Flags().StringVar(&siret, "siret", "", "Numéro SIRET")
	updateCmd.Flags().StringVar(&vatNumber, "vat-number", "", "Numéro de TVA intracommunautaire")
	updateCmd.Flags().StringVar(&notes, "notes", "", "Notes internes")
	updateCmd.Flags().StringVar(&address, "address", "", "Rue")
	updateCmd.Flags().StringVar(&city, "city", "", "Ville")
	updateCmd.Flags().StringVar(&zipCode, "zip", "", "Code postal")
	updateCmd.Flags().StringVar(&country, "country", "FR", "Code pays ISO-2")

	// 5. delete
	deleteCmd := &cobra.Command{
		Use:   "delete <organizationId>",
		Short: "Supprimer une entreprise cliente (DELETE /organization/{organizationId})",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			client, printer, err := getClientAndPrinter()
			if err != nil {
				return err
			}
			if err := client.DeleteOrganization(context.Background(), args[0]); err != nil {
				return err
			}
			if printer.Format == output.FormatJSON {
				return printer.PrintJSON(map[string]any{"deleted": true, "id": args[0]})
			}
			fmt.Printf("Organisation %s supprimée.\n", args[0])
			return nil
		},
	}

	// 6. contacts (list contacts of an organization)
	orgContactsCmd := &cobra.Command{
		Use:   "contacts <organizationId>",
		Short: "Lister les contacts rattachés à une entreprise (GET /organization/{id}/contacts)",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			client, printer, err := getClientAndPrinter()
			if err != nil {
				return err
			}
			contacts, err := client.ListOrganizationContacts(context.Background(), args[0])
			if err != nil {
				return err
			}
			return printer.PrintContacts(&api.PaginatedResponse[api.ContactSummary]{
				Docs:       contacts,
				TotalDocs:  len(contacts),
				Page:       1,
				TotalPages: 1,
			})
		},
	}

	// 7. add-contact
	var cFirstname, cLastname, cEmail, cPhone, cJobTitle string
	addContactCmd := &cobra.Command{
		Use:   "add-contact <organizationId>",
		Short: "Ajouter un contact à une entreprise (POST /organization/{id}/contact)",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if cFirstname == "" || cLastname == "" {
				return fmt.Errorf("--firstname et --lastname sont obligatoires")
			}
			client, printer, err := getClientAndPrinter()
			if err != nil {
				return err
			}
			dto := api.CreateContactDto{
				Firstname: cFirstname,
				Lastname:  cLastname,
				Phone:     cPhone,
				JobTitle:  cJobTitle,
			}
			if cEmail != "" {
				dto.Emails = []string{cEmail}
			}
			c, err := client.CreateOrganizationContact(context.Background(), args[0], dto)
			if err != nil {
				return err
			}
			if printer.Format == output.FormatJSON {
				return printer.PrintJSON(c)
			}
			fmt.Printf("Contact d'entreprise créé ! ID: %s (%s %s)\n", c.ID, c.Firstname, c.Lastname)
			return nil
		},
	}
	addContactCmd.Flags().StringVar(&cFirstname, "firstname", "", "Prénom du contact (obligatoire)")
	addContactCmd.Flags().StringVar(&cLastname, "lastname", "", "Nom du contact (obligatoire)")
	addContactCmd.Flags().StringVar(&cEmail, "email", "", "Email du contact")
	addContactCmd.Flags().StringVar(&cPhone, "phone", "", "Téléphone du contact")
	addContactCmd.Flags().StringVar(&cJobTitle, "job-title", "", "Fonction / Poste")

	// 8. set-default-contact
	setDefaultContactCmd := &cobra.Command{
		Use:   "set-default-contact <organizationId> <contactId>",
		Short: "Définir le contact par défaut d'une entreprise (PATCH /organization/{organizationId}/default-contact)",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			client, printer, err := getClientAndPrinter()
			if err != nil {
				return err
			}
			if err := client.SetOrganizationDefaultContact(context.Background(), args[0], args[1]); err != nil {
				return err
			}
			if printer.Format == output.FormatJSON {
				return printer.PrintJSON(map[string]any{
					"organizationId":   args[0],
					"defaultContactId": args[1],
					"updated":          true,
				})
			}
			fmt.Printf("Contact par défaut %s défini pour l'organisation %s.\n", args[1], args[0])
			return nil
		},
	}

	orgsCmd.AddCommand(listCmd, getCmd, createCmd, updateCmd, deleteCmd, orgContactsCmd, addContactCmd, setDefaultContactCmd)
	return orgsCmd
}
