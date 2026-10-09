package webhook

import (
	"encoding/json"
	"testing"
	"time"
)

// TestCollectionSucceededData decodes the documented
// collection.succeeded payload from
// https://docs.bachs.io/guides/webhooks/events/collection-succeeded.
func TestCollectionSucceededData(t *testing.T) {
	var event Event
	if err := json.Unmarshal([]byte(`{
		"id": "evt_3ab4e0d5d27445cf8a52ab3d8cb8f0b1",
		"type": "collection.succeeded",
		"created_at": "2026-09-22T10:41:18.402Z",
		"organization_id": "acct_7KpQ2mNv4XbR9dLc",
		"data": {
			"charge_id": "ch_1a2b3c4d5e6f",
			"checkout_id": "chk_6R7s8T9u0V1w2X3y",
			"reference": "ORD-20260922-1041",
			"status": "SUCCEEDED",
			"amount": "75000.00",
			"currency": "NGN",
			"settlement_amount": "73875.00",
			"settlement_currency": "NGN",
			"processing_fee": "1125.00",
			"processing_fee_currency": "NGN",
			"fee_bearer": "merchant",
			"product_cart": [
				{"product_id": "prod_1a2b3c", "quantity": 2},
				{"product_id": "prod_9x8y7z", "quantity": 1, "amount": "5000.00"}
			],
			"customer": {"id": "cust_xyz789", "email": "jane@example.com", "name": "Jane Doe"},
			"payment_method_details": {
				"type": "bank_transfer",
				"bank_transfer": {
					"sender_name": "JANE ADEYEMI",
					"sender_bank": "Guaranty Trust Bank",
					"sender_bank_code": "058",
					"sender_account_number": "2294879124",
					"session_id": "000013260922104115000821734502",
					"narration": "ORD-20260922-1041",
					"virtual_account": {
						"id": null,
						"account_number": "9902847361",
						"bank_name": "Example Bank",
						"type": "one_time",
						"expires_at": "2026-09-22T11:11:18.402Z"
					}
				}
			},
			"metadata": {"order_id": "ORD-20260922-1041"}
		}
	}`), &event); err != nil {
		t.Fatalf("unmarshal event: %v", err)
	}

	var data CollectionSucceededData
	if err := event.DataAs(&data); err != nil {
		t.Fatalf("DataAs: %v", err)
	}
	if data.ChargeID == nil || *data.ChargeID != "ch_1a2b3c4d5e6f" {
		t.Errorf("ChargeID = %v", data.ChargeID)
	}
	if data.Status != "SUCCEEDED" {
		t.Errorf("Status = %q", data.Status)
	}
	if len(data.ProductCart) != 2 || data.ProductCart[0].Quantity != 2 {
		t.Errorf("ProductCart = %+v", data.ProductCart)
	}
	if data.ProductCart[1].Amount == nil || *data.ProductCart[1].Amount != "5000.00" {
		t.Errorf("ProductCart[1].Amount = %v", data.ProductCart[1].Amount)
	}
	if data.Customer == nil || data.Customer.ID == nil || *data.Customer.ID != "cust_xyz789" {
		t.Errorf("Customer = %+v", data.Customer)
	}
	bt := data.PaymentMethodDetails.BankTransfer
	if bt == nil {
		t.Fatal("BankTransfer is nil")
	}
	if bt.SenderBankCode == nil || *bt.SenderBankCode != "058" {
		t.Errorf("SenderBankCode = %v", bt.SenderBankCode)
	}
	if bt.VirtualAccount == nil || bt.VirtualAccount.ExpiresAt == nil {
		t.Fatalf("VirtualAccount = %+v", bt.VirtualAccount)
	}
	if bt.VirtualAccount.ID != nil {
		t.Errorf("VirtualAccount.ID = %v, want nil for a one-time number", bt.VirtualAccount.ID)
	}
	if data.Metadata["order_id"] != "ORD-20260922-1041" {
		t.Errorf("Metadata = %v", data.Metadata)
	}
}

// TestCustomerSubscriptionData decodes the documented
// customer.subscription.created payload from
// https://docs.bachs.io/guides/webhooks/events/customer-subscription-created.
func TestCustomerSubscriptionData(t *testing.T) {
	var event Event
	if err := json.Unmarshal([]byte(`{
		"id": "evt_1a2b3c4d5e6f7g8h",
		"type": "customer.subscription.created",
		"created_at": "2026-04-27T12:00:00.000000+00:00",
		"organization_id": "acct_7KpQ2mNv4XbR9dLc",
		"data": {
			"subscription_id": "sub_1a2b3c4d5e",
			"customer": {
				"customer_id": "cust_1a2b3c4d5e6f",
				"email": "jane@example.com",
				"name": "Jane Doe",
				"phone_number": "+2348012345678",
				"metadata": {},
				"billing_address": {
					"line1": "40 Yaba Road",
					"line2": null,
					"city": "Lagos",
					"state": "Lagos",
					"postal_code": "101245",
					"country": "NG"
				},
				"created_at": "2026-03-01T12:00:00Z",
				"updated_at": "2026-03-01T12:00:00Z"
			},
			"product_id": "prod_abc123",
			"status": "active",
			"collection_method": "charge_automatically",
			"currency": "USD",
			"amount": "10.00",
			"billing_cycle": {"interval": "month", "frequency": 1},
			"quantity": 1,
			"current_period_start": "2026-04-01T00:00:00Z",
			"current_period_end": "2026-05-01T00:00:00Z",
			"next_billed_at": "2026-05-01T00:00:00Z",
			"trial_end": null,
			"cancel_at_period_end": false,
			"canceled_at": null,
			"created_at": "2026-03-01T12:00:00Z",
			"items": [],
			"metadata": {}
		}
	}`), &event); err != nil {
		t.Fatalf("unmarshal event: %v", err)
	}

	if !event.CreatedAt.Equal(time.Date(2026, 4, 27, 12, 0, 0, 0, time.UTC)) {
		t.Errorf("CreatedAt = %v (fractional + offset timestamp)", event.CreatedAt)
	}

	var data CustomerSubscriptionData
	if err := event.DataAs(&data); err != nil {
		t.Fatalf("DataAs: %v", err)
	}
	if data.SubscriptionID != "sub_1a2b3c4d5e" {
		t.Errorf("SubscriptionID = %q", data.SubscriptionID)
	}
	if data.Customer.CustomerID != "cust_1a2b3c4d5e6f" {
		t.Errorf("CustomerID = %q", data.Customer.CustomerID)
	}
	if data.Customer.BillingAddress == nil || data.Customer.BillingAddress.City == nil ||
		*data.Customer.BillingAddress.City != "Lagos" {
		t.Errorf("BillingAddress = %+v", data.Customer.BillingAddress)
	}
	if data.BillingCycle.Interval != "month" || data.BillingCycle.Frequency != 1 {
		t.Errorf("BillingCycle = %+v", data.BillingCycle)
	}
	if data.NextBilledAt == nil {
		t.Error("NextBilledAt is nil")
	}
	if data.TrialEnd != nil || data.CanceledAt != nil {
		t.Errorf("TrialEnd/CanceledAt = %v/%v, want nil", data.TrialEnd, data.CanceledAt)
	}
}
