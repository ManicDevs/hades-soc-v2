#!/usr/bin/env bash
# QEMU boot script for HADES kernel + initramfs test
# Improved: clearer CLI, safer checks, and optional flags

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
cd "$SCRIPT_DIR"

usage() {
  cat <<EOF
Usage: $(basename "$0") [--net] [--disk] [--help]

Options:
  --net, -n     Enable user-mode networking (forward port 2222 -> 22)
  --disk, -d    Attach rootfs.ext2 as a disk (root=/dev/sda)
  --help, -h    Show this help and exit

Examples:
  $(basename "$0")            # Boot with initramfs only
  $(basename "$0") --net      # Boot with networking (port 2222 -> 22)
  $(basename "$0") --disk     # Boot with rootfs.ext2 disk image
EOF
}

# Parse args
NET=0
DISK=0
while [ $# -gt 0 ]; do
  case "$1" in
    --net|-n)
      NET=1
      shift
      ;;
    --disk|-d)
      DISK=1
      shift
      ;;
    -h|--help)
      usage
      exit 0
      ;;
    *)
      echo "Unknown option: $1"
      usage
      exit 2
      ;;
  esac
done

# Files we expect
KERNEL="boot/bzImage"
INITRAMFS="initramfs.cpio.gz"
ROOTFS="rootfs.ext2"

for f in "$KERNEL" "$INITRAMFS"; do
  if [ ! -f "$f" ]; then
    echo "[-] Required file not found: $f"
    echo "    Run: ./build.sh (component may be missing)"
    exit 1
  fi
done

if [ "$DISK" -eq 1 ] && [ ! -f "$ROOTFS" ]; then
  echo "[-] Disk image requested but not found: $ROOTFS"
  exit 1
fi

# Ensure QEMU is available
if ! command -v qemu-system-x86_64 >/dev/null 2>&1; then
  echo "[-] qemu-system-x86_64 not found in PATH. Install qemu-system-x86 or adjust PATH."
  exit 1
fi

# Base QEMU options
QEMU_OPTS=(
  -m 512M
  -smp 2
  -kernel "$KERNEL"
  -initrd "$INITRAMFS"
  -append "console=ttyS0 loglevel=3 init=/init"
  -nographic
  -serial mon:stdio
)

# Optional: use disk image
if [ "$DISK" -eq 1 ]; then
  QEMU_OPTS+=(
    -drive "file=$ROOTFS,format=raw,index=0,media=disk"
    -append "console=ttyS0 loglevel=3 root=/dev/sda ro"
  )
  echo "[+] Using disk image: $ROOTFS"
fi

# Optional: networking
if [ "$NET" -eq 1 ]; then
  QEMU_OPTS+=(
    -netdev user,id=net0,hostfwd=tcp::2222-:22
    -device e1000,netdev=net0
  )
  echo "[+] Networking enabled (port 2222 forwarded -> guest:22)"
fi

# Optional: enable KVM if available and permitted
if [ -e /dev/kvm ] && [ -r /dev/kvm ] && [ -w /dev/kvm ]; then
  QEMU_OPTS+=(-enable-kvm)
  echo "[+] KVM acceleration enabled"
else
  echo "[!] /dev/kvm not available or not writable — running without KVM"
fi

echo "[*] Booting QEMU..."
echo "    Press Ctrl+A then X to exit"

eval "qemu-system-x86_64 ${QEMU_OPTS[*]}"
