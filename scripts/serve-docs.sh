#!/usr/bin/env bash
# serve-docs.sh serves the generated docs site locally and opens it in the
# default browser. Generate the site first with gen-docs.sh.
#
# Usage:
#   ./scripts/gen-docs.sh      # once
#   ./scripts/serve-docs.sh    # serves http://127.0.0.1:56789
#   DOCS_PORT=9000 ./scripts/serve-docs.sh
set -euo pipefail
cd "$(dirname "$0")/.."

if [ ! -f docs/index.html ]; then
  echo "docs/index.html not found — run ./scripts/gen-docs.sh first." >&2
  exit 1
fi

GOLDS_VERSION="v0.8.7"
DOCS_PORT="${DOCS_PORT:-56789}"

# golds can serve a directory of static files itself, so no separate web
# server is needed. It opens the browser automatically.
go run "go101.org/golds@${GOLDS_VERSION}" -dir=docs -port="$DOCS_PORT"
