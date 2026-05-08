# HADES API v2 Security Notes

## Authentication

- `POST /api/v2/auth/login` returns a JWT token signed with `HADES_JWT_SECRET`.
- `POST /api/v2/auth/refresh` requires a valid bearer token and reissues a token for the same claims.
- `GET /api/v2/auth/me` requires a valid bearer token and resolves the current user from token claims.

## Runtime Security Configuration

- `HADES_JWT_SECRET` is mandatory at API startup.
- Optional development credentials are loaded from `HADES_DEV_CREDENTIALS` as `user:password` pairs.

## Safety Governor

- Sentinel startup fails closed when orchestrator safety governor wiring is unavailable.
- Destructive isolation actions in orchestrator honey-token and lateral-movement paths are validated through the safety governor prior to execution.

## Quantum/PQC

- Quantum key generation defaults to `kyber1024` when no algorithm is specified.
- Simulated cryptographic operations are disabled unless explicitly enabled with `HADES_ALLOW_SIMULATED_CRYPTO=true` for non-production test environments.
