package bachs

import (
	"context"
	"net/http"
	"net/url"
	"time"
)

// ConversionService provides methods for converting between settlement
// currencies (for example USD to NGN): quoting a conversion before you run
// it, executing it, and reading the conversion records. Source:
// https://docs.bachs.io/for-you/payins-payouts
type ConversionService struct {
	service
}

// ConversionQuote is a short-lived quote for a conversion between settlement
// currencies.
type ConversionQuote struct {
	// QuoteID identifies the quote. Pass it to Conversions.Create to execute.
	QuoteID string `json:"quote_id"`

	// FromCurrency is the currency being converted away.
	FromCurrency string `json:"from_currency"`

	// ToCurrency is the currency being converted into.
	ToCurrency string `json:"to_currency"`

	// FromAmount is the quoted amount in the source currency.
	FromAmount string `json:"from_amount"`

	// ToAmount is the quoted amount in the destination currency.
	ToAmount string `json:"to_amount"`

	// ExchangeRate applied to the conversion.
	ExchangeRate string `json:"exchange_rate"`

	// ExpiresAt is when the quote stops being usable.
	ExpiresAt time.Time `json:"expires_at"`
}

// Conversion is a completed, pending, or failed currency conversion.
type Conversion struct {
	// ConversionID uniquely identifies the conversion.
	ConversionID string `json:"conversion_id"`

	// Status is "pending", "completed", or "failed".
	Status string `json:"status"`

	// FromCurrency is the currency being converted away.
	FromCurrency string `json:"from_currency"`

	// ToCurrency is the currency being converted into.
	ToCurrency string `json:"to_currency"`

	// FromAmount in the source currency.
	FromAmount string `json:"from_amount"`

	// ToAmount in the destination currency.
	ToAmount string `json:"to_amount"`

	// ExchangeRate applied.
	ExchangeRate string `json:"exchange_rate"`

	// CreatedAt is when the conversion was created.
	CreatedAt time.Time `json:"created_at"`

	// QuoteID the conversion was executed against, if any.
	QuoteID *string `json:"quote_id"`

	// Metadata is key-value data stored on the conversion.
	Metadata map[string]any `json:"metadata"`
}

// CreateConversionQuoteRequest is the payload for Conversions.CreateQuote.
type CreateConversionQuoteRequest struct {
	// FromCurrency is the currency being converted away.
	FromCurrency string `json:"from_currency"`

	// ToCurrency is the currency being converted into.
	ToCurrency string `json:"to_currency"`

	// Amount to convert, as a decimal string.
	Amount string `json:"amount"`
}

// CreateConversionRequest is the payload for Conversions.Create: execute a
// conversion using a valid quote ID.
type CreateConversionRequest struct {
	// FromCurrency is the currency being converted away.
	FromCurrency string `json:"from_currency"`

	// ToCurrency is the currency being converted into.
	ToCurrency string `json:"to_currency"`

	// Amount to convert, as a decimal string.
	Amount string `json:"amount"`

	// QuoteID from Conversions.CreateQuote, locking in the quoted rate.
	QuoteID string `json:"quote_id"`
}

// CreateQuote quotes a conversion between settlement currencies before you
// execute it, showing the exchange rate that will apply.
func (s *ConversionService) CreateQuote(ctx context.Context, req CreateConversionQuoteRequest, opts ...RequestOption) (*ConversionQuote, *ResponseMeta, error) {
	var out ConversionQuote
	meta, err := s.request(ctx, http.MethodPost, "/conversions/quotes", req, &out, opts...)
	if err != nil {
		return nil, meta, err
	}
	return &out, meta, nil
}

// List returns a page of conversion records for your organization, newest
// first. Filter with ListParams.FromCurrency, ToCurrency, Status
// ("pending", "completed", "failed"), StartDate, and EndDate.
func (s *ConversionService) List(ctx context.Context, params ListParams) (*Page[Conversion], *ResponseMeta, error) {
	var env pageEnvelope[Conversion]
	meta, err := s.request(ctx, http.MethodGet, queryPath("/conversions", params), nil, &env)
	if err != nil {
		return nil, meta, err
	}
	return env.page(), meta, nil
}

// Create executes a conversion between settlement currencies using a valid
// quote ID.
func (s *ConversionService) Create(ctx context.Context, req CreateConversionRequest, opts ...RequestOption) (*Conversion, *ResponseMeta, error) {
	var out Conversion
	meta, err := s.request(ctx, http.MethodPost, "/conversions", req, &out, opts...)
	if err != nil {
		return nil, meta, err
	}
	return &out, meta, nil
}

// Get returns a single conversion by ID.
func (s *ConversionService) Get(ctx context.Context, conversionID string) (*Conversion, *ResponseMeta, error) {
	var out Conversion
	meta, err := s.request(ctx, http.MethodGet, "/conversions/"+url.PathEscape(conversionID), nil, &out)
	if err != nil {
		return nil, meta, err
	}
	return &out, meta, nil
}
