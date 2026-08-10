package bachs

import (
	"context"
	"io"
	"net/http"
	"net/url"
	"time"
)

// MediaService provides methods for uploading files and retrieving or deleting
// the uploads. An upload returns an upload_id that you attach to a product via
// the media field. Source:
// https://docs.bachs.io/api-reference/media/upload-a-file
type MediaService struct {
	service
}

// Upload is metadata for a file uploaded to Bachs. Use UploadID as the value
// of the media field when creating or updating a product.
type Upload struct {
	// UploadID identifies the upload.
	UploadID string `json:"upload_id"`

	// Provider is the storage provider label.
	Provider string `json:"provider"`

	// FileName of the uploaded file.
	FileName string `json:"file_name"`

	// MIMEType detected for the uploaded file.
	MIMEType string `json:"mime_type"`

	// FileSizeBytes is the uploaded file's size in bytes.
	FileSizeBytes int `json:"file_size_bytes"`

	// URL the upload is served from, once linked.
	URL *string `json:"url"`

	// LinkedResourceType is the resource type the upload is attached to (for
	// example "product"), or null while unattached.
	LinkedResourceType *string `json:"linked_resource_type"`

	// LinkedResourceID is the resource the upload is attached to, or null
	// while unattached.
	LinkedResourceID *string `json:"linked_resource_id"`

	// CreatedAt is when the upload was created.
	CreatedAt time.Time `json:"created_at"`

	// UpdatedAt is when the upload was last updated.
	UpdatedAt time.Time `json:"updated_at"`
}

// UploadDeleteResponse is the result of Media.Delete.
type UploadDeleteResponse struct {
	// UploadID of the deleted upload.
	UploadID string `json:"upload_id"`

	// Deleted is true when the upload was removed.
	Deleted bool `json:"deleted"`
}

// Upload uploads a file (maximum 20 MB) and returns the upload metadata,
// including the upload_id to attach to a product. Scope is a logical grouping
// label — use "product-media" for product images — and defaults to "general"
// when empty. The file is streamed as multipart/form-data.
func (s *MediaService) Upload(ctx context.Context, fileName string, file io.Reader, scope string, opts ...RequestOption) (*Upload, *ResponseMeta, error) {
	body, contentType, err := multipartUpload(fileName, file, map[string]string{"scope": scope})
	if err != nil {
		return nil, nil, err
	}

	var out Upload
	meta, err := s.request(ctx, http.MethodPost, "/utilities/uploads", nil, &out, append(opts, withRawBody(body, contentType))...)
	if err != nil {
		return nil, meta, err
	}
	return &out, meta, nil
}

// Get retrieves metadata for a previously created upload by its ID.
func (s *MediaService) Get(ctx context.Context, uploadID string) (*Upload, *ResponseMeta, error) {
	var out Upload
	meta, err := s.request(ctx, http.MethodGet, "/utilities/uploads/"+url.PathEscape(uploadID), nil, &out)
	if err != nil {
		return nil, meta, err
	}
	return &out, meta, nil
}

// Delete removes an upload that has not yet been linked to any resource. The
// API returns a 409 if the upload is already attached to a product.
func (s *MediaService) Delete(ctx context.Context, uploadID string) (*UploadDeleteResponse, *ResponseMeta, error) {
	var out UploadDeleteResponse
	meta, err := s.request(ctx, http.MethodDelete, "/utilities/uploads/"+url.PathEscape(uploadID), nil, &out)
	if err != nil {
		return nil, meta, err
	}
	return &out, meta, nil
}
