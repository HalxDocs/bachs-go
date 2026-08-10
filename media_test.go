package bachs

import (
	"context"
	"io"
	"net/http"
	"strings"
	"testing"
)

// uploadExample is the exact UploadResponse example from
// https://docs.bachs.io/api-reference/media/upload-a-file
const uploadExample = `{
	"upload_id": "upl_4f3e2d1c",
	"provider": "s3",
	"file_name": "product-hero.png",
	"mime_type": "image/png",
	"file_size_bytes": 204800,
	"url": "https://cdn.bachs.io/uploads/upl_4f3e2d1c/product-hero.png",
	"linked_resource_type": null,
	"linked_resource_id": null,
	"created_at": "2026-01-24T12:00:00.000Z",
	"updated_at": "2026-01-24T12:00:00.000Z"
}`

// TestMediaUpload uses the exact example from
// https://docs.bachs.io/api-reference/media/upload-a-file and verifies the
// multipart/form-data body.
func TestMediaUpload(t *testing.T) {
	c := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("method = %s, want POST", r.Method)
		}
		if r.URL.Path != "/v1/utilities/uploads" {
			t.Errorf("path = %q, want /v1/utilities/uploads", r.URL.Path)
		}
		ct := r.Header.Get(headerContentType)
		if !strings.HasPrefix(ct, "multipart/form-data") {
			t.Fatalf("Content-Type = %q, want multipart/form-data", ct)
		}

		mr, err := r.MultipartReader()
		if err != nil {
			t.Fatalf("MultipartReader: %v", err)
		}
		seenScope, seenFile := false, false
		var fileBody strings.Builder
		for {
			part, err := mr.NextPart()
			if err == io.EOF {
				break
			}
			if err != nil {
				t.Fatalf("NextPart: %v", err)
			}
			switch part.FormName() {
			case "scope":
				b, _ := io.ReadAll(part)
				if string(b) != "product-media" {
					t.Errorf("scope = %q, want product-media", b)
				}
				seenScope = true
			case "file":
				if part.FileName() != "product-hero.png" {
					t.Errorf("file name = %q, want product-hero.png", part.FileName())
				}
				io.Copy(&fileBody, part)
				seenFile = true
			}
		}
		if !seenScope || !seenFile {
			t.Errorf("missing parts: scope=%v file=%v", seenScope, seenFile)
		}
		if fileBody.String() != "fake-png-bytes" {
			t.Errorf("file body = %q", fileBody.String())
		}

		io.WriteString(w, uploadExample)
	})

	upload, _, err := c.Media.Upload(context.Background(), "product-hero.png", strings.NewReader("fake-png-bytes"), "product-media")
	if err != nil {
		t.Fatalf("Upload returned error: %v", err)
	}

	if upload.UploadID != "upl_4f3e2d1c" {
		t.Errorf("UploadID = %q", upload.UploadID)
	}
	if upload.FileName != "product-hero.png" || upload.MIMEType != "image/png" {
		t.Errorf("FileName/MIMEType = %s/%s", upload.FileName, upload.MIMEType)
	}
	if upload.FileSizeBytes != 204800 {
		t.Errorf("FileSizeBytes = %d, want 204800", upload.FileSizeBytes)
	}
	if upload.URL == nil || *upload.URL != "https://cdn.bachs.io/uploads/upl_4f3e2d1c/product-hero.png" {
		t.Errorf("URL = %v", upload.URL)
	}
	if upload.LinkedResourceType != nil {
		t.Errorf("LinkedResourceType = %v, want nil while unattached", upload.LinkedResourceType)
	}
}

func TestMediaGet(t *testing.T) {
	c := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Errorf("method = %s, want GET", r.Method)
		}
		if r.URL.Path != "/v1/utilities/uploads/upl_4f3e2d1c" {
			t.Errorf("path = %q", r.URL.Path)
		}
		io.WriteString(w, uploadExample)
	})

	upload, _, err := c.Media.Get(context.Background(), "upl_4f3e2d1c")
	if err != nil {
		t.Fatalf("Get returned error: %v", err)
	}
	if upload.Provider != "s3" {
		t.Errorf("Provider = %q, want s3", upload.Provider)
	}
}

// TestMediaDelete uses the exact example from
// https://docs.bachs.io/api-reference/media/delete-an-upload
func TestMediaDelete(t *testing.T) {
	c := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodDelete {
			t.Errorf("method = %s, want DELETE", r.Method)
		}
		if r.URL.Path != "/v1/utilities/uploads/upl_4f3e2d1c" {
			t.Errorf("path = %q", r.URL.Path)
		}
		io.WriteString(w, `{"upload_id": "upl_4f3e2d1c", "deleted": true}`)
	})

	res, _, err := c.Media.Delete(context.Background(), "upl_4f3e2d1c")
	if err != nil {
		t.Fatalf("Delete returned error: %v", err)
	}
	if res.UploadID != "upl_4f3e2d1c" || !res.Deleted {
		t.Errorf("Delete response = %+v", res)
	}
}
