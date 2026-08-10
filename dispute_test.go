package bachs

import (
	"context"
	"io"
	"net/http"
	"strings"
	"testing"
)

const disputeSummaryExample = `{
	"dispute_id": "dsp_3e7b1c9a2f48",
	"charge_id": "chr_8f3a1c9b4e72",
	"amount": "75.00",
	"currency": "USD",
	"status": "needs_response",
	"is_response_editable": true,
	"reason": "fraudulent",
	"response_deadline_at": "2026-03-23T23:59:59.000Z",
	"created_at": "2026-03-09T08:00:00.000Z",
	"updated_at": "2026-03-09T08:00:00.000Z"
}`

func TestListDisputes(t *testing.T) {
	c := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Errorf("method = %s, want GET", r.Method)
		}
		if got := r.URL.RequestURI(); got != "/v1/disputes?from_date=2026-03-01&status=needs_response&to_date=2026-03-31" {
			t.Errorf("path = %q", got)
		}
		io.WriteString(w, `{
			"total": 1,
			"items": [`+disputeSummaryExample+`]
		}`)
	})

	page, _, err := c.Disputes.List(context.Background(), ListParams{
		Status:   "needs_response",
		FromDate: "2026-03-01",
		ToDate:   "2026-03-31",
	})
	if err != nil {
		t.Fatalf("List returned error: %v", err)
	}
	if len(page.Items) != 1 {
		t.Fatalf("len(Items) = %d, want 1", len(page.Items))
	}
	d := page.Items[0]
	if d.DisputeID != "dsp_3e7b1c9a2f48" || d.Status != "needs_response" {
		t.Errorf("Items[0] = %+v", d)
	}
	if d.Amount != "75.00" {
		t.Errorf("Amount = %q, want decimal string 75.00", d.Amount)
	}
	if d.ChargeID == nil || *d.ChargeID != "chr_8f3a1c9b4e72" {
		t.Errorf("ChargeID = %v", d.ChargeID)
	}
	if d.ResponseDeadlineAt == nil {
		t.Errorf("ResponseDeadlineAt is nil")
	}
	if page.Pagination.Total != 1 {
		t.Errorf("Pagination.Total = %d, want 1", page.Pagination.Total)
	}
}

func TestGetDispute(t *testing.T) {
	c := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Errorf("method = %s, want GET", r.Method)
		}
		if r.URL.Path != "/v1/disputes/dsp_3e7b1c9a2f48" {
			t.Errorf("path = %q", r.URL.Path)
		}
		io.WriteString(w, `{
			"dispute_id": "dsp_3e7b1c9a2f48",
			"charge_id": "chr_8f3a1c9b4e72",
			"amount": "75.00",
			"currency": "USD",
			"status": "under_review",
			"is_response_editable": false,
			"reason": "fraudulent",
			"response_deadline_at": "2026-03-23T23:59:59.000Z",
			"evidence": {
				"customer_name": "Amara Osei",
				"customer_email_address": "customer@example.com",
				"product_description": "Annual SaaS subscription, plan ID PLAN-PRO-001",
				"service_date": "2026-03-01",
				"notes": "Customer confirmed delivery via email on March 5.",
				"customer_communication_attachment_id": "upl_9f2c7b3d1e45"
			},
			"latest_submission": {
				"submission_id": "dse_a1b2c3d4e5f6",
				"status": "submitted",
				"trigger_source": "merchant_submit",
				"submitted_at": "2026-03-10T09:15:00.000Z",
				"failed_at": null,
				"attempt_sequence": 1
			},
			"created_at": "2026-03-09T08:00:00.000Z",
			"updated_at": "2026-03-10T09:15:00.000Z"
		}`)
	})

	d, _, err := c.Disputes.Get(context.Background(), "dsp_3e7b1c9a2f48")
	if err != nil {
		t.Fatalf("Get returned error: %v", err)
	}
	if d.Status != "under_review" || d.IsResponseEditable {
		t.Errorf("dispute = %+v", d)
	}
	if d.Evidence == nil {
		t.Fatal("Evidence is nil")
	}
	if d.Evidence.CustomerName == nil || *d.Evidence.CustomerName != "Amara Osei" {
		t.Errorf("Evidence.CustomerName = %v", d.Evidence.CustomerName)
	}
	if d.Evidence.CustomerCommunicationAttachmentID == nil || *d.Evidence.CustomerCommunicationAttachmentID != "upl_9f2c7b3d1e45" {
		t.Errorf("attachment = %v", d.Evidence.CustomerCommunicationAttachmentID)
	}
	if d.LatestSubmission == nil {
		t.Fatal("LatestSubmission is nil")
	}
	if d.LatestSubmission.SubmissionID != "dse_a1b2c3d4e5f6" || d.LatestSubmission.AttemptSequence != 1 {
		t.Errorf("LatestSubmission = %+v", d.LatestSubmission)
	}
}

func TestUploadDisputeDocument(t *testing.T) {
	c := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("method = %s, want POST", r.Method)
		}
		if r.URL.Path != "/v1/disputes/uploads" {
			t.Errorf("path = %q", r.URL.Path)
		}
		if ct := r.Header.Get(headerContentType); !strings.HasPrefix(ct, "multipart/form-data") {
			t.Errorf("Content-Type = %q, want multipart/form-data", ct)
		}
		io.WriteString(w, `{
			"document_id": "upl_9f2c7b3d1e45",
			"file_name": "email-screenshot.pdf",
			"mime_type": "application/pdf",
			"file_size_bytes": 204800,
			"storage_provider": "s3",
			"uploaded_at": "2026-03-09T10:00:00.000Z"
		}`)
	})

	doc, _, err := c.Disputes.UploadDocument(context.Background(), "email-screenshot.pdf", strings.NewReader("pdf-bytes"), "dispute-evidence")
	if err != nil {
		t.Fatalf("UploadDocument returned error: %v", err)
	}
	if doc.DocumentID != "upl_9f2c7b3d1e45" || doc.FileName != "email-screenshot.pdf" {
		t.Errorf("doc = %+v", doc)
	}
	if doc.FileSizeBytes != 204800 || doc.StorageProvider != "s3" {
		t.Errorf("doc = %+v", doc)
	}
}

func TestUpdateDisputeEvidence(t *testing.T) {
	c := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPatch {
			t.Errorf("method = %s, want PATCH", r.Method)
		}
		if r.URL.Path != "/v1/disputes/dsp_3e7b1c9a2f48/evidence" {
			t.Errorf("path = %q", r.URL.Path)
		}
		io.WriteString(w, `{
			"dispute_id": "dsp_3e7b1c9a2f48",
			"status": "needs_response",
			"is_response_editable": true,
			"evidence_updated_at": "2026-03-09T11:30:00.000Z"
		}`)
	})

	res, _, err := c.Disputes.UpdateEvidence(context.Background(), "dsp_3e7b1c9a2f48", DisputeEvidenceUpdateRequest{
		CustomerName:                      "Amara Osei",
		CustomerEmailAddress:              "customer@example.com",
		ProductDescription:                "Annual SaaS subscription, plan ID PLAN-PRO-001",
		ServiceDate:                       "2026-03-01",
		Notes:                             "Customer confirmed delivery via email on March 5.",
		CustomerCommunicationAttachmentID: "upl_9f2c7b3d1e45",
	})
	if err != nil {
		t.Fatalf("UpdateEvidence returned error: %v", err)
	}
	if res.DisputeID != "dsp_3e7b1c9a2f48" || !res.IsResponseEditable {
		t.Errorf("res = %+v", res)
	}
	if res.Status != "needs_response" {
		t.Errorf("Status = %q", res.Status)
	}
}

func TestSubmitDispute(t *testing.T) {
	c := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("method = %s, want POST", r.Method)
		}
		if r.URL.Path != "/v1/disputes/dsp_3e7b1c9a2f48/submit" {
			t.Errorf("path = %q", r.URL.Path)
		}
		io.WriteString(w, `{
			"dispute_id": "dsp_3e7b1c9a2f48",
			"status": "under_review",
			"is_response_editable": false,
			"submission": {
				"submission_id": "dse_a1b2c3d4e5f6",
				"submission_status": "submitted",
				"trigger_source": "merchant_submit",
				"submitted_at": "2026-03-10T09:15:00.000Z"
			}
		}`)
	})

	res, _, err := c.Disputes.Submit(context.Background(), "dsp_3e7b1c9a2f48")
	if err != nil {
		t.Fatalf("Submit returned error: %v", err)
	}
	if res.Status != "under_review" || res.IsResponseEditable {
		t.Errorf("res = %+v", res)
	}
	if res.Submission.SubmissionID != "dse_a1b2c3d4e5f6" || res.Submission.SubmissionStatus != "submitted" {
		t.Errorf("Submission = %+v", res.Submission)
	}
	if res.Submission.SubmittedAt == nil {
		t.Error("SubmittedAt is nil")
	}
}
