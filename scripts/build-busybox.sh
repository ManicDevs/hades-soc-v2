#!/usr/bin/env bash
set -euo pipefail

# Build a static BusyBox and install into qemu-test/rootfs.
# This script is best run on a dev machine with a C toolchain installed.
# It intentionally does NOT run automatically in CI because it requires compilers
# and can take a few minutes.

BUSYBOX_VERSION=${BUSYBOX_VERSION:-1.36.1}
BUSYBOX_TARBALL="busybox-${BUSYBOX_VERSION}.tar.bz2"
BUSYBOX_URL=${BUSYBOX_URL:-"https://busybox.net/downloads/${BUSYBOX_TARBALL}"}

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
ROOT_DIR="$(cd "$SCRIPT_DIR/.." && pwd)"
ROOTFS_DIR="$ROOT_DIR/qemu-test/rootfs"

if [ ! -d "$ROOTFS_DIR" ]; then
  echo "[!] rootfs directory not found: $ROOTFS_DIR"
  echo "    Create or extract a rootfs first (see qemu-test/rootfs)"
  exit 1
fi

TMPDIR=$(mktemp -d)
trap 'rm -rf "$TMPDIR"' EXIT
cd "$TMPDIR"

echo "[*] Downloading BusyBox $BUSYBOX_VERSION from $BUSYBOX_URL"
if command -v curl >/dev/null 2>&1; then
  curl -L --fail -o "$BUSYBOX_TARBALL" "$BUSYBOX_URL"
elif command -v wget >/dev/null 2>&1; then
  wget -O "$BUSYBOX_TARBALL" "$BUSYBOX_URL"
else
  echo "Error: curl or wget required"
  exit 1
fi

echo "[*] Extracting..."
if tar -tjf "$BUSYBOX_TARBALL" >/dev/null 2>&1; then
  tar -xjf "$BUSYBOX_TARBALL"
else
  tar -xzf "$BUSYBOX_TARBALL"
fi

BB_DIR=$(find . -maxdepth 2 -type d -name 'busybox*' | head -n1)
if [ -z "$BB_DIR" ]; then
  echo "[!] Could not locate busybox source folder after extract"
  exit 1
fi
cd "$BB_DIR"

echo "[*] Creating default config"
make defconfig

# Enable static busybox
if grep -q "# CONFIG_STATIC is not set" .config; then
  sed -i 's/# CONFIG_STATIC is not set/CONFIG_STATIC=y/' .config || true
fi

echo "[*] Building BusyBox (this may take a minute)..."
make -j"$(nproc)"

echo "[*] Installing BusyBox into rootfs: $ROOTFS_DIR"
install -D -m 0755 busybox "$ROOTFS_DIR/bin/busybox"

# Create common symlinks for applets (best-effort)
APPLETS=$(./busybox --list 2>/dev/null || true)
if [ -n "$APPLETS" ]; then
  echo "[*] Creating applet symlinks in $ROOTFS_DIR/bin"
  pushd "$ROOTFS_DIR/bin" >/dev/null
  for a in $APPLETS; do
    if [ ! -e "$a" ]; then
      ln -sf busybox "$a"
    fi
  done
  popd >/dev/null
else
  echo "[!] Could not enumerate busybox applets; skipping symlink creation"
fi

echo "[+] BusyBox built and installed into $ROOTFS_DIR/bin"

echo "Next steps: regenerate initramfs with os/hades-linux/scripts/build-rootfs.sh or copy the generated initramfs into qemu-test/initramfs.cpio.gz"
