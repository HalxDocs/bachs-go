package bachs

import (
	"context"
	"io"
	"net/http"
	"testing"
)

// The fixtures below mirror the live sandbox responses captured on
// 2026-10-07: reference banks (254 entries), momo providers, business
// structures, and product categories.

func TestReferenceListBanks(t *testing.T) {
	c := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		if got := r.URL.RequestURI(); got != "/v1/reference/banks?country=NG" {
			t.Errorf("request URI = %q, want the country filter", got)
		}
		io.WriteString(w, `{
			"country": "NG",
			"banks": [
				{"name": "Guaranty Trust Bank", "code": "058"},
				{"name": "IBILE Microfinance Bank", "code": "090118"}
			]
		}`)
	})

	list, _, err := c.Reference.ListBanks(context.Background(), "NG")
	if err != nil {
		t.Fatalf("ListBanks returned error: %v", err)
	}
	if list.Country != "NG" {
		t.Errorf("Country = %q, want NG", list.Country)
	}
	if len(list.Banks) != 2 || list.Banks[0].Code != "058" {
		t.Errorf("Banks = %+v", list.Banks)
	}
}

func TestReferenceListBanksWithoutCountry(t *testing.T) {
	c := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		if got := r.URL.RequestURI(); got != "/v1/reference/banks" {
			t.Errorf("request URI = %q, want no query string", got)
		}
		io.WriteString(w, `{"country": "NG", "banks": []}`)
	})

	if _, _, err := c.Reference.ListBanks(context.Background(), ""); err != nil {
		t.Fatalf("ListBanks returned error: %v", err)
	}
}

func TestReferenceListMobileMoneyProviders(t *testing.T) {
	c := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		if got := r.URL.RequestURI(); got != "/v1/reference/momo?country=GH" {
			t.Errorf("request URI = %q, want the country filter", got)
		}
		io.WriteString(w, `{"country": "GH", "providers": ["AIRTEL", "MTN"]}`)
	})

	list, _, err := c.Reference.ListMobileMoneyProviders(context.Background(), "GH")
	if err != nil {
		t.Fatalf("ListMobileMoneyProviders returned error: %v", err)
	}
	if len(list.Providers) != 2 || list.Providers[1] != "MTN" {
		t.Errorf("Providers = %v", list.Providers)
	}
}

func TestReferenceListMobileMoneyProvidersEmpty(t *testing.T) {
	c := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		io.WriteString(w, `{"country": "NG", "providers": []}`)
	})

	list, _, err := c.Reference.ListMobileMoneyProviders(context.Background(), "NG")
	if err != nil {
		t.Fatalf("ListMobileMoneyProviders returned error: %v", err)
	}
	if len(list.Providers) != 0 {
		t.Errorf("Providers = %v, want empty", list.Providers)
	}
}

func TestReferenceListBusinessStructures(t *testing.T) {
	c := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v1/reference/business-structures" {
			t.Errorf("path = %q", r.URL.Path)
		}
		io.WriteString(w, `{
			"country": "NG",
			"structures": [
				{"value": "business_name", "label": "Business Name (BN)", "description": "A sole trader or partnership registered with the CAC. No shareholders."},
				{"value": "private_incorporated", "label": "Private Company (RC / LTD)", "description": "A company limited by shares, with directors and shareholders."}
			]
		}`)
	})

	list, _, err := c.Reference.ListBusinessStructures(context.Background())
	if err != nil {
		t.Fatalf("ListBusinessStructures returned error: %v", err)
	}
	if len(list.Structures) != 2 || list.Structures[0].Value != "business_name" {
		t.Errorf("Structures = %+v", list.Structures)
	}
	if list.Structures[1].Description == "" {
		t.Error("Description is empty")
	}
}

func TestReferenceListProductCategories(t *testing.T) {
	c := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v1/reference/product-categories" {
			t.Errorf("path = %q", r.URL.Path)
		}
		io.WriteString(w, `{
			"categories": [
				{"value": "software_as_a_service_saas", "label": "Software as a service (SaaS)"}
			],
			"sections": [
				{
					"key": "software_and_digital_services",
					"label": "Software & digital services",
					"categories": [
						{"value": "software_as_a_service_saas", "label": "Software as a service (SaaS)"}
					]
				}
			]
		}`)
	})

	list, _, err := c.Reference.ListProductCategories(context.Background())
	if err != nil {
		t.Fatalf("ListProductCategories returned error: %v", err)
	}
	if len(list.Categories) != 1 || list.Categories[0].Value != "software_as_a_service_saas" {
		t.Errorf("Categories = %+v", list.Categories)
	}
	if len(list.Sections) != 1 || list.Sections[0].Key != "software_and_digital_services" {
		t.Errorf("Sections = %+v", list.Sections)
	}
}

func TestReferenceResolveBankAccount(t *testing.T) {
	c := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("method = %s, want POST", r.Method)
		}
		if r.URL.Path != "/v1/misc/bank-accounts/resolve" {
			t.Errorf("path = %q", r.URL.Path)
		}
		// Live sandbox shape for an unresolvable test number.
		io.WriteString(w, `{
			"resolved": false,
			"account_name": null,
			"account_number": null,
			"message": "Account resolution failed"
		}`)
	})

	res, _, err := c.Reference.ResolveBankAccount(context.Background(), "058", "0123456789")
	if err != nil {
		t.Fatalf("ResolveBankAccount returned error: %v", err)
	}
	if res.Resolved {
		t.Error("Resolved is true, want false for the test number")
	}
	if res.Message == nil || *res.Message != "Account resolution failed" {
		t.Errorf("Message = %v", res.Message)
	}
}
