package bachs

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"
	"time"
)

// newTestServer starts an httptest server and returns a client wired to it.
func newTestServer(t *testing.T, h http.HandlerFunc) *httptest.Server {
	t.Helper()
	srv := httptest.NewServer(h)
	t.Cleanup(srv.Close)
	return srv
}

func newTestClient(t *testing.T, h http.HandlerFunc) *Client {
	t.Helper()
	c, err := NewClient("sk_sandbox_test", WithBaseURL(newTestServer(t, h).URL))
	if err != nil {
		t.Fatalf("NewClient returned error: %v", err)
	}
	return c
}

// writeJSON is a helper that sets headers and writes a JSON body.
func writeJSON(t *testing.T, w http.ResponseWriter, status int, body any) {
	t.Helper()
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(body); err != nil {
		t.Errorf("encode response: %v", err)
	}
}

// testHeaders returns the set of response headers every Bachs response carries.
func testHeaders(requestID string) http.Header {
	h := http.Header{}
	h.Set(headerRequestID, requestID)
	h.Set(headerRateLimitLimit, "100")
	h.Set(headerRateLimitRemaining, "99")
	h.Set(headerRateLimitReset, "1783976340")
	return h
}

func TestDoGet(t *testing.T) {
	const requestID = "2f31edcd-0bba-4a89-98b9-533921e42f26"
	c := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Errorf("method = %s, want GET", r.Method)
		}
		if r.URL.Path != "/v1/things" {
			t.Errorf("path = %q, want /v1/things", r.URL.Path)
		}
		if auth := r.Header.Get(headerAuthorization); auth != "Bearer sk_sandbox_test" {
			t.Errorf("Authorization = %q, want %q", auth, "Bearer sk_sandbox_test")
		}
		if accept := r.Header.Get(headerAccept); accept != "application/json" {
			t.Errorf("Accept = %q, want application/json", accept)
		}
		w.Header().Set("Content-Type", "application/json")
		for k, vv := range testHeaders(requestID) {
			for _, v := range vv {
				w.Header().Add(k, v)
			}
		}
		w.WriteHeader(http.StatusOK)
		io.WriteString(w, `{"name": "thing"}`)
	})

	var out struct {
		Name string `json:"name"`
	}
	meta, err := c.do(context.Background(), http.MethodGet, "/things", nil, &out)
	if err != nil {
		t.Fatalf("do returned error: %v", err)
	}
	if out.Name != "thing" {
		t.Errorf("decoded Name = %q, want %q", out.Name, "thing")
	}
	if meta.RequestID != requestID {
		t.Errorf("RequestID = %q, want %q", meta.RequestID, requestID)
	}
	if meta.RateLimitLimit != 100 {
		t.Errorf("RateLimitLimit = %d, want 100", meta.RateLimitLimit)
	}
	if meta.RateLimitRemaining != 99 {
		t.Errorf("RateLimitRemaining = %d, want 99", meta.RateLimitRemaining)
	}
	wantReset := time.Unix(1783976340, 0).UTC()
	if !meta.RateLimitReset.Equal(wantReset) {
		t.Errorf("RateLimitReset = %v, want %v", meta.RateLimitReset, wantReset)
	}
}

func TestDoPost(t *testing.T) {
	c := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("method = %s, want POST", r.Method)
		}
		if ct := r.Header.Get(headerContentType); ct != "application/json" {
			t.Errorf("Content-Type = %q, want application/json", ct)
		}
		body, _ := io.ReadAll(r.Body)
		var got map[string]any
		if err := json.Unmarshal(body, &got); err != nil {
			t.Fatalf("request body is not valid JSON: %v", err)
		}
		if got["name"] != "widget" || got["amount"] != "29.00" {
			t.Errorf("request body = %s", body)
		}
		writeJSON(t, w, http.StatusCreated, map[string]any{"id": "prod_123", "name": "widget"})
	})

	var out struct {
		ID   string `json:"id"`
		Name string `json:"name"`
	}
	_, err := c.do(context.Background(), http.MethodPost, "/products", map[string]any{
		"name":   "widget",
		"amount": "29.00",
	}, &out)
	if err != nil {
		t.Fatalf("do returned error: %v", err)
	}
	if out.ID != "prod_123" {
		t.Errorf("decoded ID = %q, want prod_123", out.ID)
	}
}

func TestDoPatch(t *testing.T) {
	c := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPatch {
			t.Errorf("method = %s, want PATCH", r.Method)
		}
		body, _ := io.ReadAll(r.Body)
		if !strings.Contains(string(body), `"description"`) {
			t.Errorf("PATCH body = %s, want description field", body)
		}
		writeJSON(t, w, http.StatusOK, map[string]any{"id": "prod_123", "description": "updated"})
	})

	var out struct {
		ID          string `json:"id"`
		Description string `json:"description"`
	}
	_, err := c.do(context.Background(), http.MethodPatch, "/products/prod_123", map[string]any{
		"description": "updated",
	}, &out)
	if err != nil {
		t.Fatalf("do returned error: %v", err)
	}
	if out.Description != "updated" {
		t.Errorf("decoded Description = %q, want updated", out.Description)
	}
}

func TestDoDelete(t *testing.T) {
	c := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodDelete {
			t.Errorf("method = %s, want DELETE", r.Method)
		}
		writeJSON(t, w, http.StatusOK, map[string]any{"upload_id": "upl_1", "deleted": true})
	})

	var out struct {
		UploadID string `json:"upload_id"`
		Deleted  bool   `json:"deleted"`
	}
	_, err := c.do(context.Background(), http.MethodDelete, "/utilities/uploads/upl_1", nil, &out)
	if err != nil {
		t.Fatalf("do returned error: %v", err)
	}
	if !out.Deleted {
		t.Error("decoded Deleted = false, want true")
	}
}

func TestDoNoContentSkipsDecode(t *testing.T) {
	c := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	})

	out := struct{}{}
	meta, err := c.do(context.Background(), http.MethodDelete, "/things/x", nil, &out)
	if err != nil {
		t.Fatalf("do returned error: %v", err)
	}
	if meta == nil {
		t.Fatal("meta is nil, want non-nil ResponseMeta even on 204")
	}
}

func TestDoErrors(t *testing.T) {
	cases := []struct {
		name       string
		status     int
		code       string
		detail     string
		wantStatus int
	}{
		{"401", http.StatusUnauthorized, "UNAUTHORIZED", "Invalid API key", 401},
		{"403", http.StatusForbidden, "FORBIDDEN", "This key does not have the required scope: products:write", 403},
		{"404", http.StatusNotFound, "NOT_FOUND", "Product not found", 404},
		{"409", http.StatusConflict, "IDEMPOTENCY_CONFLICT", "Idempotency-Key was already used with a different request", 409},
		{"422", http.StatusUnprocessableEntity, "VALIDATION_ERROR", "Missing required field(s): name, price", 422},
		{"429", http.StatusTooManyRequests, "RATE_LIMITED", "Rate limit exceeded", 429},
		{"500", http.StatusInternalServerError, "INTERNAL_ERROR", "Something went wrong", 500},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			const requestID = "req-abc-123"
			c := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
				for k, vv := range testHeaders(requestID) {
					for _, v := range vv {
						w.Header().Add(k, v)
					}
				}
				writeJSON(t, w, tc.status, map[string]any{
					"detail":     tc.detail,
					"error_code": tc.code,
					"doc_url":    "https://docs.bachs.io/api-reference/error-reference#general",
				})
			})

			var out map[string]any
			meta, err := c.do(context.Background(), http.MethodGet, "/things", nil, &out)
			if err == nil {
				t.Fatal("do returned nil error, want *APIError")
			}
			var apiErr *APIError
			if !errors.As(err, &apiErr) {
				t.Fatalf("error is %T, want *APIError", err)
			}
			if apiErr.StatusCode != tc.wantStatus {
				t.Errorf("StatusCode = %d, want %d", apiErr.StatusCode, tc.wantStatus)
			}
			if apiErr.Code != tc.code {
				t.Errorf("Code = %q, want %q", apiErr.Code, tc.code)
			}
			if apiErr.Detail != tc.detail {
				t.Errorf("Detail = %q, want %q", apiErr.Detail, tc.detail)
			}
			if apiErr.DocURL == "" {
				t.Error("DocURL is empty, want the doc_url from the body")
			}
			if apiErr.RequestID != requestID {
				t.Errorf("RequestID = %q, want %q (from the x-request-id header)", apiErr.RequestID, requestID)
			}
			if meta == nil {
				t.Error("meta is nil, want ResponseMeta alongside the error")
			}
			if meta != nil && meta.RateLimitLimit != 100 {
				t.Errorf("meta.RateLimitLimit = %d, want 100 on error responses too", meta.RateLimitLimit)
			}
		})
	}
}

func TestDoValidationErrorFields(t *testing.T) {
	c := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		writeJSON(t, w, http.StatusUnprocessableEntity, map[string]any{
			"detail":     "Missing required field(s): name, price",
			"error_code": "VALIDATION_ERROR",
			"errors": []map[string]string{
				{"field": "name", "message": "Field required", "type": "missing"},
				{"field": "price", "message": "Field required", "type": "missing"},
			},
		})
	})

	var out map[string]any
	_, err := c.do(context.Background(), http.MethodPost, "/products", map[string]any{}, &out)
	if err == nil {
		t.Fatal("do returned nil error, want *APIError")
	}
	var apiErr *APIError
	if !errors.As(err, &apiErr) {
		t.Fatalf("error is %T, want *APIError", err)
	}
	if len(apiErr.Errors) != 2 {
		t.Fatalf("len(Errors) = %d, want 2", len(apiErr.Errors))
	}
	if apiErr.Errors[0].Field != "name" || apiErr.Errors[0].Type != "missing" {
		t.Errorf("Errors[0] = %+v, want field=name type=missing", apiErr.Errors[0])
	}
}

func TestDoNonJSONErrorBody(t *testing.T) {
	c := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadGateway)
		io.WriteString(w, "<html>Bad Gateway</html>")
	})

	var out map[string]any
	_, err := c.do(context.Background(), http.MethodGet, "/things", nil, &out)
	if err == nil {
		t.Fatal("do returned nil error, want error")
	}
	var apiErr *APIError
	if !errors.As(err, &apiErr) {
		t.Fatalf("error is %T, want *APIError", err)
	}
	if apiErr.StatusCode != http.StatusBadGateway {
		t.Errorf("StatusCode = %d, want 502", apiErr.StatusCode)
	}
	if apiErr.Code == "" {
		t.Error("Code is empty, want a fallback code for non-JSON bodies")
	}
}

func TestIdempotencyKeyRoundTrip(t *testing.T) {
	const key = "order_ORD-12345_attempt_1"

	// The server behaves like Bachs: it caches the 2xx response per key and
	// rejects a reused key with a different body as 409.
	type cached struct {
		requestBody  string
		responseBody string
	}
	var cache map[string]*cached

	c := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("method = %s, want POST", r.Method)
		}
		got := r.Header.Get(headerIdempotencyKey)
		if got != key {
			t.Errorf("Idempotency-Key = %q, want %q", got, key)
		}
		body, _ := io.ReadAll(r.Body)

		if cache == nil {
			cache = map[string]*cached{}
		}
		prev, ok := cache[key]
		if ok {
			if string(body) != prev.requestBody {
				writeJSON(t, w, http.StatusConflict, map[string]any{
					"detail":     "Idempotency-Key was already used with a different request",
					"error_code": "IDEMPOTENCY_CONFLICT",
				})
				return
			}
			w.WriteHeader(http.StatusCreated)
			io.WriteString(w, prev.responseBody)
			return
		}
		cache[key] = &cached{requestBody: string(body), responseBody: `{"id":"chk_1a2b3c4d5e6f"}`}
		io.WriteString(w, cache[key].responseBody)
	})

	payload := map[string]any{"customer": map[string]any{"email": "a@b.com"}}

	var out struct {
		ID string `json:"id"`
	}
	// First call with the key executes normally.
	_, err := c.do(context.Background(), http.MethodPost, "/checkout-sessions", payload, &out, WithIdempotencyKey(key))
	if err != nil {
		t.Fatalf("first POST returned error: %v", err)
	}
	if out.ID != "chk_1a2b3c4d5e6f" {
		t.Errorf("first POST ID = %q", out.ID)
	}

	// Same key, same body: the server returns the cached response.
	out.ID = ""
	_, err = c.do(context.Background(), http.MethodPost, "/checkout-sessions", payload, &out, WithIdempotencyKey(key))
	if err != nil {
		t.Fatalf("retried POST returned error: %v", err)
	}
	if out.ID != "chk_1a2b3c4d5e6f" {
		t.Errorf("cached POST ID = %q, want the cached response", out.ID)
	}

	// Same key, different body: 409 IDEMPOTENCY_CONFLICT surfaces as an error.
	_, err = c.do(context.Background(), http.MethodPost, "/checkout-sessions", map[string]any{
		"customer": map[string]any{"email": "other@b.com"},
	}, &out, WithIdempotencyKey(key))
	if err == nil {
		t.Fatal("reused key with different body returned nil error, want 409")
	}
	var apiErr *APIError
	if !errors.As(err, &apiErr) {
		t.Fatalf("error is %T, want *APIError", err)
	}
	if apiErr.StatusCode != http.StatusConflict || apiErr.Code != "IDEMPOTENCY_CONFLICT" {
		t.Errorf("got %+v, want 409 IDEMPOTENCY_CONFLICT", apiErr)
	}
}

func TestIdempotencyKeyRejectedOnNonPost(t *testing.T) {
	c := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		t.Error("request must not be sent when the idempotency key is misused")
	})

	for _, method := range []string{http.MethodGet, http.MethodPatch, http.MethodDelete} {
		var out map[string]any
		_, err := c.do(context.Background(), method, "/things", nil, &out, WithIdempotencyKey("k"))
		if err == nil {
			t.Errorf("%s with WithIdempotencyKey returned nil error, want rejection", method)
			continue
		}
		if !errors.Is(err, errIdempotencyNonPost) {
			t.Errorf("%s error = %v, want errIdempotencyNonPost", method, err)
		}
	}
}

func TestContextCancellation(t *testing.T) {
	c := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		// Hold the response open until the client context is canceled.
		<-r.Context().Done()
	})

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // canceled before the request is made

	var out map[string]any
	_, err := c.do(ctx, http.MethodGet, "/things", nil, &out)
	if err == nil {
		t.Fatal("do returned nil error for canceled context")
	}
	if !errors.Is(err, context.Canceled) {
		t.Errorf("error = %v, want it to wrap context.Canceled", err)
	}
}

func TestDoDecodeError(t *testing.T) {
	c := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		io.WriteString(w, `{"id": `) // truncated JSON
	})

	var out struct {
		ID string `json:"id"`
	}
	meta, err := c.do(context.Background(), http.MethodGet, "/things", nil, &out)
	if err == nil {
		t.Fatal("do returned nil error for an undecodable body")
	}
	if meta == nil {
		t.Error("meta is nil, want ResponseMeta even when decoding fails")
	}
}

func TestResponseMetaMissingRateLimitHeaders(t *testing.T) {
	c := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		io.WriteString(w, `{}`)
	})

	var out map[string]any
	meta, err := c.do(context.Background(), http.MethodGet, "/things", nil, &out)
	if err != nil {
		t.Fatalf("do returned error: %v", err)
	}
	if meta.RateLimitLimit != 0 || meta.RateLimitRemaining != 0 {
		t.Errorf("expected zero rate limits when headers are absent, got %+v", meta)
	}
	if !meta.RateLimitReset.IsZero() {
		t.Errorf("expected zero RateLimitReset when header is absent, got %v", meta.RateLimitReset)
	}
}

// TestDoPipelineSequence drives the full request pipeline through three
// consecutive hops on one client — a 204 (decode skipped), a 500 (surfaced
// as *APIError), and a 200 (decoded) — asserting ResponseMeta is populated
// on every hop, including the error hop.
func TestDoPipelineSequence(t *testing.T) {
	var calls int
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls++
		w.Header().Set(headerRequestID, "req_seq_"+strconv.Itoa(calls))
		w.Header().Set(headerRateLimitLimit, "100")
		w.Header().Set(headerRateLimitRemaining, strconv.Itoa(100-calls))
		w.Header().Set(headerRateLimitReset, "1783976340")
		switch calls {
		case 1:
			w.WriteHeader(http.StatusNoContent)
		case 2:
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusInternalServerError)
			io.WriteString(w, `{"detail": "boom", "error_code": "INTERNAL_ERROR"}`)
		case 3:
			writeJSON(t, w, http.StatusOK, map[string]any{"id": "pay_seq_1", "amount": "29.00"})
		default:
			t.Errorf("unexpected %dth request", calls)
		}
	}))
	t.Cleanup(srv.Close)

	c, err := NewClient("sk_sandbox_test", WithBaseURL(srv.URL))
	if err != nil {
		t.Fatalf("NewClient returned error: %v", err)
	}

	ctx := context.Background()

	// Hop 1: 204 — response body must be skipped entirely.
	type paymentOut struct {
		ID string `json:"id"`
	}
	var out paymentOut
	meta1, err := c.do(ctx, http.MethodGet, "/payments/pay_1", nil, &out)
	if err != nil {
		t.Fatalf("hop 1 (204) returned error: %v", err)
	}
	if meta1.RequestID != "req_seq_1" {
		t.Errorf("hop 1 RequestID = %q, want req_seq_1", meta1.RequestID)
	}
	if meta1.RateLimitLimit != 100 || meta1.RateLimitRemaining != 99 {
		t.Errorf("hop 1 rate limits = %+v", meta1)
	}
	if meta1.RateLimitReset.IsZero() {
		t.Error("hop 1 RateLimitReset is zero, want parsed unix time")
	}

	// Hop 2: 500 — surfaced as *APIError, but meta still populated.
	meta2, err := c.do(ctx, http.MethodGet, "/payments/pay_2", nil, &out)
	var apiErr *APIError
	if !errors.As(err, &apiErr) {
		t.Fatalf("hop 2 error = %v, want *APIError", err)
	}
	if apiErr.StatusCode != http.StatusInternalServerError || apiErr.Code != "INTERNAL_ERROR" {
		t.Errorf("hop 2 APIError = %+v", apiErr)
	}
	if apiErr.RequestID != "req_seq_2" {
		t.Errorf("hop 2 APIError.RequestID = %q, want req_seq_2", apiErr.RequestID)
	}
	if meta2 == nil {
		t.Fatal("hop 2 ResponseMeta is nil")
	}
	if meta2.RequestID != "req_seq_2" {
		t.Errorf("hop 2 meta.RequestID = %q, want req_seq_2", meta2.RequestID)
	}
	if meta2.RateLimitRemaining != 98 {
		t.Errorf("hop 2 RateLimitRemaining = %d, want 98", meta2.RateLimitRemaining)
	}

	// Hop 3: 200 — decoded into out, meta populated again.
	meta3, err := c.do(ctx, http.MethodGet, "/payments/pay_3", nil, &out)
	if err != nil {
		t.Fatalf("hop 3 (200) returned error: %v", err)
	}
	if out.ID != "pay_seq_1" {
		t.Errorf("hop 3 decoded ID = %q, want pay_seq_1", out.ID)
	}
	if meta3.RequestID != "req_seq_3" {
		t.Errorf("hop 3 RequestID = %q, want req_seq_3", meta3.RequestID)
	}
	if meta3.RateLimitRemaining != 97 {
		t.Errorf("hop 3 RateLimitRemaining = %d, want 97", meta3.RateLimitRemaining)
	}
}
