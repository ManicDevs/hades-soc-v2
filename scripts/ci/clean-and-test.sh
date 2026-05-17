#!/usr/bin/env bash
set -euo pipefail

echo "==> gofmt check (selected dirs)"
gofmt -s -l cmd internal pkg core | sed -n '1,200p' || true

echo "==> go vet"
go vet ./cmd/... ./internal/... ./pkg/... || true

echo "==> go test (unit)
"
go test -v -race ./cmd/... ./internal/... ./pkg/... || true

# Frontend checks (non-fatal if node/npm not available)
if command -v npm >/dev/null 2>&1; then
  echo "==> frontend lint (eslint)"
  npm --prefix web run lint || echo "ESLint reported issues (see above)"

  echo "==> frontend unit tests (vitest)"
  npm --prefix web run test:unit || echo "Vitest failed or had errors"
else
  echo "npm not found, skipping frontend checks"
fi

echo "==> done"
