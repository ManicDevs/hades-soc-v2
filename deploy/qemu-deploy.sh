#!/bin/bash
# HADES-V2 QEMU Deployment Script

set -e

HADES_ISO="os/hades-linux/scripts/buildroot-2024.08.1/output/images/hades-os.iso"
HADES_IMG="os/hades-linux/scripts/buildroot-2024.08.1/output/images/hades-os.img"
QEMU_MEMORY="2048"
QEMU_CPUS="2"
QEMU_NETWORK="user"
QEMU_PORTS="2222:22,8443:8443,19001:19001,19002:19002,19003:19003"

echo "🚀 HADES-V2 QEMU Deployment"
echo "=============================="

# Check if HADES OS exists
if [ ! -f "$HADES_ISO" ] && [ ! -f "$HADES_IMG" ]; then
    echo "❌ HADES OS not found!"
    echo "Please build HADES OS first:"
    echo "  cd os/hades-linux && ./scripts/build.sh"
    exit 1
fi

# Choose boot method
BOOT_METHOD="${1:-iso}"
case "$BOOT_METHOD" in
    "iso")
        if [ ! -f "$HADES_ISO" ]; then
            echo "❌ HADES ISO not found: $HADES_ISO"
            exit 1
        fi
        echo "📀 Booting from ISO: $HADES_ISO"
        QEMU_CMD="qemu-system-x86_64 -cdrom $HADES_ISO"
        ;;
    "img")
        if [ ! -f "$HADES_IMG" ]; then
            echo "❌ HADES IMG not found: $HADES_IMG"
            exit 1
        fi
        echo "💿 Booting from disk image: $HADES_IMG"
        QEMU_CMD="qemu-system-x86_64 -drive file=$HADES_IMG,format=raw,if=virtio"
        ;;
    *)
        echo "❌ Invalid boot method: $BOOT_METHOD"
        echo "Usage: $0 [iso|img]"
        exit 1
        ;;
esac

# Configure QEMU
QEMU_CMD="$QEMU_CMD \
    -m $QEMU_MEMORY \
    -smp $QEMU_CPUS \
    -netdev $QEMU_NETWORK,id=net0,hostfwd=tcp::$QEMU_PORTS \
    -device e1000,netdev=net0 \
    -serial mon:stdio \
    -enable-kvm \
    -daemonize"

echo "🔧 QEMU Configuration:"
echo "  Memory: ${QEMU_MEMORY}MB"
echo "  CPUs: $QEMU_CPUS"
echo "  Network: $QEMU_NETWORK"
echo "  Port forwarding: $QEMU_PORTS"
echo "  KVM: Enabled"
echo ""

# Start QEMU
echo "🚀 Starting HADES-V2 in QEMU..."
echo "Command: $QEMU_CMD"
echo ""

# Check if QEMU is available
if ! command -v qemu-system-x86_64; then
    echo "❌ QEMU not found. Install with:"
    echo "  sudo apt install qemu-system-x86"
    exit 1
fi

# Start QEMU VM
eval "$QEMU_CMD"

# Wait a moment for VM to start
sleep 3

echo "✅ HADES-V2 VM started!"
echo ""
echo "🌐 Access Information:"
echo "  Web Dashboard: http://localhost:8443"
echo "  API Server: http://localhost:8080"
echo "  SSH Access: ssh root@localhost -p 2222"
echo "  P2P Ports: 19001, 19002, 19003"
echo ""
echo "🔍 Monitoring:"
echo "  To monitor VM: tail -f /tmp/hades-qemu.log"
echo "  To stop VM: pkill qemu-system-x86_64"
echo ""
echo "📊 HADES-V2 is running in QEMU!"
