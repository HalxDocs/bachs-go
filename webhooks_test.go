package bachs

import (
	"context"
	"io"
	"net/http"
	"testing"
)

const webhookEndpointExample = `{
	"endpoint_id": "whe_1a2b3c4d5e",
	"name": "Production events",
	"url": "https://api.example.com/webhooks/bachs",
	"enabled": true,
	"event_types": ["collection.succeeded", "collection.failed"],
	"created_at": "2026-04-27T12:00:00Z",
	"updated_at": "2026-04-27T12:00:00Z"
}`

func TestCreateWebhookEndpoint(t *testing.T) {
	c := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("method = %s, want POST", r.Method)
		}
		if r.URL.Path != "/v1/webhooks/endpoints" {
			t.Errorf("path = %q", r.URL.Path)
		}
		io.WriteString(w, `{
			"endpoint_id": "whe_1a2b3c4d5e",
			"name": "Production events",
			"url": "https://api.example.com/webhooks/bachs",
			"enabled": true,
			"event_types": ["collection.succeeded", "collection.failed"],
			"created_at": "2026-04-27T12:00:00Z",
			"updated_at": "2026-04-27T12:00:00Z",
			"signing_secret": "whsec_1a2b3c4d5e6f7g8h"
		}`)
	})

	ep, _, err := c.Webhooks.CreateEndpoint(context.Background(), CreateWebhookEndpointRequest{
		Name:       "Production events",
		URL:        "https://api.example.com/webhooks/bachs",
		EventTypes: []string{EventTypeCollectionSucceeded, EventTypeCollectionFailed},
	})
	if err != nil {
		t.Fatalf("CreateEndpoint returned error: %v", err)
	}
	if ep.EndpointID != "whe_1a2b3c4d5e" || !ep.Enabled {
		t.Errorf("endpoint = %+v", ep)
	}
	if ep.SigningSecret != "whsec_1a2b3c4d5e6f7g8h" {
		t.Errorf("SigningSecret = %q", ep.SigningSecret)
	}
	if len(ep.EventTypes) != 2 || ep.EventTypes[0] != "collection.succeeded" {
		t.Errorf("EventTypes = %v", ep.EventTypes)
	}
}

func TestListWebhookEndpoints(t *testing.T) {
	c := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Errorf("method = %s, want GET", r.Method)
		}
		if r.URL.Path != "/v1/webhooks/endpoints" {
			t.Errorf("path = %q", r.URL.Path)
		}
		// The list returns a flat array, not an items envelope.
		io.WriteString(w, `[`+webhookEndpointExample+`]`)
	})

	eps, _, err := c.Webhooks.ListEndpoints(context.Background())
	if err != nil {
		t.Fatalf("ListEndpoints returned error: %v", err)
	}
	if len(eps) != 1 {
		t.Fatalf("len = %d, want 1", len(eps))
	}
	if eps[0].EndpointID != "whe_1a2b3c4d5e" || eps[0].URL != "https://api.example.com/webhooks/bachs" {
		t.Errorf("eps[0] = %+v", eps[0])
	}
}

func TestGetWebhookEndpoint(t *testing.T) {
	c := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Errorf("method = %s, want GET", r.Method)
		}
		if r.URL.Path != "/v1/webhooks/endpoints/whe_1a2b3c4d5e" {
			t.Errorf("path = %q", r.URL.Path)
		}
		io.WriteString(w, webhookEndpointExample)
	})

	ep, _, err := c.Webhooks.GetEndpoint(context.Background(), "whe_1a2b3c4d5e")
	if err != nil {
		t.Fatalf("GetEndpoint returned error: %v", err)
	}
	if ep.EndpointID != "whe_1a2b3c4d5e" || ep.Name != "Production events" {
		t.Errorf("endpoint = %+v", ep)
	}
}

func TestUpdateWebhookEndpoint(t *testing.T) {
	c := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPatch {
			t.Errorf("method = %s, want PATCH", r.Method)
		}
		if r.URL.Path != "/v1/webhooks/endpoints/whe_1a2b3c4d5e" {
			t.Errorf("path = %q", r.URL.Path)
		}
		io.WriteString(w, webhookEndpointExample)
	})

	ep, _, err := c.Webhooks.UpdateEndpoint(context.Background(), "whe_1a2b3c4d5e", UpdateWebhookEndpointRequest{
		Name: "Production events",
	})
	if err != nil {
		t.Fatalf("UpdateEndpoint returned error: %v", err)
	}
	if ep.EndpointID != "whe_1a2b3c4d5e" {
		t.Errorf("endpoint = %+v", ep)
	}
}

func TestDeleteWebhookEndpoint(t *testing.T) {
	c := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodDelete {
			t.Errorf("method = %s, want DELETE", r.Method)
		}
		if r.URL.Path != "/v1/webhooks/endpoints/whe_1a2b3c4d5e" {
			t.Errorf("path = %q", r.URL.Path)
		}
		io.WriteString(w, `{
			"status": "deleted",
			"endpoint_id": "whe_1a2b3c4d5e"
		}`)
	})

	res, _, err := c.Webhooks.DeleteEndpoint(context.Background(), "whe_1a2b3c4d5e")
	if err != nil {
		t.Fatalf("DeleteEndpoint returned error: %v", err)
	}
	if res.Status != "deleted" || res.EndpointID != "whe_1a2b3c4d5e" {
		t.Errorf("res = %+v", res)
	}
}

func TestGetWebhookEndpointSecret(t *testing.T) {
	c := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Errorf("method = %s, want GET", r.Method)
		}
		if r.URL.Path != "/v1/webhooks/endpoints/whe_1a2b3c4d5e/secret" {
			t.Errorf("path = %q", r.URL.Path)
		}
		io.WriteString(w, `{
			"endpoint_id": "whe_1a2b3c4d5e",
			"name": "Production events",
			"url": "https://api.example.com/webhooks/bachs",
			"enabled": true,
			"event_types": ["collection.succeeded"],
			"created_at": "2026-04-27T12:00:00Z",
			"updated_at": "2026-04-27T12:00:00Z",
			"secret": "whsec_1a2b3c4d5e6f7g8h"
		}`)
	})

	res, _, err := c.Webhooks.GetEndpointSecret(context.Background(), "whe_1a2b3c4d5e")
	if err != nil {
		t.Fatalf("GetEndpointSecret returned error: %v", err)
	}
	if res.Secret != "whsec_1a2b3c4d5e6f7g8h" || res.EndpointID != "whe_1a2b3c4d5e" {
		t.Errorf("res = %+v", res)
	}
}

func TestRotateWebhookEndpointSecret(t *testing.T) {
	c := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("method = %s, want POST", r.Method)
		}
		if r.URL.Path != "/v1/webhooks/endpoints/whe_1a2b3c4d5e/rotate-secret" {
			t.Errorf("path = %q", r.URL.Path)
		}
		io.WriteString(w, webhookEndpointExample)
	})

	ep, _, err := c.Webhooks.RotateEndpointSecret(context.Background(), "whe_1a2b3c4d5e")
	if err != nil {
		t.Fatalf("RotateEndpointSecret returned error: %v", err)
	}
	if ep.EndpointID != "whe_1a2b3c4d5e" {
		t.Errorf("endpoint = %+v", ep)
	}
}

func TestGetWebhookEndpointMetrics(t *testing.T) {
	c := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Errorf("method = %s, want GET", r.Method)
		}
		if got := r.URL.RequestURI(); got != "/v1/webhooks/endpoints/whe_1a2b3c4d5e/metrics?date_from=2026-04-27&date_to=2026-04-28&period=day" {
			t.Errorf("path = %q", got)
		}
		io.WriteString(w, `{
			"total": "145",
			"period": "day",
			"data": [
				{"date": "2026-04-27", "success": 142, "failed": 3}
			]
		}`)
	})

	m, _, err := c.Webhooks.GetEndpointMetrics(context.Background(), "whe_1a2b3c4d5e", EndpointMetricsParams{
		Period:   "day",
		DateFrom: "2026-04-27",
		DateTo:   "2026-04-28",
	})
	if err != nil {
		t.Fatalf("GetEndpointMetrics returned error: %v", err)
	}
	if m.Total != "145" || m.Period != "day" {
		t.Errorf("metrics = %+v", m)
	}
	if len(m.Data) != 1 || m.Data[0].Success != 142 || m.Data[0].Failed != 3 {
		t.Errorf("Data = %+v", m.Data)
	}
}

func TestListWebhookEndpointEvents(t *testing.T) {
	c := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Errorf("method = %s, want GET", r.Method)
		}
		if r.URL.Path != "/v1/webhooks/endpoints/whe_1a2b3c4d5e/events" {
			t.Errorf("path = %q", r.URL.Path)
		}
		io.WriteString(w, `{
			"items": [
				{
					"event_id": "evt_1a2b3c4d5e",
					"event_type": "collection.succeeded",
					"entity_id": "pay_1a2b3c4d5e",
					"attempts": 1,
					"success": 1,
					"failed": 0,
					"last_attempt_status": "delivered",
					"last_attempt_http_status": 200,
					"last_attempt_at": "2026-04-27T12:00:01Z"
				}
			],
			"total": 145,
			"limit": 50,
			"offset": 0
		}`)
	})

	page, _, err := c.Webhooks.ListEndpointEvents(context.Background(), "whe_1a2b3c4d5e", ListParams{Limit: 50})
	if err != nil {
		t.Fatalf("ListEndpointEvents returned error: %v", err)
	}
	if len(page.Items) != 1 {
		t.Fatalf("len(Items) = %d, want 1", len(page.Items))
	}
	ev := page.Items[0]
	if ev.EventID != "evt_1a2b3c4d5e" || ev.Attempts != 1 {
		t.Errorf("Items[0] = %+v", ev)
	}
	if ev.LastAttemptStatus == nil || *ev.LastAttemptStatus != "delivered" {
		t.Errorf("LastAttemptStatus = %v", ev.LastAttemptStatus)
	}
	if ev.LastAttemptHTTPStatus == nil || *ev.LastAttemptHTTPStatus != 200 {
		t.Errorf("LastAttemptHTTPStatus = %v", ev.LastAttemptHTTPStatus)
	}
	if page.Pagination.Total != 145 {
		t.Errorf("Pagination.Total = %d, want 145", page.Pagination.Total)
	}
}

func TestGetWebhookEndpointEvent(t *testing.T) {
	c := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Errorf("method = %s, want GET", r.Method)
		}
		if r.URL.Path != "/v1/webhooks/endpoints/whe_1a2b3c4d5e/events/evt_1a2b3c4d5e" {
			t.Errorf("path = %q", r.URL.Path)
		}
		io.WriteString(w, `{
			"event_id": "evt_1a2b3c4d5e",
			"event_type": "collection.succeeded",
			"entity_type": "charge",
			"entity_id": "pay_1a2b3c4d5e",
			"created_at": "2026-04-27T12:00:00Z",
			"payload": {"id": "pay_1a2b3c4d5e", "amount": "29.00"},
			"attempts": [
				{
					"attempt_id": "wha_1a2b3c4d5e",
					"attempt_no": 1,
					"status": "delivered",
					"callback_url": "https://api.example.com/webhooks/bachs",
					"http_status": 200,
					"response_snippet": "ok",
					"last_error": null,
					"created_at": "2026-04-27T12:00:01Z",
					"updated_at": "2026-04-27T12:00:01Z"
				}
			]
		}`)
	})

	ev, _, err := c.Webhooks.GetEndpointEvent(context.Background(), "whe_1a2b3c4d5e", "evt_1a2b3c4d5e")
	if err != nil {
		t.Fatalf("GetEndpointEvent returned error: %v", err)
	}
	if ev.EventID != "evt_1a2b3c4d5e" || ev.EventType != "collection.succeeded" {
		t.Errorf("event = %+v", ev)
	}
	if len(ev.Attempts) != 1 {
		t.Fatalf("len(Attempts) = %d, want 1", len(ev.Attempts))
	}
	if ev.Attempts[0].AttemptNo != 1 || ev.Attempts[0].Status != "delivered" {
		t.Errorf("Attempts[0] = %+v", ev.Attempts[0])
	}
	if v, ok := ev.Payload["amount"]; !ok || v != "29.00" {
		t.Errorf("Payload = %v", ev.Payload)
	}
}

func TestResendWebhookEndpointEvent(t *testing.T) {
	c := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("method = %s, want POST", r.Method)
		}
		if r.URL.Path != "/v1/webhooks/endpoints/whe_1a2b3c4d5e/events/evt_1a2b3c4d5e/resend" {
			t.Errorf("path = %q", r.URL.Path)
		}
		io.WriteString(w, `{
			"status": "queued",
			"attempt_id": "wha_9z8y7x6w5v"
		}`)
	})

	res, _, err := c.Webhooks.ResendEndpointEvent(context.Background(), "whe_1a2b3c4d5e", "evt_1a2b3c4d5e")
	if err != nil {
		t.Fatalf("ResendEndpointEvent returned error: %v", err)
	}
	if res.Status != "queued" || res.AttemptID != "wha_9z8y7x6w5v" {
		t.Errorf("res = %+v", res)
	}
}

func TestListWebhookEvents(t *testing.T) {
	c := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Errorf("method = %s, want GET", r.Method)
		}
		if r.URL.Path != "/v1/webhooks/events" {
			t.Errorf("path = %q", r.URL.Path)
		}
		io.WriteString(w, `{
			"items": [
				{
					"event_id": "evt_1a2b3c4d5e",
					"event_type": "collection.succeeded",
					"entity_type": "charge",
					"entity_id": "pay_1a2b3c4d5e",
					"created_at": "2026-04-27T12:00:00Z",
					"attempts": 1,
					"success": 1,
					"failed": 0,
					"last_attempt_at": "2026-04-27T12:00:01Z"
				}
			],
			"total": 320,
			"limit": 50,
			"offset": 0
		}`)
	})

	page, _, err := c.Webhooks.ListEvents(context.Background(), ListParams{Limit: 50})
	if err != nil {
		t.Fatalf("ListEvents returned error: %v", err)
	}
	if len(page.Items) != 1 {
		t.Fatalf("len(Items) = %d, want 1", len(page.Items))
	}
	ev := page.Items[0]
	if ev.EventID != "evt_1a2b3c4d5e" || ev.EventType != "collection.succeeded" {
		t.Errorf("Items[0] = %+v", ev)
	}
	if ev.Success != 1 || ev.Failed != 0 {
		t.Errorf("Items[0] = %+v", ev)
	}
	if page.Pagination.Total != 320 {
		t.Errorf("Pagination.Total = %d, want 320", page.Pagination.Total)
	}
}

func TestGetWebhookEvent(t *testing.T) {
	c := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Errorf("method = %s, want GET", r.Method)
		}
		if r.URL.Path != "/v1/webhooks/events/evt_1a2b3c4d5e" {
			t.Errorf("path = %q", r.URL.Path)
		}
		io.WriteString(w, `{
			"event_id": "evt_1a2b3c4d5e",
			"event_type": "collection.succeeded",
			"entity_type": "charge",
			"entity_id": "pay_1a2b3c4d5e",
			"created_at": "2026-04-27T12:00:00Z",
			"payload": {"id": "pay_1a2b3c4d5e"},
			"attempts": [
				{
					"attempt_id": "wha_1a2b3c4d5e",
					"attempt_no": 1,
					"status": "delivered",
					"callback_url": "https://api.example.com/webhooks/bachs",
					"http_status": 200,
					"response_snippet": "ok",
					"last_error": null,
					"created_at": "2026-04-27T12:00:01Z",
					"updated_at": "2026-04-27T12:00:01Z"
				}
			]
		}`)
	})

	ev, _, err := c.Webhooks.GetEvent(context.Background(), "evt_1a2b3c4d5e")
	if err != nil {
		t.Fatalf("GetEvent returned error: %v", err)
	}
	if ev.EventID != "evt_1a2b3c4d5e" || ev.EntityID == nil || *ev.EntityID != "pay_1a2b3c4d5e" {
		t.Errorf("event = %+v", ev)
	}
	if len(ev.Attempts) != 1 || ev.Attempts[0].ResponseSnippet == nil || *ev.Attempts[0].ResponseSnippet != "ok" {
		t.Errorf("Attempts = %+v", ev.Attempts)
	}
}

func TestReplayWebhookEvent(t *testing.T) {
	c := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("method = %s, want POST", r.Method)
		}
		if r.URL.Path != "/v1/webhooks/replay" {
			t.Errorf("path = %q", r.URL.Path)
		}
		io.WriteString(w, `{
			"success": true,
			"event_id": "evt_3ab4e0d5d27445cf8a52ab3d8cb8f0b1",
			"attempt_id": "wha_6f1e40f6bdf84c1980e1e1f6407f3f8a",
			"attempt_no": 3,
			"event_type": "collection.failed"
		}`)
	})

	res, _, err := c.Webhooks.Replay(context.Background(), ReplayWebhookEventRequest{
		EventID: "evt_3ab4e0d5d27445cf8a52ab3d8cb8f0b1",
	})
	if err != nil {
		t.Fatalf("Replay returned error: %v", err)
	}
	if !res.Success || res.EventID != "evt_3ab4e0d5d27445cf8a52ab3d8cb8f0b1" {
		t.Errorf("res = %+v", res)
	}
	if res.AttemptNo != 3 || res.EventType != "collection.failed" {
		t.Errorf("res = %+v", res)
	}
}
