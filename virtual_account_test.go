package bachs

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"testing"
)

// TestVirtualAccountCreate uses the example payload from
// https://docs.bachs.io/for-you/virtual-accounts.
func TestVirtualAccountCreate(t *testing.T) {
	c := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("method = %s, want POST", r.Method)
		}
		if r.URL.Path != "/v1/virtual-accounts" {
			t.Errorf("path = %q, want /v1/virtual-accounts", r.URL.Path)
		}

		body, _ := io.ReadAll(r.Body)
		var got map[string]any
		if err := json.Unmarshal(body, &got); err != nil {
			t.Fatalf("request body is not valid JSON: %v", err)
		}
		if got["currency"] != "NGN" {
			t.Errorf("currency = %v, want NGN", got["currency"])
		}

		io.WriteString(w, `{
			"id": "va_8Hs2kQ4mZpXv",
			"currency": "NGN",
			"account_number": "9902847361",
			"bank_name": "Example Bank",
			"bank_code": "000",
			"status": "active",
			"created_at": "2026-09-22T09:14:02.000Z"
		}`)
	})

	va, _, err := c.VirtualAccounts.Create(context.Background(), CreateVirtualAccountRequest{Currency: "NGN"})
	if err != nil {
		t.Fatalf("Create returned error: %v", err)
	}
	if va.ID != "va_8Hs2kQ4mZpXv" {
		t.Errorf("ID = %q, want va_8Hs2kQ4mZpXv", va.ID)
	}
	if va.Currency != "NGN" || va.AccountNumber != "9902847361" {
		t.Errorf("virtual account = %+v", va)
	}
	if va.BankName != "Example Bank" || va.BankCode != "000" || va.Status != "active" {
		t.Errorf("virtual account = %+v", va)
	}
	if va.CreatedAt.IsZero() {
		t.Error("CreatedAt is zero")
	}
}

func TestVirtualAccountCreateForConnectedAccount(t *testing.T) {
	c := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		if got := r.Header.Get(headerConnectedAccountID); got != "acct_3Wq8ZfT1yHnJ5sVe" {
			t.Errorf("connected account header = %q, want acct_3Wq8ZfT1yHnJ5sVe", got)
		}
		io.WriteString(w, `{
			"id": "va_8Hs2kQ4mZpXv",
			"currency": "NGN",
			"account_number": "9902847361",
			"bank_name": "Example Bank",
			"bank_code": "000",
			"status": "active",
			"created_at": "2026-09-22T09:14:02.000Z"
		}`)
	})

	va, _, err := c.VirtualAccounts.Create(
		context.Background(),
		CreateVirtualAccountRequest{Currency: "NGN"},
		WithConnectedAccount("acct_3Wq8ZfT1yHnJ5sVe"),
	)
	if err != nil {
		t.Fatalf("Create returned error: %v", err)
	}
	if va.ID != "va_8Hs2kQ4mZpXv" {
		t.Errorf("ID = %q, want va_8Hs2kQ4mZpXv", va.ID)
	}
}

func TestVirtualAccountGet(t *testing.T) {
	c := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Errorf("method = %s, want GET", r.Method)
		}
		if got := r.URL.RequestURI(); got != "/v1/virtual-accounts?currency=NGN" {
			t.Errorf("request URI = %q, want /v1/virtual-accounts?currency=NGN", got)
		}
		io.WriteString(w, `{
			"id": "va_8Hs2kQ4mZpXv",
			"currency": "NGN",
			"account_number": "9902847361",
			"bank_name": "Example Bank",
			"bank_code": "000",
			"status": "active",
			"created_at": "2026-09-22T09:14:02.000Z"
		}`)
	})

	va, _, err := c.VirtualAccounts.Get(context.Background(), "NGN")
	if err != nil {
		t.Fatalf("Get returned error: %v", err)
	}
	if va.AccountNumber != "9902847361" || va.Currency != "NGN" {
		t.Errorf("virtual account = %+v", va)
	}
}

func TestVirtualAccountGetDefaultsCurrency(t *testing.T) {
	c := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		if got := r.URL.RequestURI(); got != "/v1/virtual-accounts" {
			t.Errorf("request URI = %q, want /v1/virtual-accounts", got)
		}
		io.WriteString(w, `{
			"id": "va_8Hs2kQ4mZpXv",
			"currency": "NGN",
			"account_number": "9902847361",
			"bank_name": "Example Bank",
			"bank_code": "000",
			"status": "active",
			"created_at": "2026-09-22T09:14:02.000Z"
		}`)
	})

	if _, _, err := c.VirtualAccounts.Get(context.Background(), ""); err != nil {
		t.Fatalf("Get returned error: %v", err)
	}
}
