package bachs

import (
	"context"
	"net/http"
	"net/url"
	"time"
)

// ProductService provides methods for managing the billing catalog: products
// are sold through checkout sessions and subscriptions. Creating a product
// requires the products:write scope; reading requires products:read.
type ProductService struct {
	service
}

// Cadence describes a duration as a frequency of interval units: a billing
// cycle (how often a product bills) or a trial period (how long it lasts).
// Source: https://docs.bachs.io/api-reference/products/object
type Cadence struct {
	// Interval is "day", "week", "month", or "year".
	Interval string `json:"interval"`

	// Frequency is the number of intervals per cycle.
	Frequency int `json:"frequency"`
}

// ProductPrice is a product's price in one currency. For create and update
// requests, price_type is "fixed" (a set Amount), "free" (no charge), or
// "custom" (the customer pays what they want within optional bounds).
type ProductPrice struct {
	// Currency is the product's primary currency (for example "USD" or
	// "NGN").
	Currency string `json:"currency,omitempty"`

	// PriceType is "fixed", "free", or "custom".
	PriceType string `json:"price_type,omitempty"`

	// Amount is the price as a decimal string (for example "29.00").
	// Required when price_type is "fixed"; omit for "free" and "custom".
	Amount string `json:"amount,omitempty"`

	// PresetAmount is the amount prefilled at checkout for a custom price.
	PresetAmount *string `json:"preset_amount,omitempty"`

	// MinimumAmount is the least the customer can pay for a custom price.
	MinimumAmount *string `json:"minimum_amount,omitempty"`

	// MaximumAmount is the most the customer can pay for a custom price.
	MaximumAmount *string `json:"maximum_amount,omitempty"`

	// CurrencyOptions sets prices in additional currencies. Each entry is one
	// additional currency and cannot repeat the primary currency.
	CurrencyOptions []CurrencyOption `json:"currency_options,omitempty"`
}

// CurrencyOption is a product's price in one additional currency.
type CurrencyOption struct {
	// Currency is a supported additional currency (for example "GHS", "KES").
	Currency string `json:"currency,omitempty"`

	// Amount is the price as a decimal string. Required when the product's
	// price_type is "fixed"; omit for "free" and "custom".
	Amount string `json:"amount,omitempty"`

	// PresetAmount is the prefilled amount for a custom price in this
	// currency.
	PresetAmount *string `json:"preset_amount,omitempty"`

	// MinimumAmount is the least the customer can pay for a custom price in
	// this currency.
	MinimumAmount *string `json:"minimum_amount,omitempty"`

	// MaximumAmount is the most the customer can pay for a custom price in
	// this currency.
	MaximumAmount *string `json:"maximum_amount,omitempty"`
}

// ProductPriceSummary is one entry in a product's Prices array: the price for
// one currency.
type ProductPriceSummary struct {
	// Currency is the currency code.
	Currency string `json:"currency"`

	// Amount is the price as a decimal string.
	Amount string `json:"amount"`

	// MinimumAmount is the custom-price lower bound, when applicable.
	MinimumAmount *string `json:"minimum_amount"`

	// MaximumAmount is the custom-price upper bound, when applicable.
	MaximumAmount *string `json:"maximum_amount"`

	// IsDefault is true when this is the product's primary price.
	IsDefault bool `json:"is_default,omitempty"`
}

// MediaItem is an image attached to a product.
type MediaItem struct {
	// ID is the upload's identifier.
	ID string `json:"id"`

	// URL the media is served from.
	URL *string `json:"url"`

	// FileName of the uploaded file.
	FileName string `json:"file_name"`

	// MIMEType of the uploaded file.
	MIMEType string `json:"mime_type"`

	// FileSizeBytes is the uploaded file's size.
	FileSizeBytes int `json:"file_size_bytes"`

	// CreatedAt is when the upload was created.
	CreatedAt time.Time `json:"created_at"`
}

// Product is an item you sell, with its pricing. A BillingCycle makes the
// product recurring; omitting it makes the product one-time. Source:
// https://docs.bachs.io/api-reference/products/object
type Product struct {
	// ID is the unique identifier, prefixed with "prod_".
	ID string `json:"id"`

	// OrganizationID is the organization that owns the product.
	OrganizationID string `json:"organization_id"`

	// Name is the display name shown to customers at checkout.
	Name string `json:"name"`

	// Description is the optional description. Null when not set.
	Description *string `json:"description"`

	// Price is the primary price.
	Price ProductPrice `json:"price"`

	// Status is "active" or "archived".
	Status string `json:"status"`

	// Metadata is the key-value data attached at creation, returned unchanged.
	Metadata map[string]any `json:"metadata"`

	// Media are the images attached to the product. Empty when none are set.
	Media []MediaItem `json:"media"`

	// ActorID identifies the user or key that created the product.
	ActorID string `json:"actor_id"`

	// CreatedAt is when the product was created.
	CreatedAt time.Time `json:"created_at"`

	// UpdatedAt is when the product was last updated.
	UpdatedAt time.Time `json:"updated_at"`

	// ArchivedAt is set when the product is archived; null while active.
	ArchivedAt *time.Time `json:"archived_at"`

	// BillingCycle makes the product recurring; null for one-time products.
	BillingCycle *Cadence `json:"billing_cycle"`

	// TrialPeriod is the free trial before the first charge, if any.
	TrialPeriod *Cadence `json:"trial_period"`

	// Prices are all prices configured on the product, one per currency.
	Prices []ProductPriceSummary `json:"prices"`

	// TotalPayments is the running count of completed payments for the
	// product.
	TotalPayments int `json:"total_payments"`

	// TotalAmount is the running total collected, as a decimal string in the
	// product currency.
	TotalAmount string `json:"total_amount"`
}

// CreateProductRequest is the payload for Products.Create. Source:
// https://docs.bachs.io/api-reference/products/create-a-product
type CreateProductRequest struct {
	// Name is the display name of the product, shown to customers at
	// checkout. Required.
	Name string `json:"name"`

	// Description is an optional description shown to customers at checkout.
	Description *string `json:"description,omitempty"`

	// Price is the product's pricing. Required.
	Price ProductPrice `json:"price"`

	// Metadata is up to 20 key/value pairs for your own reference.
	Metadata map[string]any `json:"metadata,omitempty"`

	// BillingCycle makes the product recurring. Omit it for a one-time
	// product. Once set, it is immutable.
	BillingCycle *Cadence `json:"billing_cycle,omitempty"`

	// TrialPeriod is a free trial before the first charge, for example
	// {Interval: "day", Frequency: 14}.
	TrialPeriod *Cadence `json:"trial_period,omitempty"`
}

// UpdateProductRequest is the payload for Products.Update. Only the fields
// you set are changed. Source:
// https://docs.bachs.io/api-reference/products/update-a-product
type UpdateProductRequest struct {
	// Name replaces the product's display name.
	Name string `json:"name,omitempty"`

	// Description replaces the product's description.
	Description *string `json:"description,omitempty"`

	// Metadata replaces the product's key-value metadata.
	Metadata map[string]any `json:"metadata,omitempty"`

	// Media is the ordered list of upload IDs. Replaces existing media.
	Media []string `json:"media,omitempty"`

	// Price updates the price (only valid for fixed-price products).
	Price *UpdateProductPrice `json:"price,omitempty"`

	// BillingCycle can be set only if not already set; it is immutable once
	// the product has one.
	BillingCycle *Cadence `json:"billing_cycle,omitempty"`

	// TrialPeriod sets or replaces the free trial.
	TrialPeriod *Cadence `json:"trial_period,omitempty"`
}

// UpdateProductPrice is the price portion of Products.Update.
type UpdateProductPrice struct {
	// Amount is the new price as a decimal string (for example "39.00").
	Amount string `json:"amount,omitempty"`

	// CurrencyOptions fully replaces the multi-currency prices. Omit to leave
	// unchanged.
	CurrencyOptions []CurrencyOption `json:"currency_options,omitempty"`
}

// Create creates a product with its pricing. Add a BillingCycle to make it
// recurring, or omit it for a one-time product.
func (s *ProductService) Create(ctx context.Context, req CreateProductRequest, opts ...RequestOption) (*Product, *ResponseMeta, error) {
	var out Product
	meta, err := s.request(ctx, http.MethodPost, "/products", req, &out, opts...)
	if err != nil {
		return nil, meta, err
	}
	return &out, meta, nil
}

// Get retrieves a single product by its ID.
func (s *ProductService) Get(ctx context.Context, productID string) (*Product, *ResponseMeta, error) {
	var out Product
	meta, err := s.request(ctx, http.MethodGet, "/products/"+url.PathEscape(productID), nil, &out)
	if err != nil {
		return nil, meta, err
	}
	return &out, meta, nil
}

// List returns a paginated list of products, most recent first. Archived
// products are excluded unless ListParams.IncludeArchived is set.
func (s *ProductService) List(ctx context.Context, params ListParams) (*Page[Product], *ResponseMeta, error) {
	var env pageEnvelope[Product]
	meta, err := s.request(ctx, http.MethodGet, queryPath("/products", params), nil, &env)
	if err != nil {
		return nil, meta, err
	}
	return env.page(), meta, nil
}

// Update changes a product's name, description, metadata, media, price, and
// (if not yet set) its billing_cycle and trial_period. A billing_cycle is
// immutable once set.
func (s *ProductService) Update(ctx context.Context, productID string, req UpdateProductRequest) (*Product, *ResponseMeta, error) {
	var out Product
	meta, err := s.request(ctx, http.MethodPatch, "/products/"+url.PathEscape(productID), req, &out)
	if err != nil {
		return nil, meta, err
	}
	return &out, meta, nil
}

// Archive retires a product so it can no longer be used in new checkouts or
// subscriptions. Existing subscriptions keep billing. Idempotent.
func (s *ProductService) Archive(ctx context.Context, productID string) (*Product, *ResponseMeta, error) {
	var out Product
	meta, err := s.request(ctx, http.MethodPost, "/products/"+url.PathEscape(productID)+"/archive", nil, &out)
	if err != nil {
		return nil, meta, err
	}
	return &out, meta, nil
}

// Unarchive restores an archived product to active status so it can be used
// again. Idempotent.
func (s *ProductService) Unarchive(ctx context.Context, productID string) (*Product, *ResponseMeta, error) {
	var out Product
	meta, err := s.request(ctx, http.MethodPost, "/products/"+url.PathEscape(productID)+"/unarchive", nil, &out)
	if err != nil {
		return nil, meta, err
	}
	return &out, meta, nil
}
