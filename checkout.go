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

// Customer creation behavior for a checkout session that collects the
// buyer's identity on the hosted page. Source:
// https://docs.bachs.io/guides/checkout/checkout-sessions#whether-a-guest-becomes-a-customer-customer-creation
const (
	// CustomerCreationIfRequired is the default: a guest who identifies
	// themselves on the hosted page does not join the customer directory.
	// Customer stays null and no customer.created/customer.updated webhook
	// fires, while the buyer's email and name reach you on CustomerDetails.
	CustomerCreationIfRequired = "if_required"

	// CustomerCreationAlways creates a customer record for a guest who
	// identifies themselves on the hosted page, matched by email to an
	// existing customer where one exists. Ignored for a subscription or
	// setup checkout, which always create a customer.
	CustomerCreationAlways = "always"
)

// PaymentMethodOption restricts which currencies one payment method is
// offered in on a checkout. A method left out of PaymentMethodOptions is not
// offered at all; a method included with no Currencies is offered in every
// currency it supports. For crypto the currencies are asset codes such as
// "USDT_TRC20" rather than fiat codes. Restricting only ever narrows what
// the customer sees: it cannot add a method or currency the account is not
// already enabled for. Source:
// https://docs.bachs.io/guides/checkout/checkout-sessions#restrict-methods-and-currencies
type PaymentMethodOption struct {
	// Currencies the method is offered in. Omit to offer it in all of them.
	Currencies []string `json:"currencies,omitempty"`
}

// Checkout payment-method corridors. Each entry is an exact corridor, not a
// payment type: card, bank transfer, and mobile money are each split into
// one corridor per currency. Restrict a checkout to a currency by choosing
// which corridors to list. CRYPTO is one corridor covering every supported
// asset and network. Source:
// https://docs.bachs.io/guides/checkout/checkout-sessions#restrict-payment-methods
const (
	CorridorUSDCard         = "USD_CARD"
	CorridorNGNCard         = "NGN_CARD"
	CorridorNGNBankTransfer = "NGN_BANK_TRANSFER"
	CorridorMomoGHS         = "MOMO_GHS"
	CorridorMomoKES         = "MOMO_KES"
	CorridorMomoTZS         = "MOMO_TZS"
	CorridorMomoUGX         = "MOMO_UGX"
	CorridorMomoXAF         = "MOMO_XAF"
	CorridorMomoXOF         = "MOMO_XOF"
	CorridorMomoRWF         = "MOMO_RWF"
	CorridorMomoMWK         = "MOMO_MWK"
	CorridorMomoZMW         = "MOMO_ZMW"
	CorridorCrypto          = "CRYPTO"
)

// CheckoutTransferData states the account's share of a Connect destination
// charge: what the account receives, with your platform keeping the rest of
// the sale. Send exactly one of PlatformFee (fee-first) or TransferData
// (share-first) on a destination charge; sending both, or neither, is
// rejected. On a direct charge use PlatformFee only. Amounts are decimal
// strings in the sale's base currency. Source:
// https://docs.bachs.io/connect/platform-fees
type CheckoutTransferData struct {
	// Destination is the account receiving its share.
	Destination string `json:"destination"`

	// Amount the account receives, as a decimal string.
	Amount string `json:"amount"`
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

	// PaymentMethodOptions finely restricts which payment methods a checkout
	// offers, and which currencies each one is offered in. Keys are "card",
	// "bank_transfer", "mobile_money", and "crypto". A method left out is
	// not offered; a method included with no currencies is offered in all
	// of them. If a restriction leaves nothing payable the request is
	// rejected rather than creating a checkout nobody can complete.
	PaymentMethodOptions map[string]PaymentMethodOption `json:"payment_method_options,omitempty"`

	// PaymentMethodTypes restricts which payment-method corridors appear,
	// by exact corridor name (see the Corridor* constants, for example
	// CorridorUSDCard or CorridorNGNBankTransfer). A corridor left out is
	// not offered. The restriction only ever narrows: it cannot offer a
	// corridor the account is not enabled for, and leaving the checkout
	// with nothing payable fails the request with
	// CHECKOUT_RESTRICTION_LEAVES_NO_PAYMENT_METHOD.
	PaymentMethodTypes []string `json:"payment_method_types,omitempty"`

	// PlatformFee is your platform's cut of a Connect sale, as an amount in
	// the sale's base currency — never a percentage. Fee-first: the account
	// gets the rest, and a PlatformFee record is minted. Mutually exclusive
	// with TransferData.
	PlatformFee string `json:"platform_fee,omitempty"`

	// TransferData states the account's share of a Connect destination
	// charge instead of your cut (share-first). No PlatformFee record is
	// created. Mutually exclusive with PlatformFee.
	TransferData *CheckoutTransferData `json:"transfer_data,omitempty"`

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
	// Optional since the Week of Sep 14, 2026 release: omit it and the
	// hosted checkout page collects the buyer's email and name before they
	// can pay (guest checkout). It stays required for a subscription
	// checkout, which always creates a customer record.
	Customer *CheckoutCustomer `json:"customer,omitempty"`

	// CustomerCreation decides whether a buyer who identifies themselves on
	// the hosted page also becomes a customer record: "if_required" (the
	// default) keeps them out of the directory, "always" creates the
	// record. Ignored for a subscription or setup checkout, which always
	// create a customer. Use CustomerCreationIfRequired /
	// CustomerCreationAlways so typos are caught at compile time.
	CustomerCreation string `json:"customer_creation,omitempty"`

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
// session identifiers, the hosted checkout URL, and what the session will
// collect — including its mode, so a subscription checkout is recognizable
// without a follow-up Get.
type CreateCheckoutSessionResponse struct {
	// CheckoutID uniquely identifies the underlying checkout.
	CheckoutID string `json:"checkout_id"`

	// CheckoutURL is the hosted checkout URL the customer completes payment
	// at.
	CheckoutURL string `json:"checkout_url"`

	// Status is "OPEN", "COMPLETED", "EXPIRED", or "CANCELLED". New sessions
	// start in "OPEN".
	Status string `json:"status"`

	// Mode is what the session collects: "payment" for a one-time charge,
	// "subscription" for a recurring checkout.
	Mode string `json:"mode"`

	// Currency is the base currency code.
	Currency string `json:"currency"`

	// Amount is the total amount in Currency.
	Amount string `json:"amount"`

	// Recurring describes the billing cadence. Present only for a
	// subscription checkout; null for a one-time checkout.
	Recurring *CheckoutRecurring `json:"recurring"`

	// SavePaymentMethod reports whether the payment method is kept for
	// future charges. Always true on a subscription checkout.
	SavePaymentMethod bool `json:"save_payment_method"`

	// ClientSecret is the secret for client-side confirmation flows. Null
	// on hosted checkouts.
	ClientSecret *string `json:"client_secret"`

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

	// PlatformFee echoes the fee-first split sent at creation. Null when
	// the sale carries no split — or when it was share-first (read
	// DestinationAmount instead). Always on the wire; test for null rather
	// than for field presence.
	PlatformFee *string `json:"platform_fee"`

	// DestinationAmount echoes the share-first split sent at creation.
	// Null when the sale carries no split — or when it was fee-first.
	DestinationAmount *string `json:"destination_amount"`

	// Customer attached to the checkout. Null when no customer record backs
	// the checkout (for example a guest checkout with
	// customer_creation: "if_required"); read CustomerDetails for the
	// buyer's identity instead.
	Customer *CheckoutSessionCustomer `json:"customer"`

	// CustomerDetails is what the buyer supplied: email and name, present
	// whenever an identity was collected, whether or not a customer record
	// exists for it. Read it when you want the buyer's identity and do not
	// care whether a record backs it.
	CustomerDetails *CheckoutCustomerDetails `json:"customer_details"`

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

	// Amount is the recurring amount, present on the create response.
	Amount *string `json:"amount,omitempty"`

	// TrialInterval is the trial unit ("day", "week", "month", or "year"),
	// present when the checkout carries a trial.
	TrialInterval *string `json:"trial_interval"`

	// TrialIntervalCount is the number of trial units.
	TrialIntervalCount *int `json:"trial_interval_count"`
}

// CheckoutCustomerDetails is the buyer identity collected on a checkout
// session: email and name, with or without a backing customer record.
type CheckoutCustomerDetails struct {
	// Email the buyer supplied.
	Email string `json:"email"`

	// Name the buyer supplied. Null when not collected.
	Name *string `json:"name"`
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
