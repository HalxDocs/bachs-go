package bachs

import (
	"context"
	"io"
	"net/http"
	"testing"
)

const conversionExample = `{
	"conversion_id": "cvt_1a2b3c4d5e6f",
	"status": "completed",
	"from_currency": "USD",
	"to_currency": "NGN",
	"from_amount": "1000.00",
	"to_amount": "1500000.00",
	"exchange_rate": "1500.00",
	"created_at": "2026-01-24T14:30:00.000Z"
}`

func TestCreateConversionQuote(t *testing.T) {
	c := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("method = %s, want POST", r.Method)
		}
		if r.URL.Path != "/v1/conversions/quotes" {
			t.Errorf("path = %q", r.URL.Path)
		}
		io.WriteString(w, `{
			"quote_id": "cqt_1a2b3c4d5e6f",
			"from_currency": "USD",
			"to_currency": "NGN",
			"from_amount": "1000.00",
			"to_amount": "1500000.00",
			"exchange_rate": "1500.00",
			"expires_at": "2026-01-24T14:31:00.000Z"
		}`)
	})

	quote, _, err := c.Conversions.CreateQuote(context.Background(), CreateConversionQuoteRequest{
		FromCurrency: "USD",
		ToCurrency:   "NGN",
		Amount:       "1000.00",
	})
	if err != nil {
		t.Fatalf("CreateQuote returned error: %v", err)
	}
	if quote.QuoteID != "cqt_1a2b3c4d5e6f" || quote.FromAmount != "1000.00" {
		t.Errorf("quote = %+v", quote)
	}
	if quote.ToAmount != "1500000.00" || quote.ExchangeRate != "1500.00" {
		t.Errorf("quote = %+v", quote)
	}
	if quote.ExpiresAt.IsZero() {
		t.Error("ExpiresAt is zero")
	}
}

func TestListConversions(t *testing.T) {
	c := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Errorf("method = %s, want GET", r.Method)
		}
		if got := r.URL.RequestURI(); got != "/v1/conversions?from_currency=USD&status=completed&to_currency=NGN" {
			t.Errorf("path = %q", got)
		}
		io.WriteString(w, `{
			"total": 2,
			"limit": 20,
			"offset": 0,
			"items": [`+conversionExample+`]
		}`)
	})

	page, _, err := c.Conversions.List(context.Background(), ListParams{
		FromCurrency: "USD",
		ToCurrency:   "NGN",
		Status:       "completed",
	})
	if err != nil {
		t.Fatalf("List returned error: %v", err)
	}
	if len(page.Items) != 1 {
		t.Fatalf("len(Items) = %d, want 1", len(page.Items))
	}
	cv := page.Items[0]
	if cv.ConversionID != "cvt_1a2b3c4d5e6f" || cv.Status != "completed" {
		t.Errorf("Items[0] = %+v", cv)
	}
	if cv.FromAmount != "1000.00" || cv.ToAmount != "1500000.00" {
		t.Errorf("amounts = %s / %s", cv.FromAmount, cv.ToAmount)
	}
	if page.Pagination.Total != 2 {
		t.Errorf("Pagination.Total = %d, want 2", page.Pagination.Total)
	}
	if page.Pagination.Limit != 20 {
		t.Errorf("Pagination.Limit = %d, want 20", page.Pagination.Limit)
	}
}

func TestCreateConversion(t *testing.T) {
	c := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("method = %s, want POST", r.Method)
		}
		if r.URL.Path != "/v1/conversions" {
			t.Errorf("path = %q", r.URL.Path)
		}
		io.WriteString(w, conversionExample)
	})

	cv, _, err := c.Conversions.Create(context.Background(), CreateConversionRequest{
		FromCurrency: "USD",
		ToCurrency:   "NGN",
		Amount:       "1000.00",
		QuoteID:      "cqt_1a2b3c4d5e6f",
	})
	if err != nil {
		t.Fatalf("Create returned error: %v", err)
	}
	if cv.ConversionID != "cvt_1a2b3c4d5e6f" || cv.Status != "completed" {
		t.Errorf("conversion = %+v", cv)
	}
	if cv.ExchangeRate != "1500.00" {
		t.Errorf("ExchangeRate = %q", cv.ExchangeRate)
	}
}

func TestGetConversion(t *testing.T) {
	c := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Errorf("method = %s, want GET", r.Method)
		}
		if r.URL.Path != "/v1/conversions/cvt_1a2b3c4d5e6f" {
			t.Errorf("path = %q", r.URL.Path)
		}
		io.WriteString(w, `{
			"conversion_id": "cvt_1a2b3c4d5e6f",
			"status": "completed",
			"from_currency": "USD",
			"to_currency": "NGN",
			"from_amount": "1000.00",
			"to_amount": "1500000.00",
			"exchange_rate": "1500.00",
			"created_at": "2026-01-24T14:30:00.000Z",
			"quote_id": "cqt_1a2b3c4d5e6f",
			"metadata": null
		}`)
	})

	cv, _, err := c.Conversions.Get(context.Background(), "cvt_1a2b3c4d5e6f")
	if err != nil {
		t.Fatalf("Get returned error: %v", err)
	}
	if cv.ConversionID != "cvt_1a2b3c4d5e6f" {
		t.Errorf("ConversionID = %q", cv.ConversionID)
	}
	if cv.QuoteID == nil || *cv.QuoteID != "cqt_1a2b3c4d5e6f" {
		t.Errorf("QuoteID = %v", cv.QuoteID)
	}
}
