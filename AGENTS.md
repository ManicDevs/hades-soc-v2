# Hades-V2 Agent Guidelines

## Project Structure

- **Backend:** Go 1.25+ with cobra CLI (`cmd/hades/`), pure Go (no CGO)
- **Frontend:** React 18 in `web/` directory, Vite + vitest + playwright
- **Migrations:** SQL files in `migrations/` (001-006), runs via `migrate` subcommand

## Running Commands

```bash
# Go (run from repo root)
go test -v -race ./...                    # unit tests
go test -v -race -tags=integration ./tests/integration/...  # integration tests
go build -o hades ./cmd/hades             # build binary
go vet ./...                              # lint check
gofmt -s -l .                             # format check

# Frontend (run from web/ directory)
npm run dev                               # start dev server on port 3000
npm run lint                              # ESLint
npm run test:unit                         # vitest unit tests
npm run test:e2e                          # playwright e2e tests
```

## OS Components

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

## Important Notes

- Buildroot sources already downloaded in `os/hades-linux/scripts/buildroot-2024.08.1/dl/`
- Kernel config at `os/hades-linux/configs/kernel-config`
- Post-build script creates systemd services (hades, hades-selfheal, hades-antianalysis)
- `opencode.json` references `workflow.md` which does not exist — ignore this reference
- Build uses `CGO_ENABLED=0` for cross-platform static binaries
- Docker health check hits `http://localhost:8443/api/health`
- Integration tests require postgres (port 5432) + redis; see CI workflow for env vars
- Frontend-specific rules: see `web/src/AGENTS.md`