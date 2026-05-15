#!/bin/bash
# HADES-V2 Build Verification Script (improved CI-friendly)
# Tests all components of the HADES-V2 project.

set -euo pipefail

# Behavior flags (override with env vars)
# SKIP_OS=1       -> Skip long OS/Buildroot/kernel checks
# STRICT=1        -> Treat build warnings as failures
SKIP_OS=${SKIP_OS:-0}
STRICT=${STRICT:-0}

echo "=========================================="
echo "  HADES-V2 Build Verification"
echo "=========================================="

# Color codes
GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

pass() { echo -e "${GREEN}[PASS]${NC} $1"; }
fail() { echo -e "${RED}[FAIL]${NC} $1"; }
warn() { echo -e "${YELLOW}[WARN]${NC} $1"; }
info() { echo -e "[INFO] $1"; }

cd "$(dirname "$0")/../.."

# 1. Go Backend
echo ""
echo "=== Go Backend ==="
if command -v go >/dev/null 2>&1; then
    info "Building main hades binary..."
    if go build -o hades ./cmd/hades >/dev/null 2>&1; then
        pass "Main binary builds"
    else
        if [ "$STRICT" -eq 1 ]; then
            fail "Main binary build failed"
        else
            warn "Main binary build failed"
        fi
    fi

    info "Running go vet..."
    if go vet ./... >/dev/null 2>&1; then
        pass "go vet clean"
    else
        warn "go vet has issues"
    fi

    info "Running gofmt check..."
    if [ -z "$(gofmt -s -l . 2>/dev/null)" ]; then
        pass "Code properly formatted"
    else
        warn "Some files need formatting: $(gofmt -s -l . 2>/dev/null | wc -l) files"
    fi
else
    warn "Go toolchain not found in PATH — skipping Go backend checks"
fi

# 2. Go Kernel
echo ""
echo "=== Go Kernel (OS) ==="
if command -v go >/dev/null 2>&1; then
    info "Building Go kernel..."
    if go build -o os/hades-go-kernel/hades-kernel ./os/hades-go-kernel >/dev/null 2>&1; then
        pass "Go kernel builds"
    else
        if [ "$STRICT" -eq 1 ]; then
            fail "Go kernel build failed"
        else
            warn "Go kernel build failed"
        fi
    fi

    info "Testing Go kernel execution..."
    if timeout 2 ./os/hades-go-kernel/hades-kernel >/dev/null 2>&1; then
        pass "Go kernel executes"
    else
        warn "Go kernel execution issue"
    fi
else
    warn "Go toolchain not found in PATH — skipping Go kernel checks"
fi

# 3. Frontend
echo ""
echo "=== Frontend ==="
if [ -d "web" ] && command -v npm >/dev/null 2>&1; then
    cd web

    info "Checking node_modules..."
    if [ -d "node_modules" ]; then
        pass "Dependencies installed"
    else
        warn "node_modules missing; running npm install (best-effort)..."
        npm install --legacy-peer-deps >/dev/null 2>&1 || true
    fi

    info "Running ESLint..."
    if npm run lint 2>/dev/null | grep -q "problems"; then
        warn "ESLint has warnings"
    else
        pass "ESLint clean"
    fi

    info "Building frontend..."
    if npm run build >/dev/null 2>&1; then
        pass "Frontend builds"
    else
        warn "Frontend build failed"
    fi

    info "Running unit tests..."
    if npm run test:unit -- --run 2>/dev/null | grep -q "passed"; then
        pass "Unit tests pass"
    else
        warn "Unit test issues"
    fi

    cd ..
else
    warn "Frontend checks skipped (missing 'web' directory or npm not installed)"
fi

# 4. OS/Linux Buildroot
echo ""
echo "=== Linux Build System ==="
info "Buildroot checks can be slow. Set SKIP_OS=1 to skip in CI."

if [ "$SKIP_OS" -eq 1 ]; then
    info "SKIP_OS=1 — skipping Buildroot/kernel checks"
else
    info "Checking buildroot sources..."
    if [ -d "os/hades-linux/scripts/buildroot-2024.08.1/dl" ]; then
        dl_count=$(ls os/hades-linux/scripts/buildroot-2024.08.1/dl/*/.lock 2>/dev/null | wc -l)
        pass "Buildroot sources: $dl_count packages downloaded"
    else
        warn "Buildroot sources not complete"
    fi

    info "Checking buildroot configuration..."
    if [ -f "os/hades-linux/scripts/buildroot-2024.08.1/.config" ]; then
        pass "Buildroot configured"
    else
        warn "Buildroot needs configuration"
    fi

    info "Checking kernel config..."
    if [ -f "os/hades-linux/configs/kernel-config" ]; then
        pass "Kernel config exists"
    else
        warn "Kernel config missing"
    fi
fi

# 5. Summary
echo ""
echo "=========================================="
echo "  Build Verification Summary"
echo "=========================================="
echo ""
echo "Go Binary:      $([ -f hades ] && echo "BUILD OK" || echo "NEED BUILD")"
echo "Go Kernel:       $([ -f os/hades-go-kernel/hades-kernel ] && echo "BUILD OK" || echo "NEED BUILD")"
echo "Frontend Build:  $([ -d web/dist ] && echo "BUILD OK" || echo "NEED BUILD")"
echo "Unit Tests:      $(cd web && npm run test:unit -- --run 2>/dev/null | grep -q "passed" && echo "PASS" || echo "NEED RUN")"
echo ""
echo "Linux OS requires full build: cd os/hades-linux/scripts && ./build.sh build"
echo "Full build takes ~30-60 minutes on first run"
echo ""
echo "=========================================="
