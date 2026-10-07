package bachs

import (
	"context"
	"net/http"
	"net/url"
	"time"
)

// VirtualAccountService provides methods for issuing and reading fixed bank
// account numbers (virtual accounts) for your platform or a connected
// account. Each deposit into a virtual account becomes a payment reported
// through the collection.succeeded webhook with a null checkout_id. Creating
// twice for the same currency returns the same number rather than a second
// one. NGN is the only currency issued today. Source:
// https://docs.bachs.io/for-you/virtual-accounts and the Week of Sep 21,
// 2026 entry in https://docs.bachs.io/changelog/api.
//
// Your API key needs the virtual_accounts:write scope to create a number and
// virtual_accounts:read to read one back (write includes read). Existing keys
// do not gain these scopes automatically; without them the API returns 403.
// Pass WithConnectedAccount to issue or read a connected account's number;
// otherwise the call acts on your own account.
type VirtualAccountService struct {
	service
}

// VirtualAccount is a fixed bank account number issued to an account. It does
// not expire, and anyone can send money to it at any time.
type VirtualAccount struct {
	// ID uniquely identifies the virtual account (for example "va_...").
	ID string `json:"id"`

	// Currency the number receives in. "NGN" is the only currency issued
	// today.
	Currency string `json:"currency"`

	// AccountNumber is the fixed bank account number to give out. Show it
	// together with BankName.
	AccountNumber string `json:"account_number"`

	// BankName is the bank holding the number.
	BankName string `json:"bank_name"`

	// BankCode is the bank's routing code.
	BankCode string `json:"bank_code"`

	// Status of the virtual account (for example "active").
	Status string `json:"status"`

	// CreatedAt is when the number was issued.
	CreatedAt time.Time `json:"created_at"`
}

// CreateVirtualAccountRequest is the payload for VirtualAccounts.Create. Name
// the currency to issue a number in.
type CreateVirtualAccountRequest struct {
	// Currency to issue the number in. "NGN" is the only currency issued
	// today. Required.
	Currency string `json:"currency"`
}

// Create issues a fixed bank account number for the given currency. Calling
// it a second time with the same currency returns the account you already
// hold rather than a second number, so a retry after a timeout never leaves
// you with two numbers. Idempotency keys are supported: pass
// WithIdempotencyKey to make retries safe.
func (s *VirtualAccountService) Create(ctx context.Context, req CreateVirtualAccountRequest, opts ...RequestOption) (*VirtualAccount, *ResponseMeta, error) {
	var out VirtualAccount
	meta, err := s.request(ctx, http.MethodPost, "/virtual-accounts", req, &out, opts...)
	if err != nil {
		return nil, meta, err
	}
	return &out, meta, nil
}

// Get reads back the virtual account for one currency. Currency defaults to
// "NGN" when omitted. A currency with no issued number returns a 404
// NOT_FOUND error rather than a null body.
func (s *VirtualAccountService) Get(ctx context.Context, currency string, opts ...RequestOption) (*VirtualAccount, *ResponseMeta, error) {
	path := "/virtual-accounts"
	if currency != "" {
		path += "?currency=" + url.QueryEscape(currency)
	}

	var out VirtualAccount
	meta, err := s.request(ctx, http.MethodGet, path, nil, &out, opts...)
	if err != nil {
		return nil, meta, err
	}
	return &out, meta, nil
}
