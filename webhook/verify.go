package webhook

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"time"
)

// Signature verification errors. Callers can branch on these to decide how to
// respond (for example, a 401 for a bad signature versus a 200 for a stale
// replay).
var (
	// ErrInvalidTimestamp means the X-Bachs-Timestamp header was not a unix
	// timestamp in seconds.
	ErrInvalidTimestamp = errors.New("webhook: invalid X-Bachs-Timestamp header")

	// ErrTimestampTooOld means the delivery is outside the accepted tolerance
	// and is treated as a replay rather than a live event.
	ErrTimestampTooOld = errors.New("webhook: event timestamp outside tolerance")

	// ErrSignatureMismatch means the computed HMAC does not match the
	// X-Bachs-Signature header; the delivery did not come from Bachs (or the
	// body was tampered with).
	ErrSignatureMismatch = errors.New("webhook: signature verification failed")
)

// ConstructEvent verifies a webhook delivery and decodes it into an Event.
//
// rawBody must be the untouched raw request body — read the body before any
// JSON parsing and pass those exact bytes. sigHeader is the X-Bachs-Signature
// header value (hex HMAC-SHA256 of "{timestamp}.{raw_body}"), tsHeader is the
// X-Bachs-Timestamp header value (unix seconds), and secret is the endpoint's
// signing secret.
//
// tolerance bounds how old a delivery may be before it is rejected as a
// replay; pass something like 5 * time.Minute.
func ConstructEvent(rawBody []byte, sigHeader, tsHeader, secret string, tolerance time.Duration) (*Event, error) {
	ts, err := strconv.ParseInt(tsHeader, 10, 64)
	if err != nil {
		return nil, fmt.Errorf("%w: %q is not a unix timestamp in seconds", ErrInvalidTimestamp, tsHeader)
	}

	if now := time.Now().Unix(); now-ts > int64(tolerance.Seconds()) || ts-now > int64(tolerance.Seconds()) {
		return nil, fmt.Errorf("%w: header timestamp %d, now %d", ErrTimestampTooOld, ts, now)
	}

	mac := hmac.New(sha256.New, []byte(secret))
	_, _ = mac.Write([]byte(fmt.Sprintf("%d.%s", ts, rawBody)))
	expected := hex.EncodeToString(mac.Sum(nil))

	if !hmac.Equal([]byte(expected), []byte(sigHeader)) {
		return nil, ErrSignatureMismatch
	}

	var event Event
	if err := json.Unmarshal(rawBody, &event); err != nil {
		return nil, fmt.Errorf("webhook: decode event: %w", err)
	}
	return &event, nil
}
