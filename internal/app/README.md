# internal/app/ — frozen cross-component service ports (interfaces + DTOs)

> 🌐 English | [繁體中文](README.zh-TW.md)

One source file, [`ports.go`](ports.go): the frozen contracts every component talks through
(ADD §9.9, §9.10 — the ADD lives at [`docs/design/Auspex_ADD.md`](../../docs/design/Auspex_ADD.md)).
Interfaces are intentionally narrow; the package-level comment forbids widening one into a
God interface.

What `ports.go` defines:

- `TxRunner` / `TxFunc` — the single frozen storage-transaction callback shape.
- `EvaluationService` — evaluate / decide / authorize, plus `PolicyAction` (`RUN` … `BLOCK`) and the `Evaluation` / `DecisionResult` / `Authorization` DTOs.
- Predictor pipeline stage ports (ADR-041): `ScopeEstimator`, `TokenForecaster`, `QuotaForecaster`, `RiskCombiner` — each stage independently swappable.
- `ProgressTreeService`, `StateCheckpointService`, `RepositoryCheckpointService`, `GracefulPauseService` (Observe / RequestPause / ReachSafePoint / EnterSleep / Resume / Cancel).
- Provider ports segregated by capability: `ProviderDetector`, `ProviderCapabilityReader`, `HookNormalizer`, `ManagedRunner`, `LiveObserver`.

Ownership: `internal/app/ports.go` is a frozen cross-cutting file owned exclusively by
`contract-integrator` ([CONSTITUTION.md §4.3](../../CONSTITUTION.md)); no other role edits it.

Neighbors: DTOs are built from [`internal/domain/`](../domain/README.md) types. Concrete
implementations live in the owning packages (e.g. [`internal/pause/`](../pause/README.md)'s
`Service` for `GracefulPauseService`, `internal/evaluation` for `EvaluationService`) and are
composed into one container by [`./wiring/`](wiring/README.md).
[`internal/orchestrator/`](../orchestrator/README.md) sequences calls into these ports without
knowing the concrete implementations. No `doc.go`; the package comment is at the top of
`ports.go`.
