package bachs

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/url"
	"testing"
)

func TestListParamsQueryValues(t *testing.T) {
	cases := []struct {
		name   string
		params ListParams
		want   url.Values
	}{
		{
			name:   "empty",
			params: ListParams{},
			want:   url.Values{},
		},
		{
			name:   "limit and offset",
			params: ListParams{Limit: 20, Offset: 40},
			want:   url.Values{"limit": {"20"}, "offset": {"40"}},
		},
		{
			name:   "cursor wins over offset",
			params: ListParams{Limit: 20, Cursor: "cur_20", Offset: 40},
			want:   url.Values{"limit": {"20"}, "cursor": {"cur_20"}},
		},
		{
			name:   "product filters",
			params: ListParams{Limit: 10, IncludeArchived: true},
			want:   url.Values{"limit": {"10"}, "include_archived": {"true"}},
		},
		{
			name:   "customer search",
			params: ListParams{Search: "jane@example.com"},
			want:   url.Values{"search": {"jane@example.com"}},
		},
		{
			name:   "status filters and ids",
			params: ListParams{StatusFilter: "succeeded", Status: "active", CustomerID: "cust_1", ConnectedAccountID: "org_2"},
			want: url.Values{
				"status_filter":        {"succeeded"},
				"status":               {"active"},
				"customer_id":          {"cust_1"},
				"connected_account_id": {"org_2"},
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := tc.params.queryValues()
			if len(got) != len(tc.want) {
				t.Errorf("queryValues() = %v, want %v", got, tc.want)
				return
			}
			for k, vv := range tc.want {
				if got.Get(k) != vv[0] {
					t.Errorf("queryValues()[%q] = %q, want %q", k, got.Get(k), vv[0])
				}
			}
		})
	}
}

func TestPageDecodeWithCursors(t *testing.T) {
	const body = `{
		"items": [{"id": "prod_1"}, {"id": "prod_2"}],
		"pagination": {
			"next_cursor": "cur_20",
			"prev_cursor": null,
			"has_more": true,
			"limit": 20,
			"offset": 0,
			"returned": 2,
			"total": 47
		}
	}`

	c := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		io.WriteString(w, body)
	})

	var page Page[struct {
		ID string `json:"id"`
	}]
	_, err := c.do(context.Background(), http.MethodGet, "/products", nil, &page)
	if err != nil {
		t.Fatalf("do returned error: %v", err)
	}

	if len(page.Items) != 2 {
		t.Fatalf("len(Items) = %d, want 2", len(page.Items))
	}
	if page.Items[0].ID != "prod_1" {
		t.Errorf("Items[0].ID = %q, want prod_1", page.Items[0].ID)
	}
	if page.Pagination.NextCursor == nil || *page.Pagination.NextCursor != "cur_20" {
		t.Errorf("NextCursor = %v, want cur_20", page.Pagination.NextCursor)
	}
	if page.Pagination.PrevCursor != nil {
		t.Errorf("PrevCursor = %v, want null", page.Pagination.PrevCursor)
	}
	if !page.Pagination.HasMore {
		t.Error("HasMore = false, want true")
	}
	if page.Pagination.Total != 47 {
		t.Errorf("Total = %d, want 47", page.Pagination.Total)
	}
}

func TestPageDecodeLastPageNullCursors(t *testing.T) {
	const body = `{
		"items": [],
		"pagination": {
			"next_cursor": null,
			"prev_cursor": "cur_20",
			"has_more": false,
			"limit": 20,
			"offset": 40,
			"returned": 0,
			"total": 40
		}
	}`

	c := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		io.WriteString(w, body)
	})

	var page Page[map[string]any]
	_, err := c.do(context.Background(), http.MethodGet, "/customers", nil, &page)
	if err != nil {
		t.Fatalf("do returned error: %v", err)
	}

	if page.Pagination.NextCursor != nil {
		t.Errorf("NextCursor = %v, want null", *page.Pagination.NextCursor)
	}
	if page.Pagination.PrevCursor == nil || *page.Pagination.PrevCursor != "cur_20" {
		t.Errorf("PrevCursor = %v, want cur_20", page.Pagination.PrevCursor)
	}
	if page.Pagination.HasMore {
		t.Error("HasMore = true, want false")
	}
}

// TestPageEnvelopeFlatList verifies that list endpoints returning a flat
// { items, total, limit, offset } shape (refunds, transfers, connected
// accounts) still decode into the uniform Page[T].
func TestPageEnvelopeFlatList(t *testing.T) {
	c := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		io.WriteString(w, `{
			"items": [{"refund_id": "ref_1"}],
			"total": 9,
			"limit": 20,
			"offset": 0
		}`)
	})

	var env pageEnvelope[map[string]any]
	_, err := c.do(context.Background(), http.MethodGet, "/refunds", nil, &env)
	if err != nil {
		t.Fatalf("do returned error: %v", err)
	}
	page := env.page()

	if len(page.Items) != 1 {
		t.Fatalf("len(Items) = %d, want 1", len(page.Items))
	}
	if page.Pagination.Total != 9 {
		t.Errorf("Pagination.Total = %d, want 9 (folded from the flat total)", page.Pagination.Total)
	}
	if page.Pagination.Limit != 20 || page.Pagination.Offset != 0 {
		t.Errorf("Pagination = %+v, want limit 20 offset 0", page.Pagination)
	}
	if page.Pagination.Returned != 1 {
		t.Errorf("Pagination.Returned = %d, want 1", page.Pagination.Returned)
	}
}

func TestPageEnvelopeStandardList(t *testing.T) {
	c := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		io.WriteString(w, `{
			"items": [{"id": "x"}],
			"pagination": {"next_cursor": "cur_1", "has_more": true, "limit": 20, "offset": 0, "returned": 1, "total": 3}
		}`)
	})

	var env pageEnvelope[map[string]any]
	_, err := c.do(context.Background(), http.MethodGet, "/products", nil, &env)
	if err != nil {
		t.Fatalf("do returned error: %v", err)
	}
	page := env.page()

	if page.Pagination.NextCursor == nil || *page.Pagination.NextCursor != "cur_1" {
		t.Errorf("NextCursor = %v, want cur_1", page.Pagination.NextCursor)
	}
	if page.Pagination.Total != 3 {
		t.Errorf("Pagination.Total = %d, want 3", page.Pagination.Total)
	}
}

func TestPaginationJSONTags(t *testing.T) {
	// The Pagination type must serialize with the exact wire field names.
	b, err := json.Marshal(Pagination{HasMore: true, Total: 5})
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	var m map[string]any
	if err := json.Unmarshal(b, &m); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	for _, key := range []string{"next_cursor", "prev_cursor", "has_more", "limit", "offset", "returned", "total"} {
		if _, ok := m[key]; !ok {
			t.Errorf("Pagination JSON is missing key %q: %s", key, b)
		}
	}
}
