#!/usr/bin/env bash
set -euo pipefail

# Build gobox multi-call utility and place in bin/
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"
OUTPUT_DIR="$REPO_ROOT/bin"
mkdir -p "$OUTPUT_DIR"
(cd "$REPO_ROOT/cmd/gobox" && go build -o "$OUTPUT_DIR/gobox")

echo "[+] Built $OUTPUT_DIR/gobox"
