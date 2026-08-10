package bachs

import (
	"context"
	"io"
	"net/http"
	"net/url"
	"time"
)

// DisputeService provides methods for responding to chargebacks: listing and
// reading disputes, uploading supporting documents, drafting evidence, and
// submitting it for network review. See
// https://docs.bachs.io/connect/marketplaces/refunds-and-disputes for how
// reversals are absorbed between platform and connected account.
type DisputeService struct {
	service
}

// Dispute is a chargeback a customer's bank has raised against one of your
// charges.
type Dispute struct {
	// DisputeID uniquely identifies the dispute.
	DisputeID string `json:"dispute_id"`

	// ChargeID is the charge the dispute was raised against.
	ChargeID *string `json:"charge_id"`

	// Amount disputed, as a decimal string.
	Amount string `json:"amount"`

	// Currency of the disputed amount.
	Currency string `json:"currency"`

	// Status is "needs_response", "under_review", "won", "lost", or
	// "closed".
	Status string `json:"status"`

	// IsResponseEditable is true while you can still change and submit
	// evidence. It becomes false once the response is submitted.
	IsResponseEditable bool `json:"is_response_editable"`

	// Reason code the bank gave for the dispute (for example "fraudulent").
	Reason *string `json:"reason"`

	// ResponseDeadlineAt is when your evidence must be submitted by.
	ResponseDeadlineAt *time.Time `json:"response_deadline_at"`

	// Evidence is the current evidence draft. Present on the full object
	// returned by Get; list items carry it as null.
	Evidence *DisputeEvidence `json:"evidence"`

	// LatestSubmission, when a response has been submitted, describes it.
	LatestSubmission *DisputeSubmission `json:"latest_submission"`

	// CreatedAt is when the dispute was raised.
	CreatedAt time.Time `json:"created_at"`

	// UpdatedAt is when the dispute was last updated.
	UpdatedAt time.Time `json:"updated_at"`
}

// DisputeEvidence is the evidence drafted for a dispute response. Attachment
// fields hold document IDs from DisputeService.UploadDocument.
type DisputeEvidence struct {
	// AccessActivityLog is a free-text account activity log.
	AccessActivityLog string `json:"access_activity_log"`

	// BillingAddress on the card.
	BillingAddress *string `json:"billing_address"`

	// CancellationPolicyAttachmentID of the policy document.
	CancellationPolicyAttachmentID string `json:"cancellation_policy_attachment_id"`

	// CancellationPolicyDisclosure explains the policy in plain text.
	CancellationPolicyDisclosure *string `json:"cancellation_policy_disclosure"`

	// CustomerCommunicationAttachmentID of the communication log.
	CustomerCommunicationAttachmentID *string `json:"customer_communication_attachment_id"`

	// CustomerEmailAddress the customer used.
	CustomerEmailAddress *string `json:"customer_email_address"`

	// CustomerName the customer gave.
	CustomerName *string `json:"customer_name"`

	// Notes are free-text notes supporting your response.
	Notes *string `json:"notes"`

	// ProductDescription of what was purchased.
	ProductDescription *string `json:"product_description"`

	// RefundPolicyAttachmentID of the refund policy document.
	RefundPolicyAttachmentID string `json:"refund_policy_attachment_id"`

	// RefundPolicyDisclosure explains the policy in plain text.
	RefundPolicyDisclosure *string `json:"refund_policy_disclosure"`

	// RefundRefusalExplanation states why a refund was refused.
	RefundRefusalExplanation *string `json:"refund_refusal_explanation"`

	// ServiceDate when the goods or services were delivered.
	ServiceDate *string `json:"service_date"`

	// UncategorizedAttachmentID of any other supporting document.
	UncategorizedAttachmentID *string `json:"uncategorized_attachment_id"`
}

// DisputeSubmission describes a submitted dispute response.
type DisputeSubmission struct {
	// SubmissionID uniquely identifies the submission.
	SubmissionID string `json:"submission_id"`

	// Status of the submission (for example "submitted").
	Status string `json:"status"`

	// TriggerSource is who submitted it ("merchant_submit" or similar).
	TriggerSource string `json:"trigger_source"`

	// SubmittedAt is when the submission was accepted.
	SubmittedAt *time.Time `json:"submitted_at"`

	// FailedAt is when the submission failed, if it did.
	FailedAt *time.Time `json:"failed_at"`

	// AttemptSequence counts submission attempts.
	AttemptSequence int `json:"attempt_sequence"`
}

// DisputeDocument is a supporting document uploaded for a dispute, returned
// by DisputeService.UploadDocument. Use its DocumentID as an attachment
// field in the evidence.
type DisputeDocument struct {
	// DocumentID identifies the document for evidence attachment.
	DocumentID string `json:"document_id"`

	// FileName of the uploaded file.
	FileName string `json:"file_name"`

	// MIMEType of the uploaded file.
	MIMEType string `json:"mime_type"`

	// FileSizeBytes of the uploaded file.
	FileSizeBytes int64 `json:"file_size_bytes"`

	// StorageProvider holding the file.
	StorageProvider string `json:"storage_provider"`

	// UploadedAt is when the file was uploaded.
	UploadedAt time.Time `json:"uploaded_at"`
}

// DisputeEvidenceUpdateRequest is the payload for
// DisputeService.UpdateEvidence. Same shape as the evidence block on the
// full dispute object; only the fields you send are changed.
type DisputeEvidenceUpdateRequest struct {
	// AccessActivityLog is a free-text account activity log.
	AccessActivityLog string `json:"access_activity_log,omitempty"`

	// BillingAddress on the card.
	BillingAddress string `json:"billing_address,omitempty"`

	// CancellationPolicyAttachmentID of the policy document.
	CancellationPolicyAttachmentID string `json:"cancellation_policy_attachment_id,omitempty"`

	// CancellationPolicyDisclosure explains the policy in plain text.
	CancellationPolicyDisclosure string `json:"cancellation_policy_disclosure,omitempty"`

	// CustomerCommunicationAttachmentID of the communication log.
	CustomerCommunicationAttachmentID string `json:"customer_communication_attachment_id,omitempty"`

	// CustomerEmailAddress the customer used.
	CustomerEmailAddress string `json:"customer_email_address,omitempty"`

	// CustomerName the customer gave.
	CustomerName string `json:"customer_name,omitempty"`

	// Notes are free-text notes supporting your response.
	Notes string `json:"notes,omitempty"`

	// ProductDescription of what was purchased.
	ProductDescription string `json:"product_description,omitempty"`

	// RefundPolicyAttachmentID of the refund policy document.
	RefundPolicyAttachmentID string `json:"refund_policy_attachment_id,omitempty"`

	// RefundPolicyDisclosure explains the policy in plain text.
	RefundPolicyDisclosure string `json:"refund_policy_disclosure,omitempty"`

	// RefundRefusalExplanation states why a refund was refused.
	RefundRefusalExplanation string `json:"refund_refusal_explanation,omitempty"`

	// ServiceDate when the goods or services were delivered.
	ServiceDate string `json:"service_date,omitempty"`

	// UncategorizedAttachmentID of any other supporting document.
	UncategorizedAttachmentID string `json:"uncategorized_attachment_id,omitempty"`
}

// DisputeEvidenceUpdateResponse is the result of DisputeService.UpdateEvidence.
type DisputeEvidenceUpdateResponse struct {
	// DisputeID the evidence belongs to.
	DisputeID string `json:"dispute_id"`

	// Status of the dispute after the update.
	Status string `json:"status"`

	// IsResponseEditable remains true until the response is submitted.
	IsResponseEditable bool `json:"is_response_editable"`

	// EvidenceUpdatedAt is when the evidence was last changed.
	EvidenceUpdatedAt time.Time `json:"evidence_updated_at"`
}

// DisputeSubmitResponse is the result of DisputeService.Submit.
type DisputeSubmitResponse struct {
	// DisputeID that was submitted.
	DisputeID string `json:"dispute_id"`

	// Status of the dispute after submission (for example "under_review").
	Status string `json:"status"`

	// IsResponseEditable is false once submitted — the response is locked.
	IsResponseEditable bool `json:"is_response_editable"`

	// Submission describes the accepted submission.
	Submission DisputeSubmitSubmission `json:"submission"`
}

// DisputeSubmitSubmission describes an accepted dispute response submission.
type DisputeSubmitSubmission struct {
	// SubmissionID uniquely identifies the submission.
	SubmissionID string `json:"submission_id"`

	// SubmissionStatus (for example "submitted").
	SubmissionStatus string `json:"submission_status"`

	// TriggerSource is who submitted it.
	TriggerSource string `json:"trigger_source"`

	// SubmittedAt is when the submission was accepted.
	SubmittedAt *time.Time `json:"submitted_at"`
}

// List returns a page of disputes for your organization, newest first.
// Filter with ListParams.Status ("needs_response", "under_review", "won",
// "lost", "closed") and ListParams.FromDate / ToDate (YYYY-MM-DD).
func (s *DisputeService) List(ctx context.Context, params ListParams) (*Page[Dispute], *ResponseMeta, error) {
	var env pageEnvelope[Dispute]
	meta, err := s.request(ctx, http.MethodGet, queryPath("/disputes", params), nil, &env)
	if err != nil {
		return nil, meta, err
	}
	return env.page(), meta, nil
}

// Get returns full details for a single dispute, including the current
// evidence draft and latest submission metadata.
func (s *DisputeService) Get(ctx context.Context, disputeID string) (*Dispute, *ResponseMeta, error) {
	var out Dispute
	meta, err := s.request(ctx, http.MethodGet, "/disputes/"+url.PathEscape(disputeID), nil, &out)
	if err != nil {
		return nil, meta, err
	}
	return &out, meta, nil
}

// UploadDocument uploads a supporting document (maximum 20 MB) for a dispute
// and returns a DocumentID to attach as evidence. Scope is a logical grouping
// label, like the media uploads.
func (s *DisputeService) UploadDocument(ctx context.Context, fileName string, file io.Reader, scope string, opts ...RequestOption) (*DisputeDocument, *ResponseMeta, error) {
	body, contentType, err := multipartUpload(fileName, file, map[string]string{"scope": scope})
	if err != nil {
		return nil, nil, err
	}

	var out DisputeDocument
	meta, err := s.request(ctx, http.MethodPost, "/disputes/uploads", nil, &out, append(opts, withRawBody(body, contentType))...)
	if err != nil {
		return nil, meta, err
	}
	return &out, meta, nil
}

// UpdateEvidence saves or updates dispute evidence fields before final
// submission. Evidence can be changed iteratively while the dispute remains
// editable; only the fields you send are changed.
func (s *DisputeService) UpdateEvidence(ctx context.Context, disputeID string, req DisputeEvidenceUpdateRequest) (*DisputeEvidenceUpdateResponse, *ResponseMeta, error) {
	var out DisputeEvidenceUpdateResponse
	meta, err := s.request(ctx, http.MethodPatch, "/disputes/"+url.PathEscape(disputeID)+"/evidence", req, &out)
	if err != nil {
		return nil, meta, err
	}
	return &out, meta, nil
}

// Submit submits the saved dispute evidence for network review. This action
// is irreversible: the response locks and further evidence edits are
// rejected.
func (s *DisputeService) Submit(ctx context.Context, disputeID string, opts ...RequestOption) (*DisputeSubmitResponse, *ResponseMeta, error) {
	var out DisputeSubmitResponse
	meta, err := s.request(ctx, http.MethodPost, "/disputes/"+url.PathEscape(disputeID)+"/submit", struct{}{}, &out, opts...)
	if err != nil {
		return nil, meta, err
	}
	return &out, meta, nil
}
