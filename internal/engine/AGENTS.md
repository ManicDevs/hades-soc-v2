# ⚙️ Hades Engine - Specialist Rules

## 1. Concurrency & Performance
- **Context Handling:** Every function must accept `context.Context` as the first argument. No "hanging" goroutines.
- **Worker Pool:** Modifications to the worker pool must include checks for resource starvation and properly handle `SIGTERM` for graceful shutdown.
- **State:** Workers must remain stateless. All persistence must use the `internal/db` interface.

## 2. 🚨 Safety Governor (Hard Rules)
- **Rate Limiting:** Enforcement of the 5-block/hour limit happens in `governor.go`. Any change to this logic requires 100% test coverage.
- **Manual ACK:** Ensure `ActionRequest` structs always include the `RequiresApproval` boolean for destructive actions.
