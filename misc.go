package bachs

import (
	"context"
	"net/http"
	"net/url"
	"time"
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

	// HeldForDisputes is the amount reserved against open disputes in this
	// currency.
	HeldForDisputes *string `json:"held_for_disputes"`
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
	meta, err := s.request(ctx, http.MethodGet, "/balances", nil, &out)
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

// Payout schedule intervals for a currency's automatic payouts. Source:
// https://docs.bachs.io/guides/payouts/payout-schedules
const (
	// PayoutScheduleIntervalManual disables automatic payouts for a
	// currency. The balance waits for an explicit payout creation call.
	PayoutScheduleIntervalManual = "manual"

	// PayoutScheduleIntervalInstant pays out about a minute after funds
	// settle, so busy days produce several payouts and quiet days none.
	PayoutScheduleIntervalInstant = "instant"

	// PayoutScheduleIntervalDaily pays out every day at AnchorHourUTC.
	PayoutScheduleIntervalDaily = "daily"

	// PayoutScheduleIntervalWeekly pays out every week on each of
	// WeeklyPayoutDays, at AnchorHourUTC.
	PayoutScheduleIntervalWeekly = "weekly"

	// PayoutScheduleIntervalMonthly pays out every month on each of
	// MonthlyPayoutDays, at AnchorHourUTC.
	PayoutScheduleIntervalMonthly = "monthly"
)

// PayoutSchedule is one currency's automatic payout configuration: an
// interval, timing within it, and a floor below which a run pays nothing and
// the money rolls into the next run. A schedule moves settled customer
// collections only; top-ups, received transfers, conversions, and returned
// payouts are invisible to it.
type PayoutSchedule struct {
	// Currency this schedule pays out (for example "NGN").
	Currency string `json:"currency"`

	// PayoutCurrency the destination receives, when it differs from
	// Currency (for example paying a USD balance to an NGN bank account).
	PayoutCurrency *string `json:"payout_currency"`

	// Interval is manual, instant, daily, weekly, or monthly. Use the
	// PayoutScheduleInterval* constants.
	Interval string `json:"interval"`

	// WeeklyPayoutDays are weekday names ("monday" through "sunday") read by
	// the weekly interval. Null for other intervals.
	WeeklyPayoutDays []string `json:"weekly_payout_days"`

	// MonthlyPayoutDays are days of the month (1 to 31) read by the monthly
	// interval. 29-31 mean the last day of a shorter month. Null for other
	// intervals.
	MonthlyPayoutDays []int `json:"monthly_payout_days"`

	// AnchorHourUTC is the hour runs happen at, 0 to 23 UTC. Defaults to 10.
	AnchorHourUTC int `json:"anchor_hour_utc"`

	// MinimumAmount is the floor, as a decimal string: a run whose eligible
	// total is below it pays nothing. Null when no floor is set.
	MinimumAmount *string `json:"minimum_amount"`

	// NextRunAt is the next scheduled run. Always null on instant.
	NextRunAt *time.Time `json:"next_run_at"`

	// LastRunAt is when the schedule last ran. Null before the first run.
	LastRunAt *time.Time `json:"last_run_at"`

	// LastWithdrawalID is the payout the last run created, if any.
	LastWithdrawalID *string `json:"last_withdrawal_id"`

	// DisabledReason explains why three consecutive failed runs set the
	// currency back to manual. Null otherwise.
	DisabledReason *string `json:"disabled_reason"`
}

// BalanceSettings is the response of Misc.GetBalanceSettings and
// Misc.UpdateBalanceSettings: the payout schedule keyed by currency. A
// currency never scheduled is absent from the map rather than present and
// empty.
type BalanceSettings struct {
	// ScheduleByCurrency maps each configured currency to its schedule.
	ScheduleByCurrency map[string]PayoutSchedule `json:"schedule_by_currency"`
}

// UpdatePayoutScheduleRequest is one currency's schedule within
// UpdateBalanceSettingsRequest. A currency left out of the map keeps its
// schedule; a currency named is replaced in full, so send the whole
// schedule for it rather than the one field that changed — a field omitted
// here is cleared, not kept. Turn automatic payouts off with interval
// "manual" rather than by removing the currency, which changes nothing.
type UpdatePayoutScheduleRequest struct {
	// PayoutCurrency pays out in another currency (for example a USD
	// balance to an NGN destination). Only USD and stablecoin balances
	// convert outward.
	PayoutCurrency *string `json:"payout_currency,omitempty"`

	// Interval is manual, instant, daily, weekly, or monthly. Use the
	// PayoutScheduleInterval* constants.
	Interval string `json:"interval,omitempty"`

	// WeeklyPayoutDays are weekday names for the weekly interval. Sending
	// them with any other interval is rejected.
	WeeklyPayoutDays []string `json:"weekly_payout_days,omitempty"`

	// MonthlyPayoutDays are days of the month for the monthly interval.
	// Sending them with any other interval is rejected.
	MonthlyPayoutDays []int `json:"monthly_payout_days,omitempty"`

	// AnchorHourUTC is the hour runs happen at, 0 to 23 UTC.
	AnchorHourUTC *int `json:"anchor_hour_utc,omitempty"`

	// MinimumAmount is the floor below which a run pays nothing, as a
	// decimal string.
	MinimumAmount *string `json:"minimum_amount,omitempty"`
}

// UpdateBalanceSettingsRequest is the payload for
// Misc.UpdateBalanceSettings.
type UpdateBalanceSettingsRequest struct {
	// ScheduleByCurrency maps each currency being changed to its full new
	// schedule. Currencies left out keep their schedules.
	ScheduleByCurrency map[string]UpdatePayoutScheduleRequest `json:"schedule_by_currency"`
}

// GetBalanceSettings returns the payout schedule keyed by currency, so you
// never have to keep your own copy of it. Pass WithConnectedAccount to read
// a connected account's schedule. Requires the balance:read scope.
func (s *MiscService) GetBalanceSettings(ctx context.Context, opts ...RequestOption) (*BalanceSettings, *ResponseMeta, error) {
	var out BalanceSettings
	meta, err := s.request(ctx, http.MethodGet, "/balance_settings", nil, &out, opts...)
	if err != nil {
		return nil, meta, err
	}
	return &out, meta, nil
}

// UpdateBalanceSettings sets the payout schedule per currency. Each named
// currency is replaced in full; each omitted one is untouched. Pass
// WithConnectedAccount to set a connected account's schedule, for example
// once at onboarding. Requires the balance:write scope.
func (s *MiscService) UpdateBalanceSettings(ctx context.Context, req UpdateBalanceSettingsRequest, opts ...RequestOption) (*BalanceSettings, *ResponseMeta, error) {
	var out BalanceSettings
	meta, err := s.request(ctx, http.MethodPost, "/balance_settings", req, &out, opts...)
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
