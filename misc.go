package bachs

import (
	"context"
	"net/http"
	"net/url"
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

// PaymentMethod is one payment method available to the authenticated
// organization, with the currencies it supports. Source:
// https://docs.bachs.io/api-reference/payments/list-payment-methods
type PaymentMethod struct {
	// ID is the payment method identifier (for example "BANK_TRANSFER").
	ID string `json:"id"`

	// DisplayName is the human-readable payment method name.
	DisplayName string `json:"display_name"`

	// Icon is an icon or token representing the method.
	Icon string `json:"icon"`

	// Description is the developer-facing payment method description.
	Description string `json:"description"`

	// Type is the method type (for example "fiat" or "crypto").
	Type string `json:"type"`

	// EnabledByDefault is true when the method is enabled by default for
	// organizations.
	EnabledByDefault bool `json:"enabled_by_default"`

	// Currencies supported for this method.
	Currencies []string `json:"currencies"`
}

// PaymentMethods is the response of Misc.ListPaymentMethods.
type PaymentMethods struct {
	// PaymentMethods lists the methods available to the organization in the
	// current environment.
	PaymentMethods []PaymentMethod `json:"payment_methods"`
}

// PaymentRail is one available payment rail for a method and currency
// combination. Use its ID as the payment_rail parameter when creating quotes.
type PaymentRail struct {
	// ID is the rail identifier.
	ID string `json:"id"`

	// Name is the human-readable rail name.
	Name *string `json:"name"`

	// Active reports whether the rail is currently active and available.
	Active *bool `json:"active"`
}

// PaymentRails is the response of Misc.ListPaymentRails.
type PaymentRails struct {
	// PaymentMethod requested.
	PaymentMethod string `json:"payment_method"`

	// Currency requested.
	Currency string `json:"currency"`

	// CountryCode resolved for this rail lookup.
	CountryCode *string `json:"country_code"`

	// Rails available for this method and currency.
	Rails []PaymentRail `json:"rails"`
}

// SupportedCurrencies is the response of Misc.ListSupportedCurrencies: the
// fiat and cryptocurrency codes the platform supports.
type SupportedCurrencies struct {
	// Fiat currencies (for example "USD", "NGN", "GHS").
	Fiat []string `json:"fiat"`

	// Crypto codes, which may include network identifiers (for example
	// "USDT_TRC20", "USDT_ERC20").
	Crypto []string `json:"crypto"`
}

// PayoutSupportedCurrencies is the response of
// Misc.ListPayoutSupportedCurrencies: the currencies that support
// payouts/withdrawals, organized by fiat and crypto type.
type PayoutSupportedCurrencies struct {
	// Fiat currencies that support payouts.
	Fiat []string `json:"fiat"`

	// Crypto currencies that support payouts.
	Crypto []string `json:"crypto"`
}

// ListPaymentMethods returns all available payment methods and their
// supported currencies. Use this to decide which payment options to show
// customers.
func (s *MiscService) ListPaymentMethods(ctx context.Context) (*PaymentMethods, *ResponseMeta, error) {
	var out PaymentMethods
	meta, err := s.request(ctx, http.MethodGet, "/payment-methods", nil, &out)
	if err != nil {
		return nil, meta, err
	}
	return &out, meta, nil
}

// ListPaymentRails returns the available payment rails for a specific payment
// method and currency combination. paymentMethod is one of "CARD", "CRYPTO",
// "BANK_TRANSFER", or "MOBILE_MONEY"; currency is an ISO 4217 code.
// countryCode is optional. Use a rail's ID as the payment_rail parameter when
// creating quotes.
func (s *MiscService) ListPaymentRails(ctx context.Context, paymentMethod, currency, countryCode string) (*PaymentRails, *ResponseMeta, error) {
	q := url.Values{}
	q.Set("payment_method", paymentMethod)
	q.Set("currency", currency)
	if countryCode != "" {
		q.Set("country_code", countryCode)
	}

	var out PaymentRails
	meta, err := s.request(ctx, http.MethodGet, "/payment-methods/rails?"+q.Encode(), nil, &out)
	if err != nil {
		return nil, meta, err
	}
	return &out, meta, nil
}

// ListSupportedCurrencies returns all supported fiat and cryptocurrency codes.
// Use it to validate currency selections and display currency options.
func (s *MiscService) ListSupportedCurrencies(ctx context.Context) (*SupportedCurrencies, *ResponseMeta, error) {
	var out SupportedCurrencies
	meta, err := s.request(ctx, http.MethodGet, "/currencies/supported", nil, &out)
	if err != nil {
		return nil, meta, err
	}
	return &out, meta, nil
}

// ListPayoutSupportedCurrencies returns all currencies that support
// payouts/withdrawals, organized by fiat and cryptocurrency type.
func (s *MiscService) ListPayoutSupportedCurrencies(ctx context.Context) (*PayoutSupportedCurrencies, *ResponseMeta, error) {
	var out PayoutSupportedCurrencies
	meta, err := s.request(ctx, http.MethodGet, "/currencies/payout-supported", nil, &out)
	if err != nil {
		return nil, meta, err
	}
	return &out, meta, nil
}
