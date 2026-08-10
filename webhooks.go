package bachs

import (
	"context"
	"net/http"
	"net/url"
	"time"
)

// WebhookService provides the webhook management API: registering the URLs
// Bachs delivers events to, subscribing them to event types, monitoring
// delivery metrics, inspecting and resending deliveries, and replaying
// events. Verifying the signature on an incoming delivery is the separate
// webhook package's ConstructEvent — this service manages endpoints, it does
// not receive events. Requires webhooks:read / webhooks:write scopes.
type WebhookService struct {
	service
}

// Webhook event type constants, matching the API's event_types enum. Use
// these when creating or updating an endpoint so typos are caught at compile
// time.
const (
	EventTypeCollectionSucceeded         = "collection.succeeded"
	EventTypeCollectionFailed            = "collection.failed"
	EventTypeCollectionUnderpaid         = "collection.underpaid"
	EventTypeCheckoutCompleted           = "checkout.completed"
	EventTypeCheckoutExpired             = "checkout.expired"
	EventTypePayoutCreated               = "payout.created"
	EventTypePayoutPaid                  = "payout.paid"
	EventTypePayoutFailed                = "payout.failed"
	EventTypeRefundCreated               = "refund.created"
	EventTypeRefundPaid                  = "refund.paid"
	EventTypeRefundFailed                = "refund.failed"
	EventTypeConversionCompleted         = "conversion.completed"
	EventTypeConversionFailed            = "conversion.failed"
	EventTypeCustomerCreated             = "customer.created"
	EventTypeCustomerUpdated             = "customer.updated"
	EventTypeDisputeCreated              = "dispute.created"
	EventTypeDisputeUpdated              = "dispute.updated"
	EventTypeCustomerSubscriptionCreated = "customer.subscription.created"
	EventTypeCustomerSubscriptionUpdated = "customer.subscription.updated"
	EventTypeCustomerSubscriptionDeleted = "customer.subscription.deleted"
	EventTypeInvoiceCreated              = "invoice.created"
	EventTypeInvoicePaid                 = "invoice.paid"
	EventTypeInvoicePaymentFailed        = "invoice.payment_failed"
)

// WebhookEndpoint is a URL Bachs delivers events to, together with the set of
// events it is subscribed to.
type WebhookEndpoint struct {
	// EndpointID uniquely identifies the endpoint.
	EndpointID string `json:"endpoint_id"`

	// Name is a label for the endpoint.
	Name string `json:"name"`

	// URL events are delivered to.
	URL string `json:"url"`

	// Enabled is true while the endpoint receives events.
	Enabled bool `json:"enabled"`

	// EventTypes the endpoint is subscribed to.
	EventTypes []string `json:"event_types"`

	// CreatedAt is when the endpoint was created, in UTC.
	CreatedAt time.Time `json:"created_at"`

	// UpdatedAt is when the endpoint was last updated, in UTC.
	UpdatedAt time.Time `json:"updated_at"`
}

// CreateWebhookEndpointRequest is the payload for Webhooks.CreateEndpoint.
type CreateWebhookEndpointRequest struct {
	// Name is a label for the endpoint. Required.
	Name string `json:"name"`

	// URL to deliver events to. Must be HTTPS. Required.
	URL string `json:"url"`

	// EventTypes to subscribe to, from the EventType* constants. At least one
	// is required.
	EventTypes []string `json:"event_types"`
}

// CreateWebhookEndpointResponse is the result of Webhooks.CreateEndpoint. The
// signing secret is returned only once, here — store it to verify the
// X-Bachs-Signature header on deliveries.
type CreateWebhookEndpointResponse struct {
	WebhookEndpoint

	// SigningSecret is the secret used to sign deliveries to this endpoint.
	// It is shown only on creation; use Webhooks.GetEndpointSecret to read it
	// later.
	SigningSecret string `json:"signing_secret"`
}

// UpdateWebhookEndpointRequest is the payload for Webhooks.UpdateEndpoint.
// Only the fields you send are changed.
type UpdateWebhookEndpointRequest struct {
	// Name replaces the endpoint's label.
	Name string `json:"name,omitempty"`

	// URL replaces the delivery URL.
	URL string `json:"url,omitempty"`

	// EventTypes replaces the subscription list.
	EventTypes []string `json:"event_types,omitempty"`
}

// DeleteWebhookEndpointResponse is the result of Webhooks.DeleteEndpoint.
type DeleteWebhookEndpointResponse struct {
	// Status of the deletion.
	Status string `json:"status"`

	// EndpointID that was deleted.
	EndpointID string `json:"endpoint_id"`
}

// WebhookEndpointSecretResponse is the result of Webhooks.GetEndpointSecret:
// the current signing secret plus the endpoint it belongs to.
type WebhookEndpointSecretResponse struct {
	WebhookEndpoint

	// Secret is the current signing secret. Use it to verify the
	// X-Bachs-Signature header on deliveries.
	Secret string `json:"secret"`
}

// WebhookMetricsResponse is the result of Webhooks.GetEndpointMetrics:
// delivery success and failure counts over a time range.
type WebhookMetricsResponse struct {
	// Total deliveries over the range.
	Total string `json:"total"`

	// Period grouping, for example "day".
	Period string `json:"period"`

	// Data has one point per period.
	Data []WebhookMetricsDataPoint `json:"data"`
}

// WebhookMetricsDataPoint is the delivery counts for one period.
type WebhookMetricsDataPoint struct {
	// Date of the period, for example "2026-04-27".
	Date string `json:"date"`

	// Success deliveries in this period.
	Success int `json:"success"`

	// Failed deliveries in this period.
	Failed int `json:"failed"`
}

// EndpointMetricsParams are the optional time-range filters for
// Webhooks.GetEndpointMetrics.
type EndpointMetricsParams struct {
	// Period grouping, for example "day" or "week".
	Period string

	// DateFrom is the inclusive start date (YYYY-MM-DD).
	DateFrom string

	// DateTo is the inclusive end date (YYYY-MM-DD).
	DateTo string
}

// WebhookEndpointEventListItem is a summary of an event's delivery to one
// endpoint.
type WebhookEndpointEventListItem struct {
	// EventID of the delivered event.
	EventID string `json:"event_id"`

	// EventType of the delivered event.
	EventType string `json:"event_type"`

	// EntityID is the resource the event is about.
	EntityID *string `json:"entity_id"`

	// Attempts to deliver to this endpoint.
	Attempts int `json:"attempts"`

	// Success deliveries.
	Success int `json:"success"`

	// Failed deliveries.
	Failed int `json:"failed"`

	// LastAttemptStatus of the most recent attempt.
	LastAttemptStatus *string `json:"last_attempt_status"`

	// LastAttemptHTTPStatus your endpoint returned on the last attempt.
	LastAttemptHTTPStatus *int `json:"last_attempt_http_status"`

	// LastAttemptAt is when delivery was last attempted.
	LastAttemptAt *time.Time `json:"last_attempt_at"`
}

// WebhookEventListItem is a summary of a webhook event across every endpoint.
type WebhookEventListItem struct {
	// EventID uniquely identifies the event.
	EventID string `json:"event_id"`

	// EventType of the event.
	EventType string `json:"event_type"`

	// EntityType of the resource the event is about.
	EntityType *string `json:"entity_type"`

	// EntityID of the resource the event is about.
	EntityID *string `json:"entity_id"`

	// CreatedAt is when the event was generated.
	CreatedAt time.Time `json:"created_at"`

	// Attempts to deliver this event.
	Attempts int `json:"attempts"`

	// Success deliveries.
	Success int `json:"success"`

	// Failed deliveries.
	Failed int `json:"failed"`

	// LastAttemptAt is when delivery was last attempted.
	LastAttemptAt *time.Time `json:"last_attempt_at"`
}

// WebhookEventDetail is one event's full payload and delivery attempts.
type WebhookEventDetail struct {
	// EventID uniquely identifies the event.
	EventID string `json:"event_id"`

	// EventType of the event.
	EventType string `json:"event_type"`

	// EntityType of the resource the event is about.
	EntityType *string `json:"entity_type"`

	// EntityID of the resource the event is about.
	EntityID *string `json:"entity_id"`

	// CreatedAt is when the event was generated.
	CreatedAt time.Time `json:"created_at"`

	// Payload is the exact event data that was delivered.
	Payload map[string]any `json:"payload"`

	// Attempts is the delivery history for this event.
	Attempts []WebhookEventAttempt `json:"attempts"`
}

// WebhookEventAttempt is one delivery attempt of a webhook event.
type WebhookEventAttempt struct {
	// AttemptID uniquely identifies the attempt.
	AttemptID string `json:"attempt_id"`

	// AttemptNo counts attempts, starting at 1.
	AttemptNo int `json:"attempt_no"`

	// Status of the attempt (for example "delivered").
	Status string `json:"status"`

	// CallbackURL the event was delivered to.
	CallbackURL *string `json:"callback_url"`

	// HTTPStatus your endpoint returned on this attempt.
	HTTPStatus *int `json:"http_status"`

	// ResponseSnippet of your endpoint's reply.
	ResponseSnippet *string `json:"response_snippet"`

	// LastError explains a failed attempt.
	LastError *string `json:"last_error"`

	// CreatedAt is when the attempt was made.
	CreatedAt time.Time `json:"created_at"`

	// UpdatedAt is when the attempt was last updated.
	UpdatedAt time.Time `json:"updated_at"`
}

// ResendWebhookEventResponse is the result of Webhooks.ResendEndpointEvent.
type ResendWebhookEventResponse struct {
	// Status of the resend request.
	Status string `json:"status"`

	// AttemptID of the new delivery attempt.
	AttemptID string `json:"attempt_id"`
}

// ReplayWebhookEventRequest is the payload for Webhooks.Replay. Provide at
// least one identifier: an EventID replays that exact event; otherwise the
// replay target is resolved from ChargeID or Reference.
type ReplayWebhookEventRequest struct {
	// EventID of the event to replay directly.
	EventID string `json:"event_id,omitempty"`

	// ChargeID used to resolve the latest webhook event for that charge.
	ChargeID string `json:"charge_id,omitempty"`

	// Reference used to resolve the charge and replay its latest event.
	Reference string `json:"reference,omitempty"`
}

// ReplayWebhookEventResponse is the result of Webhooks.Replay.
type ReplayWebhookEventResponse struct {
	// Success is true when the replay was accepted.
	Success bool `json:"success"`

	// EventID that was replayed.
	EventID string `json:"event_id"`

	// AttemptID of the new delivery attempt.
	AttemptID string `json:"attempt_id"`

	// AttemptNo of the new attempt.
	AttemptNo int `json:"attempt_no"`

	// EventType of the replayed event.
	EventType string `json:"event_type"`
}

// CreateEndpoint registers a URL to receive webhook events and subscribes it
// to the given event types. The signing secret is returned once, in the
// response — store it to verify deliveries.
func (s *WebhookService) CreateEndpoint(ctx context.Context, req CreateWebhookEndpointRequest, opts ...RequestOption) (*CreateWebhookEndpointResponse, *ResponseMeta, error) {
	var out CreateWebhookEndpointResponse
	meta, err := s.request(ctx, http.MethodPost, "/webhooks/endpoints", req, &out, opts...)
	if err != nil {
		return nil, meta, err
	}
	return &out, meta, nil
}

// ListEndpoints lists every webhook endpoint for your organization. The API
// returns a flat array without pagination, so this returns the slice directly
// rather than a Page.
func (s *WebhookService) ListEndpoints(ctx context.Context) ([]WebhookEndpoint, *ResponseMeta, error) {
	var out []WebhookEndpoint
	meta, err := s.request(ctx, http.MethodGet, "/webhooks/endpoints", nil, &out)
	if err != nil {
		return nil, meta, err
	}
	return out, meta, nil
}

// GetEndpoint returns a single webhook endpoint by ID.
func (s *WebhookService) GetEndpoint(ctx context.Context, endpointID string) (*WebhookEndpoint, *ResponseMeta, error) {
	var out WebhookEndpoint
	meta, err := s.request(ctx, http.MethodGet, "/webhooks/endpoints/"+url.PathEscape(endpointID), nil, &out)
	if err != nil {
		return nil, meta, err
	}
	return &out, meta, nil
}

// UpdateEndpoint changes an endpoint's name, URL, or subscribed events. Only
// the fields you send are changed.
func (s *WebhookService) UpdateEndpoint(ctx context.Context, endpointID string, req UpdateWebhookEndpointRequest) (*WebhookEndpoint, *ResponseMeta, error) {
	var out WebhookEndpoint
	meta, err := s.request(ctx, http.MethodPatch, "/webhooks/endpoints/"+url.PathEscape(endpointID), req, &out)
	if err != nil {
		return nil, meta, err
	}
	return &out, meta, nil
}

// DeleteEndpoint removes a webhook endpoint. It stops receiving events
// immediately.
func (s *WebhookService) DeleteEndpoint(ctx context.Context, endpointID string) (*DeleteWebhookEndpointResponse, *ResponseMeta, error) {
	var out DeleteWebhookEndpointResponse
	meta, err := s.request(ctx, http.MethodDelete, "/webhooks/endpoints/"+url.PathEscape(endpointID), nil, &out)
	if err != nil {
		return nil, meta, err
	}
	return &out, meta, nil
}

// GetEndpointSecret returns the current signing secret for an endpoint. Use
// it to verify the X-Bachs-Signature header on deliveries.
func (s *WebhookService) GetEndpointSecret(ctx context.Context, endpointID string) (*WebhookEndpointSecretResponse, *ResponseMeta, error) {
	var out WebhookEndpointSecretResponse
	meta, err := s.request(ctx, http.MethodGet, "/webhooks/endpoints/"+url.PathEscape(endpointID)+"/secret", nil, &out)
	if err != nil {
		return nil, meta, err
	}
	return &out, meta, nil
}

// RotateEndpointSecret generates a new signing secret for an endpoint. The
// old secret stops working immediately, so update your verification before
// rotating.
func (s *WebhookService) RotateEndpointSecret(ctx context.Context, endpointID string, opts ...RequestOption) (*WebhookEndpoint, *ResponseMeta, error) {
	var out WebhookEndpoint
	meta, err := s.request(ctx, http.MethodPost, "/webhooks/endpoints/"+url.PathEscape(endpointID)+"/rotate-secret", struct{}{}, &out, opts...)
	if err != nil {
		return nil, meta, err
	}
	return &out, meta, nil
}

// GetEndpointMetrics returns delivery success and failure counts for an
// endpoint over a time range.
func (s *WebhookService) GetEndpointMetrics(ctx context.Context, endpointID string, params EndpointMetricsParams) (*WebhookMetricsResponse, *ResponseMeta, error) {
	q := url.Values{}
	if params.Period != "" {
		q.Set("period", params.Period)
	}
	if params.DateFrom != "" {
		q.Set("date_from", params.DateFrom)
	}
	if params.DateTo != "" {
		q.Set("date_to", params.DateTo)
	}

	path := "/webhooks/endpoints/" + url.PathEscape(endpointID) + "/metrics"
	if len(q) > 0 {
		path += "?" + q.Encode()
	}

	var out WebhookMetricsResponse
	meta, err := s.request(ctx, http.MethodGet, path, nil, &out)
	if err != nil {
		return nil, meta, err
	}
	return &out, meta, nil
}

// ListEndpointEvents lists the events delivered (or attempted) to a specific
// endpoint.
func (s *WebhookService) ListEndpointEvents(ctx context.Context, endpointID string, params ListParams) (*Page[WebhookEndpointEventListItem], *ResponseMeta, error) {
	var env pageEnvelope[WebhookEndpointEventListItem]
	meta, err := s.request(ctx, http.MethodGet, queryPath("/webhooks/endpoints/"+url.PathEscape(endpointID)+"/events", params), nil, &env)
	if err != nil {
		return nil, meta, err
	}
	return env.page(), meta, nil
}

// GetEndpointEvent returns one event's full payload and delivery attempts
// for a specific endpoint.
func (s *WebhookService) GetEndpointEvent(ctx context.Context, endpointID, eventID string) (*WebhookEventDetail, *ResponseMeta, error) {
	var out WebhookEventDetail
	meta, err := s.request(ctx, http.MethodGet, "/webhooks/endpoints/"+url.PathEscape(endpointID)+"/events/"+url.PathEscape(eventID), nil, &out)
	if err != nil {
		return nil, meta, err
	}
	return &out, meta, nil
}

// ResendEndpointEvent re-delivers a past event to a specific endpoint. Use it
// when your endpoint missed or rejected an earlier delivery.
func (s *WebhookService) ResendEndpointEvent(ctx context.Context, endpointID, eventID string, opts ...RequestOption) (*ResendWebhookEventResponse, *ResponseMeta, error) {
	var out ResendWebhookEventResponse
	meta, err := s.request(ctx, http.MethodPost, "/webhooks/endpoints/"+url.PathEscape(endpointID)+"/events/"+url.PathEscape(eventID)+"/resend", struct{}{}, &out, opts...)
	if err != nil {
		return nil, meta, err
	}
	return &out, meta, nil
}

// ListEvents lists webhook events for your organization, across every
// endpoint.
func (s *WebhookService) ListEvents(ctx context.Context, params ListParams) (*Page[WebhookEventListItem], *ResponseMeta, error) {
	var env pageEnvelope[WebhookEventListItem]
	meta, err := s.request(ctx, http.MethodGet, queryPath("/webhooks/events", params), nil, &env)
	if err != nil {
		return nil, meta, err
	}
	return env.page(), meta, nil
}

// GetEvent returns a single event's full payload and delivery attempts.
func (s *WebhookService) GetEvent(ctx context.Context, eventID string) (*WebhookEventDetail, *ResponseMeta, error) {
	var out WebhookEventDetail
	meta, err := s.request(ctx, http.MethodGet, "/webhooks/events/"+url.PathEscape(eventID), nil, &out)
	if err != nil {
		return nil, meta, err
	}
	return &out, meta, nil
}

// Replay creates a new outbound delivery attempt for a previously generated
// webhook event. Use it when your endpoint missed or rejected an earlier
// delivery and you need Bachs to send that event again.
func (s *WebhookService) Replay(ctx context.Context, req ReplayWebhookEventRequest, opts ...RequestOption) (*ReplayWebhookEventResponse, *ResponseMeta, error) {
	var out ReplayWebhookEventResponse
	meta, err := s.request(ctx, http.MethodPost, "/webhooks/replay", req, &out, opts...)
	if err != nil {
		return nil, meta, err
	}
	return &out, meta, nil
}
