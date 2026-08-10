package bachs

import (
	"context"
	"net/http"
	"net/url"
	"time"
)

// RefundService provides methods for refunding payments you already collected.
// A refund returns money to a customer; it settles asynchronously and is
// reported through the refund.* webhook events. Source:
// https://docs.bachs.io/api-reference/refunds/object
type RefundService struct {
	service
}

// CreateRefundRequest is the payload for Refunds.Create. Only one refund can
// be created per charge. Source:
// https://docs.bachs.io/api-reference/refunds/create-refund
type CreateRefundRequest struct {
	// ChargeID is the ID of the payment to refund. Required.
	ChargeID string `json:"charge_id"`

	// Reference is your unique identifier for this refund, unique per
	// organization and environment. Required.
	Reference string `json:"reference"`

	// RefundAddress is the destination wallet address for crypto refunds.
	// Required when the charge currency is a cryptocurrency.
	RefundAddress *string `json:"refund_address,omitempty"`

	// Amount is an optional partial refund amount in the charge settlement
	// currency. Omit to refund the full remaining refundable balance.
	Amount *string `json:"amount,omitempty"`

	// FeeBearer is who bears the refund processing fee: "org" (default) or
	// "customer".
	FeeBearer *string `json:"fee_bearer,omitempty"`

	// Reason is a human-readable reason for the refund.
	Reason *string `json:"reason,omitempty"`

	// IdempotencyKey makes the request idempotent: sending the same key twice
	// for the same charge does not create a second refund.
	IdempotencyKey *string `json:"idempotency_key,omitempty"`

	// SimulatedOutcome is test-mode only: "success" or "failed", to force a
	// specific refund outcome. Omit to use the default sandbox outcome.
	SimulatedOutcome *string `json:"simulated_outcome,omitempty"`
}

// Refund is a refund of a payment, returned by Refunds.Create, Get,
// GetByCharge, and List.
type Refund struct {
	// RefundID is the unique identifier for the refund.
	RefundID string `json:"refund_id"`

	// ChargeID is the payment this refund is associated with.
	ChargeID string `json:"charge_id"`

	// Reference is the reference supplied on creation.
	Reference string `json:"reference"`

	// Status is "processing", "success", or "failed".
	Status string `json:"status"`

	// RequestedAmount is the refund amount requested, in the charge's
	// settlement currency.
	RequestedAmount string `json:"requested_amount"`

	// RefundedAmount is the amount actually returned to the customer. Null
	// until the refund completes or partially settles.
	RefundedAmount *string `json:"refunded_amount"`

	// RefundFeeAmount is the fee charged for the refund, in the charge's
	// settlement currency. "0" when no fee applies.
	RefundFeeAmount string `json:"refund_fee_amount"`

	// FeeBearer is "org" (merchant absorbs the fee) or "customer" (deducted
	// from the refund).
	FeeBearer string `json:"fee_bearer"`

	// Reason supplied on creation, or null when none was given.
	Reason *string `json:"reason"`

	// CreatedAt is when the refund was created.
	CreatedAt time.Time `json:"created_at"`

	// UpdatedAt is when the refund was last updated.
	UpdatedAt time.Time `json:"updated_at"`

	// CompletedAt is when the refund reached a terminal status. Null while
	// still processing.
	CompletedAt *time.Time `json:"completed_at"`
}

// Create refunds a completed payment, in full or in part.
func (s *RefundService) Create(ctx context.Context, req CreateRefundRequest, opts ...RequestOption) (*Refund, *ResponseMeta, error) {
	var out Refund
	meta, err := s.request(ctx, http.MethodPost, "/refunds", req, &out, opts...)
	if err != nil {
		return nil, meta, err
	}
	return &out, meta, nil
}

// Get retrieves a single refund by its ID.
func (s *RefundService) Get(ctx context.Context, refundID string) (*Refund, *ResponseMeta, error) {
	var out Refund
	meta, err := s.request(ctx, http.MethodGet, "/refunds/"+url.PathEscape(refundID), nil, &out)
	if err != nil {
		return nil, meta, err
	}
	return &out, meta, nil
}

// GetByCharge retrieves the refund associated with a specific payment by its
// ID.
func (s *RefundService) GetByCharge(ctx context.Context, paymentID string) (*Refund, *ResponseMeta, error) {
	var out Refund
	meta, err := s.request(ctx, http.MethodGet, "/refunds/by-charge/"+url.PathEscape(paymentID), nil, &out)
	if err != nil {
		return nil, meta, err
	}
	return &out, meta, nil
}

// List returns a paginated list of refunds, most recent first. Set
// ListParams.Status to filter ("PROCESSING", "SUCCESS", or "FAILED").
func (s *RefundService) List(ctx context.Context, params ListParams) (*Page[Refund], *ResponseMeta, error) {
	var env pageEnvelope[Refund]
	meta, err := s.request(ctx, http.MethodGet, queryPath("/refunds", params), nil, &env)
	if err != nil {
		return nil, meta, err
	}
	return env.page(), meta, nil
}
