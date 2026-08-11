# Contributing to bachs-go

Thanks for considering a contribution! This guide covers how the SDK is put
together and what must be true before a change is merged.

## Development setup

You need Go 1.21 or later. There are no third-party runtime dependencies, so
no dependency installation is needed.

```sh
git clone https://github.com/HalxDocs/bachs-go
cd bachs-go
```

## The checks that must pass

Run all of these before opening a PR. CI runs the same set:

```sh
go build ./...
go vet ./...
staticcheck ./...   # install with: go install honnef.co/go/tools/cmd/staticcheck@latest
go test ./...
go test -race ./... # data races in tests would fail CI
gofmt -l .          # must print nothing
```

The test suite is entirely `httptest`-based — no live API keys are needed.
New endpoints must ship with a test that:

- asserts the exact HTTP method and path (including query parameters), and
- decodes the **exact example payload** from the endpoint's Bachs docs page
  or the OpenAPI spec (`https://docs.bachs.io/docs/openapi/openapi.json`),
  asserting the decoded fields. Payloads in tests are the source of truth for
  wire compatibility — don't invent example responses.

## Design constraints (non-negotiable)

The SDK enforces these everywhere. New code must follow them:

1. **Money is a string.** Amounts are decimal strings (`"29.00"`), never
   `float64` or any numeric type.
2. **IDs are opaque strings.** Never parse or validate prefixes
   (`prod_`, `cust_`, `chk_`, ...). Treat them as opaque.
3. **Timestamps are `time.Time`**, unmarshaled from ISO 8601 UTC strings.
4. **No global mutable state.** No package-level `http.Client`, no singleton
   config. Everything hangs off a `*Client` instance.
5. **Every network-calling method takes `context.Context` first.**
6. **Idempotency-Key is only ever attached to POST requests.** It is
   enforced in the request builder (`request.go`), not by convention.
7. **Never log, print, or include the API key** in any error, panic, or debug
   output.
8. **`SubscriptionService` has no Create method** — subscriptions are created
   only by completing a checkout for a recurring product. If you're tempted
   to add one, stop; it does not exist in the API.
9. **Base URL defaults to sandbox** (`https://sandbox-api.bachs.io`) unless
   the caller explicitly opts into production. Never default to production.
10. **Every exported type and method gets a Go doc comment** starting with the
    identifier name (golint / staticcheck clean).

## Adding an endpoint

The SDK mirrors the Bachs API one service per resource group. To add an
endpoint:

1. **Fetch the endpoint's spec first** — the doc page linked from
   `https://docs.bachs.io/llms.txt`, or the OpenAPI spec. Do not guess field
   names, paths, or status codes. If a detail genuinely isn't documented,
   implement the conservative option and leave a `// TODO(doc-gap):` comment
   explaining what's unconfirmed.
2. Add the method to the existing service file (or create a new one for a
   new resource group, wiring it into `client.go`).
3. Request/response types go in the same file as the service.
4. `List*` methods take `(ctx, ListParams)` and return
   `(*Page[T], *ResponseMeta, error)` — one shared pagination shape, no
   per-resource pagination types.
5. Non-list methods return `(*T, *ResponseMeta, error)`.
6. POSTs that create a resource accept `opts ...RequestOption` so callers can
   attach idempotency keys.
7. Add the httptest coverage described above.

## Project layout

```
client.go            # Client struct, NewClient, environment constants
options.go           # With* client options
request.go           # request pipeline, RequestOptions, ResponseMeta
errors.go            # APIError, FieldError
pagination.go        # Page[T], Pagination, ListParams
checkout.go          # one file per resource group:
product.go           #   types + *Service methods for that group
...                  #   (customer, payment, refund, subscription, transfer,
                     #    misc, media, customer_session, connected_account,
                     #    payout, dispute, conversion, organization, webhooks)
webhook/             # standalone signature-verification package (no HTTP dep)
examples/            # runnable programs (checkout, webhook-server)
scripts/             # docs generation and serving
docs/                # generated API reference (gitignored)
```

## Commits

Keep commits small and atomic — one logical change per commit, with a message
that says *why*, not just *what*.

## Pull request flow

1. Branch from `main`.
2. Make your change, with tests.
3. Run the checks above; all must pass.
4. Open a PR against `main` describing the change and linking the Bachs doc
   page or OpenAPI operation you implemented against.
5. A maintainer will review. Keep the diff focused — separate refactors from
   feature work in different PRs.

## Reporting issues

- **Security vulnerabilities**: report privately — see [SECURITY.md](SECURITY.md).
  Never open a public issue for a security problem.
- **Bugs and feature requests**: open a GitHub issue with the reproduction
  (for bugs) or the endpoint doc URL you want covered (for features).

## Code of conduct

All contributors are expected to follow the
[Code of Conduct](CODE_OF_CONDUCT.md).
