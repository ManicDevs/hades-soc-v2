# Makefile for common developer tasks

.PHONY: fmt lint test web-test ci precommit-install

fmt:
	gofmt -s -w ./cmd ./internal ./pkg

lint:
	@echo "Running go vet..."
	go vet ./cmd/... ./internal/... ./pkg/... || true
	@echo "Running frontend lint (eslint)..."
	npm --prefix web run lint || true

test:
	@echo "Running Go unit tests..."
	go test -v -race ./cmd/... ./internal/... ./pkg/...

web-test:
	@echo "Running frontend unit tests (vitest)..."
	npm --prefix web ci
	npm --prefix web run test:unit

ci:
	@echo "Running CI checks (gofmt, go vet, unit tests, frontend checks)"
	gofmt -s -l cmd internal pkg || true
	go vet ./cmd/... ./internal/... ./pkg/... || true
	go test -v -race ./cmd/... ./internal/... ./pkg/... || true
	if command -v npm >/dev/null 2>&1; then \
		npm --prefix web ci && npm --prefix web run lint || true; \
		npm --prefix web run test:unit || true; \
	else \
		echo "npm not found, skipping frontend checks"; \
	fi

precommit-install:
	@echo "Configuring repository to use .githooks for local hooks (run once)"
	git config core.hooksPath .githooks
	@echo "Done. The pre-commit hook will now run on commits."
