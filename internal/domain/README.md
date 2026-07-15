# internal/domain/ — the frozen domain model: entities, IDs, enums, and the error shape

> 🌐 English | [繁體中文](README.zh-TW.md)

Pure types only — no I/O, no service logic, no third-party imports. Every other layer
(ports, orchestrator, CLI, daemon) speaks in these types.

Key files:

- [`ids.go`](ids.go) — opaque `string`-based ID types (`RepositoryID`, `SessionID`, `TurnID`, `EvaluationID`, `TaskID`, `PauseID`, `WakeJobID`, …); UUIDv7 at generation time, never parsed for meaning.
- [`status.go`](status.go) — frozen enum wire strings (`TurnStatus`, `PauseStatus`, and friends), verified by [`status_test.go`](status_test.go).
- [`errors.go`](errors.go) — `domain.Error` with typed `ErrorCode` (`validation`, `not_found`, `conflict`, `unauthorized`, `integrity`, `unavailable`, `internal`) plus `Retryable`; the error shape every command and API endpoint renders.
- [`usage.go`](usage.go) — `UsageObservation` / quota / context observations with pointer fields: `nil` means unknown, never a substituted zero.
- [`forecast.go`](forecast.go) — `ReasonCode`, the closed vocabulary predictions and policy decisions cite (ADD §16.4).
- [`capability.go`](capability.go) — `ProviderCapabilities`; `false` means confirmed absent, not "not yet checked".
- [`clock.go`](clock.go) — `Clock` and `IDGenerator` interfaces.
- [`checkpoint.go`](checkpoint.go), [`measurement.go`](measurement.go), [`failure.go`](failure.go), [`artifact.go`](artifact.go) — checkpoint, measurement-source, failure-class, and evidence/artifact types.

Ownership: `internal/domain/**` is a shared cross-cutting path owned exclusively by the
`contract-integrator` role — no other role edits it, ever
([CONSTITUTION.md §4.3](../../CONSTITUTION.md)). Contract-level changes go through the ADR
process (Constitution §3, [`docs/adr/`](../../docs/adr)). The frozen shapes are catalogued in
[`docs/implementation/vertical-slice/CONTRACT_FREEZE.md`](../../docs/implementation/vertical-slice/CONTRACT_FREEZE.md).

Neighbors: the frozen service ports in [`internal/app/ports.go`](../app/ports.go) build their
DTOs from these types. "ADD" section references throughout the code refer to
[`docs/design/Auspex_ADD.md`](../../docs/design/Auspex_ADD.md). There is no `doc.go`; the
per-file comments carry the contracts.
