package bachs

import (
	"context"
	"errors"
	"net/http"
	"strings"
	"testing"
)

// errorCall invokes one service method against a client that always answers
// with a 402, returning the ResponseMeta and error the method surfaced. The
// out value is discarded — this test is about the error branch.
type errorCall func(c *Client) (*ResponseMeta, error)

// TestServiceErrorBranches verifies that when the API returns a non-2xx
// response, every service method surfaces it as an *APIError carrying the
// request ID, with the ResponseMeta still populated. This exercises the
// `if err != nil { return nil, meta, err }` branch of every method, which
// the happy-path tests skip. The table covers all 16 services and every
// method on each, including the Create/List/Update variants that are
// byte-identical copies of the pattern proven per service.
func TestServiceErrorBranches(t *testing.T) {
	t.Helper()

	errorHandler := func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set(headerRequestID, "req_errtest")
		w.Header().Set(headerRateLimitLimit, "100")
		w.Header().Set(headerRateLimitRemaining, "99")
		writeJSON(t, w, http.StatusPaymentRequired, map[string]any{
			"detail":     "insufficient funds",
			"error_code": "INSUFFICIENT_FUNDS",
		})
	}

	ctx := context.Background()

	tests := []struct {
		name string
		call errorCall
	}{
		// Checkouts.
		{"Checkouts.Create", func(c *Client) (*ResponseMeta, error) {
			_, meta, err := c.Checkouts.Create(ctx, CreateCheckoutSessionRequest{})
			return meta, err
		}},
		{"Checkouts.Get", func(c *Client) (*ResponseMeta, error) {
			_, meta, err := c.Checkouts.Get(ctx, "chk_1")
			return meta, err
		}},

		// Products.
		{"Products.Create", func(c *Client) (*ResponseMeta, error) {
			_, meta, err := c.Products.Create(ctx, CreateProductRequest{})
			return meta, err
		}},
		{"Products.Get", func(c *Client) (*ResponseMeta, error) {
			_, meta, err := c.Products.Get(ctx, "prod_1")
			return meta, err
		}},
		{"Products.List", func(c *Client) (*ResponseMeta, error) {
			_, meta, err := c.Products.List(ctx, ListParams{})
			return meta, err
		}},
		{"Products.Update", func(c *Client) (*ResponseMeta, error) {
			_, meta, err := c.Products.Update(ctx, "prod_1", UpdateProductRequest{})
			return meta, err
		}},
		{"Products.Archive", func(c *Client) (*ResponseMeta, error) {
			_, meta, err := c.Products.Archive(ctx, "prod_1")
			return meta, err
		}},
		{"Products.Unarchive", func(c *Client) (*ResponseMeta, error) {
			_, meta, err := c.Products.Unarchive(ctx, "prod_1")
			return meta, err
		}},

		// Customers.
		{"Customers.Create", func(c *Client) (*ResponseMeta, error) {
			_, meta, err := c.Customers.Create(ctx, CreateCustomerRequest{})
			return meta, err
		}},
		{"Customers.Get", func(c *Client) (*ResponseMeta, error) {
			_, meta, err := c.Customers.Get(ctx, "cust_1")
			return meta, err
		}},
		{"Customers.List", func(c *Client) (*ResponseMeta, error) {
			_, meta, err := c.Customers.List(ctx, ListParams{})
			return meta, err
		}},
		{"Customers.Update", func(c *Client) (*ResponseMeta, error) {
			_, meta, err := c.Customers.Update(ctx, "cust_1", UpdateCustomerRequest{})
			return meta, err
		}},

		// Customer sessions.
		{"CustomerSessions.Create", func(c *Client) (*ResponseMeta, error) {
			_, meta, err := c.CustomerSessions.Create(ctx, "cust_1")
			return meta, err
		}},

		// Payments.
		{"Payments.Get", func(c *Client) (*ResponseMeta, error) {
			_, meta, err := c.Payments.Get(ctx, "pay_1")
			return meta, err
		}},
		{"Payments.List", func(c *Client) (*ResponseMeta, error) {
			_, meta, err := c.Payments.List(ctx, ListParams{})
			return meta, err
		}},

		// Refunds.
		{"Refunds.Create", func(c *Client) (*ResponseMeta, error) {
			_, meta, err := c.Refunds.Create(ctx, CreateRefundRequest{})
			return meta, err
		}},
		{"Refunds.Get", func(c *Client) (*ResponseMeta, error) {
			_, meta, err := c.Refunds.Get(ctx, "ref_1")
			return meta, err
		}},
		{"Refunds.GetByCharge", func(c *Client) (*ResponseMeta, error) {
			_, meta, err := c.Refunds.GetByCharge(ctx, "pay_1")
			return meta, err
		}},
		{"Refunds.List", func(c *Client) (*ResponseMeta, error) {
			_, meta, err := c.Refunds.List(ctx, ListParams{})
			return meta, err
		}},

		// Subscriptions.
		{"Subscriptions.Get", func(c *Client) (*ResponseMeta, error) {
			_, meta, err := c.Subscriptions.Get(ctx, "sub_1")
			return meta, err
		}},
		{"Subscriptions.List", func(c *Client) (*ResponseMeta, error) {
			_, meta, err := c.Subscriptions.List(ctx, ListParams{})
			return meta, err
		}},
		{"Subscriptions.Update", func(c *Client) (*ResponseMeta, error) {
			// UpdateSubscriptionRequest validates exactly-one-intent
			// client-side, so pass a valid request to reach the network call.
			_, meta, err := c.Subscriptions.Update(ctx, "sub_1", UpdateSubscriptionRequest{ProductID: "prod_1"})
			return meta, err
		}},
		{"Subscriptions.Cancel", func(c *Client) (*ResponseMeta, error) {
			_, meta, err := c.Subscriptions.Cancel(ctx, "sub_1", CancelSubscriptionRequest{})
			return meta, err
		}},

		// Media.
		{"Media.Upload", func(c *Client) (*ResponseMeta, error) {
			_, meta, err := c.Media.Upload(ctx, "hero.png", strings.NewReader("png bytes"), "product-media")
			return meta, err
		}},
		{"Media.Get", func(c *Client) (*ResponseMeta, error) {
			_, meta, err := c.Media.Get(ctx, "med_1")
			return meta, err
		}},
		{"Media.Delete", func(c *Client) (*ResponseMeta, error) {
			_, meta, err := c.Media.Delete(ctx, "med_1")
			return meta, err
		}},

		// Transfers.
		{"Transfers.Create", func(c *Client) (*ResponseMeta, error) {
			_, meta, err := c.Transfers.Create(ctx, CreateTransferRequest{})
			return meta, err
		}},
		{"Transfers.Get", func(c *Client) (*ResponseMeta, error) {
			_, meta, err := c.Transfers.Get(ctx, "trf_1")
			return meta, err
		}},
		{"Transfers.List", func(c *Client) (*ResponseMeta, error) {
			_, meta, err := c.Transfers.List(ctx, ListParams{})
			return meta, err
		}},

		// Connected accounts.
		{"ConnectedAccounts.Create", func(c *Client) (*ResponseMeta, error) {
			_, meta, err := c.ConnectedAccounts.Create(ctx, CreateConnectedAccountRequest{})
			return meta, err
		}},
		{"ConnectedAccounts.Get", func(c *Client) (*ResponseMeta, error) {
			_, meta, err := c.ConnectedAccounts.Get(ctx, "org_1")
			return meta, err
		}},
		{"ConnectedAccounts.List", func(c *Client) (*ResponseMeta, error) {
			_, meta, err := c.ConnectedAccounts.List(ctx, ListParams{})
			return meta, err
		}},
		{"ConnectedAccounts.RequestCapabilities", func(c *Client) (*ResponseMeta, error) {
			_, meta, err := c.ConnectedAccounts.RequestCapabilities(ctx, "org_1", UpdateConnectedAccountRequest{})
			return meta, err
		}},
		{"ConnectedAccounts.CreateAccountLink", func(c *Client) (*ResponseMeta, error) {
			_, meta, err := c.ConnectedAccounts.CreateAccountLink(ctx, "org_1", CreateAccountLinkRequest{})
			return meta, err
		}},
		{"ConnectedAccounts.ListCapabilities", func(c *Client) (*ResponseMeta, error) {
			_, meta, err := c.ConnectedAccounts.ListCapabilities(ctx, "org_1")
			return meta, err
		}},
		{"ConnectedAccounts.GetTaskChecklist", func(c *Client) (*ResponseMeta, error) {
			_, meta, err := c.ConnectedAccounts.GetTaskChecklist(ctx, "org_1")
			return meta, err
		}},
		{"ConnectedAccounts.ListTasks", func(c *Client) (*ResponseMeta, error) {
			_, meta, err := c.ConnectedAccounts.ListTasks(ctx, "org_1", ListParams{})
			return meta, err
		}},
		{"ConnectedAccounts.GetTaskValues", func(c *Client) (*ResponseMeta, error) {
			_, meta, err := c.ConnectedAccounts.GetTaskValues(ctx, "org_1")
			return meta, err
		}},
		{"ConnectedAccounts.SubmitTaskValues", func(c *Client) (*ResponseMeta, error) {
			_, meta, err := c.ConnectedAccounts.SubmitTaskValues(ctx, "org_1", SubmitTasksRequest{})
			return meta, err
		}},
		{"ConnectedAccounts.GetReusableIdentity", func(c *Client) (*ResponseMeta, error) {
			_, meta, err := c.ConnectedAccounts.GetReusableIdentity(ctx, "org_1")
			return meta, err
		}},
		{"ConnectedAccounts.ApplyReusableIdentity", func(c *Client) (*ResponseMeta, error) {
			_, meta, err := c.ConnectedAccounts.ApplyReusableIdentity(ctx, "org_1", ApplyReusableIdentityRequest{})
			return meta, err
		}},
		{"ConnectedAccounts.ListBanks", func(c *Client) (*ResponseMeta, error) {
			_, meta, err := c.ConnectedAccounts.ListBanks(ctx, "org_1", "NG")
			return meta, err
		}},
		{"ConnectedAccounts.ListMobileMoneyProviders", func(c *Client) (*ResponseMeta, error) {
			_, meta, err := c.ConnectedAccounts.ListMobileMoneyProviders(ctx, "org_1", "KE")
			return meta, err
		}},
		{"ConnectedAccounts.ResolveBankAccount", func(c *Client) (*ResponseMeta, error) {
			_, meta, err := c.ConnectedAccounts.ResolveBankAccount(ctx, "org_1", ResolveTaskBankAccountRequest{})
			return meta, err
		}},
		{"ConnectedAccounts.UploadDocument", func(c *Client) (*ResponseMeta, error) {
			_, meta, err := c.ConnectedAccounts.UploadDocument(ctx, "org_1", "id.jpg", strings.NewReader("jpg bytes"), "identity_documents")
			return meta, err
		}},
		{"ConnectedAccounts.GetDocument", func(c *Client) (*ResponseMeta, error) {
			_, meta, err := c.ConnectedAccounts.GetDocument(ctx, "org_1", "med_1")
			return meta, err
		}},

		// Misc.
		{"Misc.GetBalances", func(c *Client) (*ResponseMeta, error) {
			_, meta, err := c.Misc.GetBalances(ctx)
			return meta, err
		}},
		{"Misc.ListPaymentMethods", func(c *Client) (*ResponseMeta, error) {
			_, meta, err := c.Misc.ListPaymentMethods(ctx)
			return meta, err
		}},
		{"Misc.ListPaymentRails", func(c *Client) (*ResponseMeta, error) {
			_, meta, err := c.Misc.ListPaymentRails(ctx, "", "", "")
			return meta, err
		}},
		{"Misc.ListSupportedCurrencies", func(c *Client) (*ResponseMeta, error) {
			_, meta, err := c.Misc.ListSupportedCurrencies(ctx)
			return meta, err
		}},
		{"Misc.ListPayoutSupportedCurrencies", func(c *Client) (*ResponseMeta, error) {
			_, meta, err := c.Misc.ListPayoutSupportedCurrencies(ctx)
			return meta, err
		}},

		// Payouts.
		{"Payouts.GetSupportedCurrencies", func(c *Client) (*ResponseMeta, error) {
			_, meta, err := c.Payouts.GetSupportedCurrencies(ctx, "BANK_TRANSFER")
			return meta, err
		}},
		{"Payouts.CreateQuote", func(c *Client) (*ResponseMeta, error) {
			_, meta, err := c.Payouts.CreateQuote(ctx, CreatePayoutQuoteRequest{})
			return meta, err
		}},
		{"Payouts.ListDestinations", func(c *Client) (*ResponseMeta, error) {
			_, meta, err := c.Payouts.ListDestinations(ctx)
			return meta, err
		}},
		{"Payouts.CreateDestination", func(c *Client) (*ResponseMeta, error) {
			_, meta, err := c.Payouts.CreateDestination(ctx, CreatePayoutDestinationRequest{})
			return meta, err
		}},
		{"Payouts.UpdateDestination", func(c *Client) (*ResponseMeta, error) {
			_, meta, err := c.Payouts.UpdateDestination(ctx, "pdst_1", UpdatePayoutDestinationRequest{})
			return meta, err
		}},
		{"Payouts.DeleteDestination", func(c *Client) (*ResponseMeta, error) {
			_, meta, err := c.Payouts.DeleteDestination(ctx, "pdst_1")
			return meta, err
		}},
		{"Payouts.ResolveBankAccount", func(c *Client) (*ResponseMeta, error) {
			_, meta, err := c.Payouts.ResolveBankAccount(ctx, "044", "0123456789")
			return meta, err
		}},
		{"Payouts.ListBanks", func(c *Client) (*ResponseMeta, error) {
			_, meta, err := c.Payouts.ListBanks(ctx, "NG")
			return meta, err
		}},
		{"Payouts.List", func(c *Client) (*ResponseMeta, error) {
			_, meta, err := c.Payouts.List(ctx, ListParams{})
			return meta, err
		}},
		{"Payouts.Get", func(c *Client) (*ResponseMeta, error) {
			_, meta, err := c.Payouts.Get(ctx, "wd_1")
			return meta, err
		}},
		{"Payouts.CreateWithdrawal", func(c *Client) (*ResponseMeta, error) {
			_, meta, err := c.Payouts.CreateWithdrawal(ctx, CreateWithdrawalRequest{})
			return meta, err
		}},

		// Disputes.
		{"Disputes.List", func(c *Client) (*ResponseMeta, error) {
			_, meta, err := c.Disputes.List(ctx, ListParams{})
			return meta, err
		}},
		{"Disputes.Get", func(c *Client) (*ResponseMeta, error) {
			_, meta, err := c.Disputes.Get(ctx, "dsp_1")
			return meta, err
		}},
		{"Disputes.UploadDocument", func(c *Client) (*ResponseMeta, error) {
			_, meta, err := c.Disputes.UploadDocument(ctx, "evidence.pdf", strings.NewReader("pdf bytes"), "dispute-evidence")
			return meta, err
		}},
		{"Disputes.UpdateEvidence", func(c *Client) (*ResponseMeta, error) {
			_, meta, err := c.Disputes.UpdateEvidence(ctx, "dsp_1", DisputeEvidenceUpdateRequest{})
			return meta, err
		}},
		{"Disputes.Submit", func(c *Client) (*ResponseMeta, error) {
			_, meta, err := c.Disputes.Submit(ctx, "dsp_1")
			return meta, err
		}},

		// Conversions.
		{"Conversions.CreateQuote", func(c *Client) (*ResponseMeta, error) {
			_, meta, err := c.Conversions.CreateQuote(ctx, CreateConversionQuoteRequest{})
			return meta, err
		}},
		{"Conversions.List", func(c *Client) (*ResponseMeta, error) {
			_, meta, err := c.Conversions.List(ctx, ListParams{})
			return meta, err
		}},
		{"Conversions.Create", func(c *Client) (*ResponseMeta, error) {
			_, meta, err := c.Conversions.Create(ctx, CreateConversionRequest{})
			return meta, err
		}},
		{"Conversions.Get", func(c *Client) (*ResponseMeta, error) {
			_, meta, err := c.Conversions.Get(ctx, "cvt_1")
			return meta, err
		}},

		// Organizations.
		{"Organizations.GetMe", func(c *Client) (*ResponseMeta, error) {
			_, meta, err := c.Organizations.GetMe(ctx)
			return meta, err
		}},
		{"Organizations.Get", func(c *Client) (*ResponseMeta, error) {
			_, meta, err := c.Organizations.Get(ctx, "org_1")
			return meta, err
		}},
		{"Organizations.GetCheckoutSettings", func(c *Client) (*ResponseMeta, error) {
			_, meta, err := c.Organizations.GetCheckoutSettings(ctx)
			return meta, err
		}},
		{"Organizations.UpdateCheckoutSettings", func(c *Client) (*ResponseMeta, error) {
			_, meta, err := c.Organizations.UpdateCheckoutSettings(ctx, UpdateCheckoutSettingsRequest{})
			return meta, err
		}},

		// Webhooks management.
		{"Webhooks.CreateEndpoint", func(c *Client) (*ResponseMeta, error) {
			_, meta, err := c.Webhooks.CreateEndpoint(ctx, CreateWebhookEndpointRequest{})
			return meta, err
		}},
		{"Webhooks.ListEndpoints", func(c *Client) (*ResponseMeta, error) {
			_, meta, err := c.Webhooks.ListEndpoints(ctx)
			return meta, err
		}},
		{"Webhooks.GetEndpoint", func(c *Client) (*ResponseMeta, error) {
			_, meta, err := c.Webhooks.GetEndpoint(ctx, "whe_1")
			return meta, err
		}},
		{"Webhooks.UpdateEndpoint", func(c *Client) (*ResponseMeta, error) {
			_, meta, err := c.Webhooks.UpdateEndpoint(ctx, "whe_1", UpdateWebhookEndpointRequest{})
			return meta, err
		}},
		{"Webhooks.DeleteEndpoint", func(c *Client) (*ResponseMeta, error) {
			_, meta, err := c.Webhooks.DeleteEndpoint(ctx, "whe_1")
			return meta, err
		}},
		{"Webhooks.GetEndpointSecret", func(c *Client) (*ResponseMeta, error) {
			_, meta, err := c.Webhooks.GetEndpointSecret(ctx, "whe_1")
			return meta, err
		}},
		{"Webhooks.RotateEndpointSecret", func(c *Client) (*ResponseMeta, error) {
			_, meta, err := c.Webhooks.RotateEndpointSecret(ctx, "whe_1")
			return meta, err
		}},
		{"Webhooks.GetEndpointMetrics", func(c *Client) (*ResponseMeta, error) {
			_, meta, err := c.Webhooks.GetEndpointMetrics(ctx, "whe_1", EndpointMetricsParams{})
			return meta, err
		}},
		{"Webhooks.ListEndpointEvents", func(c *Client) (*ResponseMeta, error) {
			_, meta, err := c.Webhooks.ListEndpointEvents(ctx, "whe_1", ListParams{})
			return meta, err
		}},
		{"Webhooks.GetEndpointEvent", func(c *Client) (*ResponseMeta, error) {
			_, meta, err := c.Webhooks.GetEndpointEvent(ctx, "whe_1", "evt_1")
			return meta, err
		}},
		{"Webhooks.ResendEndpointEvent", func(c *Client) (*ResponseMeta, error) {
			_, meta, err := c.Webhooks.ResendEndpointEvent(ctx, "whe_1", "evt_1")
			return meta, err
		}},
		{"Webhooks.ListEvents", func(c *Client) (*ResponseMeta, error) {
			_, meta, err := c.Webhooks.ListEvents(ctx, ListParams{})
			return meta, err
		}},
		{"Webhooks.GetEvent", func(c *Client) (*ResponseMeta, error) {
			_, meta, err := c.Webhooks.GetEvent(ctx, "evt_1")
			return meta, err
		}},
		{"Webhooks.Replay", func(c *Client) (*ResponseMeta, error) {
			_, meta, err := c.Webhooks.Replay(ctx, ReplayWebhookEventRequest{})
			return meta, err
		}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := newTestClient(t, errorHandler)
			meta, err := tt.call(c)
			checkServiceError(t, meta, err)
		})
	}
}

// checkServiceError asserts the error branch contract: an *APIError with the
// status code, code, detail, and request ID intact, and a non-nil ResponseMeta
// carrying the same request ID and the rate-limit headers.
func checkServiceError(t *testing.T, meta *ResponseMeta, err error) {
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
