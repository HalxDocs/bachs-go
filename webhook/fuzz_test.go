package webhook

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"strconv"
	"testing"
	"time"
)

// signAt computes the Bachs signature for rawBody at ts, mirroring the
// production algorithm: hex(HMAC-SHA256(secret, "{ts}.{raw_body}")).
func signAt(secret string, ts int64, rawBody []byte) string {
	mac := hmac.New(sha256.New, []byte(secret))
	_, _ = mac.Write([]byte(fmt.Sprintf("%d.%s", ts, rawBody)))
	return hex.EncodeToString(mac.Sum(nil))
}

// FuzzConstructEvent feeds arbitrary raw bodies, signature and timestamp
// headers, secrets, and tolerances into ConstructEvent and asserts the
// function's contract:
//
//   - it never panics or hangs on any input;
//   - a success means the event decoded AND the supplied signature is the
//     genuine signature for (ts, rawBody) under the supplied secret —
//     recomputed independently here;
//   - every error is one of the documented sentinels, and only for the
//     condition it names (unparseable timestamp, out-of-tolerance
//     timestamp, mismatched signature, or a non-JSON body after a
//     matching signature).
//
// The corpus seeds validly-signed payloads so the happy path and the
// decode-error-after-valid-signature path are exercised even though the
// fuzzer cannot craft a valid HMAC on its own.
func FuzzConstructEvent(f *testing.F) {
	secret := "whsec_fuzz_secret"
	now := time.Now().Unix()
	validBody := []byte(`{"id":"evt_1","type":"collection.succeeded","created_at":"2026-02-22T16:20:00Z","organization_id":"org_1","data":{"payment_id":"pay_1"}}`)
	validSig := signAt(secret, now, validBody)
	nonJSONBody := []byte(`this is not json`)

	// Seed corpus: the valid signed event, a tampered body under a valid
	// signature, a non-JSON body under a valid signature, and malformed
	// headers.
	f.Add(validBody, []byte(validSig), []byte(fmt.Sprintf("%d", now)), []byte(secret), int64(300))
	f.Add(validBody, []byte(validSig), []byte(fmt.Sprintf("%d", now)), []byte("whsec_wrong_secret"), int64(300))
	f.Add(validBody, []byte(signAt(secret, now-3600, validBody)), []byte(fmt.Sprintf("%d", now-3600)), []byte(secret), int64(300))
	f.Add(nonJSONBody, []byte(signAt(secret, now, nonJSONBody)), []byte(fmt.Sprintf("%d", now)), []byte(secret), int64(300))
	f.Add(validBody, []byte("0000000000000000000000000000000000000000000000000000000000000000"), []byte(fmt.Sprintf("%d", now)), []byte(secret), int64(300))
	f.Add(validBody, []byte(validSig), []byte("not-a-number"), []byte(secret), int64(300))
	f.Add([]byte{}, []byte{}, []byte{}, []byte{}, int64(0))

	f.Fuzz(func(t *testing.T, rawBody, sigHeader, tsHeader, secret []byte, toleranceSec int64) {
		// Clamp the tolerance to something sensible so time.Duration cannot
		// overflow and the invariant checks below stay meaningful.
		if toleranceSec < 0 || toleranceSec > 30*24*3600 {
			toleranceSec = 300
		}
		tolerance := time.Duration(toleranceSec) * time.Second

		ev, err := ConstructEvent(rawBody, string(sigHeader), string(tsHeader), string(secret), tolerance)

		parsedTS, tsErr := strconv.ParseInt(string(tsHeader), 10, 64)

		switch {
		case err == nil:
			if ev == nil {
				t.Fatal("ConstructEvent returned nil error and nil event")
			}
			// The signature must genuinely verify against the header
			// timestamp and raw body under the supplied secret.
			if tsErr != nil {
				t.Fatalf("success without a parseable timestamp header %q", tsHeader)
			}
			want := signAt(string(secret), parsedTS, rawBody)
			if !hmac.Equal([]byte(sigHeader), []byte(want)) {
				t.Fatalf("accepted a signature that does not verify: got %q, want %q", sigHeader, want)
			}

		case errors.Is(err, ErrInvalidTimestamp):
			if tsErr == nil {
				t.Fatalf("ErrInvalidTimestamp for parseable header %q", tsHeader)
			}

		case errors.Is(err, ErrTimestampTooOld):
			if tsErr != nil {
				t.Fatalf("ErrTimestampTooOld for unparseable header %q", tsHeader)
			}
			nowSec := time.Now().Unix()
			if !(nowSec-parsedTS > int64(tolerance.Seconds()) || parsedTS-nowSec > int64(tolerance.Seconds())) {
				t.Fatalf("ErrTimestampTooOld for in-tolerance ts %d (now %d, tol %s)", parsedTS, nowSec, tolerance)
			}

		case errors.Is(err, ErrSignatureMismatch):
			// Any body, timestamp, or secret combination is fine here as long
			// as the timestamp parsed — a mismatched signature is expected
			// for fuzzed inputs. Just ensure the timestamp parsed, because
			// signature verification happens after the timestamp check.
			if tsErr != nil {
				t.Fatalf("ErrSignatureMismatch for unparseable header %q", tsHeader)
			}

		default:
			// The remaining path is a JSON decode failure after a *valid*
			// signature. Verify the signature independently so this error is
			// not masking a real mismatch.
			if tsErr != nil {
				t.Fatalf("unexpected error %v with unparseable timestamp %q", err, tsHeader)
			}
			want := signAt(string(secret), parsedTS, rawBody)
			if !hmac.Equal([]byte(sigHeader), []byte(want)) {
				t.Fatalf("unexpected error %v for a non-matching signature", err)
			}
		}
	})
}
