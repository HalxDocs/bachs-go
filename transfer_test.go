package bachs

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"testing"
)

// transferExample is the exact TransferResponse example from the transfer doc
// pages (create, get).
const transferExample = `{
	"id": "tr_8c1e04a7b93f2d6540ab",
	"source": "org_9f2c4a1b7e3d5086",
	"destination": "org_4d81fa9c2b6e0357",
	"amount": "7000.00",
	"currency": "NGN",
	"status": "paid",
	"description": "Order #4471 seller share",
	"metadata": {},
	"transfer_group": "chg_2f8a71c4e05b",
	"created_at": "2026-08-07T11:04:22.518Z"
}`

// TestTransferCreate uses the exact request/response examples from
// https://docs.bachs.io/api-reference/transfers/create-a-transfer
func TestTransferCreate(t *testing.T) {
	c := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("method = %s, want POST", r.Method)
		}
		if r.URL.Path != "/v1/transfers" {
			t.Errorf("path = %q, want /v1/transfers", r.URL.Path)
		}

		body, _ := io.ReadAll(r.Body)
		var got map[string]any
		if err := json.Unmarshal(body, &got); err != nil {
			t.Fatalf("request body is not valid JSON: %v", err)
		}
		if got["destination"] != "org_4d81fa9c2b6e0357" || got["amount"] != "7000.00" || got["currency"] != "NGN" {
			t.Errorf("request body = %s", body)
		}
		if got["transfer_group"] != "chg_2f8a71c4e05b" || got["description"] != "Order #4471 seller share" {
			t.Errorf("request body = %s", body)
		}

		io.WriteString(w, transferExample)
	})

	transfer, _, err := c.Transfers.Create(context.Background(), CreateTransferRequest{
		Destination:   "org_4d81fa9c2b6e0357",
		Amount:        "7000.00",
		Currency:      "NGN",
		TransferGroup: stringPtr("chg_2f8a71c4e05b"),
		Description:   stringPtr("Order #4471 seller share"),
	})
	if err != nil {
		t.Fatalf("Create returned error: %v", err)
	}

	if transfer.ID != "tr_8c1e04a7b93f2d6540ab" {
		t.Errorf("ID = %q", transfer.ID)
	}
	if transfer.Source != "org_9f2c4a1b7e3d5086" || transfer.Destination != "org_4d81fa9c2b6e0357" {
		t.Errorf("Source/Destination = %s/%s", transfer.Source, transfer.Destination)
	}
	if transfer.Status != "paid" {
		t.Errorf("Status = %q, want paid", transfer.Status)
	}
	if transfer.TransferGroup == nil || *transfer.TransferGroup != "chg_2f8a71c4e05b" {
		t.Errorf("TransferGroup = %v", transfer.TransferGroup)
	}
}

// TestTransferCreateAsConnectedAccount verifies the X-Connected-Account-ID
// header is sent when acting on a connected account's behalf.
func TestTransferCreateAsConnectedAccount(t *testing.T) {
	c := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		if got := r.Header.Get(headerConnectedAccountID); got != "org_4d81fa9c2b6e0357" {
			t.Errorf("X-Connected-Account-ID = %q, want org_4d81fa9c2b6e0357", got)
		}
		io.WriteString(w, transferExample)
	})

	_, _, err := c.Transfers.Create(context.Background(), CreateTransferRequest{
		Destination: "self",
		Amount:      "1000.00",
		Currency:    "NGN",
	}, WithConnectedAccount("org_4d81fa9c2b6e0357"))
	if err != nil {
		t.Fatalf("Create returned error: %v", err)
	}
}

func TestTransferGet(t *testing.T) {
	c := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Errorf("method = %s, want GET", r.Method)
		}
		if r.URL.Path != "/v1/transfers/tr_8c1e04a7b93f2d6540ab" {
			t.Errorf("path = %q", r.URL.Path)
		}
		io.WriteString(w, transferExample)
	})

	transfer, _, err := c.Transfers.Get(context.Background(), "tr_8c1e04a7b93f2d6540ab")
	if err != nil {
		t.Fatalf("Get returned error: %v", err)
	}
	if transfer.Amount != "7000.00" || transfer.Currency != "NGN" {
		t.Errorf("Amount/Currency = %s/%s", transfer.Amount, transfer.Currency)
	}
}

// TestTransferList verifies the flat { items, total, limit, offset } list
// shape folds into Page[Transfer].
func TestTransferList(t *testing.T) {
	c := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Errorf("method = %s, want GET", r.Method)
		}
		if r.URL.Path != "/v1/transfers" {
			t.Errorf("path = %q", r.URL.Path)
		}
		if got := r.URL.Query().Get("connected_account_id"); got != "org_4d81fa9c2b6e0357" {
			t.Errorf("connected_account_id = %q", got)
		}
		io.WriteString(w, `{
			"items": [`+transferExample+`],
			"total": 2,
			"limit": 20,
			"offset": 0
		}`)
	})

	page, _, err := c.Transfers.List(context.Background(), ListParams{ConnectedAccountID: "org_4d81fa9c2b6e0357"})
	if err != nil {
		t.Fatalf("List returned error: %v", err)
	}
	if len(page.Items) != 1 {
		t.Fatalf("len(Items) = %d, want 1", len(page.Items))
	}
	if page.Items[0].ID != "tr_8c1e04a7b93f2d6540ab" {
		t.Errorf("Items[0].ID = %q", page.Items[0].ID)
	}
	if page.Pagination.Total != 2 || page.Pagination.Limit != 20 {
		t.Errorf("Pagination = %+v", page.Pagination)
	}
}
