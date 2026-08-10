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

// subscriptionExample is the exact SubscriptionResponse example from the
// subscription doc pages (get, update, cancel).
const subscriptionExample = `{
	"id": "sub_1a2b3c4d5e6f",
	"payment_method_id": "pm_7h8i9j0k",
	"status": "active",
	"collection_method": "charge_automatically",
	"currency": "USD",
	"amount": "10.00",
	"billing_cycle": {"interval": "month", "frequency": 1},
	"quantity": 1,
	"current_period_start": "2026-07-13T12:00:00Z",
	"current_period_end": "2026-08-13T12:00:00Z",
	"previously_billed_at": "2026-07-13T12:00:00Z",
	"next_billed_at": "2026-08-13T12:00:00Z",
	"trial_end": null,
	"cancel_at_period_end": false,
	"canceled_at": null,
	"created_at": "2026-07-13T12:00:00Z",
	"product": {
		"id": "prod_abc123",
		"name": "Pro plan",
		"description": "Everything in Pro.",
		"status": "active",
		"billing_cycle": {"interval": "month", "frequency": 1},
		"trial_period": null,
		"created_at": "2026-07-01T09:00:00Z",
		"updated_at": "2026-07-01T09:00:00Z"
	},
	"items": [
		{
			"id": "si_11aa22bb",
			"status": "active",
			"quantity": 1,
			"recurring": true,
			"price_type": "fixed",
			"unit_amount": "10.00",
			"currency": "USD",
			"previously_billed_at": "2026-07-13T12:00:00Z",
			"next_billed_at": "2026-08-13T12:00:00Z",
			"price": {
				"id": "price_pro_usd",
				"product_id": "prod_abc123",
				"price_type": "fixed",
				"currency": "USD",
				"unit_amount": "10.00",
				"billing_cycle": {"interval": "month", "frequency": 1},
				"trial_period": null,
				"seat_tiers": null,
				"is_archived": false,
				"created_at": "2026-07-01T09:00:00Z",
				"updated_at": "2026-07-01T09:00:00Z"
			},
			"product": {
				"id": "prod_abc123",
				"name": "Pro plan",
				"status": "active",
				"billing_cycle": {"interval": "month", "frequency": 1},
				"trial_period": null,
				"created_at": "2026-07-01T09:00:00Z",
				"updated_at": "2026-07-01T09:00:00Z"
			},
			"created_at": "2026-07-13T12:00:00Z",
			"updated_at": "2026-07-13T12:00:00Z"
		}
	],
	"customer": {
		"customer_id": "cust_xyz789",
		"email": "customer@example.com",
		"name": "Jane Doe",
		"phone_number": "+2348012345678",
		"metadata": {},
		"created_at": "2026-07-01T09:00:00Z",
		"updated_at": "2026-07-01T09:00:00Z"
	}
}`

// TestSubscriptionGet uses the exact example payload from
// https://docs.bachs.io/api-reference/subscriptions/get-subscription
func TestSubscriptionGet(t *testing.T) {
	c := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Errorf("method = %s, want GET", r.Method)
		}
		if r.URL.Path != "/v1/subscriptions/sub_1a2b3c4d5e6f" {
			t.Errorf("path = %q, want /v1/subscriptions/sub_1a2b3c4d5e6f", r.URL.Path)
		}
		io.WriteString(w, subscriptionExample)
	})

	sub, _, err := c.Subscriptions.Get(context.Background(), "sub_1a2b3c4d5e6f")
	if err != nil {
		t.Fatalf("Get returned error: %v", err)
	}

	if sub.ID != "sub_1a2b3c4d5e6f" {
		t.Errorf("ID = %q", sub.ID)
	}
	if sub.Status != "active" || sub.Amount != "10.00" || sub.Currency != "USD" {
		t.Errorf("Status/Amount/Currency = %s/%s/%s", sub.Status, sub.Amount, sub.Currency)
	}
	if sub.BillingCycle.Interval != "month" || sub.BillingCycle.Frequency != 1 {
		t.Errorf("BillingCycle = %+v", sub.BillingCycle)
	}
	if sub.Customer.CustomerID != "cust_xyz789" {
		t.Errorf("Customer.CustomerID = %q", sub.Customer.CustomerID)
	}
	if len(sub.Items) != 1 {
		t.Fatalf("len(Items) = %d, want 1", len(sub.Items))
	}
	if sub.Items[0].UnitAmount != "10.00" || !sub.Items[0].Recurring {
		t.Errorf("Items[0] = %+v", sub.Items[0])
	}
	if sub.Items[0].Price == nil || sub.Items[0].Price.ID != "price_pro_usd" {
		t.Errorf("Items[0].Price = %+v", sub.Items[0].Price)
	}
	wantPeriodEnd := time.Date(2026, 8, 13, 12, 0, 0, 0, time.UTC)
	if !sub.CurrentPeriodEnd.Equal(wantPeriodEnd) {
		t.Errorf("CurrentPeriodEnd = %v, want %v", sub.CurrentPeriodEnd, wantPeriodEnd)
	}
	if sub.TrialEnd != nil {
		t.Errorf("TrialEnd = %v, want nil", sub.TrialEnd)
	}
}

func TestSubscriptionList(t *testing.T) {
	c := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Errorf("method = %s, want GET", r.Method)
		}
		if r.URL.Path != "/v1/subscriptions" {
			t.Errorf("path = %q, want /v1/subscriptions", r.URL.Path)
		}
		if got := r.URL.Query().Get("customer_id"); got != "cust_xyz789" {
			t.Errorf("customer_id = %q, want cust_xyz789", got)
		}
		if got := r.URL.Query().Get("status"); got != "active" {
			t.Errorf("status = %q, want active", got)
		}
		io.WriteString(w, `{
			"items": [`+subscriptionExample+`],
			"pagination": {
				"next_cursor": "cur_20",
				"prev_cursor": null,
				"has_more": true,
				"limit": 20,
				"offset": 0,
				"returned": 1,
				"total": 5
			}
		}`)
	})

	page, _, err := c.Subscriptions.List(context.Background(), ListParams{CustomerID: "cust_xyz789", Status: "active"})
	if err != nil {
		t.Fatalf("List returned error: %v", err)
	}
	if len(page.Items) != 1 || page.Items[0].ID != "sub_1a2b3c4d5e6f" {
		t.Errorf("Items = %+v", page.Items)
	}
	if page.Pagination.Total != 5 {
		t.Errorf("Pagination.Total = %d, want 5", page.Pagination.Total)
	}
}

// TestSubscriptionUpdatePlanChange uses the exact "Change the plan" example
// from https://docs.bachs.io/api-reference/subscriptions/update-subscription
func TestSubscriptionUpdatePlanChange(t *testing.T) {
	c := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPatch {
			t.Errorf("method = %s, want PATCH", r.Method)
		}
		if r.URL.Path != "/v1/subscriptions/sub_1" {
			t.Errorf("path = %q", r.URL.Path)
		}
		body, _ := io.ReadAll(r.Body)
		var got map[string]any
		if err := json.Unmarshal(body, &got); err != nil {
			t.Fatalf("request body is not valid JSON: %v", err)
		}
		if got["product_id"] != "prod_premium" || got["proration_behavior"] != "invoice_now" {
			t.Errorf("request body = %s", body)
		}
		io.WriteString(w, subscriptionExample)
	})

	sub, _, err := c.Subscriptions.Update(context.Background(), "sub_1", UpdateSubscriptionRequest{
		ProductID:         "prod_premium",
		ProrationBehavior: "invoice_now",
	})
	if err != nil {
		t.Fatalf("Update returned error: %v", err)
	}
	if sub.ID != "sub_1a2b3c4d5e6f" {
		t.Errorf("ID = %q", sub.ID)
	}
}

func TestSubscriptionUpdateTrialMove(t *testing.T) {
	c := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		var got map[string]any
		if err := json.Unmarshal(body, &got); err != nil {
			t.Fatalf("request body is not valid JSON: %v", err)
		}
		if got["trial_end"] != "2026-08-10T12:00:00Z" {
			t.Errorf("trial_end = %v, want 2026-08-10T12:00:00Z", got["trial_end"])
		}
		io.WriteString(w, subscriptionExample)
	})

	trialEnd := time.Date(2026, 8, 10, 12, 0, 0, 0, time.UTC)
	if _, _, err := c.Subscriptions.Update(context.Background(), "sub_1", UpdateSubscriptionRequest{
		TrialEnd: &trialEnd,
	}); err != nil {
		t.Fatalf("Update returned error: %v", err)
	}
}

func TestSubscriptionUpdatePaymentMethod(t *testing.T) {
	c := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		if !strings.Contains(string(body), `"payment_method_id":"pm_new456"`) {
			t.Errorf("request body = %s", body)
		}
		io.WriteString(w, subscriptionExample)
	})

	if _, _, err := c.Subscriptions.Update(context.Background(), "sub_1", UpdateSubscriptionRequest{
		PaymentMethodID: "pm_new456",
	}); err != nil {
		t.Fatalf("Update returned error: %v", err)
	}
}

func TestSubscriptionUpdateMetadata(t *testing.T) {
	c := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		var got map[string]any
		if err := json.Unmarshal(body, &got); err != nil {
			t.Fatalf("request body is not valid JSON: %v", err)
		}
		meta, ok := got["metadata"].(map[string]any)
		if !ok || meta["plan"] != "pro" {
			t.Errorf("metadata = %v", got["metadata"])
		}
		io.WriteString(w, subscriptionExample)
	})

	if _, _, err := c.Subscriptions.Update(context.Background(), "sub_1", UpdateSubscriptionRequest{
		Metadata: map[string]any{"plan": "pro", "seat_count": "5"},
	}); err != nil {
		t.Fatalf("Update returned error: %v", err)
	}
}

// TestSubscriptionUpdateRejectsCombinedIntents verifies the exactly-one-intent
// rule is enforced before any request is sent.
func TestSubscriptionUpdateRejectsCombinedIntents(t *testing.T) {
	c := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		t.Error("request must not be sent when intents are combined")
	})

	_, _, err := c.Subscriptions.Update(context.Background(), "sub_1", UpdateSubscriptionRequest{
		ProductID:       "prod_1",
		PaymentMethodID: "pm_1",
	})
	if err == nil {
		t.Fatal("Update with two intents returned nil error, want rejection")
	}
	if !strings.Contains(err.Error(), "exactly one") {
		t.Errorf("error = %v, want a clear exactly-one-intent message", err)
	}
}

func TestSubscriptionUpdateRejectsNoIntent(t *testing.T) {
	c := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		t.Error("request must not be sent when no intent is set")
	})

	_, _, err := c.Subscriptions.Update(context.Background(), "sub_1", UpdateSubscriptionRequest{})
	if err == nil {
		t.Fatal("Update with no intent returned nil error, want rejection")
	}
	if !strings.Contains(err.Error(), "exactly one") {
		t.Errorf("error = %v, want a clear exactly-one-intent message", err)
	}
}

func TestSubscriptionUpdateRejectsProrationWithoutPlan(t *testing.T) {
	_, _, err := (&SubscriptionService{}).Update(context.Background(), "sub_1", UpdateSubscriptionRequest{
		ProrationBehavior: "invoice_now",
	})
	if err == nil {
		t.Fatal("Update with proration_behavior but no product_id returned nil error")
	}
	if !strings.Contains(err.Error(), "proration_behavior") {
		t.Errorf("error = %v, want a proration_behavior message", err)
	}
}

// TestSubscriptionCancel uses the DELETE endpoint documented in the OpenAPI
// spec and the cancel doc page.
func TestSubscriptionCancel(t *testing.T) {
	c := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodDelete {
			t.Errorf("method = %s, want DELETE", r.Method)
		}
		if r.URL.Path != "/v1/subscriptions/sub_1a2b3c4d5e6f" {
			t.Errorf("path = %q, want /v1/subscriptions/sub_1a2b3c4d5e6f", r.URL.Path)
		}
		body, _ := io.ReadAll(r.Body)
		var got map[string]any
		if err := json.Unmarshal(body, &got); err != nil {
			t.Fatalf("request body is not valid JSON: %v", err)
		}
		if got["cancel_at_period_end"] != true {
			t.Errorf("cancel_at_period_end = %v, want true", got["cancel_at_period_end"])
		}
		if got["reason"] != "Too expensive" {
			t.Errorf("reason = %v, want Too expensive", got["reason"])
		}
		io.WriteString(w, subscriptionExample)
	})

	reason := "Too expensive"
	sub, _, err := c.Subscriptions.Cancel(context.Background(), "sub_1a2b3c4d5e6f", CancelSubscriptionRequest{
		CancelAtPeriodEnd: true,
		Reason:            &reason,
	})
	if err != nil {
		t.Fatalf("Cancel returned error: %v", err)
	}
	if sub.Status != "active" {
		t.Errorf("Status = %q", sub.Status)
	}
}

// TestNoSubscriptionCreateMethod guards constraint #8: the service has no
// Create method (subscriptions are only created by completing a checkout for
// a recurring product).
func TestNoSubscriptionCreateMethod(t *testing.T) {
	sub := &SubscriptionService{}
	if _, ok := interface{}(sub).(interface{ Create() }); ok {
		t.Error("SubscriptionService must not have a Create method")
	}
}
