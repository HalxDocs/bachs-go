package bachs

import (
	"context"
	"net/http"
	"net/url"
	"time"
)

// PlatformFeeService provides methods for reading platform fees: your
// platform's cut of a Connect sale, taken from the account's proceeds. A
// fee-first split (platform_fee on create-checkout) mints a PlatformFee
// record; a share-first split (transfer_data.amount) never does — read what
// the account was paid off the transfer instead. A platform fee is not
// reversed on a refund. Source:
// https://docs.bachs.io/connect/platform-fees
type PlatformFeeService struct {
	service
}

// PlatformFee is the queryable, permanent record of your platform's cut of
// one sale: the charge it came from, the amount, and both parties. Both the
// account and your platform can read a fee they were party to.
type PlatformFee struct {
	// ID uniquely identifies the platform fee (for example "pf_...").
	ID string `json:"id"`

	// Charge is the sale the cut came from.
	Charge string `json:"charge"`

	// CollectedFrom is the account the cut came out of.
	CollectedFrom string `json:"collected_from"`

	// EarnedBy is your platform.
	EarnedBy string `json:"earned_by"`

	// Amount of the cut, as a decimal string in Currency.
	Amount string `json:"amount"`

	// Currency the cut is stated in (the sale's base currency).
	Currency string `json:"currency"`

	// AmountRefunded against the charge. The fee itself is not reversed on
	// a refund; see the guide for which party bears it.
	AmountRefunded string `json:"amount_refunded"`

	// Refunded reports whether the underlying charge was refunded.
	Refunded bool `json:"refunded"`

	// CreatedAt is when the fee was minted.
	CreatedAt time.Time `json:"created_at"`
}

// List returns every platform fee your platform was a party to, newest
// first. Set ListParams.Charge to see the one fee tied to a specific sale.
func (s *PlatformFeeService) List(ctx context.Context, params ListParams) (*Page[PlatformFee], *ResponseMeta, error) {
	var env pageEnvelope[PlatformFee]
	meta, err := s.request(ctx, http.MethodGet, queryPath("/platform_fees", params), nil, &env)
	if err != nil {
		return nil, meta, err
	}
	return env.page(), meta, nil
}

// Get reads one platform fee by ID.
func (s *PlatformFeeService) Get(ctx context.Context, feeID string) (*PlatformFee, *ResponseMeta, error) {
	var out PlatformFee
	meta, err := s.request(ctx, http.MethodGet, "/platform_fees/"+url.PathEscape(feeID), nil, &out)
	if err != nil {
		return nil, meta, err
	}
	return &out, meta, nil
}
