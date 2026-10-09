package bachs

import (
	"context"
	"net/http"
	"net/url"
	"time"
)

// ConnectedAccountService provides the Connect platform API: creating and
// reading connected accounts, updating them, requesting capabilities, and
// issuing hosted account links. Source:
// https://docs.bachs.io/connect/overview
//
// Onboarding reads come from Get (requirements.entries); values are
// submitted with UpdateAccount; reference data comes from the Reference
// service; ID documents are uploaded with Media.Upload and attached with
// Persons.AttachDocument. See
// https://docs.bachs.io/connect/guides/api-onboarding.
type ConnectedAccountService struct {
	service
}

// ConnectedAccount is a financial identity you created under your platform: a
// seller or contractor with its own balance, capabilities, and onboarding
// state. Populated fully by Create, Get, and RequestCapabilities; List items
// leave Capabilities and Requirements null. Source:
// https://docs.bachs.io/api-reference/connected-accounts/get-a-connected-account
type ConnectedAccount struct {
	// ID is the unique identifier for the organization.
	ID string `json:"id"`

	// Name is the organization's display name.
	Name *string `json:"name"`

	// OwnerUserID is the user that owns the organization. For a connected
	// account this is a service user Bachs created; you never authenticate as
	// it.
	OwnerUserID string `json:"owner_user_id"`

	// ParentOrganizationID is the platform this organization is connected to,
	// or null when it is a platform in its own right.
	ParentOrganizationID *string `json:"parent_organization_id"`

	// Country is the two-letter ISO 3166-1 code that decides which Tasks the
	// account is given.
	Country *string `json:"country"`

	// EntityType is "company" or "individual". Null until set.
	EntityType *string `json:"entity_type"`

	// Configuration names the personas applied to the account (for example
	// "recipient" and "merchant"), each with the capabilities nested under
	// it. Null when the account carries no Connect configuration.
	Configuration map[string]any `json:"configuration"`

	// Responsibilities records fixed arrangements for the account, such as
	// who absorbs processing fees on its own charges
	// (responsibilities.fees.collector). Null when unset.
	Responsibilities map[string]any `json:"responsibilities"`

	// FeeHandling is who absorbs processing fees on the organization's own
	// charges: "account_pays_fee" or "customer_pays_fee".
	FeeHandling string `json:"fee_handling"`

	// EnabledPaymentMethods for this organization's checkouts, keyed by
	// method.
	EnabledPaymentMethods map[string]any `json:"enabled_payment_methods"`

	// AdaptivePricing is true when customers are shown prices in their local
	// currency where one is available.
	AdaptivePricing bool `json:"adaptive_pricing"`

	// BalanceCurrencies this organization is configured to hold a balance in.
	BalanceCurrencies []string `json:"balance_currencies"`

	// Website of the organization.
	Website *string `json:"website"`

	// PhoneNumber including country code.
	PhoneNumber *string `json:"phone_number"`

	// CompanyName when the organization is a company.
	CompanyName *string `json:"company_name"`

	// EnabledCapabilities are the names of the capabilities currently active
	// on this organization.
	EnabledCapabilities []string `json:"enabled_capabilities"`

	// Capabilities maps each capability name to its status. Populated on
	// single-account reads only; null on list items.
	Capabilities map[string]CapabilityStatus `json:"capabilities"`

	// Requirements are the outstanding Tasks for this account. Populated on
	// single-account reads only; null on list items.
	Requirements *AccountRequirements `json:"requirements"`

	// FieldsNeedingResubmission counts fields with a standing rejection.
	// Returned on list items only.
	FieldsNeedingResubmission *int `json:"fields_needing_resubmission"`

	// SandboxOrgID is the matching organization in the sandbox environment.
	SandboxOrgID *string `json:"sandbox_org_id"`

	// LiveOrgID is the matching organization in production.
	LiveOrgID *string `json:"live_org_id"`

	// IsActive is false when the organization is deactivated and cannot
	// authenticate or move funds.
	IsActive bool `json:"is_active"`

	// CreatedAt is when the organization was created, ISO 8601 in UTC.
	CreatedAt time.Time `json:"created_at"`

	// UpdatedAt is when the organization was last updated, ISO 8601 in UTC.
	UpdatedAt time.Time `json:"updated_at"`

	// Controller is the fee arrangement for a connected account; null on
	// organizations that are not connected accounts.
	Controller *ControllerResponse `json:"controller"`
}

// CapabilityStatus is the state of one capability on a connected account.
type CapabilityStatus struct {
	// Status is "active", "pending", "restricted", "unrequested", or
	// "unsupported". Only "active" authorizes anything.
	Status string `json:"status"`

	// Requested reports whether the account ever requested this capability.
	Requested bool `json:"requested"`

	// StatusDetails explains why the capability is not active; null when it is
	// active.
	StatusDetails []CapabilityStatusDetail `json:"status_details"`
}

// CapabilityStatusDetail explains why a capability is not active.
type CapabilityStatusDetail struct {
	// Code is the machine-readable reason; branch on this rather than on
	// Message.
	Code string `json:"code"`

	// Resolution is what has to happen for the capability to become active.
	Resolution *string `json:"resolution"`

	// Message is a human-readable explanation, safe to show the account
	// holder.
	Message *string `json:"message"`
}

// AccountRequirements summarizes the Tasks a connected account still owes.
type AccountRequirements struct {
	// SetupStatus is "incomplete", "awaiting_review", or "complete".
	SetupStatus string `json:"setup_status"`

	// CurrentlyDue field keys required now.
	CurrentlyDue []string `json:"currently_due"`

	// EventuallyDue field keys required later, once a threshold is reached.
	EventuallyDue []string `json:"eventually_due"`

	// PastDue field keys that were required by a date now passed.
	PastDue []string `json:"past_due"`

	// PendingVerification field keys provided and being checked.
	PendingVerification []string `json:"pending_verification"`

	// Errors are fields that were provided and then rejected.
	Errors []RequirementError `json:"errors"`

	// CurrentDeadline is the nearest outstanding deadline across entries.
	// Null when no dated requirement is outstanding.
	CurrentDeadline *time.Time `json:"current_deadline"`

	// Entries is the field-level view: every outstanding field, its state,
	// and the capabilities it holds up. Render these as the onboarding
	// form.
	Entries []RequirementEntry `json:"entries"`
}

// RequirementEntry is one field in an account's requirements: its state and
// what it unlocks.
type RequirementEntry struct {
	// Field is the canonical key (for example "persons.name" or
	// "payout_destination").
	Field string `json:"field"`

	// Status is the field's state (for example "currently_due").
	Status string `json:"status"`

	// RestrictsCapabilities names the capabilities held up by this field.
	RestrictsCapabilities []string `json:"restricts_capabilities"`

	// Resolution tells how the field is satisfied ("api" or a hosted flow).
	Resolution *string `json:"resolution"`

	// Deadline is when the field is due. Null when none is set.
	Deadline *time.Time `json:"deadline"`

	// Errors on this entry, written for display to the account holder.
	Errors []RequirementError `json:"errors"`
}

// RequirementError is a field that was submitted and rejected.
type RequirementError struct {
	// Field key that was rejected.
	Field string `json:"field"`

	// Code is the machine-readable rejection reason.
	Code *string `json:"code"`

	// Reason is the human-readable rejection reason, safe to show the account
	// holder.
	Reason *string `json:"reason"`
}

// ControllerResponse is the fee arrangement of a connected account.
type ControllerResponse struct {
	// Fees describes who absorbs processing fees on the account's charges.
	Fees ControllerFeesResponse `json:"fees"`
}

// ControllerFeesResponse is the fees portion of a connected account's
// controller.
type ControllerFeesResponse struct {
	// Payer is "account": the connected account absorbs processing fees. Set
	// at creation and immutable.
	Payer string `json:"payer"`
}

// Account personas applied through Configuration. A capability is only ever
// named inside the persona object it belongs to. Source:
// https://docs.bachs.io/connect/guides/create-an-account
const (
	// AccountPersonaMerchant applies the merchant persona: the account can
	// accept payments in its own name.
	AccountPersonaMerchant = "merchant"

	// AccountPersonaRecipient applies the recipient persona: the account can
	// be paid out and move transfers. Most marketplaces want this shape.
	AccountPersonaRecipient = "recipient"
)

// Account entity types.
const (
	// AccountEntityCompany is a registered company.
	AccountEntityCompany = "company"

	// AccountEntityIndividual is a person.
	AccountEntityIndividual = "individual"
)

// Processing-fee collectors for an account's own charges. Fixed at creation.
const (
	// FeeCollectorBachs takes the processing fee out of the account's own
	// charges. The default.
	FeeCollectorBachs = "bachs"

	// FeeCollectorPlatform has the platform absorb the processing fee
	// instead.
	FeeCollectorPlatform = "platform"
)

// PersonaConfig applies one persona to an account and requests capabilities
// under it. Name every persona the account needs as a key in Configuration,
// with each capability nested under its persona's own Capabilities: a
// capability nested under the wrong persona fails with
// capability_configuration_mismatch, and it is never silently corrected.
type PersonaConfig struct {
	// Capabilities to request under this persona, keyed by capability name.
	// Omitting Capabilities on creation requests every capability the
	// persona allows; on update it only applies the persona and requests
	// nothing.
	Capabilities map[string]CapabilityRequest `json:"capabilities,omitempty"`
}

// FeeCollectorConfig sets who absorbs processing fees on an account's own
// charges.
type FeeCollectorConfig struct {
	// Collector is FeeCollectorBachs or FeeCollectorPlatform.
	Collector string `json:"collector"`
}

// ResponsibilitiesConfig sets the fixed arrangements for an account. They
// are decided at creation; there is no endpoint to change them afterward.
type ResponsibilitiesConfig struct {
	// Fees describes who absorbs processing fees on the account's charges.
	Fees FeeCollectorConfig `json:"fees"`
}

// CreateConnectedAccountRequest is the payload for ConnectedAccounts.Create.
// ContactEmail is the only required field; send Country when the account is
// not in yours, because country decides which requirements it is given. Name
// at least one persona in Configuration and request at least one capability
// under it. Source:
// https://docs.bachs.io/connect/guides/create-an-account
type CreateConnectedAccountRequest struct {
	// ContactEmail of the person or business behind the account. Trimmed and
	// lowercased before storage. Required.
	ContactEmail string `json:"contact_email"`

	// DisplayName is the name you want the account listed under.
	DisplayName *string `json:"display_name,omitempty"`

	// Country is the two-letter ISO 3166-1 code for the account; it decides
	// which requirements the account is given.
	Country *string `json:"country,omitempty"`

	// EntityType is AccountEntityCompany or AccountEntityIndividual.
	EntityType *string `json:"entity_type,omitempty"`

	// Configuration names each persona the account needs as a key, with the
	// requested capabilities nested under it.
	Configuration map[string]PersonaConfig `json:"configuration,omitempty"`

	// Responsibilities fixes who absorbs processing fees on the account's
	// own charges. Defaults to the account absorbing them.
	Responsibilities *ResponsibilitiesConfig `json:"responsibilities,omitempty"`
}

// CapabilityRequest requests one capability for a connected account.
type CapabilityRequest struct {
	// Requested is true to request the capability.
	Requested bool `json:"requested"`
}

// UpdateConnectedAccountRequest is the payload for
// ConnectedAccounts.RequestCapabilities: the same configuration shape as
// creation. Naming a persona as a key applies it if the account does not
// already have it; unlike creation, an omitted capabilities block on update
// never blanket-requests. Capabilities cannot be revoked through the API.
type UpdateConnectedAccountRequest struct {
	// Configuration names the personas to apply, with capabilities nested
	// under each.
	Configuration map[string]PersonaConfig `json:"configuration"`
}

// CreateAccountLinkRequest is the payload for
// ConnectedAccounts.CreateAccountLink. Source:
// https://docs.bachs.io/api-reference/connected-accounts/create-an-account-link
type CreateAccountLinkRequest struct {
	// Type is what the account holder is being sent to do: "onboarding" to
	// collect everything the account owes for the first time, or "update" for
	// a later change.
	Type string `json:"type"`

	// RefreshURL is where the account holder is sent when the link is no
	// longer usable, for example after it expired.
	RefreshURL string `json:"refresh_url"`

	// ReturnURL is where the account holder is sent when they finish or
	// abandon the flow. Arriving here is not proof that onboarding completed.
	ReturnURL string `json:"return_url"`

	// CollectionOptions are carried through to the hosted flow and handed back
	// unchanged when the link is opened. Omit unless you were given them.
	CollectionOptions map[string]any `json:"collection_options,omitempty"`
}

// AccountLink is a hosted link that walks a connected account through its
// outstanding Tasks. The URL is returned only when the link is created and
// cannot be read back. Source:
// https://docs.bachs.io/api-reference/connected-accounts/create-an-account-link
type AccountLink struct {
	// ID is the unique identifier for the account link.
	ID string `json:"id"`

	// Object is always "connected_account_link", so a mixed webhook or log
	// stream can be routed on type.
	Object string `json:"object"`

	// Account is the connected account this link onboards.
	Account string `json:"account"`

	// Type echoes the type sent: "onboarding" or "update".
	Type string `json:"type"`

	// Created is when the link was issued, ISO 8601 in UTC.
	Created time.Time `json:"created"`

	// ExpiresAt is when the link stops working, ISO 8601 in UTC. After this
	// the account holder lands on the refresh URL instead.
	ExpiresAt time.Time `json:"expires_at"`

	// URL is where to send the account holder. It carries a single-use
	// credential, so deliver it over a trusted channel and do not log it.
	URL string `json:"url"`

	// PreviousLinkSuperseded is true when issuing this link invalidated an
	// outstanding active link of the same type for the account.
	PreviousLinkSuperseded bool `json:"previous_link_superseded"`
}

// ConnectedAccountCapabilities is the response of
// ConnectedAccounts.ListCapabilities.
type ConnectedAccountCapabilities struct {
	// Items lists every capability applicable to the account, including ones
	// it has never requested.
	Items []ConnectedAccountCapability `json:"items"`
}

// ConnectedAccountCapability is one capability entry.
type ConnectedAccountCapability struct {
	// Name is "payouts", "transfers", "conversions", or "connect".
	Name string `json:"name"`

	// Status is "active", "pending", "restricted", "unrequested", or
	// "unsupported".
	Status string `json:"status"`

	// Requested reports whether the account ever requested this capability.
	Requested bool `json:"requested"`

	// StatusDetails explain why the capability is not active; null when it is
	// active.
	StatusDetails []CapabilityStatusDetail `json:"status_details"`
}

// Create creates a connected account under your organization. ContactEmail is
// the only required field; name at least one persona in Configuration with
// the capabilities it needs. In sandbox, a requested capability whose
// persona is applied is granted active immediately; in live it lands
// restricted until a reviewer enables it.
func (s *ConnectedAccountService) Create(ctx context.Context, req CreateConnectedAccountRequest, opts ...RequestOption) (*ConnectedAccount, *ResponseMeta, error) {
	var out ConnectedAccount
	meta, err := s.request(ctx, http.MethodPost, "/accounts", req, &out, opts...)
	if err != nil {
		return nil, meta, err
	}
	return &out, meta, nil
}

// Get reads one of your connected accounts. This is the only read that
// populates the Capabilities and Requirements blocks, because each costs an
// extra lookup; list items leave both null.
func (s *ConnectedAccountService) Get(ctx context.Context, connectedAccountID string) (*ConnectedAccount, *ResponseMeta, error) {
	var out ConnectedAccount
	meta, err := s.request(ctx, http.MethodGet, "/accounts/"+url.PathEscape(connectedAccountID), nil, &out)
	if err != nil {
		return nil, meta, err
	}
	return &out, meta, nil
}

// List returns a page of your connected accounts. Items never carry
// Capabilities or Requirements; read a single account for those.
func (s *ConnectedAccountService) List(ctx context.Context, params ListParams) (*Page[ConnectedAccount], *ResponseMeta, error) {
	var env pageEnvelope[ConnectedAccount]
	meta, err := s.request(ctx, http.MethodGet, queryPath("/accounts", params), nil, &env)
	if err != nil {
		return nil, meta, err
	}
	return env.page(), meta, nil
}

// RequestCapabilities requests additional capabilities for a connected
// account using the same configuration shape as creation: naming a persona
// as a key applies it if the account does not already have it. In sandbox
// the requested capabilities are granted active immediately; in live they
// land restricted for review. Capabilities cannot be revoked through the
// API.
func (s *ConnectedAccountService) RequestCapabilities(ctx context.Context, connectedAccountID string, req UpdateConnectedAccountRequest) (*ConnectedAccount, *ResponseMeta, error) {
	var out ConnectedAccount
	meta, err := s.request(ctx, http.MethodPost, "/accounts/"+url.PathEscape(connectedAccountID), req, &out)
	if err != nil {
		return nil, meta, err
	}
	return &out, meta, nil
}

// UpdateAccountRequest is the payload for ConnectedAccounts.UpdateAccount:
// contact and profile changes, capability requests, and requirement values
// in one round trip. Omit a field to leave it untouched. Requirement values
// go in Fields, keyed by the field keys from requirements.entries (for
// example "payout_destination" or a "persons" list); every field sent is
// validated together, and if any one is rejected nothing in Fields is saved.
// Source: https://docs.bachs.io/connect/guides/api-onboarding
type UpdateAccountRequest struct {
	// ContactEmail replaces the contact email.
	ContactEmail *string `json:"contact_email,omitempty"`

	// DisplayName replaces the listed name.
	DisplayName *string `json:"display_name,omitempty"`

	// Country replaces the account country.
	Country *string `json:"country,omitempty"`

	// EntityType replaces the entity type.
	EntityType *string `json:"entity_type,omitempty"`

	// Configuration applies personas and requests capabilities, in the
	// same shape as creation.
	Configuration map[string]PersonaConfig `json:"configuration,omitempty"`

	// Fields submits requirement values. Contact details and capability
	// requests in the same call are applied before Fields is validated, so
	// a rejection does not undo them.
	Fields map[string]any `json:"fields,omitempty"`
}

// UpdateAccount changes a connected account: contact details, capability
// requests, and requirement values together. Submitted fields move to
// pending_verification until checked; anything rejected comes back as
// currently_due with an entry in requirements errors. Verified live against
// the sandbox.
func (s *ConnectedAccountService) UpdateAccount(ctx context.Context, connectedAccountID string, req UpdateAccountRequest) (*ConnectedAccount, *ResponseMeta, error) {
	var out ConnectedAccount
	meta, err := s.request(ctx, http.MethodPost, "/accounts/"+url.PathEscape(connectedAccountID), req, &out)
	if err != nil {
		return nil, meta, err
	}
	return &out, meta, nil
}

// CreateAccountLink issues a hosted link that walks a connected account
// through its outstanding Tasks. Creating a link invalidates any outstanding
// active link of the same type for that account, so create one at the moment
// you redirect rather than on every page render.
func (s *ConnectedAccountService) CreateAccountLink(ctx context.Context, connectedAccountID string, req CreateAccountLinkRequest) (*AccountLink, *ResponseMeta, error) {
	var out AccountLink
	meta, err := s.request(ctx, http.MethodPost, "/accounts/"+url.PathEscape(connectedAccountID)+"/account-links", req, &out)
	if err != nil {
		return nil, meta, err
	}
	return &out, meta, nil
}

// ListCapabilities lists every capability applicable to a connected account,
// including ones it has never requested.
func (s *ConnectedAccountService) ListCapabilities(ctx context.Context, connectedAccountID string) (*ConnectedAccountCapabilities, *ResponseMeta, error) {
	var out ConnectedAccountCapabilities
	meta, err := s.request(ctx, http.MethodGet, "/accounts/"+url.PathEscape(connectedAccountID)+"/capabilities", nil, &out)
	if err != nil {
		return nil, meta, err
	}
	return &out, meta, nil
}
