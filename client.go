// Package bachs is an idiomatic Go SDK for the Bachs payments API.
//
// Create a client with NewClient, passing your secret API key. By default the
// client talks to the sandbox environment (https://sandbox-api.bachs.io);
// switch to production with the WithProduction option or a WithBaseURL option.
//
// Every network-calling method takes a context.Context as its first argument
// and returns the resource (or page of resources) plus a *ResponseMeta
// carrying the request ID and rate-limit headers Bachs sent back. Errors from
// the API are returned as *APIError values; inspect their Code field (for
// example "VALIDATION_ERROR") rather than parsing the human-readable Detail.
package bachs

import (
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"
)

// Environment facts, hardcoded as constants. Source:
// https://docs.bachs.io/api-reference/api-standards#base-url and
// https://docs.bachs.io/api-reference/api-standards (rate limits, headers).
const (
	// BaseURLProduction is the production API base URL.
	BaseURLProduction = "https://api.bachs.io"

	// BaseURLSandbox is the sandbox API base URL, the default for NewClient.
	BaseURLSandbox = "https://sandbox-api.bachs.io"

	// APIVersion is the versioned path segment prepended to every request
	// (requests are sent to <baseURL>/v1/<path>).
	APIVersion = "v1"

	// SandboxKeyPrefix is the prefix every sandbox API key carries. Keys are
	// opaque; the prefix is documented here for reference only.
	SandboxKeyPrefix = "sk_sandbox_"

	// LiveKeyPrefix is the prefix every production (live) API key carries.
	LiveKeyPrefix = "sk_live_"

	// RateLimitProduction is the per-key request limit in production:
	// 500 requests per minute.
	RateLimitProduction = 500

	// RateLimitSandbox is the per-key request limit in sandbox:
	// 100 requests per minute.
	RateLimitSandbox = 100

	// PaginationDefaultLimit is the default page size used by list endpoints
	// when no limit is passed (defaults vary between 20 and 50 per endpoint).
	PaginationDefaultLimit = 20

	// PaginationMaxLimit is the maximum page size. Values above it are
	// clamped by the API, not rejected.
	PaginationMaxLimit = 100
)

// HTTP header names used by the Bachs API. Source:
// https://docs.bachs.io/api-reference/api-standards (rate limits, request IDs)
// and https://docs.bachs.io/guides/idempotency (Idempotency-Key).
const (
	headerAuthorization      = "Authorization"
	headerContentType        = "Content-Type"
	headerAccept             = "Accept"
	headerIdempotencyKey     = "Idempotency-Key"
	headerConnectedAccountID = "X-Connected-Account-ID"
	headerRequestID          = "x-request-id"
	headerRateLimitLimit     = "X-RateLimit-Limit"
	headerRateLimitRemaining = "X-RateLimit-Remaining"
	headerRateLimitReset     = "X-RateLimit-Reset"
	headerRetryAfter         = "Retry-After"
)

// Client is a Bachs API client. Create one with NewClient; all configuration
// hangs off the returned *Client — there is no global state, and no shared
// package-level HTTP client.
type Client struct {
	baseURL    string
	apiKey     string
	httpClient *http.Client

	// Checkouts provides methods for creating and retrieving checkout
	// sessions.
	Checkouts *CheckoutService

	// Products provides methods for managing the billing catalog.
	Products *ProductService

	// Customers provides methods for managing customers.
	Customers *CustomerService

	// Payments provides methods for retrieving payments.
	Payments *PaymentService

	// Refunds provides methods for creating and retrieving refunds.
	Refunds *RefundService

	// Subscriptions provides methods for managing subscriptions. There is no
	// Subscriptions.Create: subscriptions are created only when a customer
	// completes a checkout for a recurring product.
	Subscriptions *SubscriptionService

	// Transfers provides methods for moving funds between your platform
	// balance and the connected accounts you own.
	Transfers *TransferService

	// Misc provides account-wide endpoints: balances, payment methods,
	// payment rails, and supported currencies.
	Misc *MiscService

	// Media provides methods for uploading files and retrieving or deleting
	// the uploads. Upload IDs attach to products via the media field.
	Media *MediaService

	// CustomerSessions provides methods for opening the customer portal as a
	// specific customer.
	CustomerSessions *CustomerSessionService

	// ConnectedAccounts provides methods for the Connect platform: creating
	// and reading connected accounts, requesting capabilities, walking their
	// onboarding Tasks, and managing account documents.
	ConnectedAccounts *ConnectedAccountService

	// Payouts provides methods for withdrawing funds to bank accounts,
	// mobile money, or crypto wallets: destinations, quotes, and withdrawals.
	Payouts *PayoutService

	// Disputes provides methods for responding to chargebacks: listing and
	// reading disputes, uploading documents, and submitting evidence.
	Disputes *DisputeService

	// Conversions provides methods for converting between settlement
	// currencies: quoting, executing, and reading conversions.
	Conversions *ConversionService

	// Organizations provides methods for reading your own organization and
	// managing its checkout configuration.
	Organizations *OrganizationService
}

// NewClient returns a Bachs API client authenticated with apiKey.
//
// The apiKey must be non-empty; an error is returned otherwise (the function
// never panics). By default the client targets the sandbox environment; pass
// WithProduction or WithBaseURL to change that. The default HTTP client has a
// 30-second timeout; pass WithHTTPClient to provide your own.
func NewClient(apiKey string, opts ...Option) (*Client, error) {
	if strings.TrimSpace(apiKey) == "" {
		return nil, errors.New("bachs: api key must not be empty")
	}

	c := &Client{
		baseURL:    BaseURLSandbox,
		apiKey:     apiKey,
		httpClient: &http.Client{Timeout: 30 * time.Second},
	}

	for _, opt := range opts {
		opt(c)
	}

	c.baseURL = strings.TrimRight(c.baseURL, "/")

	c.Checkouts = &CheckoutService{service{request: c.do}}
	c.Products = &ProductService{service{request: c.do}}
	c.Customers = &CustomerService{service{request: c.do}}
	c.Payments = &PaymentService{service{request: c.do}}
	c.Refunds = &RefundService{service{request: c.do}}
	c.Subscriptions = &SubscriptionService{service{request: c.do}}
	c.Transfers = &TransferService{service{request: c.do}}
	c.Misc = &MiscService{service{request: c.do}}
	c.Media = &MediaService{service{request: c.do}}
	c.CustomerSessions = &CustomerSessionService{service{request: c.do}}
	c.ConnectedAccounts = &ConnectedAccountService{service{request: c.do}}
	c.Payouts = &PayoutService{service{request: c.do}}
	c.Disputes = &DisputeService{service{request: c.do}}
	c.Conversions = &ConversionService{service{request: c.do}}
	c.Organizations = &OrganizationService{service{request: c.do}}

	return c, nil
}

// BaseURL returns the base URL this client sends requests to.
func (c *Client) BaseURL() string {
	return c.baseURL
}

// String returns a short description of the client without any secret
// material. The API key is never included.
func (c *Client) String() string {
	return fmt.Sprintf("bachs.Client(baseURL=%s)", c.baseURL)
}
