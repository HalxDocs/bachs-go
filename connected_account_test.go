package bachs

import (
	"context"
	"io"
	"net/http"
	"strings"
	"testing"
)

// connectedAccountExample is the OrganizationResponse example from the
// connected-account create doc page.
const connectedAccountExample = `{
	"id": "org_4d81fa9c2b6e0357",
	"name": "Ada Stores",
	"owner_user_id": "usr_5e0b74c8a213",
	"parent_organization_id": "org_9f2c4a1b7e3d5086",
	"country": "NG",
	"fee_handling": "org_pays_fee",
	"enabled_payment_methods": null,
	"adaptive_pricing": true,
	"balance_currencies": ["NGN"],
	"website": null,
	"phone_number": null,
	"company_name": null,
	"enabled_capabilities": [],
	"capabilities": {
		"payouts": {"status": "pending", "requested": true, "status_details": null},
		"transfers": {"status": "pending", "requested": true, "status_details": null}
	},
	"requirements": {
		"setup_status": "incomplete",
		"currently_due": ["persons", "company.registered_name", "payout_destination"],
		"eventually_due": [],
		"past_due": [],
		"pending_verification": [],
		"errors": []
	},
	"fields_needing_resubmission": null,
	"sandbox_org_id": null,
	"live_org_id": null,
	"is_active": true,
	"created_at": "2026-08-07T11:04:22.518Z",
	"updated_at": "2026-08-07T11:04:22.518Z",
	"controller": {
		"fees": {"payer": "account"}
	}
}`

func accountServer(t *testing.T, method, path string, body string) *Client {
	t.Helper()
	return newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != method {
			t.Errorf("method = %s, want %s", r.Method, method)
		}
		if got := r.URL.RequestURI(); got != path {
			t.Errorf("path = %q, want %q", got, path)
		}
		if body != "" {
			io.WriteString(w, body)
		} else {
			w.WriteHeader(http.StatusNoContent)
		}
	})
}

// TestConnectedAccountCreate uses the exact request example from
// https://docs.bachs.io/api-reference/connected-accounts/create-a-connected-account
func TestConnectedAccountCreate(t *testing.T) {
	c := accountServer(t, http.MethodPost, "/v1/organizations/connected-accounts", connectedAccountExample)

	acct, _, err := c.ConnectedAccounts.Create(context.Background(), CreateConnectedAccountRequest{
		ContactEmail: "ada@adastores.example",
		DisplayName:  stringPtr("Ada Stores"),
		FirstName:    stringPtr("Ada"),
		LastName:     stringPtr("Okafor"),
		Country:      stringPtr("NG"),
		EntityType:   stringPtr("company"),
		Capabilities: map[string]CapabilityRequest{
			"payouts":   {Requested: true},
			"transfers": {Requested: true},
		},
		Controller: &ControllerRequest{Fees: ControllerFeesRequest{Payer: "account"}},
	})
	if err != nil {
		t.Fatalf("Create returned error: %v", err)
	}

	if acct.ID != "org_4d81fa9c2b6e0357" {
		t.Errorf("ID = %q", acct.ID)
	}
	if acct.Country == nil || *acct.Country != "NG" {
		t.Errorf("Country = %v", acct.Country)
	}
	if acct.Requirements == nil {
		t.Fatal("Requirements is nil")
	}
	if len(acct.Requirements.CurrentlyDue) != 3 {
		t.Errorf("Requirements.CurrentlyDue = %v", acct.Requirements.CurrentlyDue)
	}
	status, ok := acct.Capabilities["payouts"]
	if !ok || status.Status != "pending" || !status.Requested {
		t.Errorf("Capabilities[payouts] = %+v", status)
	}
	if acct.Controller == nil || acct.Controller.Fees.Payer != "account" {
		t.Errorf("Controller = %+v", acct.Controller)
	}
}

func TestConnectedAccountGet(t *testing.T) {
	c := accountServer(t, http.MethodGet, "/v1/connected-accounts/org_4d81fa9c2b6e0357", connectedAccountExample)

	acct, _, err := c.ConnectedAccounts.Get(context.Background(), "org_4d81fa9c2b6e0357")
	if err != nil {
		t.Fatalf("Get returned error: %v", err)
	}
	if acct.ID != "org_4d81fa9c2b6e0357" || !acct.AdaptivePricing {
		t.Errorf("account = %+v", acct)
	}
}

func TestConnectedAccountList(t *testing.T) {
	c := accountServer(t, http.MethodGet, "/v1/organizations/connected-accounts", `{
		"items": [`+connectedAccountExample+`],
		"total": 1,
		"limit": 20,
		"offset": 0
	}`)

	page, _, err := c.ConnectedAccounts.List(context.Background(), ListParams{})
	if err != nil {
		t.Fatalf("List returned error: %v", err)
	}
	if len(page.Items) != 1 {
		t.Fatalf("len(Items) = %d, want 1", len(page.Items))
	}
	if page.Items[0].Name == nil || *page.Items[0].Name != "Ada Stores" {
		t.Errorf("Items[0].Name = %v", page.Items[0].Name)
	}
	if page.Pagination.Total != 1 {
		t.Errorf("Pagination.Total = %d, want 1", page.Pagination.Total)
	}
}

func TestConnectedAccountRequestCapabilities(t *testing.T) {
	c := accountServer(t, http.MethodPatch, "/v1/connected-accounts/org_1", connectedAccountExample)

	acct, _, err := c.ConnectedAccounts.RequestCapabilities(context.Background(), "org_1", UpdateConnectedAccountRequest{
		Capabilities: map[string]bool{"conversions": true},
	})
	if err != nil {
		t.Fatalf("RequestCapabilities returned error: %v", err)
	}
	if acct.ID != "org_4d81fa9c2b6e0357" {
		t.Errorf("ID = %q", acct.ID)
	}
}

// TestConnectedAccountCreateAccountLink uses the exact example from
// https://docs.bachs.io/api-reference/connected-accounts/create-an-account-link
func TestConnectedAccountCreateAccountLink(t *testing.T) {
	c := accountServer(t, http.MethodPost, "/v1/connected-accounts/org_4d81fa9c2b6e0357/account-links", `{
		"id": "alnk_3b7e12c9d4a05f68b1c2",
		"object": "connected_account_link",
		"account": "org_4d81fa9c2b6e0357",
		"type": "onboarding",
		"created": "2026-08-07T11:04:22.518Z",
		"expires_at": "2026-09-06T11:04:22.518Z",
		"url": "https://connect.bachs.io/onboard/alnk_3b7e12c9d4a05f68b1c2",
		"previous_link_superseded": false
	}`)

	link, _, err := c.ConnectedAccounts.CreateAccountLink(context.Background(), "org_4d81fa9c2b6e0357", CreateAccountLinkRequest{
		Type:       "onboarding",
		RefreshURL: "https://adastores.example/connect/refresh",
		ReturnURL:  "https://adastores.example/connect/return",
	})
	if err != nil {
		t.Fatalf("CreateAccountLink returned error: %v", err)
	}
	if link.ID != "alnk_3b7e12c9d4a05f68b1c2" || link.Type != "onboarding" {
		t.Errorf("link = %+v", link)
	}
	if !strings.HasPrefix(link.URL, "https://connect.bachs.io/onboard/") {
		t.Errorf("URL = %q", link.URL)
	}
}

// TestConnectedAccountListCapabilities uses the exact example from
// https://docs.bachs.io/api-reference/connected-accounts/list-capabilities
func TestConnectedAccountListCapabilities(t *testing.T) {
	c := accountServer(t, http.MethodGet, "/v1/connected-accounts/org_4d81fa9c2b6e0357/capabilities", `{
		"items": [
			{"name": "payouts", "status": "active", "requested": true, "status_details": null},
			{
				"name": "transfers",
				"status": "restricted",
				"requested": true,
				"status_details": [
					{"code": "platform_disabled", "resolution": "Contact support to re-enable this capability.", "message": "This capability was disabled by the platform."}
				]
			},
			{"name": "conversions", "status": "unrequested", "requested": false, "status_details": null},
			{"name": "connect", "status": "unrequested", "requested": false, "status_details": null}
		]
	}`)

	caps, _, err := c.ConnectedAccounts.ListCapabilities(context.Background(), "org_4d81fa9c2b6e0357")
	if err != nil {
		t.Fatalf("ListCapabilities returned error: %v", err)
	}
	if len(caps.Items) != 4 {
		t.Fatalf("len(Items) = %d, want 4", len(caps.Items))
	}
	if caps.Items[0].Name != "payouts" || caps.Items[0].Status != "active" {
		t.Errorf("Items[0] = %+v", caps.Items[0])
	}
	if len(caps.Items[1].StatusDetails) != 1 || caps.Items[1].StatusDetails[0].Code != "platform_disabled" {
		t.Errorf("Items[1].StatusDetails = %+v", caps.Items[1].StatusDetails)
	}
	if caps.Items[2].Status != "unrequested" {
		t.Errorf("Items[2].Status = %q, want unrequested", caps.Items[2].Status)
	}
}

// TestConnectedAccountGetTaskChecklist uses the exact example from
// https://docs.bachs.io/api-reference/connected-accounts/get-the-task-checklist
func TestConnectedAccountGetTaskChecklist(t *testing.T) {
	c := accountServer(t, http.MethodGet, "/v1/connected-accounts/org_4d81fa9c2b6e0357/requirements/checklist", `{
		"organization_id": "org_4d81fa9c2b6e0357",
		"entity_type": "company",
		"country": "NG",
		"currently_due": 2,
		"pending_review": 0,
		"in_verification": 1,
		"needs_attention": 1,
		"setup_status": "incomplete",
		"checklist": [
			{
				"field_key": "company.registration_number",
				"label": "Registration number",
				"group": "company",
				"state": "currently_due",
				"provided": false,
				"error_reason": null,
				"reference": {"type": "account", "resource": null, "label": null}
			}
		],
		"capabilities": []
	}`)

	cl, _, err := c.ConnectedAccounts.GetTaskChecklist(context.Background(), "org_4d81fa9c2b6e0357")
	if err != nil {
		t.Fatalf("GetTaskChecklist returned error: %v", err)
	}
	if cl.SetupStatus != "incomplete" || cl.CurrentlyDue != 2 || cl.NeedsAttention != 1 {
		t.Errorf("checklist = %+v", cl)
	}
	if len(cl.Checklist) != 1 {
		t.Fatalf("len(Checklist) = %d, want 1", len(cl.Checklist))
	}
	if cl.Checklist[0].FieldKey != "company.registration_number" || cl.Checklist[0].State != "currently_due" {
		t.Errorf("Checklist[0] = %+v", cl.Checklist[0])
	}
	if cl.Checklist[0].Reference == nil || cl.Checklist[0].Reference.Type != "account" {
		t.Errorf("Checklist[0].Reference = %+v", cl.Checklist[0].Reference)
	}
}

func TestConnectedAccountListTasks(t *testing.T) {
	c := accountServer(t, http.MethodGet, "/v1/connected-accounts/org_1/requirements/tasks?status=open", `{
		"total": 1,
		"items": [
			{
				"id": "tsk_1",
				"title": "Provide your registration number",
				"description": null,
				"type": "form_field",
				"status": "open",
				"field_ref": "company.registration_number",
				"document_type": null,
				"requirements": {},
				"response_contract": {},
				"due_date": null,
				"impacts_capability": "payouts",
				"section_key": null,
				"past_due": false,
				"rejection_reason": null,
				"created_at": "2026-08-07T11:04:22.518Z",
				"updated_at": "2026-08-07T11:04:22.518Z"
			}
		]
	}`)

	page, _, err := c.ConnectedAccounts.ListTasks(context.Background(), "org_1", ListParams{Status: "open"})
	if err != nil {
		t.Fatalf("ListTasks returned error: %v", err)
	}
	if len(page.Items) != 1 {
		t.Fatalf("len(Items) = %d, want 1", len(page.Items))
	}
	if page.Items[0].Type != "form_field" || page.Items[0].FieldRef == nil || *page.Items[0].FieldRef != "company.registration_number" {
		t.Errorf("Items[0] = %+v", page.Items[0])
	}
	if page.Pagination.Total != 1 {
		t.Errorf("Pagination.Total = %d, want 1", page.Pagination.Total)
	}
}

func TestConnectedAccountGetTaskValues(t *testing.T) {
	c := accountServer(t, http.MethodGet, "/v1/connected-accounts/org_1/requirements/values", `{
		"organization_id": "org_1",
		"entity_type": "company",
		"values": [
			{
				"field": "company.registered_name",
				"label": "Registered name",
				"group": "business_profile",
				"provided": true,
				"sensitive": false,
				"value": "Ada Stores Ltd",
				"display": "Ada Stores Ltd",
				"reference_data": null
			}
		],
		"persons": []
	}`)

	vals, _, err := c.ConnectedAccounts.GetTaskValues(context.Background(), "org_1")
	if err != nil {
		t.Fatalf("GetTaskValues returned error: %v", err)
	}
	if len(vals.Values) != 1 {
		t.Fatalf("len(Values) = %d, want 1", len(vals.Values))
	}
	if vals.Values[0].Field != "company.registered_name" || vals.Values[0].Value != "Ada Stores Ltd" {
		t.Errorf("Values[0] = %+v", vals.Values[0])
	}
	if vals.Values[0].Sensitive {
		t.Error("Values[0].Sensitive = true, want false")
	}
}

func TestConnectedAccountSubmitTaskValues(t *testing.T) {
	c := accountServer(t, http.MethodPost, "/v1/connected-accounts/org_1/requirements/submit", `{
		"organization_id": "org_1",
		"entity_type": "company",
		"country": "NG",
		"currently_due": 0,
		"pending_review": 0,
		"in_verification": 0,
		"needs_attention": 0,
		"setup_status": "awaiting_review",
		"checklist": [],
		"capabilities": []
	}`)

	cl, _, err := c.ConnectedAccounts.SubmitTaskValues(context.Background(), "org_1", SubmitTasksRequest{
		Fields: map[string]any{"company.registered_name": "Ada Stores Ltd"},
	})
	if err != nil {
		t.Fatalf("SubmitTaskValues returned error: %v", err)
	}
	if cl.SetupStatus != "awaiting_review" {
		t.Errorf("SetupStatus = %q, want awaiting_review", cl.SetupStatus)
	}
}

func TestConnectedAccountReusableIdentity(t *testing.T) {
	getServer := accountServer(t, http.MethodGet, "/v1/connected-accounts/org_1/requirements/reusable-identity", `{
		"available": true,
		"person_public_id": "per_8f3a1c9b4e72",
		"first_name": "Ada",
		"last_name": "Okafor",
		"country": "NG",
		"verification_status": "verified",
		"used_by": ["Ada Stores", "Ada Tech"]
	}`)

	id, _, err := getServer.ConnectedAccounts.GetReusableIdentity(context.Background(), "org_1")
	if err != nil {
		t.Fatalf("GetReusableIdentity returned error: %v", err)
	}
	if !id.Available || id.PersonPublicID == nil || *id.PersonPublicID != "per_8f3a1c9b4e72" {
		t.Errorf("identity = %+v", id)
	}
	if len(id.UsedBy) != 2 {
		t.Errorf("UsedBy = %v", id.UsedBy)
	}

	applyServer := accountServer(t, http.MethodPost, "/v1/connected-accounts/org_1/requirements/reusable-identity/apply", `{
		"applied": true,
		"verification_status": "verified"
	}`)

	res, _, err := applyServer.ConnectedAccounts.ApplyReusableIdentity(context.Background(), "org_1", ApplyReusableIdentityRequest{
		PersonPublicID: "per_8f3a1c9b4e72",
	})
	if err != nil {
		t.Fatalf("ApplyReusableIdentity returned error: %v", err)
	}
	if !res.Applied || res.VerificationStatus != "verified" {
		t.Errorf("apply result = %+v", res)
	}
}

func TestConnectedAccountBanks(t *testing.T) {
	c := accountServer(t, http.MethodGet, "/v1/connected-accounts/org_1/requirements/banks?country=NG", `{
		"country": "NG",
		"banks": [
			{"name": "Providus Bank", "code": "PROVIDUS"},
			{"name": "GTBank", "code": "GTB"}
		]
	}`)

	banks, _, err := c.ConnectedAccounts.ListBanks(context.Background(), "org_1", "NG")
	if err != nil {
		t.Fatalf("ListBanks returned error: %v", err)
	}
	if banks.Country != "NG" || len(banks.Banks) != 2 {
		t.Errorf("banks = %+v", banks)
	}
	if banks.Banks[0].Code != "PROVIDUS" {
		t.Errorf("Banks[0].Code = %q", banks.Banks[0].Code)
	}
}

func TestConnectedAccountMobileMoneyProviders(t *testing.T) {
	c := accountServer(t, http.MethodGet, "/v1/connected-accounts/org_1/requirements/momo?country=KE", `{
		"country": "KE",
		"providers": ["M-PESA", "Airtel Money"]
	}`)

	list, _, err := c.ConnectedAccounts.ListMobileMoneyProviders(context.Background(), "org_1", "KE")
	if err != nil {
		t.Fatalf("ListMobileMoneyProviders returned error: %v", err)
	}
	if len(list.Providers) != 2 || list.Providers[0] != "M-PESA" {
		t.Errorf("providers = %v", list.Providers)
	}
}

func TestConnectedAccountResolveBankAccount(t *testing.T) {
	c := accountServer(t, http.MethodPost, "/v1/connected-accounts/org_1/requirements/accounts/resolve", `{
		"resolved": true,
		"account_name": "ADA OKAFOR",
		"account_number": "0123456789",
		"message": null
	}`)

	res, _, err := c.ConnectedAccounts.ResolveBankAccount(context.Background(), "org_1", ResolveTaskBankAccountRequest{
		AccountNumber: "0123456789",
		BankCode:      "GTB",
		Country:       stringPtr("NG"),
	})
	if err != nil {
		t.Fatalf("ResolveBankAccount returned error: %v", err)
	}
	if !res.Resolved || res.AccountName == nil || *res.AccountName != "ADA OKAFOR" {
		t.Errorf("resolve = %+v", res)
	}
	if res.Message != nil {
		t.Errorf("Message = %v, want nil on a successful match", res.Message)
	}
}

func TestConnectedAccountDocuments(t *testing.T) {
	uploadServer := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("method = %s, want POST", r.Method)
		}
		if r.URL.Path != "/v1/connected-accounts/org_1/uploads" {
			t.Errorf("path = %q", r.URL.Path)
		}
		if ct := r.Header.Get(headerContentType); !strings.HasPrefix(ct, "multipart/form-data") {
			t.Errorf("Content-Type = %q, want multipart/form-data", ct)
		}
		io.WriteString(w, uploadExample)
	})

	up, _, err := uploadServer.ConnectedAccounts.UploadDocument(context.Background(), "org_1", "id-front.jpg", strings.NewReader("jpeg-bytes"), "identity_documents")
	if err != nil {
		t.Fatalf("UploadDocument returned error: %v", err)
	}
	if up.UploadID != "upl_4f3e2d1c" {
		t.Errorf("UploadID = %q", up.UploadID)
	}

	getServer := accountServer(t, http.MethodGet, "/v1/connected-accounts/org_1/uploads/upl_4f3e2d1c", uploadExample)
	doc, _, err := getServer.ConnectedAccounts.GetDocument(context.Background(), "org_1", "upl_4f3e2d1c")
	if err != nil {
		t.Fatalf("GetDocument returned error: %v", err)
	}
	if doc.FileName != "product-hero.png" {
		t.Errorf("FileName = %q", doc.FileName)
	}
}
