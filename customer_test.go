package bachs

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"strings"
	"testing"
	"time"
)

// customerExample is the exact example payload from the customer doc pages
// (create, retrieve, update).
const customerExample = `{
	"customer_id": "cust_1a2b3c4d5e6f7g8h",
	"email": "ada@example.com",
	"name": "Ada Lovelace",
	"phone_number": "+2348012345678",
	"metadata": {"tier": "vip"},
	"billing_address": {
		"line1": "40 Yaba Road",
		"line2": null,
		"city": "Lagos",
		"state": "Lagos",
		"postal_code": "101245",
		"country": "NG"
	},
	"created_at": "2026-07-13T14:00:00.000Z",
	"updated_at": "2026-07-13T14:00:00.000Z"
}`

// TestCustomerCreate uses the exact request/response examples from
// https://docs.bachs.io/api-reference/customers/create-a-customer
func TestCustomerCreate(t *testing.T) {
	c := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("method = %s, want POST", r.Method)
		}
		if r.URL.Path != "/v1/customers" {
			t.Errorf("path = %q, want /v1/customers", r.URL.Path)
		}

		body, _ := io.ReadAll(r.Body)
		var got map[string]any
		if err := json.Unmarshal(body, &got); err != nil {
			t.Fatalf("request body is not valid JSON: %v", err)
		}
		if got["email"] != "jane@example.com" {
			t.Errorf("email = %v, want jane@example.com", got["email"])
		}
		meta := got["metadata"].(map[string]any)
		if meta["plan"] != "pro" {
			t.Errorf("metadata = %v", meta)
		}

		io.WriteString(w, customerExample)
	})

	customer, _, err := c.Customers.Create(context.Background(), CreateCustomerRequest{
		Email:    "jane@example.com",
		Metadata: map[string]any{"plan": "pro"},
	})
	if err != nil {
		t.Fatalf("Create returned error: %v", err)
	}

	if customer.CustomerID != "cust_1a2b3c4d5e6f7g8h" {
		t.Errorf("CustomerID = %q", customer.CustomerID)
	}
	if customer.Email != "ada@example.com" {
		t.Errorf("Email = %q", customer.Email)
	}
	if customer.Name == nil || *customer.Name != "Ada Lovelace" {
		t.Errorf("Name = %v", customer.Name)
	}
	if customer.BillingAddress == nil {
		t.Fatal("BillingAddress is nil")
	}
	if customer.BillingAddress.Country == nil || *customer.BillingAddress.Country != "NG" {
		t.Errorf("BillingAddress.Country = %v", customer.BillingAddress.Country)
	}
	if customer.BillingAddress.Line2 != nil {
		t.Errorf("BillingAddress.Line2 = %v, want nil", customer.BillingAddress.Line2)
	}
	wantCreated := time.Date(2026, 7, 13, 14, 0, 0, 0, time.UTC)
	if !customer.CreatedAt.Equal(wantCreated) {
		t.Errorf("CreatedAt = %v, want %v", customer.CreatedAt, wantCreated)
	}
}

func TestCustomerGet(t *testing.T) {
	c := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Errorf("method = %s, want GET", r.Method)
		}
		if r.URL.Path != "/v1/customers/cust_1a2b3c4d5e6f" {
			t.Errorf("path = %q", r.URL.Path)
		}
		io.WriteString(w, customerExample)
	})

	customer, _, err := c.Customers.Get(context.Background(), "cust_1a2b3c4d5e6f")
	if err != nil {
		t.Fatalf("Get returned error: %v", err)
	}
	if customer.Email != "ada@example.com" || customer.Metadata["tier"] != "vip" {
		t.Errorf("customer = %+v", customer)
	}
}

// TestCustomerList uses the exact example from
// https://docs.bachs.io/api-reference/customers/list-customers
func TestCustomerList(t *testing.T) {
	c := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Errorf("method = %s, want GET", r.Method)
		}
		if r.URL.Path != "/v1/customers" {
			t.Errorf("path = %q, want /v1/customers", r.URL.Path)
		}
		if got := r.URL.Query().Get("search"); got != "ada" {
			t.Errorf("search = %q, want ada", got)
		}
		io.WriteString(w, `{
			"items": [
				{
					"customer_id": "cust_1a2b3c4d5e6f7g8h",
					"email": "ada@example.com",
					"name": "Ada Lovelace",
					"metadata": {"tier": "vip"},
					"created_at": "2026-07-13T14:00:00.000Z",
					"updated_at": "2026-07-13T14:00:00.000Z"
				}
			],
			"pagination": {
				"next_cursor": "cur_20",
				"prev_cursor": null,
				"has_more": true,
				"limit": 20,
				"offset": 0,
				"returned": 1,
				"total": 47
			}
		}`)
	})

	page, _, err := c.Customers.List(context.Background(), ListParams{Search: "ada"})
	if err != nil {
		t.Fatalf("List returned error: %v", err)
	}
	if len(page.Items) != 1 {
		t.Fatalf("len(Items) = %d, want 1", len(page.Items))
	}
	if page.Items[0].CustomerID != "cust_1a2b3c4d5e6f7g8h" {
		t.Errorf("Items[0].CustomerID = %q", page.Items[0].CustomerID)
	}
	if page.Pagination.HasMore != true || page.Pagination.Total != 47 {
		t.Errorf("Pagination = %+v", page.Pagination)
	}
}

// TestCustomerUpdate uses the exact request example from
// https://docs.bachs.io/api-reference/customers/update-a-customer
func TestCustomerUpdate(t *testing.T) {
	c := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPatch {
			t.Errorf("method = %s, want PATCH", r.Method)
		}
		if r.URL.Path != "/v1/customers/cust_1a2b3c4d5e6f" {
			t.Errorf("path = %q", r.URL.Path)
		}

		body, _ := io.ReadAll(r.Body)
		var got map[string]any
		if err := json.Unmarshal(body, &got); err != nil {
			t.Fatalf("request body is not valid JSON: %v", err)
		}
		if got["email"] != "jane.new@example.com" {
			t.Errorf("email = %v", got["email"])
		}
		meta := got["metadata"].(map[string]any)
		if meta["plan"] != "enterprise" {
			t.Errorf("metadata = %v", meta)
		}
		if _, hasAddress := got["billing_address"]; hasAddress {
			t.Errorf("billing_address should be omitted when untouched: %s", body)
		}

		io.WriteString(w, customerExample)
	})

	customer, _, err := c.Customers.Update(context.Background(), "cust_1a2b3c4d5e6f", UpdateCustomerRequest{
		Email:    "jane.new@example.com",
		Metadata: map[string]any{"plan": "enterprise"},
	})
	if err != nil {
		t.Fatalf("Update returned error: %v", err)
	}
	if customer.Email != "ada@example.com" {
		t.Errorf("Email = %q, want ada@example.com from the example response", customer.Email)
	}
}

// TestCustomerUpdateClearsBillingAddress verifies the pointer-to-pointer
// tri-state: nil leaves the field alone, &nil clears it (sends null).
func TestCustomerUpdateClearsBillingAddress(t *testing.T) {
	c := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		var got map[string]any
		if err := json.Unmarshal(body, &got); err != nil {
			t.Fatalf("request body is not valid JSON: %v", err)
		}
		v, present := got["billing_address"]
		if !present {
			t.Fatalf("billing_address missing from body: %s", body)
		}
		if v != nil {
			t.Errorf("billing_address = %v, want explicit null", v)
		}
		io.WriteString(w, customerExample)
	})

	var clear *CustomerBillingAddress
	_, _, err := c.Customers.Update(context.Background(), "cust_1", UpdateCustomerRequest{
		BillingAddress: &clear,
	})
	if err != nil {
		t.Fatalf("Update returned error: %v", err)
	}
}

// TestCustomerUpdateReplacesBillingAddress sends a full address object.
func TestCustomerUpdateReplacesBillingAddress(t *testing.T) {
	c := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		if !strings.Contains(string(body), `"line1"`) || !strings.Contains(string(body), `"country"`) {
			t.Errorf("billing_address not serialized in full: %s", body)
		}
		io.WriteString(w, customerExample)
	})

	line1 := "1 Main Street"
	country := "NG"
	addr := &CustomerBillingAddress{Line1: &line1, Country: &country}
	_, _, err := c.Customers.Update(context.Background(), "cust_1", UpdateCustomerRequest{
		BillingAddress: &addr,
	})
	if err != nil {
		t.Fatalf("Update returned error: %v", err)
	}
}
