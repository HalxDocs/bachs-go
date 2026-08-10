# Security Policy

## Supported versions

Security fixes are released for the latest tagged version. Older releases
are supported on a best-effort basis.

| Version | Supported          |
| ------- | ------------------ |
| latest  | :white_check_mark: |
| < latest | :warning: best-effort |

## Reporting a vulnerability

**Please do not open a public GitHub issue for security problems.** Report
them privately so they can be fixed before they are disclosed:

- Open a private vulnerability report on GitHub:
  https://github.com/HalxDocs/bachs-go/security/advisories/new
- Or email the maintainers (see the repository's settings for the current
  contact address).

Please include:

- the affected version(s),
- a description of the vulnerability,
- a minimal reproduction, if possible, and
- your suggested impact assessment (what an attacker could do).

You should receive an acknowledgment within 3 business days and an update on
the fix timeline. We ask that you give us reasonable time to release a fix
before public disclosure.

## What this SDK guards against

- **API key leakage** — the SDK never logs, prints, or embeds the API key in
  error messages. Reports of new leak paths are treated as vulnerabilities.
- **Webhook signature verification** — the `webhook` package uses constant-
  time comparison (`hmac.Equal`) and timestamp tolerance to reject forged or
  replayed events. Bugs that weaken this check are security issues.
- **Incorrect idempotency handling** — idempotency keys are POST-only; a
  change that could double-charge customers via retries is treated seriously.
