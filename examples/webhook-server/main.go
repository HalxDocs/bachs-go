// Command webhook-server is a minimal webhook receiver for Bachs events.
//
// It reads the raw request body before any parsing, verifies the
// X-Bachs-Signature header via webhook.ConstructEvent, and dispatches on
// event.Type. Run it with the endpoint secret from your Bachs dashboard
// (find it under Settings → Webhooks):
//
//	BACHS_WEBHOOK_SECRET=whsec_... go run .
//
// Point your Bachs webhook endpoint at http://localhost:8080/webhooks and
// send a test event. The server responds 200 for valid events and 400 for
// signature or timestamp failures.
package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/HalxDocs/bachs-go/webhook"
)

const webhookSecretEnv = "BACHS_WEBHOOK_SECRET"

// signatureTolerance is how far the webhook timestamp may drift from the
// server clock before the event is rejected.
const signatureTolerance = 5 * time.Minute

// WebhookHandler returns an http.Handler that verifies and dispatches Bachs
// webhook events. The secret must be the raw signing secret from the Bachs
// dashboard.
func WebhookHandler(secret string) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Read the raw body before touching any JSON: ConstructEvent needs
		// the exact bytes that were signed.
		rawBody, err := io.ReadAll(r.Body)
		if err != nil {
			http.Error(w, "failed to read request body", http.StatusBadRequest)
			return
		}

		event, err := webhook.ConstructEvent(
			rawBody,
			r.Header.Get("X-Bachs-Signature"),
			r.Header.Get("X-Bachs-Timestamp"),
			secret,
			signatureTolerance,
		)
		if err != nil {
			// Never log the signature header or body on failure: it is
			// secret material.
			log.Printf("webhook rejected: %v", err)
			http.Error(w, "invalid signature", http.StatusBadRequest)
			return
		}

		// Dispatch on the event type. The event data is an opaque JSON
		// object; decode the fields you care about here.
		switch event.Type {
		case "checkout.session.completed":
			var data struct {
				ID         string `json:"id"`
				PaymentID  string `json:"payment_id"`
				CustomerID string `json:"customer_id"`
			}
			if err := json.Unmarshal(event.Data, &data); err != nil {
				log.Printf("checkout.session.completed: failed to decode data: %v", err)
			} else {
				log.Printf("checkout %s completed (payment %s, customer %s)", data.ID, data.PaymentID, data.CustomerID)
			}
		case "payment.succeeded":
			var data struct {
				ID     string `json:"id"`
				Amount string `json:"amount"`
			}
			if err := json.Unmarshal(event.Data, &data); err != nil {
				log.Printf("payment.succeeded: failed to decode data: %v", err)
			} else {
				log.Printf("payment %s succeeded for %s", data.ID, data.Amount)
			}
		default:
			log.Printf("unhandled event type %q (event %s)", event.Type, event.ID)
		}

		w.WriteHeader(http.StatusOK)
		fmt.Fprintln(w, "ok")
	})
}

func main() {
	secret := os.Getenv(webhookSecretEnv)
	if secret == "" {
		log.Fatalf("%s is not set", webhookSecretEnv)
	}

	addr := ":8080"
	mux := http.NewServeMux()
	mux.Handle("/webhooks", WebhookHandler(secret))

	log.Printf("listening on %s, verify signature, then send events to /webhooks", addr)
	if err := http.ListenAndServe(addr, mux); err != nil {
		log.Fatal(err)
	}
}
