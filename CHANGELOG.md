# Changelog

All notable changes to this project are documented in this file. The format
is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/), and
this project adheres to [Semantic Versioning](https://semver.org/).

## [v1.3.0] - 2026-10-07

Catches up with the API changelog through the Week of Sep 21, 2026
(virtual accounts). The Jul 27 additions (`billing_address`, the new
dispute evidence fields) were already covered.

### Added

- **Virtual accounts** — new `VirtualAccounts` service: `Create`
  (`POST /v1/virtual-accounts`, idempotent — creating twice for the same
  currency returns the same number) and `Get`
  (`GET /v1/virtual-accounts?currency=NGN`, defaults to NGN). Pass
  `WithConnectedAccount` to issue or read a connected account's number.
  Requires the `virtual_accounts:read` / `virtual_accounts:write` scopes.
- **Guest checkout** — `CreateCheckoutSessionRequest.Customer` is now
  optional (`*CheckoutCustomer` with `omitempty`); omit it and the hosted
  page collects the buyer's email and name. New `CustomerCreation`
  field (`CustomerCreationIfRequired` / `CustomerCreationAlways`) and
  `CustomerDetails` (`{email, name}`) on the checkout session object.
- **Per-method currency restrictions** — new
  `CreateCheckoutSessionRequest.PaymentMethodOptions`
  (`card`, `bank_transfer`, `mobile_money`, `crypto`, each with optional
  `currencies`); `AllowedPaymentMethodTypes` remains for coarse
  whole-method restriction.
- **Payer details on payments** — new `Payment.PaymentMethodDetails`
  (`type` plus `bank_transfer` sender/session/virtual-account details),
  returned by `Payments.Get` and `collection.succeeded` for methods that
  provide payer details.
- **Connect webhook events** — new `EventTypeAccountUpdated`
  (`account.updated`), `EventTypeCapabilityUpdated`
  (`capability.updated`), and `EventTypeTransferCreated`
  (`transfer.created`) constants.
- **Checkout error codes** — new exported constants
  `ErrCodeAccountNotActivated`, `ErrCodeAccountPaymentMethodsRestricted`,
  `ErrCodeCheckoutHasNoPaymentMethod`,
  `ErrCodeCheckoutRestrictionLeavesNoPaymentMethod`, and
  `ErrCodePaymentMethodNotAllowed` for branching on `APIError.Code`.

### Added (additive batch)

- **Payout schedules** — `Misc.GetBalanceSettings`
  (`GET /v1/balance_settings`) and `Misc.UpdateBalanceSettings`
  (`POST /v1/balance_settings`): per-currency automatic payouts
  (`manual`/`instant`/`daily`/`weekly`/`monthly` via the
  `PayoutScheduleInterval*` constants), with `WithConnectedAccount`
  support for connected accounts.
- **Payout destination read** — `Payouts.GetDestination`
  (`GET /v1/payouts/destinations/{destination_id}`).
- **Persons** — new `Persons` service for the account identity
  subresource: `List`/`Create`/`Get`/`Update`/`Delete` plus
  `AttachDocument` for ID documents
  (`POST /v1/accounts/{account_id}/persons…`), with representative /
  owner / director roles and a `pending`/`passed`/`failed` verification
  state.
- **Platform fees** — new `PlatformFees` service: `List` (filterable with
  the new `ListParams.Charge`) and `Get` (`GET /v1/platform_fees…`).
- **Connect splits on checkout** — `CreateCheckoutSessionRequest` gained
  `PlatformFee` (fee-first) and `TransferData` (share-first); the
  checkout and payment objects echo back `PlatformFee` /
  `DestinationAmount`.
- **Corridor allowlist on checkout** — `CreateCheckoutSessionRequest`
  gained `PaymentMethodTypes` with `Corridor*` constants (`USD_CARD`,
  `NGN_BANK_TRANSFER`, `MOMO_*`, `CRYPTO`); `AllowedPaymentMethodTypes`
  is unchanged.
- **Connect error codes** — `ErrCodeInvalidPlatformFee` and
  `ErrCodeContradictoryChargeType`.

### Added (sandbox-verified batch)

Every shape below was confirmed against the live sandbox on 2026-10-07.

- **Reference data** — new `Reference` service: `ListBanks`
  (`GET /v1/reference/banks?country=NG` → `{country, banks[]}`),
  `ListMobileMoneyProviders` (`GET /v1/reference/momo` →
  `{country, providers[]}`), `ListBusinessStructures`
  (`{country, structures[]}` with value/label/description), 
  `ListProductCategories` (`{categories[], sections[]}`), and
  `ResolveBankAccount` (`POST /v1/misc/bank-accounts/resolve`).
- **Product groups** — new `ProductGroups` service: `Create`/`List`/`Get`/
  `Update`/`Delete` on `/v1/product-groups`. A group carries the resolved
  product records; deleting one leaves the products untouched.
- **Balance details** — `BalanceBucket` gained `HeldForDisputes`, and
  `ConnectedAccount` gained `EntityType`, `Configuration`, and
  `Responsibilities`, all observed on live responses.
- **Path corrections** — `Organizations` now calls `/v1/accounts/me`,
  `/v1/accounts/{id}`, and `/v1/accounts/checkout/settings`,
  `Misc.GetBalances` calls `/v1/balances`, and `ConnectedAccounts`
  `Get`/`List`/`ListCapabilities` call `/v1/accounts…`. The old
  `/organizations/*`, `/accounts/balances`, and `/connected-accounts/*`
  paths return `404` on the live API, so the previous paths never
  worked — this is a fix, not a migration. The remaining
  `ConnectedAccounts` write and requirement-task helpers keep their
  legacy paths pending a probe with the `connect` capability.

### Deferred (deliberately)

- The remaining `ConnectedAccounts` write and requirement-task helpers
  (`Create`, `RequestCapabilities`, account links, documents, tasks,
  reusable identity, bank/momo lookups, uploads) keep their legacy paths:
  verifying them needs the `connect` capability or a submitted fixture
  the probe did not cover. The payout withdrawal create
  (`POST /payouts/withdrawals` vs `POST /v1/payouts`) is likewise
  unchanged — firing it moves money. They are next once probed.

## [v1.2.0] - 2026-08-11

Hardening release: the full API surface is now under table-driven error-branch
coverage (99.6% of statements in the main package), CI enforces that floor and
fuzzes the webhook verifier, and the repo is set up for external contributors.

### Added

- **Contribution scaffolding** — rewritten README with badges, install
  instructions, and a per-resource tour; `CONTRIBUTING.md` documenting the
  design  constraints and test expectations; `SECURITY.md`; `CODE_OF_CONDUCT.md`;
  an Apache-2.0 `LICENSE` (with its explicit patent grant); and a CI
  workflow running build, vet, staticcheck, gofmt, and tests across Go
  1.21–1.24.
- **CI coverage gate** — a dedicated job fails the build if main-package
  statement coverage drops below 95% (currently 99.6%).
- **CI fuzz smoke** — a job fuzzes `webhook.ConstructEvent` for 20s on every
  push and PR, on top of the seed-corpus runs in the build matrix.
- **Race detector** — `go test -race ./...` in the CI matrix on every Go
  version, with the local command documented in `CONTRIBUTING.md`.
- **Webhook fuzz test** — `FuzzConstructEvent` feeds random raw bodies,
  signature/timestamp headers, secrets, and tolerances into the verifier,
  asserting it never panics and that any success implies a genuine
  signature under the documented sentinel-error classification.

### Changed

- The test suite now covers the error branch of every service method via a
  table-driven test, the pipeline's defensive branches (non-JSON and
  truncated error bodies, unencodable bodies, invalid base URLs), and the
  idempotent retry-after-error flow. `go test -cover ./` and a fuzz smoke
  run are documented in `CONTRIBUTING.md`.

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
