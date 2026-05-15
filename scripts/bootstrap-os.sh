#!/usr/bin/env bash
set -euo pipefail

# Bootstrap small, non-sensitive OS files so documentation and CI smoke
# checks can run without requiring the full Buildroot/kernel tree.

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"

CONFIG_DIR="$REPO_ROOT/os/hades-linux/configs"
BUILDROOT_DIR="$REPO_ROOT/os/hades-linux/scripts/buildroot-2024.08.1"

mkdir -p "$CONFIG_DIR"
mkdir -p "$(dirname "$BUILDROOT_DIR/.config")"

if [ ! -f "$CONFIG_DIR/kernel-config" ]; then
  cat > "$CONFIG_DIR/kernel-config" <<'EOF'
# Placeholder kernel-config
# This is a placeholder file to satisfy documentation and CI smoke checks.
# Replace with a real kernel .config (e.g., from `make savedefconfig`) for full builds.
CONFIG_EXAMPLE=y
EOF
  echo "[+] Created placeholder $CONFIG_DIR/kernel-config"
else
  echo "[+] Kernel config already exists: $CONFIG_DIR/kernel-config"
fi

if [ ! -f "$BUILDROOT_DIR/.config" ]; then
  mkdir -p "$BUILDROOT_DIR"
  cat > "$BUILDROOT_DIR/.config" <<'EOF'
# Placeholder buildroot .config
# Replace with a real Buildroot .config for full OS builds.
EOF
  echo "[+] Created placeholder $BUILDROOT_DIR/.config"
else
  echo "[+] Buildroot config exists: $BUILDROOT_DIR/.config"
fi

cat <<EOF
Bootstrap complete.
Note: these are placeholders only. To perform a full OS build, provide the real
Buildroot and kernel configurations, then run:
  cd os/hades-linux/scripts && ./build.sh build
EOF
