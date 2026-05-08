# ⚡ HADES-V2 Agent Directives

## 1. Project Mission & Tech Stack
HADES-V2 is a high-concurrency, quantum-resistant Enterprise Security Framework.
- **Backend:** Go 1.21+ (Primary engine, distributed workers)
- **Frontend:** React + TypeScript (SOC Dashboard)
- **Security:** Kyber1024 (PQC), manual ACK gates for destructive actions.

## 2. 🛡️ Critical Safety Guardrails (DO NOT BYPASS)
- **Safety Governor:** Hardcoded limit of **5 automated blocks per hour**. Any logic modifying `internal/engine/` MUST preserve this circuit breaker.
- **Manual ACK Policy:** Autonomous actions with high blast radii (e.g., firewall drops, credential resets) require a `ManualACK: true` flag in the payload.
- **PQC Key Rotation:** Default to **Kyber1024** in `pkg/sdk/` for all cryptographic operations. Standard RSA/ECDSA is prohibited unless for legacy SIEM ingestion.

## 3. 🏗️ Architectural Invariants
- **Zero Inbound Edges:** The `internal/` directory is encapsulated. `cmd/` or `pkg/` may NOT import directly from `internal/recon` or `internal/exploitation`. Use interfaces defined in `pkg/interfaces`.
- **Worker Isolation:** Distributed workers must be stateless. All persistence must flow through the central `internal/db` handlers.

## 4. 💻 Development Workflow
### Go Commands
- **Test:** `go test ./internal/... -v`
- **Build:** `go build -o bin/hades ./cmd/hades`
- **Audit:** `go mod tidy && go vet ./...`

### Frontend Standards
- **Syncing:** Before updating Go API handlers, verify types in `web/dashboard/src/types/schema.ts`.
- **Strict Typing:** No `any` types in TypeScript. Use discriminative unions for SIEM event types.

## 5. 🔍 Directory Map
- `/cmd/hades`: Main entry points and API server.
- `/internal/engine`: Core orchestration logic and Safety Governor.
- `/internal/recon`: Sensitive scanning and discovery modules (Encapsulated).
- `/pkg/sdk`: Shared quantum-resistant cryptographic primitives.
- `/web/dashboard`: React-based SOC interface.

## 6. ✅ Agent Definition of Done (DoD)
1. **Safety Check:** Did I verify the 5-block/hour limit?
2. **Encapsulation:** Did I avoid importing from `internal/` into an external package?
3. **Tests:** Did I include a mock for external SIEM/EDR API responses?
4. **Docs:** Did I update the `/docs/api/v2_spec.md` if the schema changed?
