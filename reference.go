package bachs

import (
	"context"
	"net/http"
	"net/url"
)

// ReferenceService provides account-independent lookup data: the banks and
// mobile money operators payouts can go to, the business structures and
// product categories onboarding accepts, and bank-account resolution. These
// are not account-scoped, so there is no account in the path. Source:
// https://docs.bachs.io/connect/guides/api-onboarding#resolve-the-reference-data
type ReferenceService struct {
	service
}

// ReferenceBank is one bank payouts can go to. Use its Code when resolving
// an account or naming it as a payout destination.
type ReferenceBank struct {
	// Name of the bank.
	Name string `json:"name"`

	// Code to send as bank_code when resolving an account or submitting a
	// payout destination.
	Code string `json:"code"`
}

// ReferenceBankList is the response of Reference.ListBanks.
type ReferenceBankList struct {
	// Country the list was resolved for, uppercased.
	Country string `json:"country"`

	// Banks available in this country.
	Banks []ReferenceBank `json:"banks"`
}

// ReferenceMobileMoneyList is the response of
// Reference.ListMobileMoneyProviders.
type ReferenceMobileMoneyList struct {
	// Country the list was resolved for, uppercased.
	Country string `json:"country"`

	// Providers available in this country, in display order. Empty when the
	// country has none.
	Providers []string `json:"providers"`
}

// BusinessStructure is one accepted business structure for onboarding.
type BusinessStructure struct {
	// Value to submit (for example "private_incorporated").
	Value string `json:"value"`

	// Label is the human-readable name, safe to show verbatim.
	Label string `json:"label"`

	// Description explains what the structure covers.
	Description string `json:"description"`
}

// BusinessStructureList is the response of
// Reference.ListBusinessStructures.
type BusinessStructureList struct {
	// Country the list was resolved for, uppercased.
	Country string `json:"country"`

	// Structures accepted for onboarding.
	Structures []BusinessStructure `json:"structures"`
}

// ProductCategory is one product category for onboarding.
type ProductCategory struct {
	// Value to submit (for example "software_as_a_service_saas").
	Value string `json:"value"`

	// Label is the human-readable name, safe to show verbatim.
	Label string `json:"label"`
}

// ProductCategorySection groups product categories under one heading.
type ProductCategorySection struct {
	// Key identifies the section.
	Key string `json:"key"`

	// Label is the human-readable section name.
	Label string `json:"label"`

	// Categories in this section.
	Categories []ProductCategory `json:"categories"`
}

// ProductCategoryList is the response of
// Reference.ListProductCategories.
type ProductCategoryList struct {
	// Categories is the flat list of accepted product categories.
	Categories []ProductCategory `json:"categories"`

	// Sections groups the same categories under headings.
	Sections []ProductCategorySection `json:"sections"`
}

// ReferenceBankResolution is the result of Reference.ResolveBankAccount.
// Check Resolved before trusting AccountName: a number that does not match
// returns Resolved: false, not an error.
type ReferenceBankResolution struct {
	// Resolved is true when the account number was matched.
	Resolved bool `json:"resolved"`

	// AccountName registered on the account. Show it back for confirmation.
	// Null when not resolved.
	AccountName *string `json:"account_name"`

	// AccountNumber as held on record, which can be normalised from what was
	// sent. Null when not resolved.
	AccountNumber *string `json:"account_number"`

	// Message explains why the lookup did not resolve. Null on a successful
	// match.
	Message *string `json:"message"`
}

// ListBanks returns the banks payouts can go to in a country. Pass country
// as a two-letter ISO 3166-1 code; it falls back to your own when omitted.
func (s *ReferenceService) ListBanks(ctx context.Context, country string) (*ReferenceBankList, *ResponseMeta, error) {
	path := "/reference/banks"
	if country != "" {
		path += "?country=" + url.QueryEscape(country)
	}

	var out ReferenceBankList
	meta, err := s.request(ctx, http.MethodGet, path, nil, &out)
	if err != nil {
		return nil, meta, err
	}
	return &out, meta, nil
}

// ListMobileMoneyProviders returns the mobile money operators payouts can go
// to in a country. Returns an empty providers array for a country with none,
// rather than an error.
func (s *ReferenceService) ListMobileMoneyProviders(ctx context.Context, country string) (*ReferenceMobileMoneyList, *ResponseMeta, error) {
	path := "/reference/momo"
	if country != "" {
		path += "?country=" + url.QueryEscape(country)
	}

	var out ReferenceMobileMoneyList
	meta, err := s.request(ctx, http.MethodGet, path, nil, &out)
	if err != nil {
		return nil, meta, err
	}
	return &out, meta, nil
}

// ListBusinessStructures returns the business structures onboarding accepts.
func (s *ReferenceService) ListBusinessStructures(ctx context.Context) (*BusinessStructureList, *ResponseMeta, error) {
	var out BusinessStructureList
	meta, err := s.request(ctx, http.MethodGet, "/reference/business-structures", nil, &out)
	if err != nil {
		return nil, meta, err
	}
	return &out, meta, nil
}

// ListProductCategories returns the product categories onboarding accepts,
// flat and grouped into sections.
func (s *ReferenceService) ListProductCategories(ctx context.Context) (*ProductCategoryList, *ResponseMeta, error) {
	var out ProductCategoryList
	meta, err := s.request(ctx, http.MethodGet, "/reference/product-categories", nil, &out)
	if err != nil {
		return nil, meta, err
	}
	return &out, meta, nil
}

// ResolveBankAccount confirms a bank account resolves to a real account name
// before it is submitted. Resolving first turns a rejected requirement days
// later into an inline error while the account holder is still on the page.
func (s *ReferenceService) ResolveBankAccount(ctx context.Context, bankCode, accountNumber string, opts ...RequestOption) (*ReferenceBankResolution, *ResponseMeta, error) {
	req := struct {
		BankCode      string `json:"bank_code"`
		AccountNumber string `json:"account_number"`
	}{BankCode: bankCode, AccountNumber: accountNumber}

	var out ReferenceBankResolution
	meta, err := s.request(ctx, http.MethodPost, "/misc/bank-accounts/resolve", req, &out, opts...)
	if err != nil {
		return nil, meta, err
	}
	return &out, meta, nil
}
