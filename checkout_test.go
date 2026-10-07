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

// TestCheckoutCreate uses the exact request/response examples from
// https://docs.bachs.io/api-reference/payments/create-checkout-session
func TestCheckoutCreate(t *testing.T) {
	c := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("method = %s, want POST", r.Method)
		}
		if r.URL.Path != "/v1/checkout-sessions" {
			t.Errorf("path = %q, want /v1/checkout-sessions", r.URL.Path)
		}

		body, _ := io.ReadAll(r.Body)
		var got map[string]any
		if err := json.Unmarshal(body, &got); err != nil {
			t.Fatalf("request body is not valid JSON: %v", err)
		}
		customer := got["customer"].(map[string]any)
		if customer["email"] != "customer@example.com" || customer["name"] != "John Doe" {
			t.Errorf("customer = %v", customer)
		}
		cart := got["product_cart"].([]any)
		if cart[0].(map[string]any)["product_id"] != "prod_abc123" {
			t.Errorf("product_cart = %v", cart)
		}

		// The exact example payload from the doc page.
		io.WriteString(w, `{
			"checkout_id": "chk_1a2b3c4d5e6f",
			"checkout_url": "https://checkout.bachs.io/c/chk_1a2b3c4d5e6f",
			"status": "open",
			"expires_at": "2026-01-24T15:30:00.000Z",
			"created_at": "2026-01-24T14:30:00.000Z",
			"reference": "order_9876"
		}`)
	})

	req := CreateCheckoutSessionRequest{
		Customer:    &CheckoutCustomer{Email: "customer@example.com", Name: "John Doe"},
		ProductCart: []ProductItemRequest{{ProductID: "prod_abc123"}},
	}
	session, _, err := c.Checkouts.Create(context.Background(), req)
	if err != nil {
		t.Fatalf("Create returned error: %v", err)
	}

	if session.CheckoutID != "chk_1a2b3c4d5e6f" {
		t.Errorf("CheckoutID = %q, want chk_1a2b3c4d5e6f", session.CheckoutID)
	}
	if session.CheckoutURL != "https://checkout.bachs.io/c/chk_1a2b3c4d5e6f" {
		t.Errorf("CheckoutURL = %q", session.CheckoutURL)
	}
	if session.Status != "open" {
		t.Errorf("Status = %q, want open", session.Status)
	}
	if session.Reference != "order_9876" {
		t.Errorf("Reference = %q, want order_9876", session.Reference)
	}
	wantExpiry := time.Date(2026, 1, 24, 15, 30, 0, 0, time.UTC)
	if !session.ExpiresAt.Equal(wantExpiry) {
		t.Errorf("ExpiresAt = %v, want %v", session.ExpiresAt, wantExpiry)
	}
}

func TestCheckoutCreateSendsAdHocPricing(t *testing.T) {
	c := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		var got map[string]any
		if err := json.Unmarshal(body, &got); err != nil {
			t.Fatalf("request body is not valid JSON: %v", err)
		}
		pricing, ok := got["pricing"].(map[string]any)
		if !ok {
			t.Fatalf("pricing missing from body: %s", body)
		}
		if pricing["currency"] != "USD" || pricing["amount"] != "100.00" || pricing["price_type"] != "custom" {
			t.Errorf("pricing = %v", pricing)
		}
		if _, hasCart := got["product_cart"]; hasCart {
			t.Errorf("product_cart should be absent when pricing is set: %s", body)
		}
		io.WriteString(w, `{"checkout_id":"chk_x","checkout_url":"https://checkout.bachs.io/c/chk_x","status":"OPEN","expires_at":"2026-01-24T15:30:00Z","created_at":"2026-01-24T14:30:00Z"}`)
	})

	req := CreateCheckoutSessionRequest{
		Customer: &CheckoutCustomer{Email: "buyer@example.com"},
		Pricing: &CheckoutPricing{
			Currency:  "USD",
			Amount:    "100.00",
			PriceType: "custom",
		},
	}
	if _, _, err := c.Checkouts.Create(context.Background(), req); err != nil {
		t.Fatalf("Create returned error: %v", err)
	}
}

// TestCheckoutGet uses the exact example payload from
// https://docs.bachs.io/api-reference/checkout-sessions/get-checkout-session
func TestCheckoutGet(t *testing.T) {
	c := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Errorf("method = %s, want GET", r.Method)
		}
		if r.URL.Path != "/v1/checkout-sessions/chk_1a2b3c4d5e6f" {
			t.Errorf("path = %q, want /v1/checkout-sessions/chk_1a2b3c4d5e6f", r.URL.Path)
		}
		io.WriteString(w, `{
			"checkout_id": "chk_1a2b3c4d5e6f",
			"status": "COMPLETED",
			"recurring": null,
			"payment_status": "succeeded",
			"source_type": "CHECKOUT_SESSION",
			"amount": "50.00",
			"currency": "USD",
			"reference": "order_9876",
			"charge": {
				"payment_id": "pay_1a2b3c4d5e",
				"billing_reason": "purchase",
				"status": "succeeded",
				"amount": "29.00",
				"currency": "USD",
				"fee_usd": "0.59"
			},
			"payment_method": "card",
			"customer": {
				"id": "cust_1a2b3c4d5e6f",
				"email": "jane@example.com",
				"name": "Jane Doe"
			},
			"success_url": "https://yourapp.com/success",
			"cancel_url": "https://yourapp.com/cancel",
			"products": [
				{
					"product_id": "prod_abc123",
					"product_name": "Premium Plan",
					"quantity": 1,
					"unit_amount": "50.00",
					"currency": "USD",
					"price_type": "fixed",
					"minimum_amount": null,
					"maximum_amount": null,
					"line_total": "50.00"
				}
			],
			"billing_currency": "NGN",
			"session_mode": "CART",
			"metadata": {"order_id": "ORD-9876"},
			"created_at": "2026-01-24T14:30:00.000Z",
			"expires_at": "2026-01-24T15:30:00.000Z",
			"completed_at": "2026-01-24T14:35:00.000Z",
			"updated_at": "2026-01-24T14:35:00.000Z"
		}`)
	})

	session, _, err := c.Checkouts.Get(context.Background(), "chk_1a2b3c4d5e6f")
	if err != nil {
		t.Fatalf("Get returned error: %v", err)
	}

	if session.Status != "COMPLETED" {
		t.Errorf("Status = %q, want COMPLETED", session.Status)
	}
	if session.Amount != "50.00" || session.Currency != "USD" {
		t.Errorf("Amount/Currency = %s/%s, want 50.00/USD", session.Amount, session.Currency)
	}
	if session.Recurring != nil {
		t.Errorf("Recurring = %+v, want nil for a one-time checkout", session.Recurring)
	}
	if session.Charge == nil {
		t.Fatal("Charge is nil, want the linked payment")
	}
	if session.Charge.PaymentID != "pay_1a2b3c4d5e" || session.Charge.FeeUSD == nil || *session.Charge.FeeUSD != "0.59" {
		t.Errorf("Charge = %+v", session.Charge)
	}
	if session.Customer == nil || session.Customer.Email != "jane@example.com" {
		t.Errorf("Customer = %+v", session.Customer)
	}
	if len(session.Products) != 1 {
		t.Fatalf("len(Products) = %d, want 1", len(session.Products))
	}
	if session.Products[0].ProductName != "Premium Plan" || session.Products[0].LineTotal != "50.00" {
		t.Errorf("Products[0] = %+v", session.Products[0])
	}
	if session.SessionMode == nil || *session.SessionMode != "CART" {
		t.Errorf("SessionMode = %v, want CART", session.SessionMode)
	}
	if session.Metadata["order_id"] != "ORD-9876" {
		t.Errorf("Metadata = %v", session.Metadata)
	}
	if session.CompletedAt == nil {
		t.Error("CompletedAt is nil, want the completion timestamp")
	}
}

func TestCheckoutGetSubscriptionRecurring(t *testing.T) {
	c := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		io.WriteString(w, `{
			"checkout_id": "chk_sub_1",
			"status": "OPEN",
			"recurring": {"interval": "month", "interval_count": 1},
			"amount": "10.00",
			"currency": "USD",
			"created_at": "2026-01-24T14:30:00Z",
			"updated_at": "2026-01-24T14:30:00Z"
		}`)
	})

	session, _, err := c.Checkouts.Get(context.Background(), "chk_sub_1")
	if err != nil {
		t.Fatalf("Get returned error: %v", err)
	}
	if session.Recurring == nil {
		t.Fatal("Recurring is nil, want the cadence for a subscription checkout")
	}
	if session.Recurring.Interval != "month" || session.Recurring.IntervalCount != 1 {
		t.Errorf("Recurring = %+v", session.Recurring)
	}
}

func TestCheckoutGetEscapesID(t *testing.T) {
	c := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		if !strings.Contains(r.URL.EscapedPath(), "chk%2Fodd") {
			t.Errorf("escaped path = %q, want the ID percent-escaped", r.URL.EscapedPath())
		}
		io.WriteString(w, `{"checkout_id":"x","status":"OPEN","amount":"1.00","currency":"USD","created_at":"2026-01-24T14:30:00Z","updated_at":"2026-01-24T14:30:00Z"}`)
	})
	if _, _, err := c.Checkouts.Get(context.Background(), "chk/odd"); err != nil {
		t.Fatalf("Get returned error: %v", err)
	}
}

// TestCheckoutCreateGuestOmitsCustomer verifies the Sep 14, 2026 guest
// checkout: omitting Customer leaves the field out of the request body so the
// hosted page collects the buyer's identity instead.
func TestCheckoutCreateGuestOmitsCustomer(t *testing.T) {
	c := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		var got map[string]any
		if err := json.Unmarshal(body, &got); err != nil {
			t.Fatalf("request body is not valid JSON: %v", err)
		}
		if _, present := got["customer"]; present {
			t.Errorf("customer should be absent for a guest checkout: %s", body)
		}
		if got["customer_creation"] != "always" {
			t.Errorf("customer_creation = %v, want always", got["customer_creation"])
		}
		io.WriteString(w, `{"checkout_id":"chk_guest","checkout_url":"https://checkout.bachs.io/c/chk_guest","status":"OPEN","expires_at":"2026-01-24T15:30:00Z","created_at":"2026-01-24T14:30:00Z"}`)
	})

	req := CreateCheckoutSessionRequest{
		CustomerCreation: CustomerCreationAlways,
		ProductCart:      []ProductItemRequest{{ProductID: "prod_abc123"}},
	}
	session, _, err := c.Checkouts.Create(context.Background(), req)
	if err != nil {
		t.Fatalf("Create returned error: %v", err)
	}
	if session.CheckoutID != "chk_guest" {
		t.Errorf("CheckoutID = %q, want chk_guest", session.CheckoutID)
	}
}

// TestCheckoutCreateSendsPaymentMethodOptions verifies the Aug 17, 2026
// payment_method_options restriction serializes as documented.
func TestCheckoutCreateSendsPaymentMethodOptions(t *testing.T) {
	c := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		var got map[string]any
		if err := json.Unmarshal(body, &got); err != nil {
			t.Fatalf("request body is not valid JSON: %v", err)
		}
		opts, ok := got["payment_method_options"].(map[string]any)
		if !ok {
			t.Fatalf("payment_method_options missing from body: %s", body)
		}
		card, ok := opts["card"].(map[string]any)
		if !ok {
			t.Fatalf("card restriction missing: %s", body)
		}
		currencies, ok := card["currencies"].([]any)
		if !ok || len(currencies) != 1 || currencies[0] != "USD" {
			t.Errorf("card.currencies = %v, want [USD]", card["currencies"])
		}
		if _, ok := opts["bank_transfer"].(map[string]any); !ok {
			t.Errorf("bank_transfer restriction missing: %s", body)
		}
		io.WriteString(w, `{"checkout_id":"chk_x","checkout_url":"https://checkout.bachs.io/c/chk_x","status":"OPEN","expires_at":"2026-01-24T15:30:00Z","created_at":"2026-01-24T14:30:00Z"}`)
	})

	req := CreateCheckoutSessionRequest{
		Customer: &CheckoutCustomer{Email: "buyer@example.com"},
		ProductCart: []ProductItemRequest{
			{ProductID: "prod_abc123"},
		},
		PaymentMethodOptions: map[string]PaymentMethodOption{
			"card":          {Currencies: []string{"USD"}},
			"bank_transfer": {},
		},
	}
	if _, _, err := c.Checkouts.Create(context.Background(), req); err != nil {
		t.Fatalf("Create returned error: %v", err)
	}
}

// TestCheckoutGetCustomerDetails verifies the Sep 14, 2026 customer_details
// field decodes on a checkout with no backing customer record.
func TestCheckoutGetCustomerDetails(t *testing.T) {
	c := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		io.WriteString(w, `{
			"checkout_id": "chk_guest_1",
			"status": "COMPLETED",
			"amount": "10.00",
			"currency": "USD",
			"customer": null,
			"customer_details": {"email": "guest@example.com", "name": "Guest Buyer"},
			"created_at": "2026-01-24T14:30:00Z",
			"updated_at": "2026-01-24T14:35:00Z"
		}`)
	})

	session, _, err := c.Checkouts.Get(context.Background(), "chk_guest_1")
	if err != nil {
		t.Fatalf("Get returned error: %v", err)
	}
	if session.Customer != nil {
		t.Errorf("Customer = %+v, want nil for a guest checkout", session.Customer)
	}
	if session.CustomerDetails == nil {
		t.Fatal("CustomerDetails is nil, want the buyer identity")
	}
	if session.CustomerDetails.Email != "guest@example.com" {
		t.Errorf("CustomerDetails.Email = %q", session.CustomerDetails.Email)
	}
	if session.CustomerDetails.Name == nil || *session.CustomerDetails.Name != "Guest Buyer" {
		t.Errorf("CustomerDetails.Name = %v", session.CustomerDetails.Name)
	}
}

// TestCheckoutCreateSendsConnectSplit verifies the Connect split fields from
// https://docs.bachs.io/connect/platform-fees: a fee-first checkout sends
// platform_fee, a share-first one sends transfer_data.
func TestCheckoutCreateSendsConnectSplit(t *testing.T) {
	for _, tt := range []struct {
		name  string
		req   CreateCheckoutSessionRequest
		check func(t *testing.T, body map[string]any)
	}{
		{
			name: "fee-first",
			req: CreateCheckoutSessionRequest{
				Customer:    &CheckoutCustomer{Email: "jane@example.com"},
				ProductCart: []ProductItemRequest{{ProductID: "prod_abc123"}},
				SuccessURL:  "https://example.com/thank-you",
				PlatformFee: "20000.00",
			},
			check: func(t *testing.T, body map[string]any) {
				if body["platform_fee"] != "20000.00" {
					t.Errorf("platform_fee = %v, want 20000.00", body["platform_fee"])
				}
				if _, present := body["transfer_data"]; present {
					t.Errorf("transfer_data should be absent on a fee-first split: %v", body)
				}
			},
		},
		{
			name: "share-first",
			req: CreateCheckoutSessionRequest{
				Customer:    &CheckoutCustomer{Email: "jane@example.com"},
				ProductCart: []ProductItemRequest{{ProductID: "prod_abc123"}},
				SuccessURL:  "https://example.com/thank-you",
				TransferData: &CheckoutTransferData{
					Destination: "acct_3Wq8ZfT1yHnJ5sVe",
					Amount:      "80000.00",
				},
			},
			check: func(t *testing.T, body map[string]any) {
				td, ok := body["transfer_data"].(map[string]any)
				if !ok {
					t.Fatalf("transfer_data missing: %v", body)
				}
				if td["destination"] != "acct_3Wq8ZfT1yHnJ5sVe" || td["amount"] != "80000.00" {
					t.Errorf("transfer_data = %v", td)
				}
				if _, present := body["platform_fee"]; present {
					t.Errorf("platform_fee should be absent on a share-first split: %v", body)
				}
			},
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			c := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
				body, _ := io.ReadAll(r.Body)
				var got map[string]any
				if err := json.Unmarshal(body, &got); err != nil {
					t.Fatalf("request body is not valid JSON: %v", err)
				}
				tt.check(t, got)
				io.WriteString(w, `{"checkout_id":"chk_x","checkout_url":"https://checkout.bachs.io/c/chk_x","status":"OPEN","expires_at":"2026-01-24T15:30:00Z","created_at":"2026-01-24T14:30:00Z"}`)
			})
			if _, _, err := c.Checkouts.Create(context.Background(), tt.req); err != nil {
				t.Fatalf("Create returned error: %v", err)
			}
		})
	}
}

// TestCheckoutCreateSendsPaymentMethodTypes verifies the corridor allowlist
// from https://docs.bachs.io/guides/checkout/checkout-sessions#restrict-payment-methods.
func TestCheckoutCreateSendsPaymentMethodTypes(t *testing.T) {
	c := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		var got map[string]any
		if err := json.Unmarshal(body, &got); err != nil {
			t.Fatalf("request body is not valid JSON: %v", err)
		}
		types, ok := got["payment_method_types"].([]any)
		if !ok || len(types) != 2 || types[0] != CorridorUSDCard || types[1] != CorridorNGNBankTransfer {
			t.Errorf("payment_method_types = %v, want [USD_CARD NGN_BANK_TRANSFER]", got["payment_method_types"])
		}
		io.WriteString(w, `{"checkout_id":"chk_x","checkout_url":"https://checkout.bachs.io/c/chk_x","status":"OPEN","expires_at":"2026-01-24T15:30:00Z","created_at":"2026-01-24T14:30:00Z"}`)
	})

	req := CreateCheckoutSessionRequest{
		Customer:           &CheckoutCustomer{Email: "jane@example.com"},
		ProductCart:        []ProductItemRequest{{ProductID: "prod_abc123"}},
		PaymentMethodTypes: []string{CorridorUSDCard, CorridorNGNBankTransfer},
	}
	if _, _, err := c.Checkouts.Create(context.Background(), req); err != nil {
		t.Fatalf("Create returned error: %v", err)
	}
}
