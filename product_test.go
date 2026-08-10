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

// productExample is the exact example payload used across the product doc
// pages (create, retrieve, list, update).
const productExample = `{
	"id": "prod_1a2b3c4d5e6f7g8h",
	"organization_id": "org_abc123",
	"name": "Pro Plan",
	"description": "Full access, billed monthly.",
	"price": {
		"currency": "USD",
		"price_type": "fixed",
		"amount": "29.00",
		"preset_amount": null,
		"minimum_amount": null,
		"maximum_amount": null,
		"currency_options": []
	},
	"prices": [
		{"currency": "USD", "amount": "29.00", "minimum_amount": null, "maximum_amount": null, "is_default": true}
	],
	"billing_cycle": {"interval": "month", "frequency": 1},
	"trial_period": null,
	"status": "active",
	"metadata": {"tier": "pro"},
	"media": [],
	"actor_id": "usr_abc123",
	"total_payments": 0,
	"total_amount": "0.00",
	"created_at": "2026-07-13T14:00:00.000Z",
	"updated_at": "2026-07-13T14:00:00.000Z",
	"archived_at": null
}`

// TestProductCreate uses the exact request/response examples from
// https://docs.bachs.io/api-reference/products/create-a-product
func TestProductCreate(t *testing.T) {
	c := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("method = %s, want POST", r.Method)
		}
		if r.URL.Path != "/v1/products" {
			t.Errorf("path = %q, want /v1/products", r.URL.Path)
		}

		body, _ := io.ReadAll(r.Body)
		var got map[string]any
		if err := json.Unmarshal(body, &got); err != nil {
			t.Fatalf("request body is not valid JSON: %v", err)
		}
		if got["name"] != "Pro Plan" {
			t.Errorf("name = %v", got["name"])
		}
		price := got["price"].(map[string]any)
		if price["currency"] != "USD" || price["amount"] != "29.00" {
			t.Errorf("price = %v", price)
		}
		opts := price["currency_options"].([]any)
		first := opts[0].(map[string]any)
		if first["currency"] != "NGN" || first["amount"] != "45000.00" {
			t.Errorf("currency_options = %v", opts)
		}
		cycle := got["billing_cycle"].(map[string]any)
		if cycle["interval"] != "month" || cycle["frequency"] != float64(1) {
			t.Errorf("billing_cycle = %v", cycle)
		}

		io.WriteString(w, productExample)
	})

	ngn := "45000.00"
	req := CreateProductRequest{
		Name:        "Pro Plan",
		Description: stringPtr("Monthly access to all Pro features."),
		Price: ProductPrice{
			Currency: "USD",
			Amount:   "29.00",
			CurrencyOptions: []CurrencyOption{
				{Currency: "NGN", Amount: ngn},
			},
		},
		BillingCycle: &Cadence{Interval: "month", Frequency: 1},
	}
	product, _, err := c.Products.Create(context.Background(), req)
	if err != nil {
		t.Fatalf("Create returned error: %v", err)
	}

	if product.ID != "prod_1a2b3c4d5e6f7g8h" {
		t.Errorf("ID = %q", product.ID)
	}
	if product.Name != "Pro Plan" {
		t.Errorf("Name = %q", product.Name)
	}
	if product.Price.Amount != "29.00" || product.Price.Currency != "USD" {
		t.Errorf("Price = %+v", product.Price)
	}
	if len(product.Prices) != 1 || !product.Prices[0].IsDefault {
		t.Errorf("Prices = %+v, want one default price", product.Prices)
	}
	if product.BillingCycle == nil || product.BillingCycle.Interval != "month" {
		t.Errorf("BillingCycle = %+v", product.BillingCycle)
	}
	if product.TotalAmount != "0.00" || product.TotalPayments != 0 {
		t.Errorf("TotalAmount/TotalPayments = %s/%d", product.TotalAmount, product.TotalPayments)
	}
	wantCreated := time.Date(2026, 7, 13, 14, 0, 0, 0, time.UTC)
	if !product.CreatedAt.Equal(wantCreated) {
		t.Errorf("CreatedAt = %v, want %v", product.CreatedAt, wantCreated)
	}
}

// TestProductGet uses the exact example from
// https://docs.bachs.io/api-reference/products/retrieve-a-product
func TestProductGet(t *testing.T) {
	c := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Errorf("method = %s, want GET", r.Method)
		}
		if r.URL.Path != "/v1/products/prod_1a2b3c4d5e6f7g8h" {
			t.Errorf("path = %q", r.URL.Path)
		}
		io.WriteString(w, productExample)
	})

	product, _, err := c.Products.Get(context.Background(), "prod_1a2b3c4d5e6f7g8h")
	if err != nil {
		t.Fatalf("Get returned error: %v", err)
	}
	if product.ID != "prod_1a2b3c4d5e6f7g8h" || product.Status != "active" {
		t.Errorf("product = %+v", product)
	}
	if product.Description == nil || *product.Description != "Full access, billed monthly." {
		t.Errorf("Description = %v", product.Description)
	}
	if product.ArchivedAt != nil {
		t.Errorf("ArchivedAt = %v, want nil while active", product.ArchivedAt)
	}
}

// TestProductList uses the exact example from
// https://docs.bachs.io/api-reference/products/list-products
func TestProductList(t *testing.T) {
	c := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Errorf("method = %s, want GET", r.Method)
		}
		if r.URL.Path != "/v1/products" {
			t.Errorf("path = %q, want /v1/products", r.URL.Path)
		}
		if got := r.URL.Query().Get("include_archived"); got != "true" {
			t.Errorf("include_archived = %q, want true", got)
		}
		if got := r.URL.Query().Get("limit"); got != "5" {
			t.Errorf("limit = %q, want 5", got)
		}
		io.WriteString(w, `{
			"items": [`+productExample+`],
			"pagination": {
				"next_cursor": "cur_20",
				"prev_cursor": null,
				"has_more": false,
				"limit": 20,
				"offset": 0,
				"returned": 1,
				"total": 47
			}
		}`)
	})

	page, _, err := c.Products.List(context.Background(), ListParams{Limit: 5, IncludeArchived: true})
	if err != nil {
		t.Fatalf("List returned error: %v", err)
	}
	if len(page.Items) != 1 {
		t.Fatalf("len(Items) = %d, want 1", len(page.Items))
	}
	if page.Items[0].Name != "Pro Plan" {
		t.Errorf("Items[0].Name = %q", page.Items[0].Name)
	}
	if page.Pagination.Total != 47 || page.Pagination.Returned != 1 {
		t.Errorf("Pagination = %+v", page.Pagination)
	}
	if page.Pagination.NextCursor == nil || *page.Pagination.NextCursor != "cur_20" {
		t.Errorf("NextCursor = %v", page.Pagination.NextCursor)
	}
}

// TestProductUpdate uses the exact example from
// https://docs.bachs.io/api-reference/products/update-a-product
func TestProductUpdate(t *testing.T) {
	c := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPatch {
			t.Errorf("method = %s, want PATCH", r.Method)
		}
		if r.URL.Path != "/v1/products/prod_1a2b3c4d5e6f7g8h" {
			t.Errorf("path = %q", r.URL.Path)
		}
		body, _ := io.ReadAll(r.Body)
		var got map[string]any
		if err := json.Unmarshal(body, &got); err != nil {
			t.Fatalf("request body is not valid JSON: %v", err)
		}
		if got["description"] != "Updated description." {
			t.Errorf("description = %v", got["description"])
		}
		meta := got["metadata"].(map[string]any)
		if meta["tier"] != "pro" {
			t.Errorf("metadata = %v", meta)
		}

		// The exact response example from the doc page.
		updated := strings.Replace(productExample, "Full access, billed monthly.", "Updated description.", 1)
		io.WriteString(w, updated)
	})

	product, _, err := c.Products.Update(context.Background(), "prod_1a2b3c4d5e6f7g8h", UpdateProductRequest{
		Description: stringPtr("Updated description."),
		Metadata:    map[string]any{"tier": "pro"},
	})
	if err != nil {
		t.Fatalf("Update returned error: %v", err)
	}
	if product.Description == nil || *product.Description != "Updated description." {
		t.Errorf("Description = %v", product.Description)
	}
}

func TestProductArchive(t *testing.T) {
	c := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("method = %s, want POST", r.Method)
		}
		if r.URL.Path != "/v1/products/prod_1/archive" {
			t.Errorf("path = %q, want /v1/products/prod_1/archive", r.URL.Path)
		}
		io.WriteString(w, strings.Replace(productExample, `"status": "active"`, `"status": "archived"`, 1))
	})

	product, _, err := c.Products.Archive(context.Background(), "prod_1")
	if err != nil {
		t.Fatalf("Archive returned error: %v", err)
	}
	if product.Status != "archived" {
		t.Errorf("Status = %q, want archived", product.Status)
	}
}

func TestProductUnarchive(t *testing.T) {
	c := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("method = %s, want POST", r.Method)
		}
		if r.URL.Path != "/v1/products/prod_1/unarchive" {
			t.Errorf("path = %q, want /v1/products/prod_1/unarchive", r.URL.Path)
		}
		io.WriteString(w, productExample)
	})

	product, _, err := c.Products.Unarchive(context.Background(), "prod_1")
	if err != nil {
		t.Fatalf("Unarchive returned error: %v", err)
	}
	if product.Status != "active" {
		t.Errorf("Status = %q, want active", product.Status)
	}
}
