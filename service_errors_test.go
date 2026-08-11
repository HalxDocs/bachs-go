package bachs

import (
	"context"
	"errors"
	"net/http"
	"strings"
	"testing"
)

// TestServiceErrorPropagation verifies that when the API returns a non-2xx
// response, every service method surfaces it as an *APIError carrying the
// request ID, with the ResponseMeta still populated. This exercises the
// error branch of each service method, which the happy-path tests skip.
func TestServiceErrorPropagation(t *testing.T) {
	t.Helper()

	// One representative call per service. Each returns the error branch of
	// its method: (*T, *ResponseMeta, *APIError) with the request ID intact.
	errorHandler := func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set(headerRequestID, "req_errtest")
		w.Header().Set(headerRateLimitLimit, "100")
		w.Header().Set(headerRateLimitRemaining, "99")
		writeJSON(t, w, http.StatusPaymentRequired, map[string]any{
			"detail":     "insufficient funds",
			"error_code": "INSUFFICIENT_FUNDS",
		})
	}

	check := func(t *testing.T, meta *ResponseMeta, err error) {
		t.Helper()
		var apiErr *APIError
		if !errors.As(err, &apiErr) {
			t.Fatalf("error = %v, want *APIError", err)
		}
		if apiErr.StatusCode != http.StatusPaymentRequired {
			t.Errorf("StatusCode = %d, want 402", apiErr.StatusCode)
		}
		if apiErr.Code != "INSUFFICIENT_FUNDS" || apiErr.Detail != "insufficient funds" {
			t.Errorf("APIError = %+v", apiErr)
		}
		if apiErr.RequestID != "req_errtest" {
			t.Errorf("RequestID = %q, want req_errtest", apiErr.RequestID)
		}
		if meta == nil {
			t.Fatal("ResponseMeta is nil")
		}
		if meta.RequestID != "req_errtest" {
			t.Errorf("meta.RequestID = %q, want req_errtest", meta.RequestID)
		}
		if meta.RateLimitLimit != 100 || meta.RateLimitRemaining != 99 {
			t.Errorf("meta rate limits = %+v", meta)
		}
	}

	ctx := context.Background()

	t.Run("Checkouts.Get", func(t *testing.T) {
		c := newTestClient(t, errorHandler)
		_, meta, err := c.Checkouts.Get(ctx, "chk_1")
		check(t, meta, err)
	})
	t.Run("Products.Get", func(t *testing.T) {
		c := newTestClient(t, errorHandler)
		_, meta, err := c.Products.Get(ctx, "prod_1")
		check(t, meta, err)
	})
	t.Run("Customers.Get", func(t *testing.T) {
		c := newTestClient(t, errorHandler)
		_, meta, err := c.Customers.Get(ctx, "cust_1")
		check(t, meta, err)
	})
	t.Run("Payments.Get", func(t *testing.T) {
		c := newTestClient(t, errorHandler)
		_, meta, err := c.Payments.Get(ctx, "pay_1")
		check(t, meta, err)
	})
	t.Run("Refunds.Get", func(t *testing.T) {
		c := newTestClient(t, errorHandler)
		_, meta, err := c.Refunds.Get(ctx, "ref_1")
		check(t, meta, err)
	})
	t.Run("Subscriptions.Get", func(t *testing.T) {
		c := newTestClient(t, errorHandler)
		_, meta, err := c.Subscriptions.Get(ctx, "sub_1")
		check(t, meta, err)
	})
	t.Run("Transfers.Get", func(t *testing.T) {
		c := newTestClient(t, errorHandler)
		_, meta, err := c.Transfers.Get(ctx, "trf_1")
		check(t, meta, err)
	})
	t.Run("Misc.GetBalances", func(t *testing.T) {
		c := newTestClient(t, errorHandler)
		_, meta, err := c.Misc.GetBalances(ctx)
		check(t, meta, err)
	})
	t.Run("Media.Get", func(t *testing.T) {
		c := newTestClient(t, errorHandler)
		_, meta, err := c.Media.Get(ctx, "med_1")
		check(t, meta, err)
	})
	t.Run("CustomerSessions.Create", func(t *testing.T) {
		c := newTestClient(t, errorHandler)
		_, meta, err := c.CustomerSessions.Create(ctx, "cust_1")
		check(t, meta, err)
	})
	t.Run("ConnectedAccounts.Get", func(t *testing.T) {
		c := newTestClient(t, errorHandler)
		_, meta, err := c.ConnectedAccounts.Get(ctx, "org_1")
		check(t, meta, err)
	})
	t.Run("Payouts.Get", func(t *testing.T) {
		c := newTestClient(t, errorHandler)
		_, meta, err := c.Payouts.Get(ctx, "wd_1")
		check(t, meta, err)
	})
	t.Run("Disputes.Get", func(t *testing.T) {
		c := newTestClient(t, errorHandler)
		_, meta, err := c.Disputes.Get(ctx, "dsp_1")
		check(t, meta, err)
	})
	t.Run("Conversions.Get", func(t *testing.T) {
		c := newTestClient(t, errorHandler)
		_, meta, err := c.Conversions.Get(ctx, "cvt_1")
		check(t, meta, err)
	})
	t.Run("Organizations.GetMe", func(t *testing.T) {
		c := newTestClient(t, errorHandler)
		_, meta, err := c.Organizations.GetMe(ctx)
		check(t, meta, err)
	})
	t.Run("Webhooks.GetEndpoint", func(t *testing.T) {
		c := newTestClient(t, errorHandler)
		_, meta, err := c.Webhooks.GetEndpoint(ctx, "whe_1")
		check(t, meta, err)
	})
}

// TestClientString verifies the client's String method never leaks the API
// key.
func TestClientString(t *testing.T) {
	c, err := NewClient("sk_sandbox_secret_key_value")
	if err != nil {
		t.Fatalf("NewClient returned error: %v", err)
	}
	got := c.String()
	if got == "" {
		t.Error("String returned empty")
	}
	if strings.Contains(got, "sk_sandbox_secret_key_value") {
		t.Errorf("String leaks the API key: %q", got)
	}
}

// failingReader returns an error partway through a read, to force the
// multipart encoding error path in Media.Upload and DisputeService.
type failingReader struct{}

func (failingReader) Read([]byte) (int, error) {
	return 0, errors.New("simulated read failure")
}

func TestMediaUploadMultipartError(t *testing.T) {
	c := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		t.Error("request should not have been sent")
	})

	_, _, err := c.Media.Upload(context.Background(), "broken.png", failingReader{}, "product-media")
	if err == nil {
		t.Fatal("Upload returned nil error, want multipart error")
	}
	if !strings.Contains(err.Error(), "copy upload file") {
		t.Errorf("error = %v, want multipart copy error", err)
	}
}

func TestDisputeUploadMultipartError(t *testing.T) {
	c := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		t.Error("request should not have been sent")
	})

	_, _, err := c.Disputes.UploadDocument(context.Background(), "broken.pdf", failingReader{}, "dispute-evidence")
	if err == nil {
		t.Fatal("UploadDocument returned nil error, want multipart error")
	}
	if !strings.Contains(err.Error(), "copy upload file") {
		t.Errorf("error = %v, want multipart copy error", err)
	}
}

func TestConnectedAccountUploadMultipartError(t *testing.T) {
	c := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		t.Error("request should not have been sent")
	})

	_, _, err := c.ConnectedAccounts.UploadDocument(context.Background(), "org_1", "broken.jpg", failingReader{}, "identity_documents")
	if err == nil {
		t.Fatal("UploadDocument returned nil error, want multipart error")
	}
	if !strings.Contains(err.Error(), "copy upload file") {
		t.Errorf("error = %v, want multipart copy error", err)
	}
}
