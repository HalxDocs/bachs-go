package bachs

import (
	"context"
	"io"
	"net/http"
	"testing"
)

const platformFeeExample = `{
	"id": "pf_8c1e04a7b93f2d6540ab1234",
	"charge": "ch_2f8a71c4e05b",
	"collected_from": "acct_7KpQ2mNv4XbR9dLc",
	"earned_by": "acct_3Wq8ZfT1yHnJ5sVe",
	"amount": "20000.00",
	"currency": "NGN",
	"amount_refunded": "0.00",
	"refunded": false,
	"created_at": "2026-08-12T10:31:00.000Z"
}`

// TestPlatformFeesList uses the example payload from
// https://docs.bachs.io/connect/platform-fees.
func TestPlatformFeesList(t *testing.T) {
	c := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Errorf("method = %s, want GET", r.Method)
		}
		if got := r.URL.RequestURI(); got != "/v1/platform_fees?charge=ch_2f8a71c4e05b" {
			t.Errorf("request URI = %q, want the charge filter", got)
		}
		io.WriteString(w, `{
			"items": [`+platformFeeExample+`],
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

	page, _, err := c.PlatformFees.List(context.Background(), ListParams{Charge: "ch_2f8a71c4e05b"})
	if err != nil {
		t.Fatalf("List returned error: %v", err)
	}
	if len(page.Items) != 1 {
		t.Fatalf("len(Items) = %d, want 1", len(page.Items))
	}
	fee := page.Items[0]
	if fee.ID != "pf_8c1e04a7b93f2d6540ab1234" {
		t.Errorf("ID = %q", fee.ID)
	}
	if fee.Amount != "20000.00" || fee.Currency != "NGN" {
		t.Errorf("Amount/Currency = %s/%s", fee.Amount, fee.Currency)
	}
	if fee.CollectedFrom != "acct_7KpQ2mNv4XbR9dLc" || fee.EarnedBy != "acct_3Wq8ZfT1yHnJ5sVe" {
		t.Errorf("parties = %s/%s", fee.CollectedFrom, fee.EarnedBy)
	}
	if fee.Refunded {
		t.Error("Refunded is true, want false")
	}
	if page.Pagination.Total != 1 {
		t.Errorf("Pagination.Total = %d, want 1", page.Pagination.Total)
	}
}

func TestPlatformFeesGet(t *testing.T) {
	c := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v1/platform_fees/pf_8c1e04a7b93f2d6540ab1234" {
			t.Errorf("path = %q", r.URL.Path)
		}
		io.WriteString(w, platformFeeExample)
	})

	fee, _, err := c.PlatformFees.Get(context.Background(), "pf_8c1e04a7b93f2d6540ab1234")
	if err != nil {
		t.Fatalf("Get returned error: %v", err)
	}
	if fee.Charge != "ch_2f8a71c4e05b" {
		t.Errorf("Charge = %q", fee.Charge)
	}
	if fee.CreatedAt.IsZero() {
		t.Error("CreatedAt is zero")
	}
}
