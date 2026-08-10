package bachs

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"time"
)

// errIdempotencyNonPost is returned when WithIdempotencyKey is used on a
// request that is not a POST. The API ignores Idempotency-Key everywhere
// except POST, so sending it on other methods would be misleading.
var errIdempotencyNonPost = errors.New("bachs: idempotency keys are only supported on POST requests")

// requestConfig carries per-request settings applied through RequestOptions.
type requestConfig struct {
	// idempotencyKey, when set, is sent as the Idempotency-Key header.
	idempotencyKey string
}

// RequestOption configures a single API request.
type RequestOption func(*requestConfig)

// WithIdempotencyKey attaches an Idempotency-Key header to the request, so a
// retried request is only applied once: Bachs caches the 2xx response for 24
// hours per key and returns it for any later request with the same key and
// body. Reusing a key with a different body returns a 409 IDEMPOTENCY_CONFLICT.
//
// Idempotency-Key is only honored on POST requests. Passing this option to a
// GET, PATCH, or DELETE request is an error, enforced by the request builder.
func WithIdempotencyKey(key string) RequestOption {
	return func(cfg *requestConfig) {
		cfg.idempotencyKey = key
	}
}

// requestFunc is the client's internal request method, injected into each
// service. Services hold this function rather than the *Client, so they can
// issue API calls without exposing the client's internals.
type requestFunc func(ctx context.Context, method, path string, body any, out any, opts ...RequestOption) (*ResponseMeta, error)

// service is embedded by every resource service. It carries only the request
// function — never the *Client.
type service struct {
	request requestFunc
}

// ResponseMeta holds per-response metadata Bachs sends in headers: the request
// ID, for support, and the current rate-limit budget.
type ResponseMeta struct {
	// RequestID is the x-request-id header value. Bachs includes it on every
	// response, success or error.
	RequestID string

	// RateLimitLimit is the maximum requests allowed in the current window
	// (X-RateLimit-Limit).
	RateLimitLimit int

	// RateLimitRemaining is the requests remaining in the current window
	// (X-RateLimit-Remaining).
	RateLimitRemaining int

	// RateLimitReset is when the rate-limit window resets, parsed from the
	// X-RateLimit-Reset header (a unix timestamp in seconds). The zero
	// time.Time means the header was absent or unparseable.
	RateLimitReset time.Time
}

// do runs one API request through the pipeline:
//
//  1. build the URL from the base URL, the API version, and path;
//  2. JSON-encode body, if non-nil;
//  3. set the Authorization, Content-Type, and Accept headers;
//  4. apply RequestOptions (the idempotency key is rejected for non-POST
//     methods here, in the request builder);
//  5. execute via the client's HTTP client;
//  6. on a non-2xx response, decode the body into an APIError, populate its
//     request ID and rate-limit fields from headers, and return it as an error;
//  7. on a 2xx response, decode the body into out (skipped entirely on 204);
//  8. always return a *ResponseMeta with the request ID and rate-limit info
//     from the headers, even on success.
func (c *Client) do(ctx context.Context, method, path string, body any, out any, opts ...RequestOption) (*ResponseMeta, error) {
	url := c.baseURL + "/" + APIVersion + path

	var bodyBytes []byte
	if body != nil {
		b, err := json.Marshal(body)
		if err != nil {
			return nil, fmt.Errorf("bachs: encode request body: %w", err)
		}
		bodyBytes = b
	}

	var cfg requestConfig
	for _, opt := range opts {
		if opt != nil {
			opt(&cfg)
		}
	}

	req, err := http.NewRequestWithContext(ctx, method, url, bytes.NewReader(bodyBytes))
	if err != nil {
		return nil, fmt.Errorf("bachs: build request: %w", err)
	}

	req.Header.Set(headerAuthorization, "Bearer "+c.apiKey)
	req.Header.Set(headerAccept, "application/json")

	if len(bodyBytes) > 0 {
		req.Header.Set(headerContentType, "application/json")
	}

	// Idempotency-Key is only ever attached to POST requests. Enforced here,
	// in the request builder, not by convention in each service.
	if cfg.idempotencyKey != "" {
		if method != http.MethodPost {
			return nil, fmt.Errorf("%w (got %s)", errIdempotencyNonPost, method)
		}
		req.Header.Set(headerIdempotencyKey, cfg.idempotencyKey)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("bachs: send request: %w", err)
	}
	defer resp.Body.Close()

	meta := responseMetaFromHeaders(resp)

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return meta, apiErrorFromResponse(resp)
	}

	if out != nil && resp.StatusCode != http.StatusNoContent {
		if err := json.NewDecoder(resp.Body).Decode(out); err != nil {
			return meta, fmt.Errorf("bachs: decode response: %w", err)
		}
	}
	return meta, nil
}

// responseMetaFromHeaders extracts the request ID and rate-limit headers from
// a response. Present on every response, success or error.
func responseMetaFromHeaders(resp *http.Response) *ResponseMeta {
	meta := &ResponseMeta{RequestID: resp.Header.Get(headerRequestID)}

	if v := resp.Header.Get(headerRateLimitLimit); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			meta.RateLimitLimit = n
		}
	}
	if v := resp.Header.Get(headerRateLimitRemaining); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			meta.RateLimitRemaining = n
		}
	}
	if v := resp.Header.Get(headerRateLimitReset); v != "" {
		if sec, err := strconv.ParseInt(v, 10, 64); err == nil {
			meta.RateLimitReset = time.Unix(sec, 0).UTC()
		}
	}
	return meta
}
