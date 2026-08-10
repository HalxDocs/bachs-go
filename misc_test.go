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
