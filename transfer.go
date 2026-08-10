package bachs

import (
	"context"
	"net/http"
	"net/url"
	"time"
)

// TransferService provides methods for moving funds between your platform
// balance and the connected accounts you own. Transfers draw on available
// balance only, move a single currency, and never take a balance below zero.
// Source: https://docs.bachs.io/connect/transfers
type TransferService struct {
	service
}

// CreateTransferRequest is the payload for Transfers.Create. Source:
// https://docs.bachs.io/api-reference/transfers/create-a-transfer
type CreateTransferRequest struct {
	// Destination is the connected account to credit, or "self" to send funds
	// back to your platform when acting as a connected account (see
	// WithConnectedAccount).
	Destination string `json:"destination"`

	// Amount to move as a decimal string in Currency, for example "7000.00".
	// Must be greater than zero.
	Amount string `json:"amount"`

	// Currency is the three-letter ISO 4217 code, for example "NGN". Both
	// balances must already hold this currency.
	Currency string `json:"currency"`

	// Description is an arbitrary string returned unchanged, useful for
	// naming the order or invoice the share belongs to.
	Description *string `json:"description,omitempty"`

	// Metadata is key-value data returned unchanged, not used for processing.
	Metadata map[string]any `json:"metadata,omitempty"`

	// TransferGroup tags the transfer as part of a group so several shares
	// funded by the same charge can be reconciled together.
	TransferGroup *string `json:"transfer_group,omitempty"`
}

// Transfer is one movement of funds between a platform and a connected
// account. Source: https://docs.bachs.io/api-reference/transfers/get-a-transfer
type Transfer struct {
	// ID is the unique identifier for the transfer.
	ID string `json:"id"`

	// Source is the organization debited: your platform when sending a share
	// to a connected account, the connected account when recovering one.
	Source string `json:"source"`

	// Destination is the organization credited.
	Destination string `json:"destination"`

	// Amount moved, as a decimal string in Currency.
	Amount string `json:"amount"`

	// Currency the transfer moved.
	Currency string `json:"currency"`

	// Status is "paid" (balances updated) or "pending".
	Status string `json:"status"`

	// Description supplied on creation, or null.
	Description *string `json:"description"`

	// Metadata supplied on creation, returned unchanged.
	Metadata map[string]any `json:"metadata"`

	// TransferGroup supplied on creation, or null when untagged.
	TransferGroup *string `json:"transfer_group"`

	// CreatedAt is when the transfer was created, ISO 8601 in UTC.
	CreatedAt time.Time `json:"created_at"`
}

// Create moves funds from your platform balance to a connected account. Pass
// WithConnectedAccount to act as a connected account (for example sending
// funds back to your platform with destination "self"). The source balance is
// debited immediately and the transfer cannot be cancelled.
func (s *TransferService) Create(ctx context.Context, req CreateTransferRequest, opts ...RequestOption) (*Transfer, *ResponseMeta, error) {
	var out Transfer
	meta, err := s.request(ctx, http.MethodPost, "/transfers", req, &out, opts...)
	if err != nil {
		return nil, meta, err
	}
	return &out, meta, nil
}

// Get retrieves a single transfer your platform was a party to.
func (s *TransferService) Get(ctx context.Context, transferID string) (*Transfer, *ResponseMeta, error) {
	var out Transfer
	meta, err := s.request(ctx, http.MethodGet, "/transfers/"+url.PathEscape(transferID), nil, &out)
	if err != nil {
		return nil, meta, err
	}
	return &out, meta, nil
}

// List returns the transfers your platform was a party to, newest first.
// Filter with ListParams.ConnectedAccountID to see one account's transfers.
func (s *TransferService) List(ctx context.Context, params ListParams) (*Page[Transfer], *ResponseMeta, error) {
	var env pageEnvelope[Transfer]
	meta, err := s.request(ctx, http.MethodGet, queryPath("/transfers", params), nil, &env)
	if err != nil {
		return nil, meta, err
	}
	return env.page(), meta, nil
}
