## What does this change?

<!-- One or two sentences: what the PR does and why. -->

## What's the API source for this change?

<!-- Every endpoint change must cite its source: the Bachs docs page URL or
the OpenAPI operation (https://docs.bachs.io/docs/openapi/openapi.json).
If a field name, path, or status code was not documented, say so and
describe the conservative choice made. -->

## Checks

- [ ] `go build ./...` passes
- [ ] `go vet ./...` passes
- [ ] `staticcheck ./...` passes
- [ ] `gofmt -l .` prints nothing
- [ ] `go test ./...` passes
- [ ] New endpoints have httptest coverage using the exact example payload
      from the cited docs

## Notes for reviewers

<!-- Anything the reviewer should know: tradeoffs, open questions,
follow-up work. -->
