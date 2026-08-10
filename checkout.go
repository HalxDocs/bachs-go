package bachs

import (
	"context"
	"net/http"
	"net/url"
	"time"
)

// CheckoutService provides methods for creating and retrieving checkout
// sessions. A checkout session ties one or more products (or a raw amount) to
// a customer and produces a hosted checkout URL; completing it collects the
// payment and, for recurring products, creates the subscription.
type CheckoutService struct {
	service
}

// CheckoutCustomer identifies the customer on a checkout session: either an
// existing customer's ID or the details of a new customer.
type CheckoutCustomer struct {
	// CustomerID is the ID of an existing customer, when the customer is
	// known.
	CustomerID string `json:"customer_id,omitempty"`

	// Email of a new customer. Required when creating a customer inline.
	Email string `json:"email,omitempty"`

	// Name of a new customer.
	Name string `json:"name,omitempty"`

	// PhoneNumber of a new customer, in E.164 format.
	PhoneNumber *string `json:"phone_number,omitempty"`
}

// ProductItemRequest is one catalog product in a checkout's cart. A product
// whose price_type is custom (pay-what-you-want) can carry a chosen Amount.
type ProductItemRequest struct {
	// ProductID of the catalog product to include.
	ProductID string `json:"product_id"`

	// Quantity of units. Defaults to 1 when omitted.
	Quantity int `json:"quantity,omitempty"`

	// Amount is the chosen amount for a pay-what-you-want price, or an
	// ad-hoc charge for a CUSTOM product.
	Amount *string `json:"amount,omitempty"`

	// Pricing is an ad-hoc price override for this checkout only; no product
	// is created or modified.
	Pricing *AdHocPricing `json:"pricing,omitempty"`
}

// AdHocPricing overrides a product's price for a single checkout, in the
// product's primary currency.
type AdHocPricing struct {
	// PriceType is "fixed", "custom", or "free".
	PriceType string `json:"price_type,omitempty"`

	// Amount is the price for a fixed ad-hoc price. Required for "fixed";
	// not valid for "custom".
	Amount string `json:"amount,omitempty"`

	// PresetAmount is the suggested starting amount for a custom ad-hoc price.
	PresetAmount string `json:"preset_amount,omitempty"`

	// MinimumAmount is the lower bound for a custom ad-hoc price.
	MinimumAmount string `json:"minimum_amount,omitempty"`

	// MaximumAmount is the upper bound for a custom ad-hoc price.
	MaximumAmount string `json:"maximum_amount,omitempty"`
}

// CheckoutPricing is the raw pricing for a product-less (pure) checkout:
// an amount and currency with no catalog products. Mutually exclusive with
// ProductCart.
type CheckoutPricing struct {
	// Currency is the base currency code (for example "USD", "NGN").
	Currency string `json:"currency,omitempty"`

	// Amount is the base amount as a decimal string. Required for a fixed
	// price; omit for a custom (buyer-entered) or free price.
	Amount string `json:"amount,omitempty"`

	// PriceType is "fixed" (default when amount is set), "custom", or "free".
	PriceType string `json:"price_type,omitempty"`

	// PresetAmount is the suggested starting amount for a custom price.
	PresetAmount string `json:"preset_amount,omitempty"`

	// MinimumAmount is the lower bound for a custom price.
	MinimumAmount string `json:"minimum_amount,omitempty"`

	// MaximumAmount is the upper bound for a custom price.
	MaximumAmount string `json:"maximum_amount,omitempty"`

	// CurrencyOptions are currency-specific pricing overrides: keys are fiat
	// currency codes, values are decimal amount strings.
	CurrencyOptions map[string]string `json:"currency_options,omitempty"`
}

// CreateCheckoutSessionRequest is the payload for Checkouts.Create. Exactly
// one of ProductCart or Pricing must be set. Source:
// https://docs.bachs.io/api-reference/payments/create-checkout-session
type CreateCheckoutSessionRequest struct {
	// BillingCurrency optionally overrides the checkout billing currency;
	// defaults to the product pricing currency when omitted.
	BillingCurrency string `json:"billing_currency,omitempty"`

	// AllowedPaymentMethodTypes restricts which payment methods the customer
	// may use: "card", "crypto", "bank_transfer", "mobile_money".
	AllowedPaymentMethodTypes []string `json:"allowed_payment_method_types,omitempty"`

	// CancelURL is where the customer is sent if they cancel or abandon the
	// checkout.
	CancelURL string `json:"cancel_url,omitempty"`

	// ReturnURL is a deprecated alias for SuccessURL, kept for backward
	// compatibility. SuccessURL wins when both are set.
	ReturnURL string `json:"return_url,omitempty"`

	// SuccessURL is where the customer is redirected after a successful
	// payment. Bachs appends "?checkout_id=<id>".
	SuccessURL string `json:"success_url,omitempty"`

	// Customer is either an existing customer or the new customer's details.
	Customer CheckoutCustomer `json:"customer"`

	// Metadata is optional key-value data (max 20 keys, max 10KB total).
	Metadata map[string]any `json:"metadata,omitempty"`

	// ProductCart lists the catalog products in the cart. Mutually exclusive
	// with Pricing.
	ProductCart []ProductItemRequest `json:"product_cart,omitempty"`

	// Pricing sets a raw amount for a product-less checkout. Mutually
	// exclusive with ProductCart.
	Pricing *CheckoutPricing `json:"pricing,omitempty"`

	// Reference is an optional client reference, unique per organization. An
	// auto-generated one is used when omitted.
	Reference string `json:"reference,omitempty"`

	// ExpiresInMinutes is how long the checkout stays open; defaults to 60.
	ExpiresInMinutes int `json:"expires_in_minutes,omitempty"`
}

// CreateCheckoutSessionResponse is the result of Checkouts.Create: the
// session identifiers and the hosted checkout URL.
type CreateCheckoutSessionResponse struct {
	// CheckoutID uniquely identifies the underlying checkout.
	CheckoutID string `json:"checkout_id"`

	// CheckoutURL is the hosted checkout URL the customer completes payment
	// at.
	CheckoutURL string `json:"checkout_url"`

	// Status is "OPEN", "COMPLETED", "EXPIRED", or "CANCELLED". New sessions
	// start in "OPEN".
	Status string `json:"status"`

	// ExpiresAt is when the checkout URL stops working.
	ExpiresAt time.Time `json:"expires_at"`

	// CreatedAt is when the checkout was created.
	CreatedAt time.Time `json:"created_at"`

	// Reference echoes the client reference supplied at creation.
	Reference string `json:"reference,omitempty"`
}

// CheckoutSession is the full checkout session object returned by
// Checkouts.Get, including resolved product line items and, once payment has
// been attempted, the linked payment.
type CheckoutSession struct {
	// CheckoutID uniquely identifies the checkout.
	CheckoutID string `json:"checkout_id"`

	// Status is "OPEN", "COMPLETED", "EXPIRED", or "CANCELLED".
	Status string `json:"status"`

	// Recurring is present only for a subscription checkout; null for a
	// one-time checkout.
	Recurring *CheckoutRecurring `json:"recurring"`

	// PaymentStatus is the payment lifecycle, for example "succeeded" or
	// "requires_payment_method". Null before payment is attempted.
	PaymentStatus *string `json:"payment_status"`

	// SourceType is what created the checkout, for example "CHECKOUT_SESSION"
	// or "API".
	SourceType *string `json:"source_type"`

	// Amount is the total amount in Currency.
	Amount string `json:"amount"`

	// Currency is the base currency code.
	Currency string `json:"currency"`

	// Reference is the merchant-supplied or auto-generated reference.
	Reference *string `json:"reference"`

	// Charge is the payment created by this checkout, once payment has been
	// attempted. Null before then.
	Charge *Payment `json:"charge"`

	// PaymentMethod is the payment method selected for the checkout, if any.
	PaymentMethod *string `json:"payment_method"`

	// Customer attached to the checkout.
	Customer *CheckoutSessionCustomer `json:"customer"`

	// SuccessURL is where the customer is redirected after payment.
	SuccessURL *string `json:"success_url"`

	// CancelURL is where the customer is redirected if they cancel.
	CancelURL *string `json:"cancel_url"`

	// Products are the resolved product line items. May be null for
	// SELECTION-mode sessions before the customer picks.
	Products []ResolvedProductItem `json:"products"`

	// BillingCurrency is the currency the customer selected for billing.
	BillingCurrency *string `json:"billing_currency"`

	// SessionMode is "CART" (a fixed set of items) or "SELECTION" (the
	// customer picks one from a group).
	SessionMode *string `json:"session_mode"`

	// Metadata attached at session creation.
	Metadata map[string]any `json:"metadata"`

	// CreatedAt is when the checkout was created.
	CreatedAt time.Time `json:"created_at"`

	// ExpiresAt is when the checkout URL stops working.
	ExpiresAt *time.Time `json:"expires_at"`

	// CompletedAt is when the session was completed.
	CompletedAt *time.Time `json:"completed_at"`

	// UpdatedAt is when the session was last updated.
	UpdatedAt time.Time `json:"updated_at"`
}

// CheckoutRecurring describes the billing cadence of a subscription checkout.
type CheckoutRecurring struct {
	// Interval is "day", "week", "month", or "year".
	Interval string `json:"interval"`

	// IntervalCount is the number of intervals per billing cycle.
	IntervalCount int `json:"interval_count"`
}

// CheckoutSessionCustomer is the customer attached to a checkout session.
type CheckoutSessionCustomer struct {
	// ID of the customer, once resolved. Null until a customer is matched or
	// created.
	ID *string `json:"id"`

	// Email of the customer.
	Email string `json:"email"`

	// Name of the customer. Null when not provided.
	Name *string `json:"name"`
}

// ResolvedProductItem is one product line item on a resolved checkout session.
type ResolvedProductItem struct {
	// ProductID identifies the product.
	ProductID string `json:"product_id"`

	// ProductName is the product's display name.
	ProductName string `json:"product_name"`

	// Quantity is the number of units.
	Quantity int `json:"quantity"`

	// UnitAmount is the price per unit in Currency.
	UnitAmount string `json:"unit_amount"`

	// Currency is the currency code for this line item.
	Currency string `json:"currency"`

	// PriceType is "fixed", "free", or "custom".
	PriceType string `json:"price_type"`

	// MinimumAmount is the minimum allowed amount for a custom price.
	MinimumAmount *string `json:"minimum_amount"`

	// MaximumAmount is the maximum allowed amount for a custom price.
	MaximumAmount *string `json:"maximum_amount"`

	// LineTotal is UnitAmount times Quantity.
	LineTotal string `json:"line_total"`
}

// Create creates a product-based or ad-hoc checkout session and returns the
// hosted checkout URL. Idempotency-Key may be passed via WithIdempotencyKey
// to make retries safe.
func (s *CheckoutService) Create(ctx context.Context, req CreateCheckoutSessionRequest, opts ...RequestOption) (*CreateCheckoutSessionResponse, *ResponseMeta, error) {
	var out CreateCheckoutSessionResponse
	meta, err := s.request(ctx, http.MethodPost, "/checkout-sessions", req, &out, opts...)
	if err != nil {
		return nil, meta, err
	}
	return &out, meta, nil
}

// Get retrieves the details of a checkout session by its ID, including
// resolved product line items and charge information.
func (s *CheckoutService) Get(ctx context.Context, checkoutID string) (*CheckoutSession, *ResponseMeta, error) {
	var out CheckoutSession
	meta, err := s.request(ctx, http.MethodGet, "/checkout-sessions/"+url.PathEscape(checkoutID), nil, &out)
	if err != nil {
		return nil, meta, err
	}
	return &out, meta, nil
}
