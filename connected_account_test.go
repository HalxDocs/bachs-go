package bachs

import (
	"context"
	"encoding/json"
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

// TestConnectedAccountCreate uses the request example from
// https://docs.bachs.io/connect/guides/create-an-account.
func TestConnectedAccountCreate(t *testing.T) {
	c := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("method = %s, want POST", r.Method)
		}
		if r.URL.Path != "/v1/accounts" {
			t.Errorf("path = %q, want /v1/accounts", r.URL.Path)
		}
		body, _ := io.ReadAll(r.Body)
		var got map[string]any
		if err := json.Unmarshal(body, &got); err != nil {
			t.Fatalf("request body is not valid JSON: %v", err)
		}
		if got["contact_email"] != "ada@adastores.example" {
			t.Errorf("contact_email = %v", got["contact_email"])
		}
		cfg, ok := got["configuration"].(map[string]any)
		if !ok {
			t.Fatalf("configuration missing: %s", body)
		}
		recipient, ok := cfg["recipient"].(map[string]any)
		if !ok {
			t.Fatalf("recipient persona missing: %s", body)
		}
		caps, ok := recipient["capabilities"].(map[string]any)
		if !ok || caps["payouts"].(map[string]any)["requested"] != true {
			t.Errorf("capabilities = %v, want payouts requested", recipient["capabilities"])
		}
		io.WriteString(w, connectedAccountExample)
	})

	acct, _, err := c.ConnectedAccounts.Create(context.Background(), CreateConnectedAccountRequest{
		ContactEmail: "ada@adastores.example",
		DisplayName:  stringPtr("Ada Stores"),
		Country:      stringPtr("NG"),
		EntityType:   stringPtr(AccountEntityIndividual),
		Configuration: map[string]PersonaConfig{
			AccountPersonaRecipient: {
				Capabilities: map[string]CapabilityRequest{
					"payouts":   {Requested: true},
					"transfers": {Requested: true},
				},
			},
		},
		Responsibilities: &ResponsibilitiesConfig{
			Fees: FeeCollectorConfig{Collector: FeeCollectorBachs},
		},
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
	c := accountServer(t, http.MethodGet, "/v1/accounts/org_4d81fa9c2b6e0357", connectedAccountExample)

	acct, _, err := c.ConnectedAccounts.Get(context.Background(), "org_4d81fa9c2b6e0357")
	if err != nil {
		t.Fatalf("Get returned error: %v", err)
	}
	if acct.ID != "org_4d81fa9c2b6e0357" || !acct.AdaptivePricing {
		t.Errorf("account = %+v", acct)
	}
}

// TestConnectedAccountGetRequirementsEntries decodes the live requirements
// shape observed on the sandbox: buckets plus an entries array with the
// field, its state, and the capabilities it holds up.
func TestConnectedAccountGetRequirementsEntries(t *testing.T) {
	c := accountServer(t, http.MethodGet, "/v1/accounts/acct_1", `{
		"id": "acct_1",
		"owner_user_id": "usr_1",
		"fee_handling": "customer_pays_fee",
		"adaptive_pricing": false,
		"balance_currencies": ["NGN"],
		"enabled_capabilities": ["payouts"],
		"is_active": true,
		"created_at": "2026-08-07T09:12:44.000Z",
		"updated_at": "2026-08-07T09:12:44.000Z",
		"requirements": {
			"currently_due": ["persons.name", "payout_destination"],
			"eventually_due": [],
			"past_due": [],
			"pending_verification": [],
			"errors": [],
			"current_deadline": null,
			"entries": [
				{
					"field": "persons.name",
					"status": "currently_due",
					"restricts_capabilities": ["payouts", "transfers"],
					"errors": [],
					"resolution": "api",
					"deadline": null
				}
			]
		}
	}`)

	acct, _, err := c.ConnectedAccounts.Get(context.Background(), "acct_1")
	if err != nil {
		t.Fatalf("Get returned error: %v", err)
	}
	reqs := acct.Requirements
	if reqs == nil {
		t.Fatal("Requirements is nil")
	}
	if len(reqs.CurrentlyDue) != 2 {
		t.Errorf("CurrentlyDue = %v", reqs.CurrentlyDue)
	}
	if len(reqs.Entries) != 1 {
		t.Fatalf("Entries = %+v", reqs.Entries)
	}
	entry := reqs.Entries[0]
	if entry.Field != "persons.name" || entry.Status != "currently_due" {
		t.Errorf("Entry = %+v", entry)
	}
	if len(entry.RestrictsCapabilities) != 2 {
		t.Errorf("RestrictsCapabilities = %v", entry.RestrictsCapabilities)
	}
}

func TestConnectedAccountList(t *testing.T) {
	c := accountServer(t, http.MethodGet, "/v1/accounts", `{
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
	c := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("method = %s, want POST", r.Method)
		}
		if r.URL.Path != "/v1/accounts/org_1" {
			t.Errorf("path = %q, want /v1/accounts/org_1", r.URL.Path)
		}
		body, _ := io.ReadAll(r.Body)
		var got map[string]any
		if err := json.Unmarshal(body, &got); err != nil {
			t.Fatalf("request body is not valid JSON: %v", err)
		}
		cfg, ok := got["configuration"].(map[string]any)
		if !ok {
			t.Fatalf("configuration missing: %s", body)
		}
		recipient, ok := cfg["recipient"].(map[string]any)
		if !ok {
			t.Fatalf("recipient persona missing: %s", body)
		}
		caps, ok := recipient["capabilities"].(map[string]any)
		if !ok || caps["conversions"].(map[string]any)["requested"] != true {
			t.Errorf("capabilities = %v, want conversions requested", recipient["capabilities"])
		}
		io.WriteString(w, connectedAccountExample)
	})

	acct, _, err := c.ConnectedAccounts.RequestCapabilities(context.Background(), "org_1", UpdateConnectedAccountRequest{
		Configuration: map[string]PersonaConfig{
			AccountPersonaRecipient: {
				Capabilities: map[string]CapabilityRequest{"conversions": {Requested: true}},
			},
		},
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
	c := accountServer(t, http.MethodPost, "/v1/accounts/org_4d81fa9c2b6e0357/account-links", `{
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
	c := accountServer(t, http.MethodGet, "/v1/accounts/org_4d81fa9c2b6e0357/capabilities", `{
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
// TestConnectedAccountUpdateAccount uses the submission shape from
// https://docs.bachs.io/connect/guides/api-onboarding: contact details,
// capability requests, and requirement fields in one call. Only the fields
// sent change; the request below updates the listed name.
func TestConnectedAccountUpdateAccount(t *testing.T) {
	c := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("method = %s, want POST", r.Method)
		}
		if r.URL.Path != "/v1/accounts/acct_1" {
			t.Errorf("path = %q, want /v1/accounts/acct_1", r.URL.Path)
		}
		body, _ := io.ReadAll(r.Body)
		var got map[string]any
		if err := json.Unmarshal(body, &got); err != nil {
			t.Fatalf("request body is not valid JSON: %v", err)
		}
		if got["display_name"] != "Ada Stores Renamed" {
			t.Errorf("display_name = %v", got["display_name"])
		}
		if _, present := got["fields"]; present {
			t.Errorf("fields should be absent when only updating contact details: %s", body)
		}
		io.WriteString(w, `{
			"id": "acct_1",
			"owner_user_id": "usr_1",
			"name": "Ada Stores Renamed",
			"fee_handling": "customer_pays_fee",
			"adaptive_pricing": false,
			"balance_currencies": ["NGN"],
			"enabled_capabilities": ["payouts"],
			"is_active": true,
			"created_at": "2026-08-07T09:12:44.000Z",
			"updated_at": "2026-10-09T12:00:00.000Z"
		}`)
	})

	acct, _, err := c.ConnectedAccounts.UpdateAccount(context.Background(), "acct_1", UpdateAccountRequest{
		DisplayName: stringPtr("Ada Stores Renamed"),
	})
	if err != nil {
		t.Fatalf("UpdateAccount returned error: %v", err)
	}
	if acct.Name == nil || *acct.Name != "Ada Stores Renamed" {
		t.Errorf("Name = %v", acct.Name)
	}
}
