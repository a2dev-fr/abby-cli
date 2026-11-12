package cli

import (
	"context"
	"fmt"

	"github.com/a2dev-fr/abby-cli/internal/api"
	"github.com/a2dev-fr/abby-cli/internal/output"
	"github.com/spf13/cobra"
)

func newCatalogCmd() *cobra.Command {
	catalogCmd := &cobra.Command{
		Use:     "catalog",
		Aliases: []string{"products", "product", "produits", "produit"},
		Short:   "Gérer le catalogue de produits et services (/v2/catalog)",
	}

	// 1. list
	var page, limit int
	var search string
	var sapCompatible bool
	listCmd := &cobra.Command{
		Use:   "list",
		Short: "Lister et rechercher des produits dans le catalogue (GET /v2/catalog)",
		RunE: func(cmd *cobra.Command, args []string) error {
			client, printer, err := getClientAndPrinter()
			if err != nil {
				return err
			}
			filter := api.ListProductsFilter{
				Page:   page,
				Limit:  limit,
				Search: search,
			}
			if cmd.Flags().Changed("sap") {
				filter.SapCompatible = &sapCompatible
			}
			res, err := client.ListProducts(context.Background(), filter)
			if err != nil {
				return err
			}
			return printer.PrintProducts(res)
		},
	}
	listCmd.Flags().IntVar(&page, "page", 1, "Numéro de page")
	listCmd.Flags().IntVar(&limit, "limit", 20, "Nombre d'éléments par page")
	listCmd.Flags().StringVarP(&search, "search", "s", "", "Recherche par désignation ou référence")
	listCmd.Flags().BoolVar(&sapCompatible, "sap", false, "Filtrer les produits compatibles Services à la Personne (SAP)")

	// 2. get
	getCmd := &cobra.Command{
		Use:   "get <productId>",
		Short: "Récupérer un produit par ID (GET /v2/catalog/product/{productId})",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			client, printer, err := getClientAndPrinter()
			if err != nil {
				return err
			}
			pr, err := client.GetProduct(context.Background(), args[0])
			if err != nil {
				return err
			}
			return printer.PrintProduct(pr)
		},
	}

	// 3. create
	var designation, description, reference, prodType, unit, vatCode string
	var amountCents int64
	var amountEuros float64
	var taxesIncluded, isDeliveryOfGood bool
	var stock int
	createCmd := &cobra.Command{
		Use:   "create",
		Short: "Créer un nouveau produit ou service dans le catalogue (POST /v2/catalog/product)",
		RunE: func(cmd *cobra.Command, args []string) error {
			if designation == "" {
				return fmt.Errorf("--designation est obligatoire")
			}
			client, printer, err := getClientAndPrinter()
			if err != nil {
				return err
			}
			cents := amountCents
			if cents == 0 && amountEuros != 0 {
				cents = output.EurosToCents(amountEuros)
			}
			dto := api.ProductDto{
				Designation:      designation,
				Description:      description,
				Reference:        reference,
				Type:             api.ProductType(prodType),
				Unit:             api.ProductUnit(unit),
				VatCode:          normalizeVatCode(vatCode),
				AmountInCents:    &cents,
				TaxesIncluded:    taxesIncluded,
				IsDeliveryOfGood: isDeliveryOfGood,
				Stock:            stock,
			}
			pr, err := client.CreateProduct(context.Background(), dto)
			if err != nil {
				return err
			}
			return printer.PrintProduct(pr)
		},
	}
	createCmd.Flags().StringVar(&designation, "designation", "", "Désignation / Nom du produit ou service (obligatoire)")
	createCmd.Flags().StringVar(&description, "description", "", "Description détaillée")
	createCmd.Flags().StringVar(&reference, "reference", "", "Référence produit / SKU")
	createCmd.Flags().StringVar(&prodType, "type", string(api.ProductTypeServiceDelivery), "Type (service_delivery, sale_of_goods, commercial_or_craft_services, sale_of_manufactured_goods, disbursement)")
	createCmd.Flags().StringVar(&unit, "unit", string(api.ProductUnitUnit), "Unité (unit, hour, day, fixed_rate, month, kilometer, etc.)")
	createCmd.Flags().StringVar(&vatCode, "vat", string(api.VatCodeFR2000), "Code TVA (FR_2000, FR_1000, FR_550, FR_210, FR_00HT)")
	createCmd.Flags().Int64Var(&amountCents, "price-cents", 0, "Prix unitaire en centimes (ex: 50000 = 500.00 €)")
	createCmd.Flags().Float64Var(&amountEuros, "price-euros", 0, "Prix unitaire en euros (converti automatiquement en centimes)")
	createCmd.Flags().BoolVar(&taxesIncluded, "taxes-included", false, "Indique si le prix renseigné est TTC (défaut: HT)")
	createCmd.Flags().BoolVar(&isDeliveryOfGood, "delivery-of-good", false, "Livraison de bien physique")
	createCmd.Flags().IntVar(&stock, "stock", 0, "Quantité en stock")

	// 4. update
	updateCmd := &cobra.Command{
		Use:   "update <productId>",
		Short: "Mettre à jour un produit du catalogue (PUT /v2/catalog/product/{productId})",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if designation == "" {
				return fmt.Errorf("--designation est obligatoire")
			}
			client, printer, err := getClientAndPrinter()
			if err != nil {
				return err
			}
			cents := amountCents
			if cents == 0 && amountEuros != 0 {
				cents = output.EurosToCents(amountEuros)
			}
			dto := api.ProductDto{
				Designation:      designation,
				Description:      description,
				Reference:        reference,
				Type:             api.ProductType(prodType),
				Unit:             api.ProductUnit(unit),
				VatCode:          normalizeVatCode(vatCode),
				AmountInCents:    &cents,
				TaxesIncluded:    taxesIncluded,
				IsDeliveryOfGood: isDeliveryOfGood,
				Stock:            stock,
			}
			pr, err := client.UpdateProduct(context.Background(), args[0], dto)
			if err != nil {
				return err
			}
			return printer.PrintProduct(pr)
		},
	}
	updateCmd.Flags().StringVar(&designation, "designation", "", "Désignation / Nom du produit (obligatoire)")
	updateCmd.Flags().StringVar(&description, "description", "", "Description détaillée")
	updateCmd.Flags().StringVar(&reference, "reference", "", "Référence produit / SKU")
	updateCmd.Flags().StringVar(&prodType, "type", string(api.ProductTypeServiceDelivery), "Type (service_delivery, sale_of_goods, commercial_or_craft_services, sale_of_manufactured_goods, disbursement)")
	updateCmd.Flags().StringVar(&unit, "unit", string(api.ProductUnitUnit), "Unité (unit, hour, day, fixed_rate, month, kilometer)")
	updateCmd.Flags().StringVar(&vatCode, "vat", string(api.VatCodeFR2000), "Code TVA (FR_2000, FR_1000, FR_550, FR_210, FR_00HT)")
	updateCmd.Flags().Int64Var(&amountCents, "price-cents", 0, "Prix unitaire en centimes")
	updateCmd.Flags().Float64Var(&amountEuros, "price-euros", 0, "Prix unitaire en euros")
	updateCmd.Flags().BoolVar(&taxesIncluded, "taxes-included", false, "Prix TTC")
	updateCmd.Flags().BoolVar(&isDeliveryOfGood, "delivery-of-good", false, "Livraison de bien physique")
	updateCmd.Flags().IntVar(&stock, "stock", 0, "Stock")

	catalogCmd.AddCommand(listCmd, getCmd, createCmd, updateCmd)
	return catalogCmd
}
