package bachs

import (
	"context"
	"io"
	"net/http"
	"testing"
)

// TestGetBalances uses the exact example payload from
// https://docs.bachs.io/api-reference/accounts/get-balances
func TestGetBalances(t *testing.T) {
	c := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Errorf("method = %s, want GET", r.Method)
		}
		if r.URL.Path != "/v1/accounts/balances" {
			t.Errorf("path = %q, want /v1/accounts/balances", r.URL.Path)
		}
		io.WriteString(w, `{
			"account_id": "org_2bdb10644eba4ec488b7e87405597e43",
			"balances": [
				{"currency": "NGN", "available_balance": "58700.00", "pending_balance": "0.00"},
				{"currency": "USD", "available_balance": "95178.20", "pending_balance": "0.00"}
			],
			"total_balance_usd": "95221.29",
			"pending_settlements_by_day": []
		}`)
	})

	balances, _, err := c.Misc.GetBalances(context.Background())
	if err != nil {
		t.Fatalf("GetBalances returned error: %v", err)
	}

	if balances.AccountID != "org_2bdb10644eba4ec488b7e87405597e43" {
		t.Errorf("AccountID = %q", balances.AccountID)
	}
	if len(balances.Balances) != 2 {
		t.Fatalf("len(Balances) = %d, want 2", len(balances.Balances))
	}
	first := balances.Balances[0]
	if first.Currency != "NGN" || first.AvailableBalance != "58700.00" || first.PendingBalance != "0.00" {
		t.Errorf("Balances[0] = %+v", first)
	}
	if balances.TotalBalanceUSD != "95221.29" {
		t.Errorf("TotalBalanceUSD = %q, want 95221.29", balances.TotalBalanceUSD)
	}
	if len(balances.PendingSettlementsByDay) != 0 {
		t.Errorf("PendingSettlementsByDay = %v, want empty", balances.PendingSettlementsByDay)
	}
}

// TestListPaymentMethods uses the exact example payload from
// https://docs.bachs.io/api-reference/payments/list-payment-methods
func TestListPaymentMethods(t *testing.T) {
	c := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Errorf("method = %s, want GET", r.Method)
		}
		if r.URL.RequestURI() != "/v1/payment-methods" {
			t.Errorf("path = %q, want /v1/payment-methods", r.URL.RequestURI())
		}
		io.WriteString(w, `{
			"payment_methods": [
				{
					"id": "BANK_TRANSFER",
					"display_name": "Bank Transfer",
					"icon": "bank",
					"description": "Pay via bank transfer",
					"type": "fiat",
					"enabled_by_default": true,
					"currencies": ["NGN", "USD"]
				},
				{
					"id": "CRYPTO",
					"display_name": "Cryptocurrency",
					"icon": "crypto",
					"description": "Pay with supported crypto assets",
					"type": "crypto",
					"enabled_by_default": true,
					"currencies": ["USDT_TRC20", "USDT_BEP20"]
				}
			]
		}`)
	})

	methods, _, err := c.Misc.ListPaymentMethods(context.Background())
	if err != nil {
		t.Fatalf("ListPaymentMethods returned error: %v", err)
	}
	if len(methods.PaymentMethods) != 2 {
		t.Fatalf("len(PaymentMethods) = %d, want 2", len(methods.PaymentMethods))
	}
	if methods.PaymentMethods[0].ID != "BANK_TRANSFER" || methods.PaymentMethods[0].Type != "fiat" {
		t.Errorf("PaymentMethods[0] = %+v", methods.PaymentMethods[0])
	}
	if len(methods.PaymentMethods[1].Currencies) != 2 || methods.PaymentMethods[1].Currencies[0] != "USDT_TRC20" {
		t.Errorf("PaymentMethods[1].Currencies = %v", methods.PaymentMethods[1].Currencies)
	}
}

// TestListPaymentRails uses the exact example payload from
// https://docs.bachs.io/api-reference/payments/list-payment-rails
func TestListPaymentRails(t *testing.T) {
	c := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Errorf("method = %s, want GET", r.Method)
		}
		if got := r.URL.RequestURI(); got != "/v1/payment-methods/rails?currency=NGN&payment_method=BANK_TRANSFER" {
			t.Errorf("path = %q", got)
		}
		io.WriteString(w, `{
			"payment_method": "BANK_TRANSFER",
			"currency": "NGN",
			"country_code": "NG",
			"rails": [
				{"id": "bank_transfer_ng", "name": "Bank Transfer Nigeria", "active": true}
			]
		}`)
	})

	rails, _, err := c.Misc.ListPaymentRails(context.Background(), "BANK_TRANSFER", "NGN", "")
	if err != nil {
		t.Fatalf("ListPaymentRails returned error: %v", err)
	}
	if rails.PaymentMethod != "BANK_TRANSFER" || rails.Currency != "NGN" {
		t.Errorf("rails = %+v", rails)
	}
	if len(rails.Rails) != 1 || rails.Rails[0].ID != "bank_transfer_ng" {
		t.Errorf("Rails = %+v", rails.Rails)
	}
	if rails.Rails[0].Active == nil || !*rails.Rails[0].Active {
		t.Errorf("Rails[0].Active = %v, want true", rails.Rails[0].Active)
	}
}

// TestListPaymentRailsWithCountry verifies the optional country_code query
// parameter is appended.
func TestListPaymentRailsWithCountry(t *testing.T) {
	c := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		if got := r.URL.RequestURI(); got != "/v1/payment-methods/rails?country_code=NG&currency=NGN&payment_method=BANK_TRANSFER" {
			t.Errorf("path = %q", got)
		}
		io.WriteString(w, `{"payment_method": "BANK_TRANSFER", "currency": "NGN", "rails": []}`)
	})

	_, _, err := c.Misc.ListPaymentRails(context.Background(), "BANK_TRANSFER", "NGN", "NG")
	if err != nil {
		t.Fatalf("ListPaymentRails returned error: %v", err)
	}
}

// TestListSupportedCurrencies uses the exact example payload from
// https://docs.bachs.io/api-reference/payments/list-supported-currencies
func TestListSupportedCurrencies(t *testing.T) {
	c := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.RequestURI() != "/v1/currencies/supported" {
			t.Errorf("path = %q", r.URL.RequestURI())
		}
		io.WriteString(w, `{
			"fiat": ["USD", "NGN", "GHS", "KES", "ZAR"],
			"crypto": ["USDT_TRC20", "USDT_ERC20", "BTC"]
		}`)
	})

	cur, _, err := c.Misc.ListSupportedCurrencies(context.Background())
	if err != nil {
		t.Fatalf("ListSupportedCurrencies returned error: %v", err)
	}
	if len(cur.Fiat) != 5 || cur.Fiat[0] != "USD" {
		t.Errorf("Fiat = %v", cur.Fiat)
	}
	if len(cur.Crypto) != 3 || cur.Crypto[2] != "BTC" {
		t.Errorf("Crypto = %v", cur.Crypto)
	}
}

// TestListPayoutSupportedCurrencies uses the exact example payload from
// https://docs.bachs.io/api-reference/payments/list-payout-supported-currencies
func TestListPayoutSupportedCurrencies(t *testing.T) {
	c := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.RequestURI() != "/v1/currencies/payout-supported" {
			t.Errorf("path = %q", r.URL.RequestURI())
		}
		io.WriteString(w, `{
			"fiat": ["NGN", "USD", "GHS"],
			"crypto": ["USDT_TRC20", "USDT_ERC20"]
		}`)
	})

	cur, _, err := c.Misc.ListPayoutSupportedCurrencies(context.Background())
	if err != nil {
		t.Fatalf("ListPayoutSupportedCurrencies returned error: %v", err)
	}
	if len(cur.Fiat) != 3 || cur.Fiat[1] != "USD" {
		t.Errorf("Fiat = %v", cur.Fiat)
	}
	if len(cur.Crypto) != 2 || cur.Crypto[0] != "USDT_TRC20" {
		t.Errorf("Crypto = %v", cur.Crypto)
	}
}
