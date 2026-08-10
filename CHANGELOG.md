# Changelog

All notable changes to this project are documented in this file. The format
is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/), and
this project adheres to [Semantic Versioning](https://semver.org/).

## [Unreleased]

### Added

- **Contribution scaffolding** — rewritten README with badges, install
  instructions, and a per-resource tour; `CONTRIBUTING.md` documenting the
  design constraints and test expectations; `SECURITY.md`; `CODE_OF_CONDUCT.md`;
  MIT `LICENSE`; and a CI workflow running build, vet, staticcheck, gofmt,
  and tests across Go 1.21–1.24.

## [v1.1.0] - 2026-08-11

Adds the resource groups that were out of scope for v1.0: payouts, disputes,
conversions, organizations, and the webhook management API.

### Added

- **Payouts** — supported currencies, quotes, payout destination CRUD,
  bank-account resolution and bank listings, and withdrawal create/get/list.
- **Disputes** — list/get with status and date filters, multipart document
  uploads for evidence, incremental evidence updates, and the irreversible
  submit action.
- **Conversions** — quoting, executing against a quote ID, and reading
  conversion records with currency and status filters.
- **Organizations** — `GetMe`, `Get` by ID, and checkout-settings
  read/update.
- **Webhook management** — endpoint create/list/get/update/delete, signing
  secret read and rotation, delivery metrics, per-endpoint and org-wide event
  listing with payloads and attempt history, resending a delivery, and
  replaying an event. Event types are exposed as typed `EventType*`
  constants.
- **Offline API reference** — `scripts/gen-docs.sh` renders the module's doc
  comments into a static pkg.go.dev-style site in `docs/`, served locally
  with `scripts/serve-docs.sh`.

### Changed

- `ListParams` gained `FromDate`/`ToDate`, `StartDate`/`EndDate`, and
  `FromCurrency`/`ToCurrency` filters used by the disputes and conversions
  list endpoints.

## [v1.0.0] - 2026-08-10

Initial release of the Bachs Go SDK, covering the full v1 API surface:

### Added

- **Foundation** — `NewClient` with `WithBaseURL`, `WithSandbox`,
  `WithProduction`, and `WithHTTPClient` options; defaults to the sandbox
  environment; context-first request pipeline with `Authorization: Bearer`
  auth, `ResponseMeta` (request ID + rate-limit headers), typed `APIError`
  with per-field validation errors, and POST-only idempotency keys enforced
  in the request builder.
- **Checkouts** — create and retrieve checkout sessions.
- **Products** — create, get, list, update, archive, and unarchive products.
- **Customers** — create, get, list, and update customers.
- **Payments** — get and list payments.
- **Refunds** — create, get, get-by-charge, and list refunds.
- **Subscriptions** — get, list, update (exactly-one-intent enforced
  client-side), and cancel. Deliberately no `Create` — subscriptions are only
  created by completing a checkout session for a recurring product.
- **Transfers** — create, get, and list transfers.
- **Balances** — get account balances (Misc).
- **Media** — upload, get, and delete media.
- **Customer sessions** — create portal sessions.
- **Connected accounts** — create/get/list accounts, request capabilities,
  account links, task checklists/values/submission, reusable identity,
  bank and mobile-money lookups, document uploads.
- **Misc** — payment methods, payment rails, supported currencies, and
  payout-supported currencies.
- **Webhooks** — `webhook.ConstructEvent` with HMAC-SHA256 signature
  verification, timestamp tolerance, and constant-time comparison.
- **Examples** — `examples/checkout` (product → checkout session → URL) and
  `examples/webhook-server` (raw-body verification + event dispatch).

### Constraints honored

- Money is always a decimal string; IDs are opaque strings; timestamps are
  `time.Time`.
- No global mutable state; everything hangs off `*Client`.
- The API key never appears in errors, logs, or debug output.
