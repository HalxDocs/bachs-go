# bachs-go

An idiomatic Go SDK for the [Bachs payments API](https://docs.bachs.io).

```go
import "github.com/HalxDocs/bachs-go"
```

## Sandbox setup

1. Grab a sandbox API key from the Bachs dashboard (`sk_sandbox_...`).
2. Point the SDK at the sandbox environment — this is the default, so you can
   skip it, or be explicit:

```go
client, err := bachs.NewClient("sk_sandbox_...", bachs.WithSandbox())
```

Every call takes a `context.Context` as its first argument. Money amounts are
decimal strings (`"29.00"`), IDs are opaque strings, and timestamps are
`time.Time`.

## Authentication

Pass your API key to `NewClient`. The SDK sends it as
`Authorization: Bearer <key>` on every request. To move to production, opt in
explicitly — the SDK never defaults to production:

```go
client, err := bachs.NewClient("sk_live_...", bachs.WithProduction())
```

## Checkouts

Create a product, then complete a checkout session to collect a payment:

```go
product, _, err := client.Products.Create(ctx, bachs.CreateProductRequest{
    Name: "Premium plan",
    Price: bachs.ProductPrice{
        Currency:  "USD",
        PriceType: "fixed",
        Amount:    "29.00",
    },
})

session, _, err := client.Checkouts.Create(ctx, bachs.CreateCheckoutSessionRequest{
    Customer: bachs.CheckoutCustomer{Email: "customer@example.com"},
    ProductCart: []bachs.ProductItemRequest{
        {ProductID: product.ID, Quantity: 1},
    },
    SuccessURL: "https://example.com/success",
    CancelURL:  "https://example.com/cancel",
})
// redirect the customer to session.CheckoutURL
```

A full runnable example lives in [`examples/checkout`](examples/checkout).

## Products & customers

```go
product, _, _ := client.Products.Get(ctx, "prod_...")
page, _, _ := client.Products.List(ctx, bachs.ListParams{Limit: 20})
product, _, _ = client.Products.Update(ctx, "prod_...", bachs.UpdateProductRequest{Name: "New name"})
product, _, _ = client.Products.Archive(ctx, "prod_...")
product, _, _ = client.Products.Unarchive(ctx, "prod_...")

customer, _, _ := client.Customers.Create(ctx, bachs.CreateCustomerRequest{Email: "ada@example.com"})
customer, _, _ = client.Customers.Get(ctx, "cust_...")
page, _, _ = client.Customers.List(ctx, bachs.ListParams{})
customer, _, _ = client.Customers.Update(ctx, "cust_...", bachs.UpdateCustomerRequest{Email: "new@example.com"})
```

## Payments & refunds

```go
payment, _, _ := client.Payments.Get(ctx, "pay_...")
page, _, _ := client.Payments.List(ctx, bachs.ListParams{})

refund, _, _ := client.Refunds.Create(ctx, bachs.CreateRefundRequest{
    ChargeID:  "pay_...",
    Reference: "refund-123",
})
refund, _, _ = client.Refunds.Get(ctx, "ref_...")
refund, _, _ = client.Refunds.GetByCharge(ctx, "pay_...")
page, _, _ = client.Refunds.List(ctx, bachs.ListParams{})
```

## Subscriptions

> **Subscriptions have no Create method.** Subscriptions are created only by
> completing a checkout session for a recurring product — see the
> [subscriptions guide](https://docs.bachs.io/guides/subscriptions). There is
> no `POST /subscriptions` endpoint, so the SDK deliberately omits
> `Subscriptions.Create`.

```go
sub, _, _ := client.Subscriptions.Get(ctx, "sub_...")
page, _, _ := client.Subscriptions.List(ctx, bachs.ListParams{})
// Move the subscription to another plan, extend its trial, swap its payment
// method, or update metadata — exactly one intent per call (enforced
// client-side):
sub, _, _ = client.Subscriptions.Update(ctx, "sub_...", bachs.UpdateSubscriptionRequest{
    ProductID: "prod_...",
})
// Cancel at the current period end, or immediately:
sub, _, _ = client.Subscriptions.Cancel(ctx, "sub_...", bachs.CancelSubscriptionRequest{
    CancelAtPeriodEnd: true,
})
```

## Transfers & balances

```go
transfer, _, _ := client.Transfers.Create(ctx, bachs.CreateTransferRequest{...})
transfer, _, _ = client.Transfers.Get(ctx, "trf_...")
page, _, _ := client.Transfers.List(ctx, bachs.ListParams{})

balances, _, _ := client.Misc.GetBalances(ctx)
```

## Customer sessions & connected accounts

```go
// A portal session lets a customer manage their own subscriptions and cards:
session, _, _ := client.CustomerSessions.Create(ctx, "cust_...")

account, _, _ := client.ConnectedAccounts.Create(ctx, bachs.CreateConnectedAccountRequest{...})
link, _, _ := client.ConnectedAccounts.CreateAccountLink(ctx, account.ID, bachs.CreateAccountLinkRequest{...})
```

## Media & misc

```go
// Scope is a logical grouping label, e.g. "product-media":
upload, _, _ := client.Media.Upload(ctx, "hero.png", file, "product-media")
media, _, _ := client.Media.Get(ctx, upload.UploadID)
_, _, _ = client.Media.Delete(ctx, upload.UploadID)

methods, _, _ := client.Misc.ListPaymentMethods(ctx)
rails, _, _ := client.Misc.ListPaymentRails(ctx, "BANK_TRANSFER", "NGN", "")
currencies, _, _ := client.Misc.ListSupportedCurrencies(ctx)
payoutCurrencies, _, _ := client.Misc.ListPayoutSupportedCurrencies(ctx)
```

## Payouts

Withdraw funds from your balance to a bank account, mobile money, or crypto
wallet:

```go
// Quote first to see fees and the exchange rate, then create the withdrawal:
quote, _, _ := client.Payouts.CreateQuote(ctx, bachs.CreatePayoutQuoteRequest{
    FromCurrency: "USD",
    ToCurrency:   "NGN",
    Amount:       "100.00",
})

destination, _, _ := client.Payouts.CreateDestination(ctx, bachs.CreatePayoutDestinationRequest{
    DestinationType: "bank_account",
    Currency:        "NGN",
    AccountNumber:   "0123456789",
    AccountName:     "John Doe",
    BankCode:        "058",
})

withdrawal, _, _ := client.Payouts.CreateWithdrawal(ctx, bachs.CreateWithdrawalRequest{
    FromCurrency:       "USD",
    ToCurrency:         "NGN",
    Amount:             "100.00",
    PaymentMethod:      "BANK_TRANSFER",
    Reference:          "WD-001",
    Email:              "ops@example.com",
    PayoutDestinationID: destination.ID,
})

payout, _, _ := client.Payouts.Get(ctx, withdrawal.WithdrawalID)
page, _, _ := client.Payouts.List(ctx, bachs.ListParams{StatusFilter: "processing"})

banks, _, _ := client.Payouts.ListBanks(ctx, "NG")
resolved, _, _ := client.Payouts.ResolveBankAccount(ctx, "058", "0123456789")
```

## Disputes

Respond to chargebacks before the deadline:

```go
page, _, _ := client.Disputes.List(ctx, bachs.ListParams{Status: "needs_response"})

// Upload a supporting document, attach it as evidence, then submit:
doc, _, _ := client.Disputes.UploadDocument(ctx, "email-screenshot.pdf", file, "dispute-evidence")
_, _, _ = client.Disputes.UpdateEvidence(ctx, "dsp_...", bachs.DisputeEvidenceUpdateRequest{
    CustomerName:                     "Amara Osei",
    CustomerCommunicationAttachmentID: doc.DocumentID,
})
submitted, _, _ := client.Disputes.Submit(ctx, "dsp_...") // irreversible
```

## Conversions

Convert between settlement currencies, quoting first to lock in the rate:

```go
quote, _, _ := client.Conversions.CreateQuote(ctx, bachs.CreateConversionQuoteRequest{
    FromCurrency: "USD",
    ToCurrency:   "NGN",
    Amount:       "1000.00",
})
conversion, _, _ := client.Conversions.Create(ctx, bachs.CreateConversionRequest{
    FromCurrency: "USD",
    ToCurrency:   "NGN",
    Amount:       "1000.00",
    QuoteID:      quote.QuoteID,
})
page, _, _ := client.Conversions.List(ctx, bachs.ListParams{Status: "completed"})
```

## Organizations

Read your own organization and manage its checkout configuration:

```go
me, _, _ := client.Organizations.GetMe(ctx) // capabilities/requirements left null
org, _, _ := client.Organizations.Get(ctx, "org_...") // populated blocks
settings, _, _ := client.Organizations.GetCheckoutSettings(ctx)
settings, _, _ = client.Organizations.UpdateCheckoutSettings(ctx, bachs.UpdateCheckoutSettingsRequest{
    FeePreference: "org_pays",
})
```

## Webhook management

Register delivery endpoints, monitor delivery health, and replay missed
events. Requires `webhooks:read` / `webhooks:write` scopes:

```go
endpoint, _, _ := client.Webhooks.CreateEndpoint(ctx, bachs.CreateWebhookEndpointRequest{
    Name:       "Production events",
    URL:        "https://api.example.com/webhooks/bachs",
    EventTypes: []string{bachs.EventTypeCollectionSucceeded, bachs.EventTypeCollectionFailed},
})
// The signing secret is returned only once, at creation — store it.

metrics, _, _ := client.Webhooks.GetEndpointMetrics(ctx, endpoint.EndpointID, bachs.EndpointMetricsParams{
    Period: "day",
})
page, _, _ := client.Webhooks.ListEvents(ctx, bachs.ListParams{})

// Redeliver a missed event to one endpoint, or replay it across the org:
_, _, _ = client.Webhooks.ResendEndpointEvent(ctx, endpoint.EndpointID, "evt_...")
_, _, _ = client.Webhooks.Replay(ctx, bachs.ReplayWebhookEventRequest{EventID: "evt_..."})
```

## Webhooks

Verify webhook signatures with the `webhook` package — pass the **untouched
raw request body**:

```go
event, err := webhook.ConstructEvent(rawBody, sigHeader, tsHeader, secret, 5*time.Minute)
switch event.Type {
case "checkout.session.completed":
    // fulfill the order
}
```

A runnable webhook server lives in [`examples/webhook-server`](examples/webhook-server).

## Errors

Non-2xx responses surface as `*bachs.APIError` with the status code, error
code, detail, doc URL, per-field validation errors, and the request ID:

```go
if err != nil {
    var apiErr *bachs.APIError
    if errors.As(err, &apiErr) {
        fmt.Println(apiErr.Code, apiErr.Detail)
    }
}
```

## Pagination

Every `List*` method takes `bachs.ListParams` and returns a `*bachs.Page[T]`.
`Cursor` takes precedence over `Offset` when both are set:

```go
page, _, err := client.Products.List(ctx, bachs.ListParams{Limit: 50})
for page.Pagination.HasMore {
    page, _, err = client.Products.List(ctx, bachs.ListParams{Limit: 50, Cursor: page.Pagination.NextCursor})
}
```

## Idempotency

Attach an idempotency key to POSTs so retries don't double-charge. The SDK
enforces that keys are only sent on POST requests:

```go
session, _, err := client.Checkouts.Create(ctx, bachs.CreateCheckoutSessionRequest{...},
    bachs.WithIdempotencyKey("session-12345"))
```

## API reference (offline docs)

The full API reference is browsable without pkg.go.dev: `scripts/gen-docs.sh`
generates a static, pkg.go.dev-style site into `docs/` using
[golds](https://go101.org/golds) (pinned to v0.8.7). Because golds renders
the doc comments and source of the module directly, the reference can never
drift from the code.

```sh
./scripts/gen-docs.sh      # generates docs/ (~11 MB, gitignored)
./scripts/serve-docs.sh    # serves it at http://127.0.0.1:56789
```

Or open `docs/index.html` directly, or serve the folder with any static file
server (e.g. `python -m http.server --directory docs`). Set `DOCS_PORT` to
change the serve port.

## Contributing

Run the checks before pushing:

```sh
go build ./...
go vet ./...
staticcheck ./...
go test ./...
```
