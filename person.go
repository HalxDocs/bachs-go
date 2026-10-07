package bachs

import (
	"context"
	"net/http"
	"net/url"
	"time"
)

// PersonService provides the persons subresource of a Connect account: the
// representative, owners, and directors who carry the account's identity.
// Submit a person's details with their ID numbers, attach an ID document,
// and read the verification result from the person. Source:
// https://docs.bachs.io/connect/guides/identity-verification
type PersonService struct {
	service
}

// PersonRelationship is the set of roles one person holds on an account. A
// founder is commonly representative, owner, and director at once.
type PersonRelationship struct {
	// Representative is true for the person who acts for the account.
	Representative bool `json:"representative"`

	// Owner is true for a person with an ownership stake.
	Owner bool `json:"owner"`

	// Director is true for a company director.
	Director bool `json:"director"`

	// PercentOwnership is the ownership share, when the person is an owner.
	PercentOwnership *int `json:"percent_ownership"`
}

// PersonVerification is a person's identity verification state.
type PersonVerification struct {
	// Status is "pending" until a reviewer decides, then "passed" or
	// "failed".
	Status string `json:"status"`

	// DocumentProvided is true once an ID document has been attached. What
	// was supplied is reported without echoing the number or the file.
	DocumentProvided bool `json:"document_provided"`

	// FailureReason explains a "failed" status. Null otherwise.
	FailureReason *string `json:"failure_reason"`
}

// Person is one human on an account: the representative, an owner, or a
// director.
type Person struct {
	// ID uniquely identifies the person (for example "per_...").
	ID string `json:"id"`

	// FirstName of the person. Null when not supplied.
	FirstName *string `json:"first_name"`

	// LastName of the person. Null when not supplied.
	LastName *string `json:"last_name"`

	// DOB is the date of birth (YYYY-MM-DD). Null when not supplied.
	DOB *string `json:"dob"`

	// Phone of the person. Null when not supplied.
	Phone *string `json:"phone"`

	// Email of the person. Null when not supplied.
	Email *string `json:"email"`

	// IDNumberProvided is true once an ID number has been supplied, without
	// echoing the number itself.
	IDNumberProvided bool `json:"id_number_provided"`

	// Relationship is the set of roles the person holds.
	Relationship PersonRelationship `json:"relationship"`

	// Verification is the person's identity verification state.
	Verification PersonVerification `json:"verification"`

	// CreatedAt is when the person was created. Null on responses that do
	// not carry it, such as a single-person read.
	CreatedAt *time.Time `json:"created_at"`

	// UpdatedAt is when the person was last updated.
	UpdatedAt *time.Time `json:"updated_at"`
}

// PersonRelationshipRequest sets the roles a person holds. Flags are pointers
// so an update can distinguish "leave alone" (nil) from an explicit value.
type PersonRelationshipRequest struct {
	// Representative marks the person who acts for the account.
	Representative *bool `json:"representative,omitempty"`

	// Owner marks a person with an ownership stake.
	Owner *bool `json:"owner,omitempty"`

	// Director marks a company director.
	Director *bool `json:"director,omitempty"`

	// PercentOwnership is the ownership share.
	PercentOwnership *int `json:"percent_ownership,omitempty"`
}

// PersonIDNumberRequest is one government or scheme identifier for a person.
// Each names its own scheme under Type ("nin", "bvn", "passport", a driver's
// licence) and carries its own IssuingCountry.
type PersonIDNumberRequest struct {
	// Type is the identifier scheme (for example "nin" or "bvn").
	Type string `json:"type"`

	// Value is the identifier itself. It is never echoed back: reads report
	// IDNumberProvided instead.
	Value string `json:"value"`

	// IssuingCountry is the two-letter country code of the issuer.
	IssuingCountry *string `json:"issuing_country,omitempty"`
}

// CreatePersonRequest is the payload for Persons.Create and, with only the
// changing fields set, Persons.Update.
type CreatePersonRequest struct {
	// FirstName of the person.
	FirstName *string `json:"first_name,omitempty"`

	// LastName of the person.
	LastName *string `json:"last_name,omitempty"`

	// DOB is the date of birth (YYYY-MM-DD).
	DOB *string `json:"dob,omitempty"`

	// Email of the person.
	Email *string `json:"email,omitempty"`

	// Phone of the person.
	Phone *string `json:"phone,omitempty"`

	// Relationship flags the person as representative, owner, and/or
	// director. The virtual_accounts capability requires the BVN of the
	// person marked as representative.
	Relationship *PersonRelationshipRequest `json:"relationship,omitempty"`

	// IDNumbers are the person's identifiers, each with its own scheme and
	// issuing country.
	IDNumbers []PersonIDNumberRequest `json:"id_numbers,omitempty"`
}

// UpdatePersonRequest is the payload for Persons.Update. Only the fields you
// send change; everything absent is left alone.
type UpdatePersonRequest = CreatePersonRequest

// Document slots a person document can be attached to.
const (
	// PersonDocumentPrimary is a government ID.
	PersonDocumentPrimary = "primary_verification"

	// PersonDocumentSecondary is address evidence.
	PersonDocumentSecondary = "secondary_verification"
)

// AttachPersonDocumentRequest is the payload for Persons.AttachDocument.
// Upload the file first (for example via Media.Upload with scope
// "identity_document"), then point a document slot at the returned upload
// ID. A two-sided card is two files, each attached with its own Side.
type AttachPersonDocumentRequest struct {
	// File is the upload ID of the previously uploaded file.
	File string `json:"file"`

	// Document is the slot: PersonDocumentPrimary for a government ID,
	// PersonDocumentSecondary for address evidence.
	Document string `json:"document"`

	// Side is "front" or "back" for a two-sided document.
	Side *string `json:"side,omitempty"`
}

// PersonDocument is a document attached to a person.
type PersonDocument struct {
	// ID uniquely identifies the attachment (for example "doc_...").
	ID string `json:"id"`

	// Person is the person the document was attached to.
	Person string `json:"person"`

	// DocumentType is the slot the file fills.
	DocumentType string `json:"document_type"`

	// FileName of the attached file.
	FileName string `json:"file_name"`

	// UploadedAt is when the document was attached.
	UploadedAt time.Time `json:"uploaded_at"`
}

// List returns the people on an account. A returning account holder should
// confirm rather than re-type: find the entry whose
// Relationship.Representative is true. An empty list means no representative
// exists yet.
func (s *PersonService) List(ctx context.Context, accountID string, params ListParams) (*Page[Person], *ResponseMeta, error) {
	var env pageEnvelope[Person]
	meta, err := s.request(ctx, http.MethodGet, queryPath("/accounts/"+url.PathEscape(accountID)+"/persons", params), nil, &env)
	if err != nil {
		return nil, meta, err
	}
	return env.page(), meta, nil
}

// Create submits a person on an account and flags their roles, for example
// the representative. The same person is updated later via Update.
func (s *PersonService) Create(ctx context.Context, accountID string, req CreatePersonRequest, opts ...RequestOption) (*Person, *ResponseMeta, error) {
	var out Person
	meta, err := s.request(ctx, http.MethodPost, "/accounts/"+url.PathEscape(accountID)+"/persons", req, &out, opts...)
	if err != nil {
		return nil, meta, err
	}
	return &out, meta, nil
}

// Get reads one person on its own, including the verification result. A
// person passing is not the same as a capability being enabled: wait for
// capability.updated before unlocking anything.
func (s *PersonService) Get(ctx context.Context, accountID, personID string) (*Person, *ResponseMeta, error) {
	var out Person
	meta, err := s.request(ctx, http.MethodGet, "/accounts/"+url.PathEscape(accountID)+"/persons/"+url.PathEscape(personID), nil, &out)
	if err != nil {
		return nil, meta, err
	}
	return &out, meta, nil
}

// Update changes a person. Address the person by ID and send only what is
// changing; anything absent is left alone.
func (s *PersonService) Update(ctx context.Context, accountID, personID string, req UpdatePersonRequest) (*Person, *ResponseMeta, error) {
	var out Person
	meta, err := s.request(ctx, http.MethodPost, "/accounts/"+url.PathEscape(accountID)+"/persons/"+url.PathEscape(personID), req, &out)
	if err != nil {
		return nil, meta, err
	}
	return &out, meta, nil
}

// AttachDocument points one of a person's document slots at a previously
// uploaded file. Attaching a document does not verify it: a reviewer
// accepting it is what moves Verification.Status to "passed".
func (s *PersonService) AttachDocument(ctx context.Context, accountID, personID string, req AttachPersonDocumentRequest, opts ...RequestOption) (*PersonDocument, *ResponseMeta, error) {
	var out PersonDocument
	meta, err := s.request(ctx, http.MethodPost, "/accounts/"+url.PathEscape(accountID)+"/persons/"+url.PathEscape(personID)+"/documents", req, &out, opts...)
	if err != nil {
		return nil, meta, err
	}
	return &out, meta, nil
}

// Delete removes a person from an account.
func (s *PersonService) Delete(ctx context.Context, accountID, personID string) (*ResponseMeta, error) {
	meta, err := s.request(ctx, http.MethodDelete, "/accounts/"+url.PathEscape(accountID)+"/persons/"+url.PathEscape(personID), nil, nil)
	if err != nil {
		return meta, err
	}
	return meta, nil
}
