package bachs

import (
	"net/url"
	"strconv"
)

// Pagination describes one page of a list response. Source:
// https://docs.bachs.io/guides/pagination
type Pagination struct {
	// NextCursor is the cursor to pass on the next request to fetch the
	// following page, or null on the last page.
	NextCursor *string `json:"next_cursor"`

	// PrevCursor is the cursor to pass to fetch the previous page, or null on
	// the first page.
	PrevCursor *string `json:"prev_cursor"`

	// HasMore is true when more results exist beyond this page.
	HasMore bool `json:"has_more"`

	// Limit is the page size that was actually applied, after clamping.
	Limit int `json:"limit"`

	// Offset is the record offset this page starts from.
	Offset int `json:"offset"`

	// Returned is the number of items in this page.
	Returned int `json:"returned"`

	// Total is the total number of records matching the query.
	Total int `json:"total"`
}

// Page is the response of every list endpoint: the items on this page plus
// the pagination details. One shape is reused by every List* method.
type Page[T any] struct {
	// Items are the resources on this page.
	Items []T `json:"items"`

	// Pagination describes how to page further.
	Pagination Pagination `json:"pagination"`
}

// ListParams holds the pagination (and, where an endpoint defines them,
// filtering) query parameters for a List* method.
//
// Cursor takes precedence over Offset: when Cursor is set, Offset is ignored,
// matching the API. Limit is clamped server-side to a maximum of 100.
type ListParams struct {
	// Limit is the requested page size. Zero means "use the endpoint default".
	Limit int

	// Cursor is an opaque next_cursor or prev_cursor from a previous page.
	// When set, it takes precedence over Offset.
	Cursor string

	// Offset is the record offset to start from. Ignored when Cursor is set.
	Offset int

	// IncludeArchived includes archived resources in the response. Used by
	// Products.List; the products endpoint excludes archived products unless
	// this is true.
	IncludeArchived bool

	// Search filters by email or name substring. Used by Customers.List.
	Search string

	// StatusFilter filters by resource status. Used by Payments.List, which
	// sends it as the status_filter query parameter.
	StatusFilter string

	// Status filters by resource status. Used by Refunds.List (PROCESSING,
	// SUCCESS, FAILED) and Subscriptions.List (trialing, active, past_due,
	// unpaid, canceled), which send it as the status query parameter.
	Status string

	// CustomerID returns only resources belonging to this customer. Used by
	// Subscriptions.List.
	CustomerID string

	// ConnectedAccountID returns only resources involving this connected
	// account. Used by Transfers.List.
	ConnectedAccountID string
}

// queryValues renders the params as URL query values. Fields that are unset
// are omitted; cursor wins over offset.
func (p ListParams) queryValues() url.Values {
	v := url.Values{}
	if p.Limit > 0 {
		v.Set("limit", strconv.Itoa(p.Limit))
	}
	if p.Cursor != "" {
		v.Set("cursor", p.Cursor)
	} else if p.Offset > 0 {
		v.Set("offset", strconv.Itoa(p.Offset))
	}
	if p.IncludeArchived {
		v.Set("include_archived", "true")
	}
	if p.Search != "" {
		v.Set("search", p.Search)
	}
	if p.StatusFilter != "" {
		v.Set("status_filter", p.StatusFilter)
	}
	if p.Status != "" {
		v.Set("status", p.Status)
	}
	if p.CustomerID != "" {
		v.Set("customer_id", p.CustomerID)
	}
	if p.ConnectedAccountID != "" {
		v.Set("connected_account_id", p.ConnectedAccountID)
	}
	return v
}

// queryPath appends the ListParams as a query string to path, if any are set.
// Used by every List* method to build its request URL.
func queryPath(path string, params ListParams) string {
	if q := params.queryValues().Encode(); q != "" {
		return path + "?" + q
	}
	return path
}

// pageEnvelope is the internal decode target for list responses. Most list
// endpoints return { items, pagination }; a few (refunds, transfers, connected
// accounts) return top-level { items, total, limit, offset } instead. This
// captures both so every List* method can surface a uniform *Page[T].
type pageEnvelope[T any] struct {
	Items      []T        `json:"items"`
	Pagination Pagination `json:"pagination"`

	// Top-level counts, present on list endpoints without a pagination object.
	Total  *int `json:"total"`
	Limit  *int `json:"limit"`
	Offset *int `json:"offset"`
}

// page converts the envelope into the uniform Page[T] shape, folding the
// top-level counts into Pagination when no pagination object was present.
func (e *pageEnvelope[T]) page() *Page[T] {
	p := &Page[T]{Items: e.Items, Pagination: e.Pagination}
	if e.Total != nil {
		p.Pagination.Total = *e.Total
		if e.Limit != nil {
			p.Pagination.Limit = *e.Limit
		}
		if e.Offset != nil {
			p.Pagination.Offset = *e.Offset
		}
		p.Pagination.Returned = len(e.Items)
	}
	return p
}
