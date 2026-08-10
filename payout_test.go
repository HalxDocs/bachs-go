package bachs

import (
	"context"
	"io"
	"net/http"
	"testing"
)

const payoutExample = `{
	"withdrawal_id": "wd_1a2b3c4d5e6f",
	"organization_id": "org_abc123",
	"amount": "100.00",
	"currency": "USD",
	"reference": "WD-20260222-001",
	"from_currency": "USD",
	"to_currency": "NGN",
	"from_amount": "100.00",
	"to_amount": "150000.00",
	"status": "processing",
	"payout_method": "BANK_TRANSFER",
	"destination": "058 • 0123456789",
	"created_at": "2026-02-22T12:00:00.000Z",
	"completed_at": null
}`

func TestGetSupportedCurrencies(t *testing.T) {
	c := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Errorf("method = %s, want GET", r.Method)
		}
		if got := r.URL.RequestURI(); got != "/v1/payouts/supported-currencies?method=BANK_TRANSFER" {
			t.Errorf("path = %q", got)
		}
		io.WriteString(w, `{
			"method": "BANK_TRANSFER",
			"currencies": ["NGN"]
		}`)
	})

	out, _, err := c.Payouts.GetSupportedCurrencies(context.Background(), "BANK_TRANSFER")
	if err != nil {
		t.Fatalf("GetSupportedCurrencies returned error: %v", err)
	}
	if out.Method != "BANK_TRANSFER" || len(out.Currencies) != 1 || out.Currencies[0] != "NGN" {
		t.Errorf("out = %+v", out)
	}
}

func TestCreateQuote(t *testing.T) {
	c := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("method = %s, want POST", r.Method)
		}
		if r.URL.Path != "/v1/payouts/quotes" {
			t.Errorf("path = %q", r.URL.Path)
		}
		io.WriteString(w, `{
			"quote_id": "pqt_1a2b3c4d5e6f",
			"from_currency": "USD",
			"to_currency": "NGN",
			"from_amount": "100.00",
			"to_amount": "150000.00",
			"exchange_rate": "1500.00",
			"expires_at": "2026-02-22T12:31:00+00:00"
		}`)
	})

	quote, _, err := c.Payouts.CreateQuote(context.Background(), CreatePayoutQuoteRequest{
		FromCurrency: "USD",
		ToCurrency:   "NGN",
		Amount:       "100.00",
	})
	if err != nil {
		t.Fatalf("CreateQuote returned error: %v", err)
	}
	if quote.QuoteID != "pqt_1a2b3c4d5e6f" || quote.ExchangeRate != "1500.00" {
		t.Errorf("quote = %+v", quote)
	}
	if quote.ToAmount != "150000.00" {
		t.Errorf("ToAmount = %q, want 150000.00", quote.ToAmount)
	}
}

func TestListDestinations(t *testing.T) {
	c := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Errorf("method = %s, want GET", r.Method)
		}
		if r.URL.Path != "/v1/payouts/destinations" {
			t.Errorf("path = %q", r.URL.Path)
		}
		io.WriteString(w, `{
			"destinations": [
				{
					"id": "dest_1a2b3c4d5e6f",
					"organization_id": "org_abc123",
					"env": "live",
					"destination_type": "bank_account",
					"currency": "NGN",
					"label": "My GTBank Savings",
					"account_number": "0123456789",
					"account_name": "JOHN DOE",
					"bank_code": "058",
					"bank_name": "Guaranty Trust Bank",
					"phone_number": null,
					"mobile_provider": null,
					"wallet_address": null,
					"network": null,
					"is_active": true,
					"metadata": null,
					"created_at": "2026-01-24T14:30:00.000Z",
					"updated_at": "2026-01-24T14:30:00.000Z"
				}
			],
			"total": 1
		}`)
	})

	list, _, err := c.Payouts.ListDestinations(context.Background())
	if err != nil {
		t.Fatalf("ListDestinations returned error: %v", err)
	}
	if len(list.Destinations) != 1 || list.Total != 1 {
		t.Fatalf("list = %+v", list)
	}
	d := list.Destinations[0]
	if d.ID != "dest_1a2b3c4d5e6f" || d.DestinationType != "bank_account" {
		t.Errorf("destination = %+v", d)
	}
	if d.BankCode == nil || *d.BankCode != "058" {
		t.Errorf("BankCode = %v", d.BankCode)
	}
}

func TestCreateDestination(t *testing.T) {
	c := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("method = %s, want POST", r.Method)
		}
		if r.URL.Path != "/v1/payouts/destinations" {
			t.Errorf("path = %q", r.URL.Path)
		}
		io.WriteString(w, `{
			"id": "dest_1a2b3c4d5e6f",
			"organization_id": "org_abc123",
			"env": "live",
			"destination_type": "bank_account",
			"currency": "NGN",
			"label": "My GTBank Savings",
			"account_number": "0123456789",
			"account_name": "JOHN DOE",
			"bank_code": "058",
			"bank_name": "Guaranty Trust Bank",
			"is_active": true,
			"created_at": "2026-01-24T14:30:00.000Z",
			"updated_at": "2026-01-24T14:30:00.000Z"
		}`)
	})

	dest, _, err := c.Payouts.CreateDestination(context.Background(), CreatePayoutDestinationRequest{
		DestinationType: "bank_account",
		Currency:        "NGN",
		AccountNumber:   "0123456789",
		AccountName:     "John Doe",
		BankCode:        "058",
	})
	if err != nil {
		t.Fatalf("CreateDestination returned error: %v", err)
	}
	if dest.ID != "dest_1a2b3c4d5e6f" || dest.Currency != "NGN" {
		t.Errorf("dest = %+v", dest)
	}
	if dest.AccountName == nil || *dest.AccountName != "JOHN DOE" {
		t.Errorf("AccountName = %v", dest.AccountName)
	}
}

func TestUpdateDestination(t *testing.T) {
	c := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPut {
			t.Errorf("method = %s, want PUT", r.Method)
		}
		if r.URL.Path != "/v1/payouts/destinations/dest_1a2b3c4d5e6f" {
			t.Errorf("path = %q", r.URL.Path)
		}
		io.WriteString(w, `{
			"id": "dest_1a2b3c4d5e6f",
			"organization_id": "org_abc123",
			"env": "live",
			"destination_type": "bank_account",
			"currency": "NGN",
			"label": "Treasury NGN Account",
			"account_number": "0123456789",
			"account_name": "JOHN DOE",
			"bank_code": "058",
			"bank_name": "Guaranty Trust Bank",
			"is_active": true,
			"created_at": "2026-01-24T14:30:00.000Z",
			"updated_at": "2026-02-22T14:30:00.000Z"
		}`)
	})

	dest, _, err := c.Payouts.UpdateDestination(context.Background(), "dest_1a2b3c4d5e6f", UpdatePayoutDestinationRequest{
		DestinationType: "bank_account",
		Currency:        "NGN",
		AccountNumber:   "0123456789",
		AccountName:     "JOHN DOE",
		BankCode:        "058",
	})
	if err != nil {
		t.Fatalf("UpdateDestination returned error: %v", err)
	}
	if dest.Label != "Treasury NGN Account" {
		t.Errorf("Label = %q", dest.Label)
	}
}

func TestDeleteDestination(t *testing.T) {
	c := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodDelete {
			t.Errorf("method = %s, want DELETE", r.Method)
		}
		if r.URL.Path != "/v1/payouts/destinations/dest_1a2b3c4d5e6f" {
			t.Errorf("path = %q", r.URL.Path)
		}
		io.WriteString(w, `{
			"success": true,
			"message": "Payout destination deleted"
		}`)
	})

	res, _, err := c.Payouts.DeleteDestination(context.Background(), "dest_1a2b3c4d5e6f")
	if err != nil {
		t.Fatalf("DeleteDestination returned error: %v", err)
	}
	if !res.Success || res.Message != "Payout destination deleted" {
		t.Errorf("res = %+v", res)
	}
}

func TestResolveBankAccount(t *testing.T) {
	c := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("method = %s, want POST", r.Method)
		}
		if r.URL.Path != "/v1/payouts/resolve-account" {
			t.Errorf("path = %q", r.URL.Path)
		}
		io.WriteString(w, `{
			"status": true,
			"message": "Account resolved successfully",
			"data": {
				"account_number": "0123456789",
				"account_name": "JOHN DOE",
				"bank_code": "058",
				"bank_name": "Guaranty Trust Bank"
			},
			"error": null
		}`)
	})

	res, _, err := c.Payouts.ResolveBankAccount(context.Background(), "058", "0123456789")
	if err != nil {
		t.Fatalf("ResolveBankAccount returned error: %v", err)
	}
	if !res.Status || res.Data == nil {
		t.Fatalf("res = %+v", res)
	}
	if res.Data.AccountName != "JOHN DOE" || res.Data.BankCode != "058" {
		t.Errorf("data = %+v", res.Data)
	}
	if res.Error != nil {
		t.Errorf("Error = %v, want nil on success", *res.Error)
	}
}

func TestListBanks(t *testing.T) {
	c := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Errorf("method = %s, want GET", r.Method)
		}
		if got := r.URL.RequestURI(); got != "/v1/payouts/banks?country_code=NG" {
			t.Errorf("path = %q", got)
		}
		io.WriteString(w, `{
			"status": true,
			"message": "Banks retrieved successfully",
			"data": [
				{
					"name": "Guaranty Trust Bank",
					"slug": "gtbank",
					"code": "058",
					"nibss_bank_code": "058",
					"country": "NG"
				}
			],
			"error": null
		}`)
	})

	list, _, err := c.Payouts.ListBanks(context.Background(), "NG")
	if err != nil {
		t.Fatalf("ListBanks returned error: %v", err)
	}
	if !list.Status || len(list.Data) != 1 {
		t.Fatalf("list = %+v", list)
	}
	if list.Data[0].Code != "058" || list.Data[0].Slug != "gtbank" {
		t.Errorf("Data[0] = %+v", list.Data[0])
	}
	if list.Data[0].NIBSSBankCode == nil || *list.Data[0].NIBSSBankCode != "058" {
		t.Errorf("NIBSSBankCode = %v", list.Data[0].NIBSSBankCode)
	}
}

func TestListPayouts(t *testing.T) {
	c := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Errorf("method = %s, want GET", r.Method)
		}
		if got := r.URL.RequestURI(); got != "/v1/payouts?status_filter=processing" {
			t.Errorf("path = %q", got)
		}
		io.WriteString(w, `{
			"total": 1,
			"items": [`+payoutExample+`]
		}`)
	})

	page, _, err := c.Payouts.List(context.Background(), ListParams{StatusFilter: "processing"})
	if err != nil {
		t.Fatalf("List returned error: %v", err)
	}
	if len(page.Items) != 1 {
		t.Fatalf("len(Items) = %d, want 1", len(page.Items))
	}
	if page.Items[0].WithdrawalID != "wd_1a2b3c4d5e6f" || page.Items[0].Status != "processing" {
		t.Errorf("Items[0] = %+v", page.Items[0])
	}
	if page.Items[0].Amount != "100.00" {
		t.Errorf("Amount = %q, want decimal string 100.00", page.Items[0].Amount)
	}
	if page.Pagination.Total != 1 {
		t.Errorf("Pagination.Total = %d, want 1", page.Pagination.Total)
	}
}

func TestGetPayout(t *testing.T) {
	c := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Errorf("method = %s, want GET", r.Method)
		}
		if r.URL.Path != "/v1/payouts/wd_1a2b3c4d5e6f" {
			t.Errorf("path = %q", r.URL.Path)
		}
		io.WriteString(w, payoutExample)
	})

	p, _, err := c.Payouts.Get(context.Background(), "wd_1a2b3c4d5e6f")
	if err != nil {
		t.Fatalf("Get returned error: %v", err)
	}
	if p.WithdrawalID != "wd_1a2b3c4d5e6f" || p.ToAmount == nil || *p.ToAmount != "150000.00" {
		t.Errorf("payout = %+v", p)
	}
	if p.Reference == nil || *p.Reference != "WD-20260222-001" {
		t.Errorf("Reference = %v", p.Reference)
	}
	if p.CompletedAt != nil {
		t.Errorf("CompletedAt = %v, want nil", p.CompletedAt)
	}
}

func TestCreateWithdrawal(t *testing.T) {
	c := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("method = %s, want POST", r.Method)
		}
		if r.URL.Path != "/v1/payouts/withdrawals" {
			t.Errorf("path = %q", r.URL.Path)
		}
		io.WriteString(w, `{
			"withdrawal_id": "wd_1a2b3c4d5e6f",
			"status": "pending",
			"provider_reference": null
		}`)
	})

	res, _, err := c.Payouts.CreateWithdrawal(context.Background(), CreateWithdrawalRequest{
		FromCurrency:  "USD",
		ToCurrency:    "NGN",
		Amount:        "100.00",
		PaymentMethod: "BANK_TRANSFER",
		Reference:     "WD-20260222-001",
		Email:         "ops@example.com",
		AccountNumber: "0123456789",
		BankCode:      "058",
	})
	if err != nil {
		t.Fatalf("CreateWithdrawal returned error: %v", err)
	}
	if res.WithdrawalID != "wd_1a2b3c4d5e6f" || res.Status != "pending" {
		t.Errorf("res = %+v", res)
	}
	if res.ProviderReference != nil {
		t.Errorf("ProviderReference = %v, want nil", *res.ProviderReference)
	}
}
