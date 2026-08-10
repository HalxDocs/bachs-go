package webhook

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"testing"
	"time"
)

// sign produces a valid signature for rawBody at ts, exactly like Bachs does:
// hex(HMAC-SHA256(secret, "{ts}.{raw_body}")).
func sign(t *testing.T, secret string, ts int64, rawBody []byte) string {
	t.Helper()
	mac := hmac.New(sha256.New, []byte(secret))
	_, _ = mac.Write([]byte(fmt.Sprintf("%d.%s", ts, rawBody)))
	return hex.EncodeToString(mac.Sum(nil))
}

func TestConstructEventValidSignature(t *testing.T) {
	secret := "whsec_test_secret"
	ts := time.Now().Unix()
	rawBody := []byte(`{"id":"evt_1","type":"collection.succeeded","created_at":"2026-02-22T16:20:00Z","organization_id":"org_1","data":{"payment_id":"pay_1"}}`)
	sig := sign(t, secret, ts, rawBody)

	ev, err := ConstructEvent(rawBody, sig, fmt.Sprintf("%d", ts), secret, 5*time.Minute)
	if err != nil {
		t.Fatalf("ConstructEvent returned error: %v", err)
	}
	if ev.ID != "evt_1" {
		t.Errorf("ID = %q, want evt_1", ev.ID)
	}
	if ev.Type != "collection.succeeded" {
		t.Errorf("Type = %q, want collection.succeeded", ev.Type)
	}
	if ev.OrganizationID != "org_1" {
		t.Errorf("OrganizationID = %q, want org_1", ev.OrganizationID)
	}
	if ev.Account != "" {
		t.Errorf("Account = %q, want empty for a non-Connect event", ev.Account)
	}
	if string(ev.Data) != `{"payment_id":"pay_1"}` {
		t.Errorf("Data = %s, want the raw payload", ev.Data)
	}
}

func TestConstructEventRejectsTamperedBody(t *testing.T) {
	secret := "whsec_test_secret"
	ts := time.Now().Unix()
	rawBody := []byte(`{"id":"evt_1","type":"collection.succeeded","data":{}}`)
	sig := sign(t, secret, ts, rawBody)

	// Same timestamp and signature header, different (tampered) body bytes.
	tampered := []byte(`{"id":"evt_1","type":"collection.failed","data":{}}`)
	_, err := ConstructEvent(tampered, sig, fmt.Sprintf("%d", ts), secret, 5*time.Minute)
	if !errors.Is(err, ErrSignatureMismatch) {
		t.Fatalf("error = %v, want ErrSignatureMismatch", err)
	}
}

func TestConstructEventRejectsWrongSecret(t *testing.T) {
	ts := time.Now().Unix()
	rawBody := []byte(`{"id":"evt_1","type":"collection.succeeded","data":{}}`)
	sig := sign(t, "whsec_correct", ts, rawBody)

	_, err := ConstructEvent(rawBody, sig, fmt.Sprintf("%d", ts), "whsec_wrong", 5*time.Minute)
	if !errors.Is(err, ErrSignatureMismatch) {
		t.Fatalf("error = %v, want ErrSignatureMismatch", err)
	}
}

func TestConstructEventRejectsStaleTimestamp(t *testing.T) {
	secret := "whsec_test_secret"
	old := time.Now().Add(-1 * time.Hour).Unix() // outside the 5-minute tolerance
	rawBody := []byte(`{"id":"evt_1","type":"collection.succeeded","data":{}}`)
	sig := sign(t, secret, old, rawBody)

	_, err := ConstructEvent(rawBody, sig, fmt.Sprintf("%d", old), secret, 5*time.Minute)
	if !errors.Is(err, ErrTimestampTooOld) {
		t.Fatalf("error = %v, want ErrTimestampTooOld", err)
	}
}

func TestConstructEventRejectsFutureTimestamp(t *testing.T) {
	secret := "whsec_test_secret"
	future := time.Now().Add(1 * time.Hour).Unix()
	rawBody := []byte(`{"id":"evt_1","type":"collection.succeeded","data":{}}`)
	sig := sign(t, secret, future, rawBody)

	_, err := ConstructEvent(rawBody, sig, fmt.Sprintf("%d", future), secret, 5*time.Minute)
	if !errors.Is(err, ErrTimestampTooOld) {
		t.Fatalf("error = %v, want ErrTimestampTooOld", err)
	}
}

func TestConstructEventRejectsMalformedTimestamp(t *testing.T) {
	secret := "whsec_test_secret"
	rawBody := []byte(`{"id":"evt_1","type":"collection.succeeded","data":{}}`)

	for _, bad := range []string{"", "not-a-number", "123.45", "0x1F", "   "} {
		_, err := ConstructEvent(rawBody, "whatever", bad, secret, 5*time.Minute)
		if !errors.Is(err, ErrInvalidTimestamp) {
			t.Errorf("tsHeader %q: error = %v, want ErrInvalidTimestamp", bad, err)
		}
	}
}

func TestConstructEventRejectsMalformedSignature(t *testing.T) {
	secret := "whsec_test_secret"
	ts := time.Now().Unix()
	rawBody := []byte(`{"id":"evt_1","type":"collection.succeeded","data":{}}`)

	// A syntactically valid hex string that is not the right digest fails the
	// constant-time comparison.
	_, err := ConstructEvent(rawBody, "0000000000000000000000000000000000000000000000000000000000000000", fmt.Sprintf("%d", ts), secret, 5*time.Minute)
	if !errors.Is(err, ErrSignatureMismatch) {
		t.Fatalf("error = %v, want ErrSignatureMismatch", err)
	}
}

func TestConstructEventConnectEvent(t *testing.T) {
	secret := "whsec_test_secret"
	ts := time.Now().Unix()
	rawBody := []byte(`{
		"id": "evt_connect_1",
		"type": "transfer.created",
		"created_at": "2026-03-01T10:00:00Z",
		"organization_id": "org_platform",
		"account": "org_seller_1",
		"data": {"id": "tr_1", "amount": "100.00"}
	}`)
	sig := sign(t, secret, ts, rawBody)

	ev, err := ConstructEvent(rawBody, sig, fmt.Sprintf("%d", ts), secret, 5*time.Minute)
	if err != nil {
		t.Fatalf("ConstructEvent returned error: %v", err)
	}
	if ev.Account != "org_seller_1" {
		t.Errorf("Account = %q, want org_seller_1 (Connect events carry the account)", ev.Account)
	}
	if ev.Type != "transfer.created" {
		t.Errorf("Type = %q, want transfer.created", ev.Type)
	}
	var data struct {
		ID     string `json:"id"`
		Amount string `json:"amount"`
	}
	if err := json.Unmarshal(ev.Data, &data); err != nil {
		t.Fatalf("unmarshal Data: %v", err)
	}
	if data.Amount != "100.00" {
		t.Errorf("data.Amount = %q, want 100.00", data.Amount)
	}
}

func TestConstructEventRejectsNonJSONBody(t *testing.T) {
	secret := "whsec_test_secret"
	ts := time.Now().Unix()
	rawBody := []byte(`this is not json`)
	sig := sign(t, secret, ts, rawBody)

	_, err := ConstructEvent(rawBody, sig, fmt.Sprintf("%d", ts), secret, 5*time.Minute)
	if err == nil {
		t.Fatal("ConstructEvent returned nil error for a non-JSON body")
	}
	if errors.Is(err, ErrSignatureMismatch) {
		t.Errorf("error = %v, want a JSON decode error, not a signature error", err)
	}
}

func TestConstructEventCreatedAtParsing(t *testing.T) {
	// Bachs sends timestamps with microsecond precision and a numeric offset.
	secret := "whsec_test_secret"
	ts := time.Now().Unix()
	rawBody := []byte(`{"id":"evt_1","type":"collection.succeeded","created_at":"2026-02-22T16:20:00.123456+00:00","organization_id":"org_1","data":{}}`)
	sig := sign(t, secret, ts, rawBody)

	ev, err := ConstructEvent(rawBody, sig, fmt.Sprintf("%d", ts), secret, 5*time.Minute)
	if err != nil {
		t.Fatalf("ConstructEvent returned error: %v", err)
	}
	want := time.Date(2026, 2, 22, 16, 20, 0, 123456000, time.UTC)
	if !ev.CreatedAt.Equal(want) {
		t.Errorf("CreatedAt = %v, want %v", ev.CreatedAt, want)
	}
}
