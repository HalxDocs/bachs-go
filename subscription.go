package bachs

import (
	"context"
	"errors"
	"net/http"
	"net/url"
	"time"
)

// SubscriptionService provides methods for managing subscriptions.
//
// There is no Subscriptions.Create method, and there never will be: the Bachs
// API has no create-subscription endpoint. A subscription is created only
// when a customer completes a checkout session for a recurring product (one
// with a billing_cycle). Create a recurring product, send the customer
// through a checkout, and react to the customer.subscription.created webhook
// instead. See https://docs.bachs.io/guides/subscriptions/overview.
type SubscriptionService struct {
	service
}

// Subscription is a customer's ongoing relationship with a recurring product.
// It is created when the customer completes a checkout for a recurring
// product and renews automatically until canceled. Source:
// https://docs.bachs.io/api-reference/subscriptions/object
type Subscription struct {
	// ID is the unique identifier for the subscription.
	ID string `json:"id"`

	// Customer is the customer this subscription bills.
	Customer Customer `json:"customer"`

	// PaymentMethodID is the saved payment method billed on each renewal.
	// Null until a payment method is attached.
	PaymentMethodID *string `json:"payment_method_id"`

	// Status is "trialing", "active", "past_due", "unpaid", "canceled", or
	// "paused".
	Status string `json:"status"`

	// CollectionMethod is how renewals are collected; always
	// "charge_automatically" today.
	CollectionMethod string `json:"collection_method"`

	// Currency the subscription is billed in (USD only today).
	Currency string `json:"currency"`

	// Amount is the recurring amount as a decimal string.
	Amount string `json:"amount"`

	// BillingCycle is the recurring cadence.
	BillingCycle Cadence `json:"billing_cycle"`

	// Quantity is the total billable quantity across the line items.
	Quantity int `json:"quantity"`

	// CurrentPeriodStart is the start of the period being billed for, UTC.
	CurrentPeriodStart time.Time `json:"current_period_start"`

	// CurrentPeriodEnd is the end of the period being billed for, UTC.
	CurrentPeriodEnd time.Time `json:"current_period_end"`

	// PreviouslyBilledAt is the start of the period that was last billed.
	PreviouslyBilledAt *time.Time `json:"previously_billed_at"`

	// NextBilledAt is the next scheduled charge date.
	NextBilledAt *time.Time `json:"next_billed_at"`

	// TrialEnd is when the free trial ends and billing begins. Null when not
	// trialing.
	TrialEnd *time.Time `json:"trial_end"`

	// CancelAtPeriodEnd is true when the subscription stays active until
	// CurrentPeriodEnd and is not renewed.
	CancelAtPeriodEnd bool `json:"cancel_at_period_end"`

	// CanceledAt is when the subscription was canceled. Null when it has not
	// been canceled.
	CanceledAt *time.Time `json:"canceled_at"`

	// CreatedAt is when the subscription was created, UTC.
	CreatedAt time.Time `json:"created_at"`

	// Product is the recurring product this subscription bills for.
	Product *SubscriptionCatalogProduct `json:"product"`

	// Items are the line items that make up the subscription.
	Items []SubscriptionItem `json:"items"`

	// Metadata attached at creation, returned unchanged.
	Metadata map[string]any `json:"metadata"`
}

// SubscriptionCatalogProduct is the recurring product a subscription bills
// for.
type SubscriptionCatalogProduct struct {
	// ID is the unique identifier for the product.
	ID string `json:"id"`

	// Name shown to customers at checkout.
	Name string `json:"name"`

	// Description, or null when none was set.
	Description *string `json:"description"`

	// Status is "active" or "archived".
	Status string `json:"status"`

	// BillingCycle of the product.
	BillingCycle *Cadence `json:"billing_cycle"`

	// TrialPeriod of the product, if any.
	TrialPeriod *Cadence `json:"trial_period"`

	// CreatedAt is when the product was created, UTC.
	CreatedAt time.Time `json:"created_at"`

	// UpdatedAt is when the product was last updated, UTC.
	UpdatedAt time.Time `json:"updated_at"`
}

// SubscriptionItem is one line item of a subscription, tying a product and
// its price to a billed quantity.
type SubscriptionItem struct {
	// ID is the unique identifier for the line item.
	ID string `json:"id"`

	// Status follows the parent subscription's status.
	Status string `json:"status"`

	// Quantity is the billed quantity for this item.
	Quantity int `json:"quantity"`

	// Recurring is always true for subscription items.
	Recurring bool `json:"recurring"`

	// PriceType is "fixed", "free", or "custom".
	PriceType string `json:"price_type"`

	// UnitAmount is the price for one unit, as a decimal string.
	UnitAmount string `json:"unit_amount"`

	// Currency this item is billed in.
	Currency string `json:"currency"`

	// PreviouslyBilledAt is when this item was last billed. Null if it has
	// not been billed yet.
	PreviouslyBilledAt *time.Time `json:"previously_billed_at"`

	// NextBilledAt is when this item will next be billed.
	NextBilledAt *time.Time `json:"next_billed_at"`

	// Price is the price this item bills at.
	Price *SubscriptionItemPrice `json:"price"`

	// Product is the product this item belongs to.
	Product *SubscriptionCatalogProduct `json:"product"`

	// CreatedAt is when the item was created, UTC.
	CreatedAt time.Time `json:"created_at"`

	// UpdatedAt is when the item was last updated, UTC.
	UpdatedAt time.Time `json:"updated_at"`
}

// SubscriptionItemPrice is the price a subscription line item bills at.
type SubscriptionItemPrice struct {
	// ID is the unique identifier for the price.
	ID string `json:"id"`

	// ProductID is the product this price belongs to.
	ProductID string `json:"product_id"`

	// PriceType is "fixed", "free", or "custom".
	PriceType string `json:"price_type"`

	// Currency of this price.
	Currency string `json:"currency"`

	// UnitAmount is the price as a decimal string.
	UnitAmount string `json:"unit_amount"`

	// BillingCycle of this price, if recurring.
	BillingCycle *Cadence `json:"billing_cycle"`

	// TrialPeriod of this price, if any.
	TrialPeriod *Cadence `json:"trial_period"`

	// SeatTiers is reserved for seat-based pricing; null today.
	SeatTiers map[string]any `json:"seat_tiers"`

	// IsArchived reports whether the price has been archived. Archived prices
	// keep billing existing subscribers but are not offered for new checkouts.
	IsArchived bool `json:"is_archived"`

	// CreatedAt is when the price was created, UTC.
	CreatedAt time.Time `json:"created_at"`

	// UpdatedAt is when the price was last updated, UTC.
	UpdatedAt time.Time `json:"updated_at"`
}

// UpdateSubscriptionRequest is the payload for Subscriptions.Update. Exactly
// one intent must be set per request: change the plan (ProductID), move the
// trial (TrialEnd), change the payment method (PaymentMethodID), or update
// metadata (Metadata). Combining intents — or sending none — returns an error
// before any request is made, matching the API's "combining intents returns
// 400" rule. Source:
// https://docs.bachs.io/api-reference/subscriptions/update-subscription
type UpdateSubscriptionRequest struct {
	// ProductID moves the subscription to this product (plan). The price is
	// resolved from the product for the subscription's currency.
	ProductID string `json:"product_id,omitempty"`

	// TrialEnd moves the trial: a future time adds or extends it; a past or
	// present time ends it and bills now.
	TrialEnd *time.Time `json:"trial_end,omitempty"`

	// PaymentMethodID points the subscription at a different saved card. If
	// past_due or unpaid, retries immediately. Stands alone.
	PaymentMethodID string `json:"payment_method_id,omitempty"`

	// Metadata merges key-value metadata (up to 20 keys total). Sent keys are
	// added or overwritten; a key sent with an empty-string value is removed;
	// set Metadata to the empty string ("") to clear all metadata. Stands
	// alone.
	Metadata any `json:"metadata,omitempty"`

	// ProrationBehavior is how a plan change is settled: "invoice_now"
	// (default), "next_cycle", or "none". Only valid with ProductID.
	ProrationBehavior string `json:"proration_behavior,omitempty"`
}

// validate enforces the exactly-one-intent rule locally so a request with
// combined or missing intents fails fast with a clear error instead of
// reaching the API.
func (r UpdateSubscriptionRequest) validate() error {
	if r.ProrationBehavior != "" && r.ProductID == "" {
		return errors.New("bachs: proration_behavior only applies to a plan change (product_id)")
	}

	intents := 0
	if r.ProductID != "" {
		intents++
	}
	if r.TrialEnd != nil {
		intents++
	}
	if r.PaymentMethodID != "" {
		intents++
	}
	if r.Metadata != nil {
		intents++
	}

	switch {
	case intents == 0:
		return errors.New("bachs: UpdateSubscriptionRequest needs exactly one intent: set product_id, trial_end, payment_method_id, or metadata")
	case intents > 1:
		return errors.New("bachs: UpdateSubscriptionRequest combines multiple intents: set exactly one of product_id, trial_end, payment_method_id, or metadata")
	}
	return nil
}

// CancelSubscriptionRequest is the payload for Subscriptions.Cancel. Source:
// https://docs.bachs.io/api-reference/subscriptions/cancel-subscription
type CancelSubscriptionRequest struct {
	// CancelAtPeriodEnd is true to cancel at the current period end, false to
	// cancel immediately.
	CancelAtPeriodEnd bool `json:"cancel_at_period_end"`

	// Reason is an optional free-text note (max 255 characters).
	Reason *string `json:"reason,omitempty"`
}

// Get retrieves a single subscription with its product, price, items, and
// billing dates.
func (s *SubscriptionService) Get(ctx context.Context, subscriptionID string) (*Subscription, *ResponseMeta, error) {
	var out Subscription
	meta, err := s.request(ctx, http.MethodGet, "/subscriptions/"+url.PathEscape(subscriptionID), nil, &out)
	if err != nil {
		return nil, meta, err
	}
	return &out, meta, nil
}

// List returns a paginated list of subscriptions, newest first. Filter with
// ListParams.CustomerID or ListParams.Status ("trialing", "active",
// "past_due", "unpaid", "canceled").
func (s *SubscriptionService) List(ctx context.Context, params ListParams) (*Page[Subscription], *ResponseMeta, error) {
	var env pageEnvelope[Subscription]
	meta, err := s.request(ctx, http.MethodGet, queryPath("/subscriptions", params), nil, &env)
	if err != nil {
		return nil, meta, err
	}
	return env.page(), meta, nil
}

// Update changes a subscription. The request must carry exactly one intent
// (change the plan, move the trial, change the payment method, or update
// metadata); combining intents is rejected before the request is sent.
func (s *SubscriptionService) Update(ctx context.Context, subscriptionID string, req UpdateSubscriptionRequest) (*Subscription, *ResponseMeta, error) {
	if err := req.validate(); err != nil {
		return nil, nil, err
	}

	var out Subscription
	meta, err := s.request(ctx, http.MethodPatch, "/subscriptions/"+url.PathEscape(subscriptionID), req, &out)
	if err != nil {
		return nil, meta, err
	}
	return &out, meta, nil
}

// Cancel cancels a subscription immediately, or at the end of the current
// period when CancelAtPeriodEnd is true. Returns the updated subscription.
func (s *SubscriptionService) Cancel(ctx context.Context, subscriptionID string, req CancelSubscriptionRequest) (*Subscription, *ResponseMeta, error) {
	var out Subscription
	meta, err := s.request(ctx, http.MethodDelete, "/subscriptions/"+url.PathEscape(subscriptionID), req, &out)
	if err != nil {
		return nil, meta, err
	}
	return &out, meta, nil
}
