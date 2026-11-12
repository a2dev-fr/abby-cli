package api

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
)

type recordedRequest struct {
	Method string
	Path   string
	Query  string
	Body   string
}

func setupMockServer(t *testing.T, handler func(w http.ResponseWriter, r *http.Request)) (*Client, *httptest.Server, *recordedRequest) {
	t.Helper()
	rec := &recordedRequest{}
	srv := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		rec.Method = r.Method
		rec.Path = r.URL.Path
		rec.Query = r.URL.RawQuery
		if r.Body != nil {
			b, _ := io.ReadAll(r.Body)
			rec.Body = string(b)
		}
		handler(w, r)
	}))
	t.Cleanup(srv.Close)

	client, err := NewClient(srv.URL, "suk-test-key", WithHTTPClient(srv.Client()))
	if err != nil {
		t.Fatalf("NewClient err: %v", err)
	}
	return client, srv, rec
}

func TestServices_CompanyAndClients(t *testing.T) {
	ctx := context.Background()
	client, _, rec := setupMockServer(t, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch r.URL.Path {
		case "/v2/company/me":
			_, _ = w.Write([]byte(`{"company":{"id":"comp_1","name":"Ma Société"},"user":{"id":"u_1","firstname":"Alex"}}`))
		case "/contacts":
			_, _ = w.Write([]byte(`{"docs":[{"id":"c_1","firstname":"Jean","lastname":"Martin"}],"totalDocs":1,"page":1,"totalPages":1}`))
		case "/contact", "/contact/c_1":
			_, _ = w.Write([]byte(`{"id":"c_1","firstname":"Jean","lastname":"Martin"}`))
		case "/organizations":
			_, _ = w.Write([]byte(`{"docs":[{"id":"org_1","name":"Dupont SARL"}],"totalDocs":1,"page":1,"totalPages":1}`))
		case "/organization", "/organization/org_1":
			_, _ = w.Write([]byte(`{"id":"org_1","name":"Dupont SARL"}`))
		case "/organization/org_1/contacts":
			_, _ = w.Write([]byte(`[{"id":"c_1","firstname":"Jean","lastname":"Martin"}]`))
		case "/organization/org_1/contact":
			_, _ = w.Write([]byte(`{"id":"c_2","firstname":"Marie","lastname":"Curie"}`))
		case "/organization/org_1/default-contact":
			w.WriteHeader(http.StatusOK)
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	})

	// 1. GetMe
	me, _, err := client.GetMe(ctx)
	if err != nil || me.Company.Name != "Ma Société" || rec.Method != http.MethodGet {
		t.Fatalf("GetMe échec: %v / %+v", err, me)
	}

	// 2. Contacts CRUD
	archived := false
	contacts, err := client.ListContacts(ctx, ListContactsFilter{Page: 1, Limit: 10, Search: "Jean", Archived: &archived})
	if err != nil || len(contacts.EffectiveItems()) != 1 {
		t.Fatalf("ListContacts échec: %v", err)
	}
	if _, err := client.GetContact(ctx, "c_1"); err != nil || rec.Path != "/contact/c_1" {
		t.Fatalf("GetContact échec: %v", err)
	}
	if _, err := client.CreateContact(ctx, CreateContactDto{Firstname: "Jean", Lastname: "Martin"}); err != nil || rec.Method != http.MethodPost {
		t.Fatalf("CreateContact échec: %v", err)
	}
	if _, err := client.UpdateContact(ctx, "c_1", CreateContactDto{Firstname: "Jean", Lastname: "Martin"}); err != nil || rec.Method != http.MethodPut {
		t.Fatalf("UpdateContact échec: %v", err)
	}
	if err := client.DeleteContact(ctx, "c_1"); err != nil || rec.Method != http.MethodDelete {
		t.Fatalf("DeleteContact échec: %v", err)
	}

	// 3. Organizations CRUD
	orgs, err := client.ListOrganizations(ctx, ListOrganizationsFilter{Page: 1, Limit: 10, Search: "Dupont", Archived: &archived})
	if err != nil || len(orgs.EffectiveItems()) != 1 {
		t.Fatalf("ListOrganizations échec: %v", err)
	}
	if _, err := client.GetOrganization(ctx, "org_1"); err != nil || rec.Path != "/organization/org_1" {
		t.Fatalf("GetOrganization échec: %v", err)
	}
	if _, err := client.CreateOrganization(ctx, CreateOrganizationDto{Name: "Dupont SARL"}); err != nil || rec.Method != http.MethodPost {
		t.Fatalf("CreateOrganization échec: %v", err)
	}
	if _, err := client.UpdateOrganization(ctx, "org_1", CreateOrganizationDto{Name: "Dupont SAS"}); err != nil || rec.Method != http.MethodPut {
		t.Fatalf("UpdateOrganization échec: %v", err)
	}
	if list, err := client.ListOrganizationContacts(ctx, "org_1"); err != nil || len(list) != 1 {
		t.Fatalf("ListOrganizationContacts échec: %v", err)
	}
	if _, err := client.CreateOrganizationContact(ctx, "org_1", CreateContactDto{Firstname: "Marie", Lastname: "Curie"}); err != nil || rec.Method != http.MethodPost {
		t.Fatalf("CreateOrganizationContact échec: %v", err)
	}
	if err := client.SetOrganizationDefaultContact(ctx, "org_1", "c_2"); err != nil || rec.Method != http.MethodPatch {
		t.Fatalf("SetOrganizationDefaultContact échec: %v", err)
	}
	if err := client.DeleteOrganization(ctx, "org_1"); err != nil || rec.Method != http.MethodDelete {
		t.Fatalf("DeleteOrganization échec: %v", err)
	}
}

func TestServices_InvoicingEstimatesAssetsAdvances(t *testing.T) {
	ctx := context.Background()
	client, _, rec := setupMockServer(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/v2/billing/doc_1/download" {
			w.Header().Set("Content-Type", "application/pdf")
			w.Header().Set("Content-Disposition", `attachment; filename="FA-2026-001.pdf"`)
			_, _ = w.Write([]byte("%PDF-1.4 mock"))
			return
		}
		w.Header().Set("Content-Type", "application/json")
		if r.URL.Path == "/v2/billings" {
			_, _ = w.Write([]byte(`{"docs":[{"id":"doc_1","type":"invoice","state":"draft"}],"totalDocs":1}`))
			return
		}
		_ = json.NewEncoder(w).Encode(BillingDocument{
			ID:    "doc_1",
			Type:  BillingTypeInvoice,
			State: BillingStateDraft,
		})
	})

	// 1. Billings common
	late := true
	archived := false
	list, err := client.ListBillings(ctx, ListBillingsFilter{
		Page:           1,
		Limit:          10,
		Search:         "FA",
		Types:          []BillingType{BillingTypeInvoice},
		States:         []BillingState{BillingStateDraft},
		CustomerID:     "c_1",
		OrganizationID: "org_1",
		Late:           &late,
		Archived:       &archived,
		FromDate:       "2026-01-01",
		ToDate:         "2026-12-31",
	})
	if err != nil || len(list.EffectiveItems()) != 1 {
		t.Fatalf("ListBillings échec: %v", err)
	}
	if _, err := client.GetBillingByID(ctx, "doc_1"); err != nil || rec.Path != "/v2/billing/doc_1" {
		t.Fatalf("GetBillingByID échec: %v", err)
	}
	if _, err := client.UpdateBillingLines(ctx, "doc_1", UpdateBillingLinesDto{
		Lines: []UpdateLineDto{{Designation: "Dev", UnitPrice: 35000, Quantity: 1, QuantityUnit: ProductUnitDay, Type: ProductTypeServiceDelivery, VatCode: VatCodeFR2000}},
	}); err != nil || rec.Method != http.MethodPatch || rec.Path != "/v2/billing/doc_1/lines" {
		t.Fatalf("UpdateBillingLines échec: %v", err)
	}
	if _, err := client.UpdateBillingTitle(ctx, "doc_1", "Titre"); err != nil || rec.Path != "/v2/billing/doc_1/title" {
		t.Fatalf("UpdateBillingTitle échec: %v", err)
	}
	if err := client.ArchiveBilling(ctx, "doc_1"); err != nil || rec.Path != "/v2/billing/doc_1/archive" {
		t.Fatalf("ArchiveBilling échec: %v", err)
	}
	if err := client.UnarchiveBilling(ctx, "doc_1"); err != nil || rec.Path != "/v2/billing/doc_1/unarchive" {
		t.Fatalf("UnarchiveBilling échec: %v", err)
	}
	if err := client.SendBillingByEmail(ctx, "doc_1", SendBillingByEmailDto{Emails: []string{"a@b.fr"}, Subject: "S", Content: "C"}); err != nil || rec.Path != "/v2/billing/doc_1/send-by-email" {
		t.Fatalf("SendBillingByEmail échec: %v", err)
	}
	pdfBytes, disp, err := client.DownloadBillingPDF(ctx, "doc_1")
	if err != nil || string(pdfBytes) != "%PDF-1.4 mock" || disp == "" {
		t.Fatalf("DownloadBillingPDF échec: %v / %s", err, string(pdfBytes))
	}
	if err := client.DeleteBillingDocument(ctx, "doc_1"); err != nil || rec.Method != http.MethodDelete {
		t.Fatalf("DeleteBillingDocument échec: %v", err)
	}

	// 2. Invoices
	if _, err := client.CreateInvoice(ctx, "c_1"); err != nil || rec.Path != "/v2/billing/invoice/c_1" {
		t.Fatalf("CreateInvoice échec: %v", err)
	}
	if _, err := client.UpdateInvoiceGeneralInfo(ctx, "inv_1", UpdateGeneralInformationsDto{HeaderNote: "Note"}); err != nil || rec.Path != "/v2/billing/invoice/inv_1/general-informations" {
		t.Fatalf("UpdateInvoiceGeneralInfo échec: %v", err)
	}
	if _, err := client.UpdateInvoiceTimeline(ctx, "inv_1", UpdateTimelineDto{EmittedAt: 1700000000, PaymentDelay: PaymentDelayThirtyDays}); err != nil || rec.Path != "/v2/billing/invoice/inv_1/timeline" {
		t.Fatalf("UpdateInvoiceTimeline échec: %v", err)
	}
	if _, err := client.UpdateInvoiceLocale(ctx, "inv_1", "fr"); err != nil || rec.Path != "/v2/billing/invoice/inv_1/locale" {
		t.Fatalf("UpdateInvoiceLocale échec: %v", err)
	}
	if _, err := client.UpdateInvoiceCurrency(ctx, "inv_1", "EUR"); err != nil || rec.Path != "/v2/billing/invoice/inv_1/currency" {
		t.Fatalf("UpdateInvoiceCurrency échec: %v", err)
	}
	if _, err := client.ActivateInvoiceOnlinePayment(ctx, "inv_1"); err != nil || rec.Path != "/v2/billing/invoice/inv_1/activate-online-payment" {
		t.Fatalf("ActivateInvoiceOnlinePayment échec: %v", err)
	}
	if _, err := client.CreateAssetFromInvoice(ctx, "inv_1"); err != nil || rec.Path != "/v2/billing/invoice/inv_1/asset" {
		t.Fatalf("CreateAssetFromInvoice échec: %v", err)
	}
	if _, err := client.CancelInvoice(ctx, "inv_1", true); err != nil || rec.Path != "/v2/billing/invoice/inv_1/cancel" {
		t.Fatalf("CancelInvoice échec: %v", err)
	}
	if err := client.MarkInvoiceAsPaid(ctx, "inv_1", ReconciliateInvoiceDto{
		Payments: []PaymentDto{{Amount: 120000, ReceivedAt: "2026-11-12", Method: PaymentMethodTransfer}},
	}); err != nil || rec.Path != "/v2/accounting-billing/invoice/inv_1/reconciliate" {
		t.Fatalf("MarkInvoiceAsPaid échec: %v", err)
	}
	if err := client.MarkInvoiceAsUnpaid(ctx, "inv_1"); err != nil || rec.Path != "/v2/accounting-billing/invoice/inv_1/mark-as-unpaid" {
		t.Fatalf("MarkInvoiceAsUnpaid échec: %v", err)
	}

	// 3. Estimates & Advances
	if _, err := client.CreateEstimate(ctx, "c_1", ""); err != nil || rec.Path != "/v2/billing/estimate/c_1" {
		t.Fatalf("CreateEstimate échec: %v", err)
	}
	if _, err := client.UpdateEstimateGeneralInfo(ctx, "est_1", UpdateGeneralInformationsDto{FooterNote: "Merci"}); err != nil {
		t.Fatalf("UpdateEstimateGeneralInfo échec: %v", err)
	}
	if _, err := client.UpdateEstimateTimeline(ctx, "est_1", UpdateTimelineDto{EmittedAt: 1700000000, PaymentDelay: PaymentDelayThirtyDays}); err != nil {
		t.Fatalf("UpdateEstimateTimeline échec: %v", err)
	}
	if _, err := client.SignEstimate(ctx, "est_1"); err != nil || rec.Path != "/v2/billing/estimate/est_1/sign" {
		t.Fatalf("SignEstimate échec: %v", err)
	}
	if _, err := client.UnsignEstimate(ctx, "est_1"); err != nil || rec.Path != "/v2/billing/estimate/est_1/unsign" {
		t.Fatalf("UnsignEstimate échec: %v", err)
	}
	if _, err := client.RefuseEstimate(ctx, "est_1"); err != nil || rec.Path != "/v2/billing/estimate/est_1/refuse" {
		t.Fatalf("RefuseEstimate échec: %v", err)
	}
	if _, err := client.UnrefuseEstimate(ctx, "est_1"); err != nil || rec.Path != "/v2/billing/estimate/est_1/unrefuse" {
		t.Fatalf("UnrefuseEstimate échec: %v", err)
	}
	if _, err := client.CreateAdvanceFromEstimate(ctx, "est_1", &CreateAdvanceDto{Mode: "PERCENTAGE", Percentage: 30}); err != nil || rec.Path != "/v2/billing/estimate/est_1/advance" {
		t.Fatalf("CreateAdvanceFromEstimate échec: %v", err)
	}
	if _, err := client.CreateFinalInvoiceFromEstimate(ctx, "est_1"); err != nil || rec.Path != "/v2/billing/estimate/est_1/final-invoice" {
		t.Fatalf("CreateFinalInvoiceFromEstimate échec: %v", err)
	}
	if _, err := client.GetAdvance(ctx, "adv_1"); err != nil || rec.Path != "/v2/billing/advance/adv_1" {
		t.Fatalf("GetAdvance échec: %v", err)
	}
	if _, err := client.UpdateAdvanceAmount(ctx, "adv_1", CreateAdvanceDto{Mode: "PERCENTAGE", Percentage: 50}); err != nil || rec.Path != "/v2/billing/advance/adv_1/amount" {
		t.Fatalf("UpdateAdvanceAmount échec: %v", err)
	}
	if _, err := client.UpdateAdvanceGeneralInfo(ctx, "adv_1", UpdateGeneralInformationsDto{HeaderNote: "Acompte"}); err != nil {
		t.Fatalf("UpdateAdvanceGeneralInfo échec: %v", err)
	}
	if _, err := client.ActivateAdvanceOnlinePayment(ctx, "adv_1"); err != nil {
		t.Fatalf("ActivateAdvanceOnlinePayment échec: %v", err)
	}
	if _, err := client.CancelAdvance(ctx, "adv_1", false); err != nil || rec.Path != "/v2/billing/advance/adv_1/cancel" {
		t.Fatalf("CancelAdvance échec: %v", err)
	}
	if err := client.MarkAdvanceAsPaid(ctx, "adv_1", ReconciliateInvoiceDto{}); err != nil || rec.Path != "/v2/accounting-billing/advance/adv_1/reconciliate" {
		t.Fatalf("MarkAdvanceAsPaid échec: %v", err)
	}

	// 4. Assets
	if _, err := client.GetAsset(ctx, "ass_1"); err != nil || rec.Path != "/v2/billing/asset/ass_1" {
		t.Fatalf("GetAsset échec: %v", err)
	}
	if _, err := client.CreateBlankAsset(ctx, "c_1"); err != nil || rec.Path != "/v2/billing/asset/c_1" {
		t.Fatalf("CreateBlankAsset échec: %v", err)
	}
	if _, err := client.CreateAssetFromAdvance(ctx, "adv_1"); err != nil || rec.Path != "/v2/billing/advance/adv_1/asset" {
		t.Fatalf("CreateAssetFromAdvance échec: %v", err)
	}
	if _, err := client.UpdateAssetGeneralInfo(ctx, "ass_1", UpdateGeneralInformationsDto{}); err != nil {
		t.Fatalf("UpdateAssetGeneralInfo échec: %v", err)
	}
	if _, err := client.UpdateAssetTimeline(ctx, "ass_1", UpdateTimelineDto{EmittedAt: 1700000000, PaymentDelay: PaymentDelayAtReception}); err != nil {
		t.Fatalf("UpdateAssetTimeline échec: %v", err)
	}
}

func TestServices_CatalogAccountingAndOpportunities(t *testing.T) {
	ctx := context.Background()
	client, _, rec := setupMockServer(t, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch r.URL.Path {
		case "/v2/catalog":
			_, _ = w.Write([]byte(`{"docs":[{"id":"prod_1","designation":"TJM Go","unitPrice":35000}],"totalDocs":1}`))
		case "/v2/catalog/product", "/v2/catalog/product/prod_1":
			_, _ = w.Write([]byte(`{"id":"prod_1","designation":"TJM Go","unitPrice":35000}`))
		case "/incomeBook":
			_, _ = w.Write([]byte(`{"_id":"inc_1","reference":"REC-1","priceTotalTax":120000}`))
		case "/v3/purchaseRegister/list":
			_, _ = w.Write([]byte(`{"docs":[{"id":"pur_1","label":"OVH","amount":5999}],"totalDocs":1}`))
		case "/v3/purchaseRegister", "/v3/purchaseRegister/pur_1", "/v2/purchaseRegister":
			_, _ = w.Write([]byte(`{"id":"pur_1","label":"OVH","amount":5999}`))
		case "/providers":
			_, _ = w.Write([]byte(`{"docs":[{"id":"prov_1","name":"OVH"}],"totalDocs":1}`))
		case "/provider", "/provider/prov_1":
			_, _ = w.Write([]byte(`{"id":"prov_1","name":"OVH"}`))
		case "/v2/accounting-accounts":
			_, _ = w.Write([]byte(`[{"id":"acc_1","number":606300,"name":"Fournitures"}]`))
		case "/v2/transactions/tx_1":
			_, _ = w.Write([]byte(`{"id":"tx_1","amount":5999}`))
		case "/v2/opportunity/categories":
			_, _ = w.Write([]byte(`[{"id":"cat_1","name":"En cours"}]`))
		case "/v2/opportunity", "/v2/opportunity/opp_1":
			_, _ = w.Write([]byte(`{"id":"opp_1","name":"Projet Web"}`))
		default:
			w.WriteHeader(http.StatusOK)
		}
	})

	// 1. Catalog
	sap := true
	prods, err := client.ListProducts(ctx, ListProductsFilter{Page: 1, Limit: 10, Search: "Go", SapCompatible: &sap})
	if err != nil || len(prods.EffectiveItems()) != 1 {
		t.Fatalf("ListProducts échec: %v", err)
	}
	if _, err := client.GetProduct(ctx, "prod_1"); err != nil || rec.Path != "/v2/catalog/product/prod_1" {
		t.Fatalf("GetProduct échec: %v", err)
	}
	if _, err := client.CreateProduct(ctx, ProductDto{Designation: "TJM Go"}); err != nil || rec.Method != http.MethodPost {
		t.Fatalf("CreateProduct échec: %v", err)
	}
	if _, err := client.UpdateProduct(ctx, "prod_1", ProductDto{Designation: "TJM Go"}); err != nil || rec.Method != http.MethodPut {
		t.Fatalf("UpdateProduct échec: %v", err)
	}

	// 2. Accounting (IncomeBook, PurchaseRegister v3 & v2, Providers, Accounts, Transaction)
	if inc, err := client.CreateIncomeBookItem(ctx, CreateIncomeBookItemDto{Client: "Dupont", PriceTotalTax: 120000}); err != nil || inc.ID != "inc_1" {
		t.Fatalf("CreateIncomeBookItem échec: %v", err)
	}
	if err := client.DeleteIncomeBookItem(ctx, "inc_1"); err != nil || rec.Path != "/incomeBook/inc_1" {
		t.Fatalf("DeleteIncomeBookItem échec: %v", err)
	}

	minC := int64(100)
	maxC := int64(10000)
	if purs, err := client.ListPurchaseRegisterV3(ctx, ListPurchaseRegisterFilter{
		Page: 1, Limit: 10, Search: "OVH", ProviderID: "prov_1", FromDate: "2026-01-01", ToDate: "2026-12-31", MinCents: &minC, MaxCents: &maxC,
	}); err != nil || len(purs.EffectiveItems()) != 1 {
		t.Fatalf("ListPurchaseRegisterV3 échec: %v", err)
	}
	if _, err := client.GetPurchaseRegisterV3(ctx, "pur_1"); err != nil {
		t.Fatalf("GetPurchaseRegisterV3 échec: %v", err)
	}
	if _, err := client.CreatePurchaseRegisterV3(ctx, CreatePurchaseRegisterDto{Label: "OVH", Amount: 5999}); err != nil {
		t.Fatalf("CreatePurchaseRegisterV3 échec: %v", err)
	}
	if _, err := client.CreatePurchaseRegisterV2(ctx, CreatePurchaseRegisterDto{Label: "OVH", Amount: 5999}); err != nil {
		t.Fatalf("CreatePurchaseRegisterV2 échec: %v", err)
	}
	if _, err := client.UpdatePurchaseRegisterV3(ctx, "pur_1", CreatePurchaseRegisterDto{Label: "OVH", Amount: 5999}); err != nil {
		t.Fatalf("UpdatePurchaseRegisterV3 échec: %v", err)
	}
	if err := client.DeletePurchaseRegisterV3(ctx, "pur_1"); err != nil {
		t.Fatalf("DeletePurchaseRegisterV3 échec: %v", err)
	}
	if err := client.DeletePurchaseRegisterV2(ctx, "pur_1"); err != nil {
		t.Fatalf("DeletePurchaseRegisterV2 échec: %v", err)
	}

	if provs, err := client.ListProviders(ctx, 1, 10, "OVH"); err != nil || len(provs.EffectiveItems()) != 1 {
		t.Fatalf("ListProviders échec: %v", err)
	}
	if _, err := client.CreateProvider(ctx, ProviderDto{Name: "OVH"}); err != nil {
		t.Fatalf("CreateProvider échec: %v", err)
	}
	if _, err := client.UpdateProvider(ctx, "prov_1", ProviderDto{Name: "OVH"}); err != nil {
		t.Fatalf("UpdateProvider échec: %v", err)
	}
	if err := client.DeleteProvider(ctx, "prov_1"); err != nil {
		t.Fatalf("DeleteProvider échec: %v", err)
	}

	if accs, err := client.ListAccountingAccounts(ctx, false); err != nil || len(accs) != 1 {
		t.Fatalf("ListAccountingAccounts échec: %v", err)
	}
	if _, err := client.GetTransaction(ctx, "tx_1"); err != nil {
		t.Fatalf("GetTransaction échec: %v", err)
	}

	// 3. Opportunities
	if _, err := client.ListOpportunityCategories(ctx); err != nil {
		t.Fatalf("ListOpportunityCategories échec: %v", err)
	}
	if _, err := client.GetOpportunity(ctx, "opp_1"); err != nil {
		t.Fatalf("GetOpportunity échec: %v", err)
	}
	if _, err := client.CreateOpportunity(ctx, OpportunityDto{Name: "Projet Web"}); err != nil {
		t.Fatalf("CreateOpportunity échec: %v", err)
	}
	if _, err := client.UpdateOpportunity(ctx, "opp_1", OpportunityDto{Name: "Projet Web"}); err != nil {
		t.Fatalf("UpdateOpportunity échec: %v", err)
	}
	if err := client.DeleteOpportunity(ctx, "opp_1"); err != nil {
		t.Fatalf("DeleteOpportunity échec: %v", err)
	}
}
