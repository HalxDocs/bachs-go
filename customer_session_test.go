package bachs

import (
	"context"
	"io"
	"net/http"
	"testing"
)

// TestCustomerSessionCreate uses the exact example payload from
// https://docs.bachs.io/api-reference/customer-sessions/create-a-customer-portal-session
func TestCustomerSessionCreate(t *testing.T) {
	c := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("method = %s, want POST", r.Method)
		}
		if r.URL.Path != "/v1/customers/cust_1a2b3c4d5e6f/portal-sessions" {
			t.Errorf("path = %q, want /v1/customers/cust_1a2b3c4d5e6f/portal-sessions", r.URL.Path)
		}
		io.WriteString(w, `{
			"id": "psn_9f2c4a7b1d3e",
			"url": "https://portal.bachs.io/s/6Yc0nQpR2vX1sK7fLbA9tE"
		}`)
	})

	session, _, err := c.CustomerSessions.Create(context.Background(), "cust_1a2b3c4d5e6f")
	if err != nil {
		t.Fatalf("Create returned error: %v", err)
	}
	if session.ID != "psn_9f2c4a7b1d3e" {
		t.Errorf("ID = %q, want psn_9f2c4a7b1d3e", session.ID)
	}
	if session.URL != "https://portal.bachs.io/s/6Yc0nQpR2vX1sK7fLbA9tE" {
		t.Errorf("URL = %q", session.URL)
	}
}
