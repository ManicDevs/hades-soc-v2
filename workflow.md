# Hades-V2 Development Workflow

## Project Overview
Hades-V2 is a security operating system with Go backend, React frontend, and Linux kernel components.

## Development Commands

### Backend (Go)
```bash
# Run from repository root
go test -v -race ./...                           # Unit tests
go test -v -race -tags=integration ./tests/integration/...  # Integration tests
go build -o hades ./cmd/hades                    # Build binary
go vet ./...                                     # Lint check
gofmt -s -l .                                    # Format check
```

### Frontend (React)
```bash
# Run from web/ directory
npm run dev                                      # Start dev server on port 3000
npm run lint                                     # ESLint
npm run test:unit                                # Vitest unit tests
npm run test:e2e                                 # Playwright e2e tests
```

### OS Components
```bash
# Go Kernel (userspace simulation)
cd /home/cerberus/Desktop/hades
go build -o os/hades-go-kernel/hades-kernel ./os/hades-go-kernel
./os/hades-go-kernel/hades-kernel

# Linux Build (Buildroot - requires ~30-60 min first build)
cd os/hades-linux/scripts
sudo ./build.sh build     # Full build
sudo ./build.sh config    # Menuconfig
sudo ./build.sh clean     # Clean
# Outputs: output/images/bzImage, rootfs.ext4, hades-v2.img
```

## Key Development Notes

- Buildroot sources already downloaded in `os/hades-linux/scripts/buildroot-2024.08.1/dl/`
- Kernel config at `os/hades-linux/configs/kernel-config`
- Post-build script creates systemd services (hades, hades-selfheal, hades-antianalysis)
- Use `CGO_ENABLED=0` for cross-platform static binaries
- Docker health check hits `http://localhost:8443/api/health`
- Integration tests require postgres (port 5432) + redis

## Testing Strategy
- Unit tests: `go test -v -race ./...`
- Integration tests: `go test -v -race -tags=integration ./tests/integration/...`
- Frontend unit: `npm run test:unit` (from web/)
- Frontend e2e: `npm run test:e2e` (from web/)

## Code Quality
- Go formatting: `gofmt -s -l .`
- Go linting: `go vet ./...`
- Frontend linting: `npm run lint` (from web/)
