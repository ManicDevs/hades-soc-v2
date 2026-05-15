# HADES-V2

HADES-V2 is a high-concurrency, quantum-resistant Enterprise Security Framework.

## Overview

HADES-V2 provides comprehensive security operations center (SOC) capabilities with AI-powered threat detection, quantum-resistant cryptography, and distributed worker architecture.

## Features

- **AI-Powered Threat Detection**: Advanced threat analysis with adversarial AI protection
- **Quantum-Resistant Cryptography**: Kyber1024 lattice-based cryptography for PQC compliance
- **Distributed Workers**: Scalable worker architecture with stateless processing
- **Anti-Analysis**: Comprehensive protection against reverse engineering and analysis
- **Zero-Trust Network**: Network segmentation and zero-trust security model
- **Blockchain Audit**: Immutable audit logging using blockchain technology
- **Multi-Region Support**: Geographic distribution with automatic failover

## Project Structure

```
hades-v2/
├── cmd/              # Entry points and CLI commands
│   ├── hades/        # Main HADES CLI
│   ├── sentinel/     # Background service
│   └── ...
├── internal/         # Private application code
│   ├── engine/       # Core orchestration and Safety Governor
│   ├── recon/        # Reconnaissance modules (encapsulated)
│   ├── exploitation/  # Exploitation modules (encapsulated)
│   ├── api/          # HTTP API handlers
│   ├── database/     # Database abstraction layer
│   └── ...
├── pkg/              # Public library code
│   └── sdk/          # Quantum-resistant cryptographic primitives
├── web/              # React-based SOC dashboard
├── os/               # OS components (kernel, bootloader)
├── docs/             # Documentation
└── scripts/          # Build and deployment scripts
```

## Architectural Invariants

- **Zero Inbound Edges**: The `internal/` directory is encapsulated. External packages may NOT import directly from `internal/recon` or `internal/exploitation`. Use interfaces defined in `pkg/interfaces`.
- **Worker Isolation**: Distributed workers must be stateless. All persistence flows through the central `internal/database` handlers.
- **Safety Governor**: Hardcoded limit of 5 automated blocks per hour in `internal/engine/`.

## Quick Start

### Prerequisites

- Go 1.25+
- Node.js 18+
- PostgreSQL 14+
- Redis 6+

### Installation

```bash
# Clone the repository
git clone https://github.com/your-org/hades-v2.git
cd hades-v2

# Install Go dependencies
go mod download

# Install frontend dependencies
cd web
npm install
cd ..
```

### Running

```bash
# Build the CLI
go build -o bin/hades ./cmd/hades

# Run the server
./bin/hades serve

# Run the background service
./bin/hades sentinel
```

### Development

```bash
# Run tests
go test ./internal/... -v

# Run frontend dev server
cd web
npm run dev

# Build for production
go build -o bin/hades ./cmd/hades
cd web && npm run build
```

## Documentation

- [User Guide](docs/USER_GUIDE.md)
- [Deployment Guide](docs/DEPLOYMENT.md)
- [API Documentation](docs/API_V2_DOCUMENTATION.md)
- [Architecture](docs/architecture/)
- [QEMU Boot Instructions](QEMU-README.md)

## OS Components

HADES-V2 includes custom OS components for bare-metal deployment:

- **hades-go-kernel**: Pure Go kernel (userspace simulation)
- **hades-linux**: Security-hardened Linux distribution (Buildroot)
- **hades-boot-kernel**: Custom bootloader with GRUB

See [QEMU-README.md](QEMU-README.md) for boot instructions.

## License

HADES-V2 License - See LICENSE file for details.
