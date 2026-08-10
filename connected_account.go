package bachs

import (
	"context"
	"io"
	"net/http"
	"net/url"
	"time"
)

// ConnectedAccountService provides the Connect platform API: creating and
// reading connected accounts, requesting capabilities, walking a connected
// account through its onboarding Tasks, managing account documents, and
// issuing hosted account links. Requires the connect capability on your own
// organization. Source: https://docs.bachs.io/connect/overview
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

	// FeeHandling is who absorbs processing fees at checkout:
	// "org_pays_fee" or "customer_pays_fee".
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

// CreateConnectedAccountRequest is the payload for ConnectedAccounts.Create.
// The account starts with nothing enabled; the capabilities you request here
// decide which Tasks it is given. Source:
// https://docs.bachs.io/api-reference/connected-accounts/create-a-connected-account
type CreateConnectedAccountRequest struct {
	// ContactEmail of the person or business behind the account. Trimmed and
	// lowercased before storage. Required.
	ContactEmail string `json:"contact_email"`

	// DisplayName is the name you want the account listed under.
	DisplayName *string `json:"display_name,omitempty"`

	// FirstName of the person being onboarded.
	FirstName *string `json:"first_name,omitempty"`

	// LastName of the person being onboarded.
	LastName *string `json:"last_name,omitempty"`

	// Country is the two-letter ISO 3166-1 code for the account; it decides
	// which Tasks the account is given.
	Country *string `json:"country,omitempty"`

	// EntityType is "company", "individual", or "business" (a legacy alias
	// stored as "company").
	EntityType *string `json:"entity_type,omitempty"`

	// Capabilities to request, keyed by capability name. Omitting the field
	// requests every capability the account is eligible for.
	Capabilities map[string]CapabilityRequest `json:"capabilities,omitempty"`

	// Controller sets the fee arrangement. Defaults to the account absorbing
	// its own processing fees.
	Controller *ControllerRequest `json:"controller,omitempty"`
}

// CapabilityRequest requests one capability for a connected account.
type CapabilityRequest struct {
	// Requested is true to request the capability.
	Requested bool `json:"requested"`
}

// ControllerRequest sets the fee arrangement when creating a connected
// account.
type ControllerRequest struct {
	// Fees describes who absorbs processing fees on the account's charges.
	Fees ControllerFeesRequest `json:"fees"`
}

// ControllerFeesRequest is the fees portion of the create request.
type ControllerFeesRequest struct {
	// Payer is "account": the connected account absorbs processing fees on
	// its own charges.
	Payer string `json:"payer"`
}

// UpdateConnectedAccountRequest is the payload for
// ConnectedAccounts.RequestCapabilities. Capabilities cannot be revoked
// through the API. Source:
// https://docs.bachs.io/api-reference/connected-accounts/request-capabilities-on-a-connected-account
type UpdateConnectedAccountRequest struct {
	// Capabilities to request, keyed by capability name. Only names set to
	// true are acted on; false entries are ignored.
	Capabilities map[string]bool `json:"capabilities"`
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

// TaskChecklist is the response of ConnectedAccounts.GetTaskChecklist and
// ConnectedAccounts.SubmitTaskValues: every field the account owes, both flat
// and grouped by capability. Source:
// https://docs.bachs.io/api-reference/connected-accounts/get-the-task-checklist
type TaskChecklist struct {
	// OrganizationID is the connected account this checklist belongs to.
	OrganizationID string `json:"organization_id"`

	// EntityType is "company" or "individual". Null until set.
	EntityType *string `json:"entity_type"`

	// Country the checklist was computed for. Null until the account has a
	// country.
	Country *string `json:"country"`

	// CurrentlyDue is how many fields are currently due.
	CurrentlyDue int `json:"currently_due"`

	// PendingReview is how many fields are waiting on a person to decide.
	PendingReview int `json:"pending_review"`

	// InVerification is how many fields are pending verification.
	InVerification int `json:"in_verification"`

	// NeedsAttention is how many fields were rejected and must be
	// resubmitted.
	NeedsAttention int `json:"needs_attention"`

	// SetupStatus is "incomplete", "awaiting_review", or "complete".
	SetupStatus string `json:"setup_status"`

	// Checklist is every field the account owes, deduplicated across
	// capabilities.
	Checklist []TaskFieldItem `json:"checklist"`

	// Capabilities groups the same fields by the capability that caused them.
	Capabilities []TaskCapabilityGroup `json:"capabilities"`
}

// TaskFieldItem is one field in a Task checklist.
type TaskFieldItem struct {
	// FieldKey is the canonical key to send the value back under when
	// submitting. Nested keys are dotted.
	FieldKey string `json:"field_key"`

	// Label is a human-readable name for the field, safe to show the account
	// holder verbatim.
	Label string `json:"label"`

	// Group names which part of the account the field belongs to (for example
	// "identity", "representative", "company").
	Group *string `json:"group"`

	// State is "currently_due", "eventually_due", "pending_verification",
	// "pending_review", "satisfied", or "past_due".
	State string `json:"state"`

	// Provided reports whether a value has been submitted for this field.
	Provided bool `json:"provided"`

	// ErrorReason explains why the submitted value was rejected, safe to show
	// the account holder. Null unless the field was rejected.
	ErrorReason *string `json:"error_reason"`

	// Reference names the resource this field belongs to, for person-level
	// fields. Null for account-level fields.
	Reference *TaskFieldReference `json:"reference"`
}

// TaskFieldReference names the resource a Task field belongs to.
type TaskFieldReference struct {
	// Type is "account" or "person".
	Type string `json:"type"`

	// Resource is the identifier of the resource the Task is about.
	Resource *string `json:"resource"`

	// Label is a human-readable name for the resource.
	Label *string `json:"label"`
}

// TaskCapabilityGroup is one capability's fields within a Task checklist.
type TaskCapabilityGroup struct {
	// CapabilityName is "payouts", "transfers", "conversions", or "connect".
	CapabilityName string `json:"capability_name"`

	// Description summarizes what the capability lets the account do.
	Description *string `json:"description"`

	// Category is a display grouping hint.
	Category *string `json:"category"`

	// State is "requested", "pending_review", or "enabled".
	State string `json:"state"`

	// Satisfied is true when every field the capability needs is satisfied.
	// The account is eligible, not enabled: a person still decides.
	Satisfied bool `json:"satisfied"`

	// Fields this capability needs, in the same shape as the checklist.
	Fields []TaskFieldItem `json:"fields"`
}

// Task is one item in a connected account's worklist: a thing the account
// holder has to do, each with its own open-to-done lifecycle. Source:
// https://docs.bachs.io/api-reference/connected-accounts/list-tasks
type Task struct {
	// ID is the unique identifier for the Task.
	ID string `json:"id"`

	// Title is a short heading, safe to show the account holder verbatim.
	Title string `json:"title"`

	// Description explains what the account holder has to do.
	Description *string `json:"description"`

	// Type is "form_field", "document", "action", or "edit_section".
	Type string `json:"type"`

	// Status is "open", "in_review", "completed", or "rejected".
	Status string `json:"status"`

	// FieldRef is the canonical field key the Task points at.
	FieldRef *string `json:"field_ref"`

	// DocumentType is which document a document Task expects.
	DocumentType *string `json:"document_type"`

	// Requirements constrain the answer. Empty on every Task returned today.
	Requirements map[string]any `json:"requirements"`

	// ResponseContract describes how the answer is shaped.
	ResponseContract map[string]any `json:"response_contract"`

	// DueDate of the Task, ISO 8601 in UTC. Null when no deadline is set.
	DueDate *time.Time `json:"due_date"`

	// ImpactsCapability is the capability paused if the Task is left unmet
	// past its deadline. Null when nothing is on the line.
	ImpactsCapability *string `json:"impacts_capability"`

	// SectionKey is which section an edit_section Task reopens.
	SectionKey *string `json:"section_key"`

	// PastDue is true when the Task is still open or rejected past its due
	// date.
	PastDue bool `json:"past_due"`

	// RejectionReason is the plain-language reason a submission was turned
	// down. Null unless Status is "rejected".
	RejectionReason *string `json:"rejection_reason"`

	// CreatedAt is when the Task was raised, ISO 8601 in UTC.
	CreatedAt time.Time `json:"created_at"`

	// UpdatedAt is when the Task last changed, ISO 8601 in UTC.
	UpdatedAt time.Time `json:"updated_at"`
}

// TaskValues is the response of ConnectedAccounts.GetTaskValues: the field
// values a connected account has already provided. Source:
// https://docs.bachs.io/api-reference/connected-accounts/get-submitted-task-values
type TaskValues struct {
	// OrganizationID is the connected account these values belong to.
	OrganizationID string `json:"organization_id"`

	// EntityType is "company" or "individual". Null until set.
	EntityType *string `json:"entity_type"`

	// Values are the account-level fields and their current values. People
	// are returned separately under Persons.
	Values []TaskValueItem `json:"values"`

	// Persons has one entry per person on the account, rolled up rather than
	// split into fields. Each entry's shape is not fully documented by the
	// API, so entries are kept as raw JSON objects.
	Persons []map[string]any `json:"persons"`
}

// TaskValueItem is one field value on a connected account.
type TaskValueItem struct {
	// Field is the canonical key for the field, the same key it is submitted
	// under.
	Field string `json:"field"`

	// Label is a human-readable name for the field.
	Label string `json:"label"`

	// Group is the onboarding section the field is shown in.
	Group *string `json:"group"`

	// Provided reports whether the account has a value for this field.
	Provided bool `json:"provided"`

	// Sensitive is true when the value is never echoed back: Value stays null
	// and Display reads "Provided".
	Sensitive bool `json:"sensitive"`

	// Value is the raw value, suitable for prefilling an edit form. Always
	// null when Sensitive is true.
	Value any `json:"value"`

	// Display is a one-line summary of the value for a review card.
	Display *string `json:"display"`

	// ReferenceData holds resolved labels for a value stored as a code.
	ReferenceData map[string]any `json:"reference_data"`
}

// SubmitTasksRequest is the payload for ConnectedAccounts.SubmitTaskValues.
// Source:
// https://docs.bachs.io/api-reference/connected-accounts/submit-task-values
type SubmitTasksRequest struct {
	// Fields provides the values being submitted, keyed by the canonical
	// field keys from the checklist. Send only what you are changing;
	// anything absent is left alone.
	Fields map[string]any `json:"fields"`

	// Draft is true to persist partial values, with validation problems
	// returned on the refreshed checklist instead of failing the request.
	Draft bool `json:"draft,omitempty"`
}

// ReusableIdentity is the response of ConnectedAccounts.GetReusableIdentity:
// whether the person behind this account already verified on another account
// under the same owner. Source:
// https://docs.bachs.io/api-reference/connected-accounts/get-a-reusable-identity
type ReusableIdentity struct {
	// Available is true when there is an identity to reuse. When false, every
	// other field is null or empty.
	Available bool `json:"available"`

	// PersonPublicID identifies the verified person. Send it back to apply
	// the identity.
	PersonPublicID *string `json:"person_public_id"`

	// FirstName on the verified identity.
	FirstName *string `json:"first_name"`

	// LastName on the verified identity.
	LastName *string `json:"last_name"`

	// Country the identity was verified in.
	Country *string `json:"country"`

	// VerificationStatus of the offered identity; only a "verified" identity
	// is ever offered.
	VerificationStatus *string `json:"verification_status"`

	// UsedBy lists the owner's other businesses already using this identity.
	UsedBy []string `json:"used_by"`
}

// ApplyReusableIdentityRequest is the payload for
// ConnectedAccounts.ApplyReusableIdentity.
type ApplyReusableIdentityRequest struct {
	// PersonPublicID is the verified person to copy onto this account, taken
	// from the reusable identity. Only an identity belonging to the same
	// owner can be applied.
	PersonPublicID string `json:"person_public_id"`
}

// ApplyReusableIdentityResponse is the result of applying a reusable identity.
type ApplyReusableIdentityResponse struct {
	// Applied is true when the identity was copied onto this account's
	// representative.
	Applied bool `json:"applied"`

	// VerificationStatus is the representative's state after the copy.
	VerificationStatus string `json:"verification_status"`
}

// TaskBankList is the response of ConnectedAccounts.ListBanks: the banks a
// connected account can name as its payout destination.
type TaskBankList struct {
	// Country the list was resolved for, uppercased.
	Country string `json:"country"`

	// Banks available as payout destinations for this country.
	Banks []TaskBank `json:"banks"`
}

// TaskBank is one bank a connected account can name as its payout
// destination.
type TaskBank struct {
	// Name to show the account holder.
	Name string `json:"name"`

	// Code to send as bank_code when resolving an account or submitting a
	// payout destination.
	Code string `json:"code"`
}

// TaskMobileMoneyList is the response of ConnectedAccounts.ListMobileMoneyProviders.
type TaskMobileMoneyList struct {
	// Country the list was resolved for, uppercased.
	Country string `json:"country"`

	// Providers available in this country, deduplicated and in display order.
	// Empty when the country has none.
	Providers []string `json:"providers"`
}

// ResolveTaskBankAccountRequest is the payload for
// ConnectedAccounts.ResolveBankAccount.
type ResolveTaskBankAccountRequest struct {
	// AccountNumber to look up, digits only and exactly as typed.
	AccountNumber string `json:"account_number"`

	// BankCode of the bank holding the account, from the bank list.
	BankCode string `json:"bank_code"`

	// Country to resolve in, two-letter ISO 3166-1. Falls back to the
	// connected account's country.
	Country *string `json:"country,omitempty"`
}

// ResolveTaskBankAccountResponse is the result of resolving a bank account.
// Check Resolved before trusting AccountName.
type ResolveTaskBankAccountResponse struct {
	// Resolved is true when the account number was matched. False covers a
	// wrong number, an unsupported country, and an unresolved lookup.
	Resolved bool `json:"resolved"`

	// AccountName registered on the account. Show it back for confirmation.
	// Null when not resolved.
	AccountName *string `json:"account_name"`

	// AccountNumber as held on record, which can be normalised from what was
	// sent. Null when not resolved.
	AccountNumber *string `json:"account_number"`

	// Message explains why the lookup did not resolve, safe to show the
	// account holder. Null on a successful match.
	Message *string `json:"message"`
}

// Create creates a connected account under your organization. The account
// starts with nothing enabled; the capabilities you request decide which Tasks
// it is given. Requires an active connect capability on your own organization.
func (s *ConnectedAccountService) Create(ctx context.Context, req CreateConnectedAccountRequest, opts ...RequestOption) (*ConnectedAccount, *ResponseMeta, error) {
	var out ConnectedAccount
	meta, err := s.request(ctx, http.MethodPost, "/organizations/connected-accounts", req, &out, opts...)
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
	meta, err := s.request(ctx, http.MethodGet, "/connected-accounts/"+url.PathEscape(connectedAccountID), nil, &out)
	if err != nil {
		return nil, meta, err
	}
	return &out, meta, nil
}

// List returns a page of your connected accounts. Items never carry
// Capabilities or Requirements; read a single account for those.
func (s *ConnectedAccountService) List(ctx context.Context, params ListParams) (*Page[ConnectedAccount], *ResponseMeta, error) {
	var env pageEnvelope[ConnectedAccount]
	meta, err := s.request(ctx, http.MethodGet, queryPath("/organizations/connected-accounts", params), nil, &env)
	if err != nil {
		return nil, meta, err
	}
	return env.page(), meta, nil
}

// RequestCapabilities requests additional capabilities for a connected
// account. Each newly requested capability lands as "pending" and surfaces
// its Tasks; requesting authorizes nothing — a person enables the capability
// once the Tasks are satisfied. Capabilities cannot be revoked through the
// API.
func (s *ConnectedAccountService) RequestCapabilities(ctx context.Context, connectedAccountID string, req UpdateConnectedAccountRequest) (*ConnectedAccount, *ResponseMeta, error) {
	var out ConnectedAccount
	meta, err := s.request(ctx, http.MethodPatch, "/connected-accounts/"+url.PathEscape(connectedAccountID), req, &out)
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
	meta, err := s.request(ctx, http.MethodPost, "/connected-accounts/"+url.PathEscape(connectedAccountID)+"/account-links", req, &out)
	if err != nil {
		return nil, meta, err
	}
	return &out, meta, nil
}

// ListCapabilities lists every capability applicable to a connected account,
// including ones it has never requested.
func (s *ConnectedAccountService) ListCapabilities(ctx context.Context, connectedAccountID string) (*ConnectedAccountCapabilities, *ResponseMeta, error) {
	var out ConnectedAccountCapabilities
	meta, err := s.request(ctx, http.MethodGet, "/connected-accounts/"+url.PathEscape(connectedAccountID)+"/capabilities", nil, &out)
	if err != nil {
		return nil, meta, err
	}
	return &out, meta, nil
}

// GetTaskChecklist returns every field a connected account owes, with the
// state each field is in, both flat and grouped by capability.
func (s *ConnectedAccountService) GetTaskChecklist(ctx context.Context, connectedAccountID string) (*TaskChecklist, *ResponseMeta, error) {
	var out TaskChecklist
	meta, err := s.request(ctx, http.MethodGet, "/connected-accounts/"+url.PathEscape(connectedAccountID)+"/requirements/checklist", nil, &out)
	if err != nil {
		return nil, meta, err
	}
	return &out, meta, nil
}

// ListTasks lists a connected account's worklist: one item per thing the
// account holder has to do. Filter with ListParams.Status.
func (s *ConnectedAccountService) ListTasks(ctx context.Context, connectedAccountID string, params ListParams) (*Page[Task], *ResponseMeta, error) {
	var env pageEnvelope[Task]
	meta, err := s.request(ctx, http.MethodGet, queryPath("/connected-accounts/"+url.PathEscape(connectedAccountID)+"/requirements/tasks", params), nil, &env)
	if err != nil {
		return nil, meta, err
	}
	return env.page(), meta, nil
}

// GetTaskValues reads back what a connected account has already provided, so
// you can show a review card or prefill an edit form. Identity documents and
// bank account numbers come back marked sensitive with no value.
func (s *ConnectedAccountService) GetTaskValues(ctx context.Context, connectedAccountID string) (*TaskValues, *ResponseMeta, error) {
	var out TaskValues
	meta, err := s.request(ctx, http.MethodGet, "/connected-accounts/"+url.PathEscape(connectedAccountID)+"/requirements/values", nil, &out)
	if err != nil {
		return nil, meta, err
	}
	return &out, meta, nil
}

// SubmitTaskValues provides values for a connected account's Tasks and
// returns the refreshed checklist in the same round trip. Send only the
// fields you are changing; anything absent is left alone. Submitting collects
// data, it does not grant capabilities — a person still decides.
func (s *ConnectedAccountService) SubmitTaskValues(ctx context.Context, connectedAccountID string, req SubmitTasksRequest) (*TaskChecklist, *ResponseMeta, error) {
	var out TaskChecklist
	meta, err := s.request(ctx, http.MethodPost, "/connected-accounts/"+url.PathEscape(connectedAccountID)+"/requirements/submit", req, &out)
	if err != nil {
		return nil, meta, err
	}
	return &out, meta, nil
}

// GetReusableIdentity checks whether the person behind this connected account
// already verified on another account under the same owner. Returns
// Available: false with every other field empty when there is nothing to
// reuse.
func (s *ConnectedAccountService) GetReusableIdentity(ctx context.Context, connectedAccountID string) (*ReusableIdentity, *ResponseMeta, error) {
	var out ReusableIdentity
	meta, err := s.request(ctx, http.MethodGet, "/connected-accounts/"+url.PathEscape(connectedAccountID)+"/requirements/reusable-identity", nil, &out)
	if err != nil {
		return nil, meta, err
	}
	return &out, meta, nil
}

// ApplyReusableIdentity copies a previously verified identity onto this
// connected account's representative, so the person skips the identity check.
// Business details, the payout destination, and terms acceptance are still
// collected per account.
func (s *ConnectedAccountService) ApplyReusableIdentity(ctx context.Context, connectedAccountID string, req ApplyReusableIdentityRequest) (*ApplyReusableIdentityResponse, *ResponseMeta, error) {
	var out ApplyReusableIdentityResponse
	meta, err := s.request(ctx, http.MethodPost, "/connected-accounts/"+url.PathEscape(connectedAccountID)+"/requirements/reusable-identity/apply", req, &out)
	if err != nil {
		return nil, meta, err
	}
	return &out, meta, nil
}

// ListBanks lists the banks a connected account can name as its payout
// destination. Use the Code from this list when resolving an account number
// or submitting a payout destination. Pass country to override the account's
// country.
func (s *ConnectedAccountService) ListBanks(ctx context.Context, connectedAccountID, country string) (*TaskBankList, *ResponseMeta, error) {
	var out TaskBankList
	path := "/connected-accounts/" + url.PathEscape(connectedAccountID) + "/requirements/banks"
	if country != "" {
		path += "?country=" + url.QueryEscape(country)
	}
	meta, err := s.request(ctx, http.MethodGet, path, nil, &out)
	if err != nil {
		return nil, meta, err
	}
	return &out, meta, nil
}

// ListMobileMoneyProviders lists the mobile money providers available for a
// connected account's country. Returns an empty providers array for a country
// with none, rather than an error.
func (s *ConnectedAccountService) ListMobileMoneyProviders(ctx context.Context, connectedAccountID, country string) (*TaskMobileMoneyList, *ResponseMeta, error) {
	var out TaskMobileMoneyList
	path := "/connected-accounts/" + url.PathEscape(connectedAccountID) + "/requirements/momo"
	if country != "" {
		path += "?country=" + url.QueryEscape(country)
	}
	meta, err := s.request(ctx, http.MethodGet, path, nil, &out)
	if err != nil {
		return nil, meta, err
	}
	return &out, meta, nil
}

// ResolveBankAccount looks up the name registered on a bank account before it
// is submitted as a connected account's payout destination. A number that
// does not match returns Resolved: false, not an error.
func (s *ConnectedAccountService) ResolveBankAccount(ctx context.Context, connectedAccountID string, req ResolveTaskBankAccountRequest) (*ResolveTaskBankAccountResponse, *ResponseMeta, error) {
	var out ResolveTaskBankAccountResponse
	meta, err := s.request(ctx, http.MethodPost, "/connected-accounts/"+url.PathEscape(connectedAccountID)+"/requirements/accounts/resolve", req, &out)
	if err != nil {
		return nil, meta, err
	}
	return &out, meta, nil
}

// UploadDocument uploads a file (at most 20 MB) against a connected account.
// The returned upload_id is the value of the document field you are
// satisfying when submitting Tasks; uploading on its own satisfies nothing.
func (s *ConnectedAccountService) UploadDocument(ctx context.Context, connectedAccountID, fileName string, file io.Reader, scope string, opts ...RequestOption) (*Upload, *ResponseMeta, error) {
	body, contentType, err := multipartUpload(fileName, file, map[string]string{"scope": scope})
	if err != nil {
		return nil, nil, err
	}

	var out Upload
	meta, err := s.request(ctx, http.MethodPost, "/connected-accounts/"+url.PathEscape(connectedAccountID)+"/uploads", nil, &out, append(opts, withRawBody(body, contentType))...)
	if err != nil {
		return nil, meta, err
	}
	return &out, meta, nil
}

// GetDocument reads the metadata for a file uploaded against a connected
// account: its name, size, type, and the URL it is served from.
func (s *ConnectedAccountService) GetDocument(ctx context.Context, connectedAccountID, uploadID string) (*Upload, *ResponseMeta, error) {
	var out Upload
	meta, err := s.request(ctx, http.MethodGet, "/connected-accounts/"+url.PathEscape(connectedAccountID)+"/uploads/"+url.PathEscape(uploadID), nil, &out)
	if err != nil {
		return nil, meta, err
	}
	return &out, meta, nil
}
