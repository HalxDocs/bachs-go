package bachs

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"testing"
)

const personExample = `{
	"id": "per_3a91c0d7f6e2b8149a05",
	"first_name": "Ada",
	"last_name": "Obi",
	"dob": "1990-04-12",
	"phone": "+2348012345678",
	"email": "ada@example.com",
	"id_number_provided": true,
	"relationship": { "representative": true, "owner": true, "director": false },
	"verification": { "status": "pending", "document_provided": false },
	"created_at": "2026-08-07T09:20:00.000Z",
	"updated_at": "2026-08-07T09:20:00.000Z"
}`

// TestPersonsList uses the example payload from
// https://docs.bachs.io/connect/guides/identity-verification.
func TestPersonsList(t *testing.T) {
	c := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Errorf("method = %s, want GET", r.Method)
		}
		if r.URL.Path != "/v1/accounts/acct_3Wq8ZfT1yHnJ5sVe/persons" {
			t.Errorf("path = %q", r.URL.Path)
		}
		io.WriteString(w, `{"items": [`+personExample+`], "total": 1, "limit": 20, "offset": 0}`)
	})

	page, _, err := c.Persons.List(context.Background(), "acct_3Wq8ZfT1yHnJ5sVe", ListParams{})
	if err != nil {
		t.Fatalf("List returned error: %v", err)
	}
	if len(page.Items) != 1 {
		t.Fatalf("len(Items) = %d, want 1", len(page.Items))
	}
	p := page.Items[0]
	if p.ID != "per_3a91c0d7f6e2b8149a05" {
		t.Errorf("ID = %q", p.ID)
	}
	if !p.Relationship.Representative || !p.Relationship.Owner {
		t.Errorf("Relationship = %+v", p.Relationship)
	}
	if p.Verification.Status != "pending" || p.Verification.DocumentProvided {
		t.Errorf("Verification = %+v", p.Verification)
	}
	if !p.IDNumberProvided {
		t.Error("IDNumberProvided is false, want true")
	}
	if page.Pagination.Total != 1 {
		t.Errorf("Pagination.Total = %d, want 1", page.Pagination.Total)
	}
}

func TestPersonsCreate(t *testing.T) {
	rep := true
	c := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("method = %s, want POST", r.Method)
		}
		if r.URL.Path != "/v1/accounts/acct_3Wq8ZfT1yHnJ5sVe/persons" {
			t.Errorf("path = %q", r.URL.Path)
		}
		body, _ := io.ReadAll(r.Body)
		var got map[string]any
		if err := json.Unmarshal(body, &got); err != nil {
			t.Fatalf("request body is not valid JSON: %v", err)
		}
		if got["first_name"] != "Ada" || got["last_name"] != "Obi" {
			t.Errorf("name = %v/%v", got["first_name"], got["last_name"])
		}
		ids, ok := got["id_numbers"].([]any)
		if !ok || len(ids) != 1 || ids[0].(map[string]any)["type"] != "nin" {
			t.Errorf("id_numbers = %v, want one nin entry", got["id_numbers"])
		}
		io.WriteString(w, personExample)
	})

	p, _, err := c.Persons.Create(context.Background(), "acct_3Wq8ZfT1yHnJ5sVe", CreatePersonRequest{
		FirstName: strptr("Ada"),
		LastName:  strptr("Obi"),
		DOB:       strptr("1990-04-12"),
		Email:     strptr("ada@example.com"),
		Phone:     strptr("+2348012345678"),
		Relationship: &PersonRelationshipRequest{
			Representative: &rep,
			Owner:          &rep,
		},
		IDNumbers: []PersonIDNumberRequest{
			{Type: "nin", Value: "12345678901", IssuingCountry: strptr("NG")},
		},
	})
	if err != nil {
		t.Fatalf("Create returned error: %v", err)
	}
	if p.ID != "per_3a91c0d7f6e2b8149a05" {
		t.Errorf("ID = %q", p.ID)
	}
}

func TestPersonsGet(t *testing.T) {
	c := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v1/accounts/acct_3Wq8ZfT1yHnJ5sVe/persons/per_3a91c0d7f6e2b8149a05" {
			t.Errorf("path = %q", r.URL.Path)
		}
		io.WriteString(w, `{
			"id": "per_3a91c0d7f6e2b8149a05",
			"id_number_provided": true,
			"relationship": { "representative": true, "owner": true, "director": false },
			"verification": { "status": "passed", "document_provided": true, "failure_reason": null },
			"updated_at": "2026-08-07T10:31:12.000Z"
		}`)
	})

	p, _, err := c.Persons.Get(context.Background(), "acct_3Wq8ZfT1yHnJ5sVe", "per_3a91c0d7f6e2b8149a05")
	if err != nil {
		t.Fatalf("Get returned error: %v", err)
	}
	if p.Verification.Status != "passed" {
		t.Errorf("Verification.Status = %q, want passed", p.Verification.Status)
	}
	if p.CreatedAt != nil {
		t.Errorf("CreatedAt = %v, want nil when absent", p.CreatedAt)
	}
}

func TestPersonsUpdate(t *testing.T) {
	c := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("method = %s, want POST", r.Method)
		}
		if r.URL.Path != "/v1/accounts/acct_1/persons/per_1" {
			t.Errorf("path = %q", r.URL.Path)
		}
		body, _ := io.ReadAll(r.Body)
		var got map[string]any
		if err := json.Unmarshal(body, &got); err != nil {
			t.Fatalf("request body is not valid JSON: %v", err)
		}
		if len(got) != 1 || got["phone"] != "+2348098765432" {
			t.Errorf("body = %s, want only the changed phone", body)
		}
		io.WriteString(w, personExample)
	})

	p, _, err := c.Persons.Update(context.Background(), "acct_1", "per_1", UpdatePersonRequest{
		Phone: strptr("+2348098765432"),
	})
	if err != nil {
		t.Fatalf("Update returned error: %v", err)
	}
	if p.ID != "per_3a91c0d7f6e2b8149a05" {
		t.Errorf("ID = %q", p.ID)
	}
}

func TestPersonsAttachDocument(t *testing.T) {
	c := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("method = %s, want POST", r.Method)
		}
		if r.URL.Path != "/v1/accounts/acct_1/persons/per_1/documents" {
			t.Errorf("path = %q", r.URL.Path)
		}
		io.WriteString(w, `{
			"id": "doc_5f0339fac1ae",
			"person": "per_3a91c0d7f6e2b8149a05",
			"document_type": "primary_verification",
			"file_name": "ada-nin.jpg",
			"uploaded_at": "2026-08-07T09:23:00.000Z"
		}`)
	})

	side := "front"
	doc, _, err := c.Persons.AttachDocument(context.Background(), "acct_1", "per_1", AttachPersonDocumentRequest{
		File:     "upl_7c2f9a10bd4e",
		Document: PersonDocumentPrimary,
		Side:     &side,
	})
	if err != nil {
		t.Fatalf("AttachDocument returned error: %v", err)
	}
	if doc.ID != "doc_5f0339fac1ae" || doc.DocumentType != PersonDocumentPrimary {
		t.Errorf("document = %+v", doc)
	}
	if doc.UploadedAt.IsZero() {
		t.Error("UploadedAt is zero")
	}
}

func TestPersonsDelete(t *testing.T) {
	c := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodDelete {
			t.Errorf("method = %s, want DELETE", r.Method)
		}
		if r.URL.Path != "/v1/accounts/acct_1/persons/per_1" {
			t.Errorf("path = %q", r.URL.Path)
		}
		w.WriteHeader(http.StatusNoContent)
	})

	meta, err := c.Persons.Delete(context.Background(), "acct_1", "per_1")
	if err != nil {
		t.Fatalf("Delete returned error: %v", err)
	}
	if meta == nil {
		t.Error("ResponseMeta is nil")
	}
}
