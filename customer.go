package bachs

import (
	"context"
	"net/http"
	"net/url"
	"time"
)

// CustomerService provides methods for managing the customers you bill
// through the Bachs API. A customer groups a buyer's payments, subscriptions,
// and saved payment methods under one record. Source:
// https://docs.bachs.io/guides/customers
type CustomerService struct {
	service
}

// CustomerBillingAddress is a customer's billing address. Line1 and Country
// are required whenever an address is supplied.
type CustomerBillingAddress struct {
	// Line1 is the street address.
	Line1 *string `json:"line1,omitempty"`

	// Line2 is the apartment, suite, unit, etc.
	Line2 *string `json:"line2,omitempty"`

	// City is the city, district, or suburb.
	City *string `json:"city,omitempty"`

	// State is the state, province, or region.
	State *string `json:"state,omitempty"`

	// PostalCode is the ZIP or postal code.
	PostalCode *string `json:"postal_code,omitempty"`

	// Country is the two-letter ISO 3166-1 alpha-2 country code.
	Country *string `json:"country,omitempty"`
}

// CreateCustomerRequest is the payload for Customers.Create. Only Email is
// required. Source:
// https://docs.bachs.io/api-reference/customers/create-a-customer
type CreateCustomerRequest struct {
	// Email of the customer, used to identify them and send receipts.
	// Required.
	Email string `json:"email"`

	// Name is the customer's full name.
	Name *string `json:"name,omitempty"`

	// PhoneNumber in E.164 format, for example "+2348012345678".
	PhoneNumber *string `json:"phone_number,omitempty"`

	// Metadata is key-value data returned unchanged.
	Metadata map[string]any `json:"metadata,omitempty"`

	// BillingAddress is optional on create. Line1 and Country are required
	// whenever an address is supplied.
	BillingAddress *CustomerBillingAddress `json:"billing_address,omitempty"`
}

// UpdateCustomerRequest is the payload for Customers.Update. Only the fields
// you send are changed. Source:
// https://docs.bachs.io/api-reference/customers/update-a-customer
type UpdateCustomerRequest struct {
	// Email of the customer.
	Email string `json:"email,omitempty"`

	// Name is the customer's full name.
	Name *string `json:"name,omitempty"`

	// PhoneNumber in E.164 format.
	PhoneNumber *string `json:"phone_number,omitempty"`

	// Metadata replaces the key-value data.
	Metadata map[string]any `json:"metadata,omitempty"`

	// BillingAddress replaces the billing address in full. Leave nil to keep
	// the address untouched. To explicitly clear it, assign a pointer to a
	// nil pointer — BillingAddress is a pointer-to-pointer precisely so the
	// request can distinguish "leave alone" from "clear":
	//
	//	var clear *CustomerBillingAddress
	//	req.BillingAddress = &clear // marshals as "billing_address": null
	BillingAddress **CustomerBillingAddress `json:"billing_address,omitempty"`
}

// Customer is the full customer object returned by Customers.Create, Get, and
// Update.
type Customer struct {
	// CustomerID is the unique identifier, prefixed with "cust_".
	CustomerID string `json:"customer_id"`

	// Email of the customer.
	Email string `json:"email"`

	// Name is the full name. Null when not set.
	Name *string `json:"name"`

	// PhoneNumber in E.164 format. Null when not set.
	PhoneNumber *string `json:"phone_number"`

	// Metadata attached to the customer.
	Metadata map[string]any `json:"metadata"`

	// CreatedAt is when the customer was created.
	CreatedAt time.Time `json:"created_at"`

	// UpdatedAt is when the customer was last updated.
	UpdatedAt time.Time `json:"updated_at"`

	// BillingAddress, or null when none is set.
	BillingAddress *CustomerBillingAddress `json:"billing_address"`
}

// CustomerListItem is the summary form of a customer returned by
// Customers.List.
type CustomerListItem struct {
	// CustomerID is the unique identifier, prefixed with "cust_".
	CustomerID string `json:"customer_id"`

	// Email of the customer.
	Email string `json:"email"`

	// Name is the full name. Null when not set.
	Name *string `json:"name"`

	// Metadata attached to the customer.
	Metadata map[string]any `json:"metadata"`

	// CreatedAt is when the customer was created.
	CreatedAt time.Time `json:"created_at"`
}

// Create creates a customer. Only Email is required.
func (s *CustomerService) Create(ctx context.Context, req CreateCustomerRequest, opts ...RequestOption) (*Customer, *ResponseMeta, error) {
	var out Customer
	meta, err := s.request(ctx, http.MethodPost, "/customers", req, &out, opts...)
	if err != nil {
		return nil, meta, err
	}
	return &out, meta, nil
}

// Get retrieves a single customer by ID.
func (s *CustomerService) Get(ctx context.Context, customerID string) (*Customer, *ResponseMeta, error) {
	var out Customer
	meta, err := s.request(ctx, http.MethodGet, "/customers/"+url.PathEscape(customerID), nil, &out)
	if err != nil {
		return nil, meta, err
	}
	return &out, meta, nil
}

// List returns a paginated list of customers, most recent first. Set
// ListParams.Search to filter by email or name substring.
func (s *CustomerService) List(ctx context.Context, params ListParams) (*Page[CustomerListItem], *ResponseMeta, error) {
	var env pageEnvelope[CustomerListItem]
	meta, err := s.request(ctx, http.MethodGet, queryPath("/customers", params), nil, &env)
	if err != nil {
		return nil, meta, err
	}
	return env.page(), meta, nil
}

// Update changes a customer. Only the fields you send are changed.
func (s *CustomerService) Update(ctx context.Context, customerID string, req UpdateCustomerRequest) (*Customer, *ResponseMeta, error) {
	var out Customer
	meta, err := s.request(ctx, http.MethodPatch, "/customers/"+url.PathEscape(customerID), req, &out)
	if err != nil {
		return nil, meta, err
	}
	return &out, meta, nil
}
