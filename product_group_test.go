package bachs

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"testing"
)

const productGroupExample = `{
	"id": "pgrp_1afe3b5a393c422f8ec4",
	"organization_id": "acct_xourxlv2BkwUfwyK",
	"name": "probe-group",
	"products": [
		{
			"id": "prod_c29c67fa3310496482aa",
			"organization_id": "acct_xourxlv2BkwUfwyK",
			"name": "probe-product",
			"description": null,
			"price": {"currency": "USD", "price_type": "fixed", "amount": "1.00"},
			"status": "active",
			"metadata": null,
			"media": [],
			"actor_id": "usr_17f1f3a09aae48dfb21cf08a29b83f05",
			"total_payments": 0,
			"total_amount": "0.00",
			"created_at": "2026-10-07T13:46:24.630394Z",
			"updated_at": "2026-10-07T13:46:24.630399Z",
			"archived_at": null,
			"billing_cycle": null,
			"trial_period": null,
			"prices": []
		}
	],
	"created_at": "2026-10-07T13:46:25.661151Z",
	"updated_at": "2026-10-07T13:46:25.661154Z"
}`

// The fixtures mirror the live sandbox responses captured on 2026-10-07.
func TestProductGroupCreate(t *testing.T) {
	c := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("method = %s, want POST", r.Method)
		}
		if r.URL.Path != "/v1/product-groups" {
			t.Errorf("path = %q, want /v1/product-groups", r.URL.Path)
		}
		body, _ := io.ReadAll(r.Body)
		var got map[string]any
		if err := json.Unmarshal(body, &got); err != nil {
			t.Fatalf("request body is not valid JSON: %v", err)
		}
		if got["name"] != "probe-group" {
			t.Errorf("name = %v", got["name"])
		}
		io.WriteString(w, productGroupExample)
	})

	group, _, err := c.ProductGroups.Create(context.Background(), CreateProductGroupRequest{
		Name:       "probe-group",
		ProductIDs: []string{"prod_c29c67fa3310496482aa"},
	})
	if err != nil {
		t.Fatalf("Create returned error: %v", err)
	}
	if group.ID != "pgrp_1afe3b5a393c422f8ec4" || group.Name != "probe-group" {
		t.Errorf("group = %+v", group)
	}
	if len(group.Products) != 1 || group.Products[0].ID != "prod_c29c67fa3310496482aa" {
		t.Errorf("Products = %+v", group.Products)
	}
	if group.CreatedAt.IsZero() {
		t.Error("CreatedAt is zero")
	}
}

func TestProductGroupList(t *testing.T) {
	c := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v1/product-groups" {
			t.Errorf("path = %q", r.URL.Path)
		}
		io.WriteString(w, `{"items": [`+productGroupExample+`]}`)
	})

	page, _, err := c.ProductGroups.List(context.Background(), ListParams{})
	if err != nil {
		t.Fatalf("List returned error: %v", err)
	}
	if len(page.Items) != 1 || page.Items[0].ID != "pgrp_1afe3b5a393c422f8ec4" {
		t.Errorf("Items = %+v", page.Items)
	}
}

func TestProductGroupGet(t *testing.T) {
	c := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v1/product-groups/pgrp_1afe3b5a393c422f8ec4" {
			t.Errorf("path = %q", r.URL.Path)
		}
		io.WriteString(w, productGroupExample)
	})

	group, _, err := c.ProductGroups.Get(context.Background(), "pgrp_1afe3b5a393c422f8ec4")
	if err != nil {
		t.Fatalf("Get returned error: %v", err)
	}
	if group.OrganizationID != "acct_xourxlv2BkwUfwyK" {
		t.Errorf("OrganizationID = %q", group.OrganizationID)
	}
}

func TestProductGroupUpdate(t *testing.T) {
	c := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPatch {
			t.Errorf("method = %s, want PATCH", r.Method)
		}
		if r.URL.Path != "/v1/product-groups/pgrp_1afe3b5a393c422f8ec4" {
			t.Errorf("path = %q", r.URL.Path)
		}
		body, _ := io.ReadAll(r.Body)
		var got map[string]any
		if err := json.Unmarshal(body, &got); err != nil {
			t.Fatalf("request body is not valid JSON: %v", err)
		}
		if got["name"] != "probe-group-renamed" {
			t.Errorf("name = %v", got["name"])
		}
		if _, present := got["product_ids"]; present {
			t.Errorf("product_ids should be absent when only renaming: %v", got)
		}
		io.WriteString(w, productGroupExample)
	})

	group, _, err := c.ProductGroups.Update(context.Background(), "pgrp_1afe3b5a393c422f8ec4", UpdateProductGroupRequest{
		Name: strptr("probe-group-renamed"),
	})
	if err != nil {
		t.Fatalf("Update returned error: %v", err)
	}
	if group.ID != "pgrp_1afe3b5a393c422f8ec4" {
		t.Errorf("ID = %q", group.ID)
	}
}

func TestProductGroupDelete(t *testing.T) {
	c := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodDelete {
			t.Errorf("method = %s, want DELETE", r.Method)
		}
		if r.URL.Path != "/v1/product-groups/pgrp_1afe3b5a393c422f8ec4" {
			t.Errorf("path = %q", r.URL.Path)
		}
		// The live API answers with 204 and no body.
		w.WriteHeader(http.StatusNoContent)
	})

	meta, err := c.ProductGroups.Delete(context.Background(), "pgrp_1afe3b5a393c422f8ec4")
	if err != nil {
		t.Fatalf("Delete returned error: %v", err)
	}
	if meta == nil {
		t.Error("ResponseMeta is nil")
	}
}
