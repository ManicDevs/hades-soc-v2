#!/usr/bin/env bash
set -euo pipefail

# Package a minimal pure-Go OS image containing hades-kernel and hades-server
# Output: os/go-os/hades-go-image.tar.gz

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_ROOT="$(cd "$SCRIPT_DIR/../.." && pwd)"
OUT="$SCRIPT_DIR/hades-go-image.tar.gz"
WORK="$(mktemp -d)"
trap 'rm -rf "$WORK"' EXIT

echo "[*] Packaging hades-go image into $OUT"

# Build hades-kernel and copy
(cd "$REPO_ROOT/os/hades-go-kernel" && ./build.sh)
cp "$REPO_ROOT/os/hades-go-kernel/hades-kernel" "$WORK/"

# Copy hades-server if present
if [ -f "$REPO_ROOT/hades-server" ]; then
  cp "$REPO_ROOT/hades-server" "$WORK/"
else
  echo "[!] hades-server not found at $REPO_ROOT/hades-server — image will not include server"
fi

# Include default config dir if present
if [ -d "$REPO_ROOT/config" ]; then
  cp -r "$REPO_ROOT/config" "$WORK/"
fi

# If gobox exists in bin/, include it and create common symlinks
if [ -f "$REPO_ROOT/bin/gobox" ]; then
  mkdir -p "$WORK/bin"
  cp "$REPO_ROOT/bin/gobox" "$WORK/bin/gobox"
  chmod +x "$WORK/bin/gobox"
  pushd "$WORK/bin" >/dev/null
  for a in ls cat hostname echo sleep whoami ps ifconfig uptime; do
    if [ ! -e "$a" ]; then ln -s gobox "$a"; fi
  done
  popd >/dev/null
fi

# Create startup script
cat > "$WORK/start.sh" <<'EOF'
#!/usr/bin/env bash
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
cd "$SCRIPT_DIR"

# Run hades-kernel in foreground; it will attempt to auto-start services found in the image
./hades-kernel
EOF
chmod +x "$WORK/start.sh"

# Tar up
tar -C "$WORK" -czf "$OUT" .

echo "[+] Created $OUT"
