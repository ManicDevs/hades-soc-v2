# Hades Go Kernel (Userspace Simulation)

This component is a pure-Go userspace kernel simulator. It runs as a normal
userland process and provides an animated boot sequence and an interactive
`hades` shell with a small set of OS-like commands.

Quick start (from repository root):

```bash
# Build the binary
cd os/hades-go-kernel
./build.sh

# Run the kernel simulation (runs an interactive shell)
./hades-kernel
```

Available commands inside the shell:
- `help`       — list commands
- `uname`      — show OS / architecture
- `ps`         — list processes (best-effort)
- `meminfo`    — memory usage
- `cpuinfo`    — CPU info
- `uptime`     — uptime since simulated boot
- `dmesg`      — show kernel messages captured during boot
- `ls [path]`  — list files
- `cat <file>` — print file contents
- `ifconfig`   — list network interfaces
- `clear`      — clear screen
- `shutdown`   — exit the simulator
- `exit`       — exit shell

Design notes:
- The simulator intentionally uses the standard library only (pure Go, no cgo).
- Where possible it will read from `/proc` for richer information when available.
- This is intended as a rapid development environment and not as a drop-in
  replacement for a real kernel; it is useful for testing the HADES control
  plane, UI and orchestration logic without needing QEMU or a VM.
