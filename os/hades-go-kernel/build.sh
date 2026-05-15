#!/usr/bin/env bash
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
cd "$SCRIPT_DIR"

echo "[*] Building Go userspace kernel (hades-kernel)..."
go build -o hades-kernel .
if [ -f hades-kernel ]; then
  echo "[+] Built: $SCRIPT_DIR/hades-kernel"
else
  echo "[!] Build failed"
  exit 1
fi
