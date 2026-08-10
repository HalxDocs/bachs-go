package bachs

import (
	"context"
	"net/http"
	"net/url"
)

// OrganizationService provides methods for reading your own organization and
// managing its checkout configuration. Source:
// https://docs.bachs.io/api-reference/organizations
type OrganizationService struct {
	service
}

// Organization is the shape of an organization as returned by the
// Organizations endpoints: your own organization via GetMe, or any
// organization you may read by ID. The wire shape is identical to
// ConnectedAccount — the API returns the same schema for both — so the types
// are aliased. GetMe leaves Capabilities and Requirements null; Get populates
// them.
type Organization = ConnectedAccount

// CheckoutSettings is the checkout configuration for an organization context:
// which payment methods are enabled, per-method currency toggles, and the fee
// preference.
type CheckoutSettings struct {
	// OrganizationID these settings belong to.
	OrganizationID string `json:"organization_id"`

	// EnabledPaymentMethods keyed by method, each holding an enabled flag and
	// currency toggles.
	EnabledPaymentMethods map[string]any `json:"enabled_payment_methods"`

	// FeePreference is "customer_pays" or "org_pays".
	FeePreference string `json:"fee_preference"`

	// AvailableCurrencies keyed by method, listing the currencies the method
	// can accept. Present on GetCheckoutSettings; absent from the update
	// response, which carries a message instead.
	AvailableCurrencies map[string]any `json:"available_currencies,omitempty"`

	// Message is a human-readable confirmation on the update response.
	Message string `json:"message,omitempty"`
}

// UpdateCheckoutSettingsRequest is the payload for
// Organizations.UpdateCheckoutSettings. Only the fields you send are changed.
type UpdateCheckoutSettingsRequest struct {
	// EnabledPaymentMethods keyed by method, each holding an enabled flag and
	// currency toggles.
	EnabledPaymentMethods map[string]any `json:"enabled_payment_methods,omitempty"`

	// FeePreference is "customer_pays" or "org_pays".
	FeePreference string `json:"fee_preference,omitempty"`
}

// GetMe returns the organization your API key belongs to, including its
// capability names, checkout payment methods, and balance currencies. The
// capabilities and requirements blocks are not populated here; read the
// organization by ID for those. Use this to confirm your platform holds an
// active connect capability before creating connected accounts.
func (s *OrganizationService) GetMe(ctx context.Context) (*Organization, *ResponseMeta, error) {
	var out ConnectedAccount
	meta, err := s.request(ctx, http.MethodGet, "/organizations/me", nil, &out)
	if err != nil {
		return nil, meta, err
	}
	return &out, meta, nil
}

// Get reads an organization by ID. An API key may only read the organization
// it belongs to, or a connected account named in the X-Connected-Account-ID
// header; any other ID returns 403. Unlike GetMe, this response populates
// the capabilities and requirements blocks.
func (s *OrganizationService) Get(ctx context.Context, organizationID string) (*Organization, *ResponseMeta, error) {
	var out ConnectedAccount
	meta, err := s.request(ctx, http.MethodGet, "/organizations/"+url.PathEscape(organizationID), nil, &out)
	if err != nil {
		return nil, meta, err
	}
	return &out, meta, nil
}

// GetCheckoutSettings returns the checkout configuration for your
// organization context: enabled payment methods, per-method currency
// toggles, and fee preference. Pass WithConnectedAccount to read a connected
// account's context.
func (s *OrganizationService) GetCheckoutSettings(ctx context.Context, opts ...RequestOption) (*CheckoutSettings, *ResponseMeta, error) {
	var out CheckoutSettings
	meta, err := s.request(ctx, http.MethodGet, "/organizations/checkout/settings", nil, &out, opts...)
	if err != nil {
		return nil, meta, err
	}
	return &out, meta, nil
}

// UpdateCheckoutSettings updates the checkout configuration for your
// organization context: enabled payment methods and fee preference. Pass
// WithConnectedAccount to update a connected account's context.
func (s *OrganizationService) UpdateCheckoutSettings(ctx context.Context, req UpdateCheckoutSettingsRequest, opts ...RequestOption) (*CheckoutSettings, *ResponseMeta, error) {
	var out CheckoutSettings
	meta, err := s.request(ctx, http.MethodPut, "/organizations/checkout/settings", req, &out, opts...)
	if err != nil {
		return nil, meta, err
	}
	return &out, meta, nil
}
