package bachs

import (
	"context"
	"io"
	"net/http"
	"testing"
)

// TestPaymentGet uses the exact example payload from
// https://docs.bachs.io/api-reference/payments/get-payment
func TestPaymentGet(t *testing.T) {
	c := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Errorf("method = %s, want GET", r.Method)
		}
		if r.URL.Path != "/v1/payments/chr_1a2b3c4d5e6f" {
			t.Errorf("path = %q, want /v1/payments/chr_1a2b3c4d5e6f", r.URL.Path)
		}
		io.WriteString(w, `{
			"reference": "ord_12903",
			"payment_id": "chr_1a2b3c4d5e6f",
			"checkout_id": "chk_1a2b3c4d5e6f",
			"status": "succeeded",
			"is_refundable": true,
			"amount": "75000.00",
			"amount_paid": "75000.00",
			"amount_remaining": "0.00",
			"currency": "NGN",
			"fee_usd": "0.59",
			"merchant_bears_cost": false,
			"payment_method": "BANK_TRANSFER",
			"channel": "api",
			"narration": "Order payment ORD-12903",
			"meta": {"order_id": "ORD-12903"},
			"message": "Successful",
			"customer": {
				"name": "Jane Doe",
				"email": "jane@example.com"
			},
			"created_at": "2026-02-22T12:00:00.000Z",
			"updated_at": "2026-02-22T12:01:30.000Z",
			"completed_at": "2026-02-22T12:01:30.000Z"
		}`)
	})

	payment, _, err := c.Payments.Get(context.Background(), "chr_1a2b3c4d5e6f")
	if err != nil {
		t.Fatalf("Get returned error: %v", err)
	}

	if payment.PaymentID != "chr_1a2b3c4d5e6f" {
		t.Errorf("PaymentID = %q", payment.PaymentID)
	}
	if payment.Status != "succeeded" {
		t.Errorf("Status = %q, want succeeded", payment.Status)
	}
	if payment.Amount != "75000.00" || payment.Currency != "NGN" {
		t.Errorf("Amount/Currency = %s/%s", payment.Amount, payment.Currency)
	}
	if payment.IsRefundable == nil || !*payment.IsRefundable {
		t.Errorf("IsRefundable = %v, want true", payment.IsRefundable)
	}
	if payment.FeeUSD == nil || *payment.FeeUSD != "0.59" {
		t.Errorf("FeeUSD = %v, want 0.59", payment.FeeUSD)
	}
	if payment.Customer == nil || payment.Customer.Name == nil || *payment.Customer.Name != "Jane Doe" {
		t.Errorf("Customer = %+v", payment.Customer)
	}
	if payment.Meta["order_id"] != "ORD-12903" {
		t.Errorf("Meta = %v", payment.Meta)
	}
	if payment.CompletedAt == nil {
		t.Error("CompletedAt is nil")
	}
}

// TestPaymentList uses the exact example payload from
// https://docs.bachs.io/api-reference/payments/list-payments
func TestPaymentList(t *testing.T) {
	c := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Errorf("method = %s, want GET", r.Method)
		}
		if r.URL.Path != "/v1/payments" {
			t.Errorf("path = %q, want /v1/payments", r.URL.Path)
		}
		if got := r.URL.Query().Get("status_filter"); got != "succeeded" {
			t.Errorf("status_filter = %q, want succeeded", got)
		}
		io.WriteString(w, `{
			"items": [
				{
					"id": "chrg_1a2b3c4d5e",
					"reference": "order_9876",
					"status": "succeeded",
					"is_refundable": true,
					"amount": "10.00",
					"customer_name": "Jane Doe",
					"customer_email": "customer@example.com",
					"amount_paid": "10.00",
					"amount_remaining": "0.00",
					"settlement_amount": "10.00",
					"fee": null,
					"vat": null,
					"currency": "USD",
					"settlement_currency": "USD",
					"meta": null,
					"transaction_date": "2026-04-27T12:00:00Z",
					"completed_at": "2026-04-27T12:00:05Z"
				}
			],
			"pagination": {
				"next_cursor": null,
				"prev_cursor": null,
				"has_more": false,
				"limit": 50,
				"offset": 0,
				"returned": 1,
				"total": 1
			}
		}`)
	})

	page, _, err := c.Payments.List(context.Background(), ListParams{StatusFilter: "succeeded"})
	if err != nil {
		t.Fatalf("List returned error: %v", err)
	}
	if len(page.Items) != 1 {
		t.Fatalf("len(Items) = %d, want 1", len(page.Items))
	}
	item := page.Items[0]
	if item.ID == nil || *item.ID != "chrg_1a2b3c4d5e" {
		t.Errorf("ID = %v", item.ID)
	}
	if item.Amount != "10.00" || item.Currency != "USD" {
		t.Errorf("Amount/Currency = %s/%s", item.Amount, item.Currency)
	}
	if item.CustomerName != "Jane Doe" || item.CustomerEmail != "customer@example.com" {
		t.Errorf("CustomerName/CustomerEmail = %s/%s", item.CustomerName, item.CustomerEmail)
	}
	if page.Pagination.Limit != 50 || page.Pagination.HasMore {
		t.Errorf("Pagination = %+v", page.Pagination)
	}
}

// TestPaymentGetMethodDetails verifies the Sep 21, 2026
// payment_method_details block decodes on a virtual account deposit: a bank
// transfer with sender details and the receiving virtual account.
func TestPaymentGetMethodDetails(t *testing.T) {
	c := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		io.WriteString(w, `{
			"reference": null,
			"payment_id": "ch_1a2b3c4d5e6f",
			"billing_reason": "purchase",
			"checkout_id": null,
			"status": "succeeded",
			"is_refundable": false,
			"amount": "250000.00",
			"currency": "NGN",
			"payment_method": "NGN_BANK_TRANSFER",
			"payment_method_details": {
				"type": "bank_transfer",
				"bank_transfer": {
					"sender_name": "JANE ADEYEMI",
					"sender_bank": "Guaranty Trust Bank",
					"sender_account_number": "2294879124",
					"session_id": "000013260922104115000821734502",
					"virtual_account": {
						"id": "va_8Hs2kQ4mZpXv",
						"account_number": "9902847361",
						"bank_name": "Example Bank",
						"type": "permanent",
						"expires_at": null
					}
				}
			},
			"created_at": "2026-09-22T10:41:18.000Z",
			"updated_at": "2026-09-22T10:41:18.000Z"
		}`)
	})

	payment, _, err := c.Payments.Get(context.Background(), "ch_1a2b3c4d5e6f")
	if err != nil {
		t.Fatalf("Get returned error: %v", err)
	}
	if payment.CheckoutID != nil {
		t.Errorf("CheckoutID = %v, want nil for a virtual account deposit", payment.CheckoutID)
	}
	if payment.IsRefundable == nil || *payment.IsRefundable {
		t.Errorf("IsRefundable = %v, want false for a deposit", payment.IsRefundable)
	}
	details := payment.PaymentMethodDetails
	if details == nil {
		t.Fatal("PaymentMethodDetails is nil, want the bank transfer details")
	}
	if details.Type != "bank_transfer" {
		t.Errorf("Type = %q, want bank_transfer", details.Type)
	}
	bt := details.BankTransfer
	if bt == nil {
		t.Fatal("BankTransfer is nil")
	}
	if bt.SenderName == nil || *bt.SenderName != "JANE ADEYEMI" {
		t.Errorf("SenderName = %v", bt.SenderName)
	}
	if bt.SessionID == nil || *bt.SessionID != "000013260922104115000821734502" {
		t.Errorf("SessionID = %v", bt.SessionID)
	}
	if bt.VirtualAccount == nil || bt.VirtualAccount.AccountNumber != "9902847361" {
		t.Errorf("VirtualAccount = %+v", bt.VirtualAccount)
	}
}
