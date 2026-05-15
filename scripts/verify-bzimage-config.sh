#!/usr/bin/env bash
set -euo pipefail

# verify-bzimage-config.sh
# Try to extract a kernel .config from a bzImage via strings and compare to a
# repository kernel config file. This is heuristic-only (works when .config is
# embedded in the image); if extraction fails the script exits with code 2.

if [ "$#" -lt 2 ]; then
  echo "Usage: $0 <path-to-bzImage> <path-to-config-to-compare>"
  exit 2
fi

BZIMAGE="$1"
REF_CONFIG="$2"

if [ ! -f "$BZIMAGE" ]; then
  echo "bzImage not found: $BZIMAGE"
  exit 2
fi
if [ ! -f "$REF_CONFIG" ]; then
  echo "Reference config not found: $REF_CONFIG"
  exit 2
fi

TMPDIR=$(mktemp -d)
trap 'rm -rf "$TMPDIR"' EXIT

EXTRACTED="$TMPDIR/extracted.config"

# Try to extract using strings heuristic. Many kernels embed the .config near
# the end of the image; searching for lines beginning with CONFIG_ is a simple
# but sometimes effective approach.

if command -v strings >/dev/null 2>&1; then
  echo "[*] Extracting candidate config from bzImage using strings..."
  strings "$BZIMAGE" | sed -n '/^CONFIG_/p' > "$EXTRACTED" || true
else
  echo "[!] 'strings' command not available; cannot extract embedded config"
  exit 2
fi

if [ ! -s "$EXTRACTED" ]; then
  echo "[!] Could not extract any CONFIG_* lines from $BZIMAGE. Embedded config not found or too compressed."
  exit 2
fi

# Normalize both files and compare
# Remove possible surrounding non-config lines and comments for a fair diff
awk '/^CONFIG_/ { print }' "$REF_CONFIG" | sort > "$TMPDIR/ref.sorted"
sort "$EXTRACTED" > "$TMPDIR/ex.sorted"

if diff -u "$TMPDIR/ref.sorted" "$TMPDIR/ex.sorted" >/dev/null 2>&1; then
  echo "[+] Embedded config in $BZIMAGE matches reference config: $REF_CONFIG"
  exit 0
else
  echo "[-] Mismatch between embedded config and reference config"
  echo "--- Reference (sorted) : $REF_CONFIG"
  echo "+++ Extracted (sorted) : embedded in $BZIMAGE"
  echo
  diff -u "$TMPDIR/ref.sorted" "$TMPDIR/ex.sorted" || true
  exit 1
fi
