# HADES QEMU Boot Instructions

This document provides instructions for booting HADES OS components using QEMU.

## Prerequisites

```bash
sudo apt-get install qemu-system-x86 qemu-kvm grub-pc-bin xorriso mtools
```

## Available QEMU Boot Scripts

### 1. HADES QEMU Test (Kernel + Initramfs) ✅ WORKING
**Location:** `qemu-test/run-qemu.sh`

Boots the pre-built Linux 6.10.10 kernel with BusyBox initramfs directly
(no bootloader layer — QEMU loads the kernel itself).

```bash
cd qemu-test
./run-qemu.sh          # Boot with initramfs only
./run-qemu.sh --net    # Boot with networking (port 2222 → 22 forwarded)
./run-qemu.sh --disk   # Boot with rootfs.ext2 disk image
```

**Status:** ✅ Working — Shows HADES banner and drops into an interactive BusyBox shell.

---

### 2. HADES-V2 Linux Distribution (hades-linux) ✅ WORKING
**Location:** `os/hades-linux/scripts/run-qemu.sh`

Full HADES-V2 Linux OS boot using the pre-built kernel and rebuilt initramfs.

```bash
cd os/hades-linux/scripts
./run-qemu.sh          # Boot without networking
./run-qemu.sh --net    # Boot with networking (ports 2222, 8443 forwarded)
./run-qemu.sh --disk   # Boot with rootfs disk image

# Rebuild initramfs from rootfs/ source tree:
./build-rootfs.sh
```

**Status:** ✅ Working — Uses Linux 6.10.10 kernel with BusyBox initramfs; interactive shell on ttyS0.

---

### 3. HADES Boot Kernel (GRUB Bootloader → Kernel → OS) ✅ WORKING
**Location:** `os/hades-boot-kernel/scripts/run-qemu.sh`

Full bootloader chain: **GRUB → Linux 6.10.10 → BusyBox initramfs**.
GRUB is embedded in a bootable ISO image.

```bash
cd os/hades-boot-kernel/scripts
./build.sh             # Build the bootable ISO (~22 MB)
./run-qemu.sh          # Boot without networking
./run-qemu.sh --net    # Boot with networking (ports 2222, 8443 forwarded)
./clean.sh             # Remove build artefacts
```

**Boot sequence:**
1. QEMU BIOS loads the CD-ROM ISO
2. GRUB displays the HADES-V2 Secure Bootloader menu (5-second countdown)
3. Kernel entry: `bzImage console=ttyS0,115200 loglevel=3 init=/init quiet`
4. Linux kernel boots, mounts `/proc`, `/sys`, `/dev`
5. HADES banner is displayed; interactive `ash` shell starts

**Status:** ✅ Working — Full GRUB → kernel → HADES OS chain verified in QEMU.

---

### 4. HADES Pure Go Kernel (Userspace Simulation) ✅ WORKING
**Location:** `os/hades-go-kernel/`

Userspace simulation of a kernel/OS environment — runs directly without QEMU.
Provides an animated boot sequence and an interactive shell with 14 OS commands.

```bash
cd os/hades-go-kernel
./build.sh             # Compile the binary
./run.sh               # Build & run immediately

# Or manually:
go build -o hades-kernel .
./hades-kernel
```

**Available shell commands:** `help`, `uname`, `ps`, `meminfo`, `cpuinfo`,
`uptime`, `dmesg`, `ls`, `cat`, `ifconfig`, `clear`, `shutdown`, `poweroff`, `exit`

**Status:** ✅ Working — Userspace kernel simulation with animated boot and interactive shell.

---

## QEMU Options

All QEMU boot scripts support the following features:

| Feature | Description |
|---------|-------------|
| **KVM Acceleration** | Automatically enabled when `/dev/kvm` is available |
| **Networking** | Use `--net` flag to enable user-mode networking |
| **Serial Console** | All output goes to serial console (`ttyS0`) |
| **Memory** | 256–512 MB depending on component |
| **Exit** | Press **Ctrl+A then X** to exit QEMU |
| **Monitor** | Press **Ctrl+A then C** for the QEMU monitor |

## Troubleshooting

### QEMU not found
```bash
sudo apt-get install qemu-system-x86 qemu-kvm
```

### Permission denied on /dev/kvm
```bash
sudo usermod -aG kvm $USER
# Log out and back in
```

### Boot hangs
- Check that kernel/initramfs files exist in `qemu-test/boot/` and `qemu-test/`
- Try without KVM: remove `-enable-kvm` from the run script
- Increase memory: change `-m 256M` to `-m 512M`

### GRUB ISO missing
```bash
cd os/hades-boot-kernel/scripts && ./build.sh
```

### Rebuild initramfs
```bash
cd os/hades-linux/scripts && ./build-rootfs.sh
```

## Quick Start

**Fastest (no bootloader):**
```bash
cd qemu-test && ./run-qemu.sh
```

**Full bootloader chain (GRUB → Kernel → OS):**
```bash
cd os/hades-boot-kernel/scripts && ./run-qemu.sh
```

**Go userspace simulation (no QEMU needed):**
```bash
cd os/hades-go-kernel && ./run.sh
```

## Component Summary

| Component | Status | Notes |
|-----------|--------|-------|
| qemu-test | Working | Kernel + initramfs, boots successfully |
| hades-linux | Working | Uses Buildroot kernel with qemu-test initramfs |
| hades-go-kernel | Working | Userspace simulation, no QEMU needed |
| hades-boot-kernel | Working | Boots from ISO with GRUB bootloader |
