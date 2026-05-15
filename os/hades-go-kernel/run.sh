#!/usr/bin/env bash
# run.sh — Build and launch the HADES Go Kernel simulation.
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
cd "$SCRIPT_DIR"

bash build.sh
echo ""
exec ./hades-kernel
