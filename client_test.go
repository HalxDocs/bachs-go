package bachs

import (
	"context"
	"io"
	"net/http"
	"strings"
	"testing"
	"time"
)

func TestNewClientDefaults(t *testing.T) {
	c, err := NewClient("sk_sandbox_test")
	if err != nil {
		t.Fatalf("NewClient returned error: %v", err)
	}
	if got := c.BaseURL(); got != BaseURLSandbox {
		t.Errorf("default BaseURL = %q, want %q (sandbox is the default)", got, BaseURLSandbox)
	}
	if c.httpClient == nil {
		t.Fatal("default httpClient is nil")
	}
	if c.httpClient.Timeout != 30*time.Second {
		t.Errorf("default httpClient.Timeout = %v, want 30s", c.httpClient.Timeout)
	}
}

func TestNewClientRejectsEmptyKey(t *testing.T) {
	if _, err := NewClient(""); err == nil {
		t.Fatal("NewClient with empty key returned no error, want error")
	}
	if _, err := NewClient("   "); err == nil {
		t.Fatal("NewClient with whitespace key returned no error, want error")
	}
}

func TestWithProduction(t *testing.T) {
	c, err := NewClient("sk_sandbox_test", WithProduction())
	if err != nil {
		t.Fatalf("NewClient returned error: %v", err)
	}
	if got := c.BaseURL(); got != BaseURLProduction {
		t.Errorf("WithProduction BaseURL = %q, want %q", got, BaseURLProduction)
	}
}

func TestWithSandbox(t *testing.T) {
	c, err := NewClient("sk_sandbox_test", WithSandbox())
	if err != nil {
		t.Fatalf("NewClient returned error: %v", err)
	}
	if got := c.BaseURL(); got != BaseURLSandbox {
		t.Errorf("WithSandbox BaseURL = %q, want %q", got, BaseURLSandbox)
	}
}

func TestWithBaseURL(t *testing.T) {
	c, err := NewClient("sk_sandbox_test", WithBaseURL("https://example.test/"))
	if err != nil {
		t.Fatalf("NewClient returned error: %v", err)
	}
	// Trailing slash is normalized away.
	if got := c.BaseURL(); got != "https://example.test" {
		t.Errorf("WithBaseURL BaseURL = %q, want %q", got, "https://example.test")
	}
}

// recordingTransport is a RoundTripper that records the request and returns a
// canned response, so a custom HTTP client can be injected without a server.
type recordingTransport struct {
	lastReq *http.Request
}

func (rt *recordingTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	rt.lastReq = req
	return &http.Response{
		StatusCode: http.StatusOK,
		Status:     "200 OK",
		Header:     http.Header{},
		Body:       io.NopCloser(strings.NewReader(`{}`)),
		Request:    req,
	}, nil
}

func TestWithHTTPClient(t *testing.T) {
	rt := &recordingTransport{}
	c, err := NewClient(
		"sk_sandbox_test",
		WithBaseURL("https://example.test"),
		WithHTTPClient(&http.Client{Transport: rt, Timeout: 5 * time.Second}),
	)
	if err != nil {
		t.Fatalf("NewClient returned error: %v", err)
	}

	var out map[string]any
	_, err = c.do(context.Background(), http.MethodGet, "/ping", nil, &out)
	if err != nil {
		t.Fatalf("do returned error: %v", err)
	}
	if rt.lastReq == nil {
		t.Fatal("custom transport was not used")
	}
	if got := rt.lastReq.URL.String(); got != "https://example.test/v1/ping" {
		t.Errorf("request URL = %q, want %q", got, "https://example.test/v1/ping")
	}
	if auth := rt.lastReq.Header.Get(headerAuthorization); auth != "Bearer sk_sandbox_test" {
		t.Errorf("Authorization = %q, want %q", auth, "Bearer sk_sandbox_test")
	}
}

func TestCustomBaseURL(t *testing.T) {
	srv := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		if !strings.HasPrefix(r.URL.Path, "/v1/") {
			t.Errorf("request path %q does not include the /v1/ version segment", r.URL.Path)
		}
		io.WriteString(w, `{}`)
	})
	c, err := NewClient("sk_sandbox_test", WithBaseURL(srv.URL))
	if err != nil {
		t.Fatalf("NewClient returned error: %v", err)
	}
	var out map[string]any
	if _, err := c.do(context.Background(), http.MethodGet, "/ping", nil, &out); err != nil {
		t.Fatalf("do returned error: %v", err)
	}
}
