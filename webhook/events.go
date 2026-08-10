// Package webhook verifies and decodes Bachs webhook deliveries. It has no
// dependency on the main bachs package or on any HTTP client, so it can be
// used inside an http.Handler without pulling in the SDK's client machinery.
//
// Bachs signs every delivery with an HMAC-SHA256 of "{timestamp}.{raw_body}".
// Use ConstructEvent to verify the signature before trusting the payload, and
// pass the untouched raw request body — read it before any JSON parsing, or
// whitespace and byte-order changes will break verification.
package webhook

import (
	"encoding/json"
	"time"
)

// Event is the envelope of a webhook delivery. The Data field carries the
// resource payload as raw JSON; unmarshal it into a resource type (or a
// map[string]any) after checking the event Type.
type Event struct {
	// ID uniquely identifies the delivery. Bachs guarantees at-least-once
	// delivery, so use this field to deduplicate events that arrive more than
	// once.
	ID string `json:"id"`

	// Type is the event type, for example "collection.succeeded" or
	// "customer.subscription.created".
	Type string `json:"type"`

	// CreatedAt is when the event occurred, as an ISO 8601 UTC timestamp.
	CreatedAt time.Time `json:"created_at"`

	// OrganizationID is the organization the event belongs to.
	OrganizationID string `json:"organization_id"`

	// Account identifies the connected account an event happened on. Present
	// only on Connect events (endpoints with event_source "connect" or "all");
	// empty otherwise.
	Account string `json:"account,omitempty"`

	// Data is the event payload. Its shape depends on Type.
	Data json.RawMessage `json:"data"`
}
