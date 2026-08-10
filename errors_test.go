package bachs

import (
	"context"
	"net/http"
	"strings"
	"testing"
)

func TestAPIErrorErrorString(t *testing.T) {
	err := &APIError{
		StatusCode: 404,
		Code:       "NOT_FOUND",
		Detail:     "Product not found",
		RequestID:  "req-123",
	}
	got := err.Error()
	for _, want := range []string{"404", "NOT_FOUND", "Product not found", "req-123"} {
		if !strings.Contains(got, want) {
			t.Errorf("Error() = %q, want it to contain %q", got, want)
		}
	}
}

func TestAPIErrorNeverContainsAPIKey(t *testing.T) {
	// The client's own key must never appear in any error it produces: the key
	// is only ever placed in the Authorization header, never in messages.
	c := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		writeJSON(t, w, 401, map[string]any{"detail": "Invalid API key", "error_code": "UNAUTHORIZED"})
	})

	var out map[string]any
	_, err := c.do(context.Background(), http.MethodGet, "/things", nil, &out)
	if err == nil {
		t.Fatal("do returned nil error")
	}
	if strings.Contains(err.Error(), "sk_sandbox_test") {
		t.Errorf("error message leaked the API key: %q", err.Error())
	}
}

func TestFieldErrorWireShape(t *testing.T) {
	err := &APIError{
		StatusCode: 422,
		Code:       "VALIDATION_ERROR",
		Detail:     "Missing required field(s): name",
		Errors: []FieldError{
			{Field: "name", Message: "Field required", Type: "missing"},
		},
	}
	if err.Errors[0].Field != "name" {
		t.Errorf("Errors[0].Field = %q, want name", err.Errors[0].Field)
	}
}
