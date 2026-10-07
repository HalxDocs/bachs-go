package bachs

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

// Error codes for checkouts that can offer no payment method. Branch on
// these rather than on Detail, which is human-readable and may change.
// Source: the Week of Sep 7, 2026 entry in
// https://docs.bachs.io/changelog/api and
// https://docs.bachs.io/guides/payments/payment-method-support#why-a-checkout-has-no-payment-method.
const (
	// ErrCodeCheckoutHasNoPaymentMethod fires when the account has no
	// payment method configured at all. Before the Week of Sep 7, 2026
	// release it also covered the two cases below.
	ErrCodeCheckoutHasNoPaymentMethod = "CHECKOUT_HAS_NO_PAYMENT_METHOD"

	// ErrCodeAccountNotActivated fires when a live checkout can offer no
	// payment method because the account is not live yet. Sandbox is
	// unaffected: every method is available there.
	ErrCodeAccountNotActivated = "ACCOUNT_NOT_ACTIVATED"

	// ErrCodeAccountPaymentMethodsRestricted fires when a live checkout can
	// offer no payment method because the account's methods have been
	// restricted.
	ErrCodeAccountPaymentMethodsRestricted = "ACCOUNT_PAYMENT_METHODS_RESTRICTED"

	// ErrCodeCheckoutRestrictionLeavesNoPaymentMethod fires when a checkout's
	// own payment-method restriction leaves nothing payable.
	ErrCodeCheckoutRestrictionLeavesNoPaymentMethod = "CHECKOUT_RESTRICTION_LEAVES_NO_PAYMENT_METHOD"

	// ErrCodePaymentMethodNotAllowed fires when a payment names a method or
	// currency the checkout's payment_method_options restriction excluded.
	// The restriction is enforced when the payment is priced and confirmed,
	// not only on the hosted page.
	ErrCodePaymentMethodNotAllowed = "PAYMENT_METHOD_NOT_ALLOWED"

	// ErrCodeInvalidPlatformFee fires when a Connect split breaks the
	// platform-fee bounds: a non-positive fee or share, a fee or share at
	// or above the gross amount, a destination charge with neither term, or
	// a value that rounds to zero at the currency's precision.
	ErrCodeInvalidPlatformFee = "INVALID_PLATFORM_FEE"

	// ErrCodeContradictoryChargeType fires when a checkout create sends both
	// X-Account-Id and transfer_data. The header alone makes the sale the
	// account's; naming a destination makes it the platform's; a request
	// cannot be both at once.
	ErrCodeContradictoryChargeType = "CONTRADICTORY_CHARGE_TYPE"
)

// apiErrorBody is the wire shape of a Bachs error response. Source:
// https://docs.bachs.io/errors
type apiErrorBody struct {
	Detail string       `json:"detail"`
	Code   string       `json:"error_code"`
	DocURL string       `json:"doc_url,omitempty"`
	Errors []FieldError `json:"errors,omitempty"`
}

// FieldError describes a single field that failed validation. Only present on
// VALIDATION_ERROR responses.
type FieldError struct {
	// Field is the name of the field that failed validation.
	Field string `json:"field"`

	// Message is a human-readable description of the failure.
	Message string `json:"message"`

	// Type is the validation error type identifier (for example "missing").
	Type string `json:"type"`
}

// APIError is an error returned by the Bachs API. Branch on Code (a stable,
// machine-readable code such as "VALIDATION_ERROR" or "NOT_FOUND") rather
// than on Detail, which is a human-readable message that may change.
type APIError struct {
	// StatusCode is the HTTP status code of the response (401, 403, 404, 409,
	// 422, 429, 500, ...).
	StatusCode int

	// Code is the machine-readable error_code from the response body.
	Code string

	// Detail is the human-readable error message from the response body.
	Detail string

	// DocURL links to the error's entry in the error reference, when present.
	DocURL string

	// RequestID is the x-request-id header of the failed request. It comes
	// from the response headers, never the body.
	RequestID string

	// Errors holds per-field validation failures. Only populated on
	// VALIDATION_ERROR responses.
	Errors []FieldError
}

// Error implements the error interface. The API key is never part of any
// error message.
func (e *APIError) Error() string {
	msg := fmt.Sprintf("bachs: %d %s: %s", e.StatusCode, e.Code, e.Detail)
	if e.RequestID != "" {
		msg += fmt.Sprintf(" (request id: %s)", e.RequestID)
	}
	return msg
}

// apiErrorFromResponse decodes a non-2xx response body into an *APIError and
// fills in the request ID from the response headers.
func apiErrorFromResponse(resp *http.Response) error {
	apiErr := &APIError{StatusCode: resp.StatusCode, RequestID: resp.Header.Get(headerRequestID)}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		apiErr.Detail = "failed to read error response body"
		return apiErr
	}

	var wire apiErrorBody
	if err := json.Unmarshal(body, &wire); err != nil || wire.Code == "" {
		// Not a JSON error body (for example a proxy or gateway page): keep a
		// generic code so callers can still branch on the status.
		apiErr.Code = http.StatusText(resp.StatusCode)
		if apiErr.Code == "" {
			apiErr.Code = fmt.Sprintf("HTTP %d", resp.StatusCode)
		}
		if wire.Detail != "" {
			apiErr.Detail = wire.Detail
		} else if apiErr.Detail == "" {
			apiErr.Detail = "request failed"
		}
		return apiErr
	}

	apiErr.Code = wire.Code
	apiErr.Detail = wire.Detail
	apiErr.DocURL = wire.DocURL
	apiErr.Errors = wire.Errors
	return apiErr
}
