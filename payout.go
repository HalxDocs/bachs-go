package bachs

import (
	"context"
	"net/http"
	"net/url"
	"time"
)

// PayoutService provides methods for withdrawing funds from your organization
// balance to a bank account, mobile-money wallet, or crypto wallet: managing
// payout destinations, quoting and creating withdrawals, and resolving bank
// accounts before you pay them. Source: https://docs.bachs.io/for-you/payins-payouts
type PayoutService struct {
	service
}

// SupportedPayoutCurrencies is the response of Payouts.GetSupportedCurrencies:
// the currencies a given payout method can withdraw into.
type SupportedPayoutCurrencies struct {
	// Method requested, normalized to uppercase (for example
	// "BANK_TRANSFER", "MOBILE_MONEY", "CRYPTO").
	Method string `json:"method"`

	// Currencies available for this payout method (for example "NGN").
	Currencies []string `json:"currencies"`
}

// PayoutQuote is a short-lived quote for a payout before it is created,
// showing the fees and exchange rate baked into the amounts.
type PayoutQuote struct {
	// QuoteID identifies the quote. Pass it when creating the withdrawal to
	// lock in the quoted amounts.
	QuoteID string `json:"quote_id"`

	// FromCurrency is the currency debited from your balance.
	FromCurrency string `json:"from_currency"`

	// ToCurrency is the destination currency the recipient receives.
	ToCurrency string `json:"to_currency"`

	// FromAmount is the quoted amount in the source currency.
	FromAmount string `json:"from_amount"`

	// ToAmount is the quoted amount in the destination currency.
	ToAmount string `json:"to_amount"`

	// ExchangeRate applied to the quote.
	ExchangeRate string `json:"exchange_rate"`

	// ExpiresAt is when the quote stops being usable.
	ExpiresAt time.Time `json:"expires_at"`
}

// PayoutDestination is a configured payout destination (bank account, mobile
// money, or crypto wallet) your organization can withdraw to.
type PayoutDestination struct {
	// ID of the destination.
	ID string `json:"id"`

	// OrganizationID the destination belongs to.
	OrganizationID string `json:"organization_id"`

	// Env is "test" or "live".
	Env string `json:"env"`

	// DestinationType is "bank_account", "mobile_money", or "crypto_wallet".
	DestinationType string `json:"destination_type"`

	// Currency the destination receives (for example "NGN").
	Currency string `json:"currency"`

	// Label is the human-readable name the owner gave this destination.
	Label string `json:"label"`

	// AccountNumber for bank accounts, masked when sensitive.
	AccountNumber *string `json:"account_number"`

	// AccountName registered on the bank account.
	AccountName *string `json:"account_name"`

	// BankCode for bank accounts (for example "058").
	BankCode *string `json:"bank_code"`

	// BankName for bank accounts.
	BankName *string `json:"bank_name"`

	// PhoneNumber for mobile-money destinations.
	PhoneNumber *string `json:"phone_number"`

	// MobileProvider for mobile-money destinations (for example "MTN").
	MobileProvider *string `json:"mobile_provider"`

	// WalletAddress for crypto destinations.
	WalletAddress *string `json:"wallet_address"`

	// Network for crypto destinations (for example "TRC20").
	Network *string `json:"network"`

	// IsActive is true while the destination accepts payouts.
	IsActive bool `json:"is_active"`

	// Metadata is key-value data stored on the destination.
	Metadata map[string]any `json:"metadata"`

	// CreatedAt is when the destination was created.
	CreatedAt time.Time `json:"created_at"`

	// UpdatedAt is when the destination was last updated.
	UpdatedAt time.Time `json:"updated_at"`
}

// PayoutDestinationList is the response of Payouts.ListDestinations. The
// endpoint returns a flat destinations array with a total rather than the
// standard items/pagination envelope.
type PayoutDestinationList struct {
	// Destinations on this page.
	Destinations []PayoutDestination `json:"destinations"`

	// Total number of configured destinations.
	Total int `json:"total"`
}

// CreatePayoutDestinationRequest is the payload for
// Payouts.CreateDestination and Payouts.UpdateDestination. Only the fields
// relevant to the DestinationType are used; the rest are ignored.
type CreatePayoutDestinationRequest struct {
	// DestinationType is "bank_account", "mobile_money", or "crypto_wallet".
	DestinationType string `json:"destination_type"`

	// Currency the destination receives (for example "NGN").
	Currency string `json:"currency"`

	// Label is a human-readable name for this destination.
	Label string `json:"label"`

	// AccountNumber for bank accounts.
	AccountNumber string `json:"account_number,omitempty"`

	// AccountName registered on the bank account.
	AccountName string `json:"account_name,omitempty"`

	// BankCode for bank accounts.
	BankCode string `json:"bank_code,omitempty"`

	// BankName for bank accounts.
	BankName string `json:"bank_name,omitempty"`

	// PhoneNumber for mobile-money destinations.
	PhoneNumber string `json:"phone_number,omitempty"`

	// MobileProvider for mobile-money destinations.
	MobileProvider string `json:"mobile_provider,omitempty"`

	// WalletAddress for crypto destinations.
	WalletAddress string `json:"wallet_address,omitempty"`

	// Network for crypto destinations.
	Network string `json:"network,omitempty"`

	// Metadata is key-value data stored on the destination.
	Metadata map[string]any `json:"metadata,omitempty"`
}

// UpdatePayoutDestinationRequest is the payload for
// Payouts.UpdateDestination. It has the same shape as the create request.
type UpdatePayoutDestinationRequest = CreatePayoutDestinationRequest

// DeletePayoutDestinationResponse is the result of Payouts.DeleteDestination.
type DeletePayoutDestinationResponse struct {
	// Success is true when the destination was deleted.
	Success bool `json:"success"`

	// Message explains the result.
	Message string `json:"message"`
}

// BankAccountResolution is the resolved name on a bank account, returned by
// Payouts.ResolveBankAccount when the account number and bank code match.
type BankAccountResolution struct {
	// AccountNumber resolved.
	AccountNumber string `json:"account_number"`

	// AccountName registered on the account.
	AccountName string `json:"account_name"`

	// BankCode resolved.
	BankCode string `json:"bank_code"`

	// BankName resolved.
	BankName string `json:"bank_name"`
}

// ResolveBankAccountResponse is the result of Payouts.ResolveBankAccount.
// Data is nil when the account could not be resolved; Error carries why.
type ResolveBankAccountResponse struct {
	// Status is true when the lookup succeeded.
	Status bool `json:"status"`

	// Message explains the result.
	Message string `json:"message"`

	// Data holds the resolved account details on a successful match.
	Data *BankAccountResolution `json:"data"`

	// Error explains a failed lookup.
	Error *string `json:"error"`
}

// Bank is one supported bank for bank-account payouts, returned by
// Payouts.ListBanks. Use its Code when resolving or paying an account.
type Bank struct {
	// Name of the bank.
	Name string `json:"name"`

	// Slug is a URL-safe identifier for the bank.
	Slug string `json:"slug"`

	// Code to use in resolve-account and withdrawal requests.
	Code string `json:"code"`

	// NIBSSBankCode is the central-bank routing code, when published.
	NIBSSBankCode *string `json:"nibss_bank_code"`

	// Country the bank operates in (two-letter ISO 3166-1).
	Country string `json:"country"`
}

// BankList is the response of Payouts.ListBanks.
type BankList struct {
	// Status is true when the list was retrieved.
	Status bool `json:"status"`

	// Message explains the result.
	Message string `json:"message"`

	// Data holds the banks on a successful response.
	Data []Bank `json:"data"`

	// Error explains a failed response.
	Error *string `json:"error"`
}

// Payout is a withdrawal of funds from your organization balance to one of
// your payout destinations.
type Payout struct {
	// WithdrawalID uniquely identifies the withdrawal.
	WithdrawalID string `json:"withdrawal_id"`

	// OrganizationID the withdrawal belongs to.
	OrganizationID string `json:"organization_id"`

	// Reference is the client reference supplied at creation.
	Reference *string `json:"reference"`

	// Amount withdrawn in Currency.
	Amount string `json:"amount"`

	// Currency debited from the balance.
	Currency string `json:"currency"`

	// Status is "pending_submission", "pending_collection", "processing",
	// "manual_review", "successful", "failed", "cancelled", "expired", or
	// "reconciled".
	Status string `json:"status"`

	// FromCurrency, for multi-currency withdrawals, is the debited currency.
	FromCurrency *string `json:"from_currency"`

	// ToCurrency, for multi-currency withdrawals, is the paid-out currency.
	ToCurrency *string `json:"to_currency"`

	// FromAmount, for multi-currency withdrawals, is the debited amount.
	FromAmount *string `json:"from_amount"`

	// ToAmount, for multi-currency withdrawals, is the paid-out amount.
	ToAmount *string `json:"to_amount"`

	// ExchangeRate applied to a multi-currency withdrawal.
	ExchangeRate *string `json:"exchange_rate"`

	// Destination is a display string for where the funds went (for example
	// "058 • 0123456789").
	Destination *string `json:"destination"`

	// QuoteID of the quote this withdrawal was created against, if any.
	QuoteID *string `json:"quote_id"`

	// PayoutMethod used (for example "BANK_TRANSFER").
	PayoutMethod *string `json:"payout_method"`

	// ProviderReference from the payout provider, once assigned.
	ProviderReference *string `json:"provider_reference"`

	// Metadata is key-value data stored on the withdrawal.
	Metadata map[string]any `json:"metadata"`

	// RejectionReason explains a rejected withdrawal.
	RejectionReason *string `json:"rejection_reason"`

	// ParentWithdrawalID is set when this withdrawal is part of a batch.
	ParentWithdrawalID *string `json:"parent_withdrawal_id"`

	// BatchIndex is this withdrawal's position in its batch.
	BatchIndex *int `json:"batch_index"`

	// BatchTotal is the size of the batch this withdrawal belongs to.
	BatchTotal *int `json:"batch_total"`

	// CreatedAt is when the withdrawal was created.
	CreatedAt time.Time `json:"created_at"`

	// ApprovedAt is when the withdrawal was approved, if it was.
	ApprovedAt *time.Time `json:"approved_at"`

	// CompletedAt is when the withdrawal finished, if it did.
	CompletedAt *time.Time `json:"completed_at"`

	// UpdatedAt is when the withdrawal was last updated.
	UpdatedAt *time.Time `json:"updated_at"`
}

// CreateWithdrawalRequest is the payload for Payouts.CreateWithdrawal. Either
// PayoutDestinationID or the inline destination details (AccountNumber and
// BankCode for banks, and so on) select where the funds go.
type CreateWithdrawalRequest struct {
	// FromCurrency is the currency debited from your balance.
	FromCurrency string `json:"from_currency"`

	// ToCurrency is the currency the destination receives.
	ToCurrency string `json:"to_currency"`

	// Amount to withdraw, as a decimal string.
	Amount string `json:"amount"`

	// PaymentMethod is the payout method (for example "BANK_TRANSFER").
	PaymentMethod string `json:"payment_method"`

	// QuoteID locks in the amounts quoted by Payouts.CreateQuote, if any.
	QuoteID string `json:"quote_id,omitempty"`

	// Reference is your unique identifier for this withdrawal.
	Reference string `json:"reference"`

	// Email for receipts and status notifications.
	Email string `json:"email"`

	// PaymentRail, when a specific rail is required, from Misc.ListPaymentRails.
	PaymentRail string `json:"payment_rail,omitempty"`

	// Metadata is key-value data stored on the withdrawal.
	Metadata map[string]any `json:"metadata,omitempty"`

	// PayoutDestinationID selects a saved destination to withdraw to.
	PayoutDestinationID string `json:"payout_destination_id,omitempty"`

	// AccountNumber, with BankCode, names a bank destination inline.
	AccountNumber string `json:"account_number,omitempty"`

	// BankCode, with AccountNumber, names a bank destination inline.
	BankCode string `json:"bank_code,omitempty"`

	// PhoneNumber names a mobile-money destination inline.
	PhoneNumber string `json:"phone_number,omitempty"`

	// WalletAddress names a crypto destination inline.
	WalletAddress string `json:"wallet_address,omitempty"`

	// Network for a crypto destination (for example "TRC20").
	Network string `json:"network,omitempty"`

	// Memo is an optional internal note.
	Memo string `json:"memo,omitempty"`

	// Description is an optional human-readable description.
	Description string `json:"description,omitempty"`
}

// CreateWithdrawalResponse is the result of Payouts.CreateWithdrawal.
type CreateWithdrawalResponse struct {
	// WithdrawalID identifies the created withdrawal.
	WithdrawalID string `json:"withdrawal_id"`

	// Status the withdrawal starts in.
	Status string `json:"status"`

	// ProviderReference, once assigned by the payout provider.
	ProviderReference *string `json:"provider_reference"`
}

// CreatePayoutQuoteRequest is the payload for Payouts.CreateQuote.
type CreatePayoutQuoteRequest struct {
	// FromCurrency is the currency debited from your balance.
	FromCurrency string `json:"from_currency"`

	// ToCurrency is the currency the destination receives.
	ToCurrency string `json:"to_currency"`

	// Amount to withdraw, as a decimal string.
	Amount string `json:"amount"`

	// PayoutMethod, when quoting for a specific method (for example
	// "BANK_TRANSFER").
	PayoutMethod string `json:"payout_method,omitempty"`
}

// GetSupportedCurrencies returns the currencies a given payout method can
// withdraw into. method is one of "BANK_TRANSFER", "MOBILE_MONEY", or
// "CRYPTO".
func (s *PayoutService) GetSupportedCurrencies(ctx context.Context, method string) (*SupportedPayoutCurrencies, *ResponseMeta, error) {
	q := url.Values{}
	q.Set("method", method)

	var out SupportedPayoutCurrencies
	meta, err := s.request(ctx, http.MethodGet, "/payouts/supported-currencies?"+q.Encode(), nil, &out)
	if err != nil {
		return nil, meta, err
	}
	return &out, meta, nil
}

// CreateQuote quotes a payout before you create it, showing the fees and
// exchange rate. The quote expires; create the withdrawal while it is valid
// to lock in the quoted amounts.
func (s *PayoutService) CreateQuote(ctx context.Context, req CreatePayoutQuoteRequest, opts ...RequestOption) (*PayoutQuote, *ResponseMeta, error) {
	var out PayoutQuote
	meta, err := s.request(ctx, http.MethodPost, "/payouts/quotes", req, &out, opts...)
	if err != nil {
		return nil, meta, err
	}
	return &out, meta, nil
}

// ListDestinations lists the payout destinations configured for your
// organization.
func (s *PayoutService) ListDestinations(ctx context.Context) (*PayoutDestinationList, *ResponseMeta, error) {
	var out PayoutDestinationList
	meta, err := s.request(ctx, http.MethodGet, "/payouts/destinations", nil, &out)
	if err != nil {
		return nil, meta, err
	}
	return &out, meta, nil
}

// CreateDestination adds a payout destination (bank account, mobile money, or
// crypto wallet) where you can withdraw funds.
func (s *PayoutService) CreateDestination(ctx context.Context, req CreatePayoutDestinationRequest, opts ...RequestOption) (*PayoutDestination, *ResponseMeta, error) {
	var out PayoutDestination
	meta, err := s.request(ctx, http.MethodPost, "/payouts/destinations", req, &out, opts...)
	if err != nil {
		return nil, meta, err
	}
	return &out, meta, nil
}

// UpdateDestination updates an existing payout destination.
func (s *PayoutService) UpdateDestination(ctx context.Context, destinationID string, req UpdatePayoutDestinationRequest) (*PayoutDestination, *ResponseMeta, error) {
	var out PayoutDestination
	meta, err := s.request(ctx, http.MethodPut, "/payouts/destinations/"+url.PathEscape(destinationID), req, &out)
	if err != nil {
		return nil, meta, err
	}
	return &out, meta, nil
}

// DeleteDestination removes a payout destination. Withdrawals already in
// flight are unaffected.
func (s *PayoutService) DeleteDestination(ctx context.Context, destinationID string) (*DeletePayoutDestinationResponse, *ResponseMeta, error) {
	var out DeletePayoutDestinationResponse
	meta, err := s.request(ctx, http.MethodDelete, "/payouts/destinations/"+url.PathEscape(destinationID), nil, &out)
	if err != nil {
		return nil, meta, err
	}
	return &out, meta, nil
}

// ResolveBankAccount validates a bank account (account number plus bank code)
// before you create a payout destination for it. Data is nil on a mismatch,
// not an error.
func (s *PayoutService) ResolveBankAccount(ctx context.Context, bankCode, accountNumber string, opts ...RequestOption) (*ResolveBankAccountResponse, *ResponseMeta, error) {
	req := struct {
		BankCode      string `json:"bank_code"`
		AccountNumber string `json:"account_number"`
	}{BankCode: bankCode, AccountNumber: accountNumber}

	var out ResolveBankAccountResponse
	meta, err := s.request(ctx, http.MethodPost, "/payouts/resolve-account", req, &out, opts...)
	if err != nil {
		return nil, meta, err
	}
	return &out, meta, nil
}

// ListBanks lists the banks supported for bank-account payouts, optionally
// filtered by country code (for example "NG"). Use a bank's Code when
// resolving an account or naming it as a destination.
func (s *PayoutService) ListBanks(ctx context.Context, countryCode string) (*BankList, *ResponseMeta, error) {
	path := "/payouts/banks"
	if countryCode != "" {
		q := url.Values{}
		q.Set("country_code", countryCode)
		path += "?" + q.Encode()
	}

	var out BankList
	meta, err := s.request(ctx, http.MethodGet, path, nil, &out)
	if err != nil {
		return nil, meta, err
	}
	return &out, meta, nil
}

// List returns a page of payout withdrawals for your organization, newest
// first. Filter with ListParams.StatusFilter ("requested", "pending",
// "processing", "approved", "rejected", "completed", or "failed").
func (s *PayoutService) List(ctx context.Context, params ListParams) (*Page[Payout], *ResponseMeta, error) {
	var env pageEnvelope[Payout]
	meta, err := s.request(ctx, http.MethodGet, queryPath("/payouts", params), nil, &env)
	if err != nil {
		return nil, meta, err
	}
	return env.page(), meta, nil
}

// Get returns a single payout withdrawal by ID.
func (s *PayoutService) Get(ctx context.Context, withdrawalID string) (*Payout, *ResponseMeta, error) {
	var out Payout
	meta, err := s.request(ctx, http.MethodGet, "/payouts/"+url.PathEscape(withdrawalID), nil, &out)
	if err != nil {
		return nil, meta, err
	}
	return &out, meta, nil
}

// CreateWithdrawal creates a payout withdrawal request and starts payout
// processing. Idempotency keys are supported, so retrying a timed-out
// request cannot double-pay.
func (s *PayoutService) CreateWithdrawal(ctx context.Context, req CreateWithdrawalRequest, opts ...RequestOption) (*CreateWithdrawalResponse, *ResponseMeta, error) {
	var out CreateWithdrawalResponse
	meta, err := s.request(ctx, http.MethodPost, "/payouts/withdrawals", req, &out, opts...)
	if err != nil {
		return nil, meta, err
	}
	return &out, meta, nil
}
