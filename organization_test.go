package bachs

import (
	"context"
	"io"
	"net/http"
	"testing"
)

// TestGetMeOrganization uses the exact example payload from
// https://docs.bachs.io/api-reference/organizations/get-my-organization
func TestGetMeOrganization(t *testing.T) {
	c := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Errorf("method = %s, want GET", r.Method)
		}
		if r.URL.Path != "/v1/organizations/me" {
			t.Errorf("path = %q", r.URL.Path)
		}
		io.WriteString(w, `{
			"id": "org_9f2c4a1b7e3d5086",
			"name": "Ada Stores",
			"owner_user_id": "usr_7b3e19d24c0a",
			"parent_organization_id": null,
			"country": "NG",
			"fee_handling": "org_pays_fee",
			"enabled_payment_methods": {
				"bank_transfer": {
					"enabled": true,
					"currencies": {"NGN": true, "USD": false}
				},
				"mobile_money": {
					"enabled": true,
					"currencies": {"GHS": true}
				}
			},
			"adaptive_pricing": true,
			"balance_currencies": ["NGN", "USD"],
			"website": null,
			"phone_number": null,
			"company_name": null,
			"enabled_capabilities": ["payouts", "conversions", "connect"],
			"capabilities": null,
			"requirements": null,
			"fields_needing_resubmission": null,
			"sandbox_org_id": "org_2c7d40b81ea6",
			"live_org_id": null,
			"is_active": true,
			"created_at": "2026-08-01T09:12:44.000Z",
			"updated_at": "2026-08-07T11:04:22.518Z",
			"controller": null
		}`)
	})

	org, _, err := c.Organizations.GetMe(context.Background())
	if err != nil {
		t.Fatalf("GetMe returned error: %v", err)
	}
	if org.ID != "org_9f2c4a1b7e3d5086" {
		t.Errorf("ID = %q", org.ID)
	}
	if org.Name == nil || *org.Name != "Ada Stores" {
		t.Errorf("Name = %v", org.Name)
	}
	if len(org.EnabledCapabilities) != 3 || org.EnabledCapabilities[0] != "payouts" {
		t.Errorf("EnabledCapabilities = %v", org.EnabledCapabilities)
	}
	if len(org.BalanceCurrencies) != 2 {
		t.Errorf("BalanceCurrencies = %v", org.BalanceCurrencies)
	}
	// GetMe deliberately leaves these null.
	if org.Capabilities != nil {
		t.Errorf("Capabilities = %v, want nil on GetMe", org.Capabilities)
	}
	if org.Requirements != nil {
		t.Errorf("Requirements = %v, want nil on GetMe", org.Requirements)
	}
	if !org.IsActive {
		t.Error("IsActive = false, want true")
	}
}

// TestGetOrganization uses the exact example payload from
// https://docs.bachs.io/api-reference/organizations/get-an-organization
func TestGetOrganization(t *testing.T) {
	c := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Errorf("method = %s, want GET", r.Method)
		}
		if r.URL.Path != "/v1/organizations/org_9f2c4a1b7e3d5086" {
			t.Errorf("path = %q", r.URL.Path)
		}
		io.WriteString(w, `{
			"id": "org_9f2c4a1b7e3d5086",
			"name": "Ada Stores",
			"owner_user_id": "usr_7b3e19d24c0a",
			"parent_organization_id": null,
			"country": "NG",
			"fee_handling": "org_pays_fee",
			"enabled_payment_methods": null,
			"adaptive_pricing": true,
			"balance_currencies": ["NGN", "USD"],
			"website": null,
			"phone_number": null,
			"company_name": null,
			"enabled_capabilities": ["payouts", "conversions", "connect"],
			"capabilities": {
				"payouts": {"status": "active", "requested": true, "status_details": null},
				"conversions": {"status": "active", "requested": true, "status_details": null},
				"connect": {"status": "active", "requested": true, "status_details": null},
				"transfers": {
					"status": "restricted",
					"requested": true,
					"status_details": [
						{
							"code": "platform_disabled",
							"resolution": "Contact support to re-enable this capability.",
							"message": "This capability was disabled by the platform."
						}
					]
				}
			},
			"requirements": {
				"setup_status": "complete",
				"currently_due": [],
				"eventually_due": [],
				"past_due": [],
				"pending_verification": [],
				"errors": []
			},
			"fields_needing_resubmission": null,
			"sandbox_org_id": "org_2c7d40b81ea6",
			"live_org_id": null,
			"is_active": true,
			"created_at": "2026-08-01T09:12:44.000Z",
			"updated_at": "2026-08-07T11:04:22.518Z",
			"controller": null
		}`)
	})

	org, _, err := c.Organizations.Get(context.Background(), "org_9f2c4a1b7e3d5086")
	if err != nil {
		t.Fatalf("Get returned error: %v", err)
	}
	if org.ID != "org_9f2c4a1b7e3d5086" {
		t.Errorf("ID = %q", org.ID)
	}
	// Unlike GetMe, Get populates the capabilities and requirements blocks.
	if org.Capabilities == nil {
		t.Fatal("Capabilities is nil, want populated")
	}
	if len(org.Capabilities) != 4 {
		t.Errorf("len(Capabilities) = %d, want 4", len(org.Capabilities))
	}
	if org.Capabilities["transfers"].Status != "restricted" {
		t.Errorf("transfers status = %q", org.Capabilities["transfers"].Status)
	}
	if org.Requirements == nil {
		t.Fatal("Requirements is nil, want populated")
	}
	if org.Requirements.SetupStatus != "complete" {
		t.Errorf("SetupStatus = %q", org.Requirements.SetupStatus)
	}
}

func TestGetCheckoutSettings(t *testing.T) {
	c := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Errorf("method = %s, want GET", r.Method)
		}
		if r.URL.Path != "/v1/organizations/checkout/settings" {
			t.Errorf("path = %q", r.URL.Path)
		}
		io.WriteString(w, `{
			"organization_id": "org_9f2c4a1b7e3d5086",
			"enabled_payment_methods": {
				"bank_transfer": {"enabled": true, "currencies": {"NGN": true}}
			},
			"fee_preference": "customer_pays",
			"available_currencies": {
				"bank_transfer": ["NGN", "USD"]
			}
		}`)
	})

	settings, _, err := c.Organizations.GetCheckoutSettings(context.Background())
	if err != nil {
		t.Fatalf("GetCheckoutSettings returned error: %v", err)
	}
	if settings.OrganizationID != "org_9f2c4a1b7e3d5086" {
		t.Errorf("OrganizationID = %q", settings.OrganizationID)
	}
	if settings.FeePreference != "customer_pays" {
		t.Errorf("FeePreference = %q", settings.FeePreference)
	}
	if settings.EnabledPaymentMethods == nil {
		t.Error("EnabledPaymentMethods is nil")
	}
	if settings.AvailableCurrencies == nil {
		t.Error("AvailableCurrencies is nil")
	}
}

func TestUpdateCheckoutSettings(t *testing.T) {
	c := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPut {
			t.Errorf("method = %s, want PUT", r.Method)
		}
		if r.URL.Path != "/v1/organizations/checkout/settings" {
			t.Errorf("path = %q", r.URL.Path)
		}
		io.WriteString(w, `{
			"organization_id": "org_9f2c4a1b7e3d5086",
			"enabled_payment_methods": {
				"bank_transfer": {"enabled": true, "currencies": {"NGN": true, "USD": false}}
			},
			"fee_preference": "org_pays",
			"message": "Checkout settings updated"
		}`)
	})

	settings, _, err := c.Organizations.UpdateCheckoutSettings(context.Background(), UpdateCheckoutSettingsRequest{
		FeePreference: "org_pays",
	})
	if err != nil {
		t.Fatalf("UpdateCheckoutSettings returned error: %v", err)
	}
	if settings.FeePreference != "org_pays" {
		t.Errorf("FeePreference = %q", settings.FeePreference)
	}
	if settings.Message != "Checkout settings updated" {
		t.Errorf("Message = %q", settings.Message)
	}
}
