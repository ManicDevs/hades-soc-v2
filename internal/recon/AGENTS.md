# 🔍 Recon Module - Privacy Rules

- **Encapsulation:** Never expose scanner internals via public variables. Use the `Scanner` interface.
- **Data Scrubbing:** Ensure all PII or sensitive target data is scrubbed before being logged to the engine's stdout.
- **No Direct Imports:** This module must never import from `cmd/` or `pkg/ui`.
