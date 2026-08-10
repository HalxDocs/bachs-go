package bachs

import (
	"context"
	"net/http"
)

// MiscService provides account-wide endpoints that do not belong to a single
// resource: balances, payment methods, payment rails, and supported
// currencies.
type MiscService struct {
	service
}

// BalanceBucket is one currency's balance on an account.
type BalanceBucket struct {
	// Currency code for this bucket (ISO 4217 or a configured crypto code).
	Currency string `json:"currency"`

	// AvailableBalance is the amount currently available for new operations
	// in this currency.
	AvailableBalance string `json:"available_balance"`

	// PendingBalance is the in-flight amount not yet available for spending.
	PendingBalance string `json:"pending_balance"`
}

// AccountBalances is the response of Misc.GetBalances: the organization's
// balance buckets by currency plus a consolidated USD total. Source:
// https://docs.bachs.io/api-reference/accounts/get-balances
type AccountBalances struct {
	// AccountID is the organization these balances belong to.
	AccountID string `json:"account_id"`

	// Balances has one entry per currency bucket.
	Balances []BalanceBucket `json:"balances"`

	// TotalBalanceUSD is the aggregate of available and pending balances
	// converted to USD.
	TotalBalanceUSD string `json:"total_balance_usd"`

	// PendingSettlementsByDay lists upcoming settlements grouped by day.
	// Empty when no settlements are pending. Each entry's shape is not
	// documented by the API, so entries are kept as raw JSON objects.
	PendingSettlementsByDay []map[string]any `json:"pending_settlements_by_day"`
}

// GetBalances returns the organization's balance buckets by currency,
// including available, locked, and pending amounts, plus a consolidated USD
// total.
func (s *MiscService) GetBalances(ctx context.Context) (*AccountBalances, *ResponseMeta, error) {
	var out AccountBalances
	meta, err := s.request(ctx, http.MethodGet, "/accounts/balances", nil, &out)
	if err != nil {
		return nil, meta, err
	}
	return &out, meta, nil
}
