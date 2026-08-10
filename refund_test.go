package bachs

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"testing"
)

// refundExample is the RefundResponse example shape from the refund doc pages.
const refundExample = `{
	"refund_id": "ref_1a2b3c4d5e6f7g8h",
	"charge_id": "chr_8f3a1c9b4e72",
	"reference": "rf_20260309_001",
	"status": "processing",
	"requested_amount": "25.00",
	"refunded_amount": null,
	"refund_fee_amount": "0",
	"fee_bearer": "org",
	"reason": "Customer requested",
	"created_at": "2026-03-09T10:30:00.000Z",
	"updated_at": "2026-03-09T10:30:00.000Z",
	"completed_at": null
}`

// TestRefundCreate uses the exact request example from
// https://docs.bachs.io/api-reference/refunds/create-refund
func TestRefundCreate(t *testing.T) {
	c := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("method = %s, want POST", r.Method)
		}
		if r.URL.Path != "/v1/refunds" {
			t.Errorf("path = %q, want /v1/refunds", r.URL.Path)
		}
		body, _ := io.ReadAll(r.Body)
		var got map[string]any
		if err := json.Unmarshal(body, &got); err != nil {
			t.Fatalf("request body is not valid JSON: %v", err)
		}
		if got["charge_id"] != "chr_8f3a1c9b4e72" || got["reference"] != "rf_20260309_001" {
			t.Errorf("request body = %s", body)
		}
		io.WriteString(w, refundExample)
	})

	refund, _, err := c.Refunds.Create(context.Background(), CreateRefundRequest{
		ChargeID:  "chr_8f3a1c9b4e72",
		Reference: "rf_20260309_001",
	})
	if err != nil {
		t.Fatalf("Create returned error: %v", err)
	}
	if refund.RefundID != "ref_1a2b3c4d5e6f7g8h" {
		t.Errorf("RefundID = %q", refund.RefundID)
	}
	if refund.Status != "processing" {
		t.Errorf("Status = %q, want processing", refund.Status)
	}
	if refund.RequestedAmount != "25.00" || refund.RefundFeeAmount != "0" {
		t.Errorf("RequestedAmount/RefundFeeAmount = %s/%s", refund.RequestedAmount, refund.RefundFeeAmount)
	}
	if refund.RefundedAmount != nil {
		t.Errorf("RefundedAmount = %v, want nil while processing", refund.RefundedAmount)
	}
	if refund.CompletedAt != nil {
		t.Errorf("CompletedAt = %v, want nil while processing", refund.CompletedAt)
	}
}

func TestRefundCreateSendsFullRequest(t *testing.T) {
	c := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		var got map[string]any
		if err := json.Unmarshal(body, &got); err != nil {
			t.Fatalf("request body is not valid JSON: %v", err)
		}
		if got["amount"] != "10.00" || got["fee_bearer"] != "org" || got["simulated_outcome"] != "success" {
			t.Errorf("request body = %s", body)
		}
		io.WriteString(w, refundExample)
	})

	amount := "10.00"
	feeBearer := "org"
	outcome := "success"
	if _, _, err := c.Refunds.Create(context.Background(), CreateRefundRequest{
		ChargeID:         "chr_1",
		Reference:        "ref_1",
		Amount:           &amount,
		FeeBearer:        &feeBearer,
		SimulatedOutcome: &outcome,
	}); err != nil {
		t.Fatalf("Create returned error: %v", err)
	}
}

func TestRefundGet(t *testing.T) {
	c := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Errorf("method = %s, want GET", r.Method)
		}
		if r.URL.Path != "/v1/refunds/ref_1a2b3c4d5e6f" {
			t.Errorf("path = %q, want /v1/refunds/ref_1a2b3c4d5e6f", r.URL.Path)
		}
		io.WriteString(w, refundExample)
	})

	refund, _, err := c.Refunds.Get(context.Background(), "ref_1a2b3c4d5e6f")
	if err != nil {
		t.Fatalf("Get returned error: %v", err)
	}
	if refund.ChargeID != "chr_8f3a1c9b4e72" {
		t.Errorf("ChargeID = %q", refund.ChargeID)
	}
}

func TestRefundGetByCharge(t *testing.T) {
	c := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Errorf("method = %s, want GET", r.Method)
		}
		if r.URL.Path != "/v1/refunds/by-charge/chr_8f3a1c9b4e72" {
			t.Errorf("path = %q, want /v1/refunds/by-charge/chr_8f3a1c9b4e72", r.URL.Path)
		}
		io.WriteString(w, refundExample)
	})

	refund, _, err := c.Refunds.GetByCharge(context.Background(), "chr_8f3a1c9b4e72")
	if err != nil {
		t.Fatalf("GetByCharge returned error: %v", err)
	}
	if refund.RefundID != "ref_1a2b3c4d5e6f7g8h" {
		t.Errorf("RefundID = %q", refund.RefundID)
	}
}

// TestRefundList verifies the flat { items, total } list shape folds into
// Page[Refund].
func TestRefundList(t *testing.T) {
	c := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Errorf("method = %s, want GET", r.Method)
		}
		if r.URL.Path != "/v1/refunds" {
			t.Errorf("path = %q, want /v1/refunds", r.URL.Path)
		}
		if got := r.URL.Query().Get("status"); got != "SUCCESS" {
			t.Errorf("status = %q, want SUCCESS", got)
		}
		io.WriteString(w, `{
			"total": 3,
			"items": [`+refundExample+`]
		}`)
	})

	page, _, err := c.Refunds.List(context.Background(), ListParams{Status: "SUCCESS"})
	if err != nil {
		t.Fatalf("List returned error: %v", err)
	}
	if len(page.Items) != 1 {
		t.Fatalf("len(Items) = %d, want 1", len(page.Items))
	}
	if page.Items[0].Reference != "rf_20260309_001" {
		t.Errorf("Items[0].Reference = %q", page.Items[0].Reference)
	}
	if page.Pagination.Total != 3 {
		t.Errorf("Pagination.Total = %d, want 3 (folded from the flat total)", page.Pagination.Total)
	}
}
