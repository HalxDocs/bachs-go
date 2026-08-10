package bachs

import (
	"context"
	"net/http"
	"net/url"
)

// CustomerSessionService provides methods for opening the customer portal as
// a specific customer. Source:
// https://docs.bachs.io/guides/customer-portal/create-portal-session
type CustomerSessionService struct {
	service
}

// CustomerPortalSession is a short-lived, pre-authenticated session that opens
// the customer portal as one specific customer.
type CustomerPortalSession struct {
	// ID is the session identifier, prefixed with "psn_". Use it to correlate
	// the session with your own logs; it is not a credential.
	ID string `json:"id"`

	// URL opens the portal as this customer. It carries the session
	// credential, so it works on any device and must not be logged or shared.
	URL string `json:"url"`
}

// Create mints a pre-authenticated customer portal session for the customer
// and returns the URL to redirect them to. Sessions are short-lived; create a
// fresh one each time a customer asks to manage their billing.
func (s *CustomerSessionService) Create(ctx context.Context, customerID string) (*CustomerPortalSession, *ResponseMeta, error) {
	var out CustomerPortalSession
	meta, err := s.request(ctx, http.MethodPost, "/customers/"+url.PathEscape(customerID)+"/portal-sessions", nil, &out)
	if err != nil {
		return nil, meta, err
	}
	return &out, meta, nil
}
