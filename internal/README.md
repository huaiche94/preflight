# internal/ — Auspex's private Go packages: frozen contracts, role-owned services, and adapters

> 🌐 English | [繁體中文](README.zh-TW.md)

Everything under `internal/` is private to the `github.com/huaiche94/auspex` module. The frozen
cross-role surfaces live in [`domain`](domain/) (entities) and [`app`](app/) (ports), per the
import-path table in
[docs/implementation/vertical-slice/CONTRACT_FREEZE.md](../docs/implementation/vertical-slice/CONTRACT_FREEZE.md);
the other packages are role-owned implementations behind those contracts.

## Layering rule

Auspex_ADD.md §7.6 "Core dependency direction" (the ADD lives at
[docs/design/Auspex_ADD.md](../docs/design/Auspex_ADD.md)) fixes the dependency direction:

```text
cli/http/hooks
    ↓
application/orchestrator
    ↓
domain services + interfaces
    ↓
adapters: sqlite/git/providers/filesystem
```

The domain layer must not import provider, SQLite, Cobra, or VS Code-specific types.

## Package index

Where a package has a `doc.go`, that file is the package contract — read it before changing the package.

| Package | Role |
|---|---|
| [app](app/) | Frozen cross-component ports (ADD §9.9, §9.10): narrow service interfaces, request/response DTOs, and the `TxRunner` transaction boundary. `app/wiring` composes real implementations into a runnable App. |
| [artifacts](artifacts/) | Checkpoint Part A's artifact validator layer: concrete checks (checksum, file-exists, heading, fence balance) that turn claimed evidence into verified evidence before `progress` records it ("Completed means evidenced", Constitution §6.2). |
| [buildinfo](buildinfo/) | Minimal build/version metadata behind `auspex version`; a fixed development placeholder until real release packaging exists. See [buildinfo/README.md](buildinfo/README.md). |
| [cli](cli/) | Cobra command-tree constructors (`NewRootCmd` and friends) for the `auspex` CLI; constructors, not package-level singletons, so the tree is testable without `os.Exit`. See its `doc.go`. |
| [clock](clock/) | Real wall-clock implementation of `domain.Clock`; production code depends on the interface so tests can substitute a fake. See [clock/README.md](clock/README.md). |
| [config](config/) | Layered YAML configuration loading per ADD §26.1's precedence chain, with the `schema_version` envelope and unknown-field warn/strict validation. See [config/README.md](config/README.md). |
| [daemon](daemon/) | M6 daemon lifecycle: singleton lock, per-restart bearer token, dynamic loopback listener, runtime metadata file, resident worker loop, and the in-memory SSE event broker. |
| [domain](domain/) | Frozen domain entities: opaque UUIDv7-backed ID types, status enums, the frozen `domain.Error` shape, forecast/usage/capability types, and the `Clock`/`IDGenerator`/`ProcessRunner` ports. |
| [evaluation](evaluation/) | Implements `app.EvaluationService`: runs the ADR-041 predictor pipeline for one turn, persists feature-vector/prediction/decision rows in one transaction, and enforces exactly-once authorizations. See its `doc.go`. |
| [features](features/) | Derives prediction-input signals from prompts, repositories, sessions, and Progress Trees (ADD §14.2); the privacy boundary where raw prompt text enters and never leaves. See its `doc.go`. |
| [gitx](gitx/) | Git plumbing for repository checkpointing: repository/worktree resolution and `git status --porcelain=v2 -z` parsing, argv-only through `domain.ProcessRunner` (never a shell string). |
| [hooks](hooks/) | Provider lifecycle-hook payload parsing. `hooks/claude` parses Claude Code hook stdin payloads (UserPromptSubmit, Stop, StopFailure) and encodes the provider-compatible stdout responses. |
| [httpapi](httpapi/) | The daemon's authenticated loopback HTTP/JSON + SSE surface (ADD §23.2–23.5): health/version/capabilities/status/jobs, the live event stream, and the scheduler-job cancel endpoint. |
| [idgen](idgen/) | UUIDv7 implementation of `domain.IDGenerator`; every Auspex-owned entity ID is a UUIDv7, never parsed for meaning. See [idgen/README.md](idgen/README.md). |
| [integrationtest](integrationtest/) | qa-owned cross-role integration and end-to-end tests: the high-risk fixture flow, restart-on-same-DB, privacy/leakage scans, scheduler double-worker contention, and more. |
| [lock](lock/) | Single-machine, PID-file-style advisory file lock asserting exclusive ownership of a runtime directory; deliberately not a distributed or network lock. See [lock/README.md](lock/README.md). |
| [managed](managed/) | Managed one-shot runner behind `auspex run` (ADD §8.1): pre-prompt gate, provider subprocess lifecycle (`claude -p … stream-json`), and terminal-outcome attribution. See its `doc.go`. |
| [orchestrator](orchestrator/) | Sequences the day-one evaluate/decide pipeline (ADD §13) across the frozen `app` ports; implements no prediction/policy/checkpoint logic itself. See its `doc.go`. |
| [paths](paths/) | Per-OS resolution of the global config/data/cache/runtime directories with an injectable environment. See [paths/README.md](paths/README.md). |
| [pause](pause/) | Pure state-transition validator for the Graceful Pause state machine (ADD §20) over the twelve frozen `domain.PauseStatus` values; no I/O. See its `doc.go`. |
| [policy](policy/) | Terminal predictor-pipeline stage: combines the risk score and runway forecast into a frozen `app.PolicyAction` under ADD §17.3's fixed priority order; never labels an uncalibrated score a probability. See its `doc.go`. |
| [predictor](predictor/) | Deterministic, explainable prediction primitives (ADD §15–§16) above `features`: risk scores and quantile estimates, never calibrated probabilities on day one. See its `doc.go`. |
| [pricing](pricing/) | Local, hand-maintained per-model price table (ADR-043) turning a token forecast into an estimated USD cost range; never fetched at runtime, always labeled an estimate. |
| [progress](progress/) | Progress Tree domain service (Constitution §6, ADD §18): node/edge/artifact/task stores, the node state machine, and the stage/verify/commit CompleteNode protocol. |
| [providers](providers/) | Provider-native payload parsing. `providers/claude` parses Claude Code status-line JSON into intermediate, unknown-field-tolerant structs. |
| [redact](redact/) | Content-based secret detection (ADD §19.5, §27.8): filename patterns plus fixed content detectors used by the untracked-file archive policy; documented as non-exhaustive. See its `doc.go`. |
| [repocheckpoint](repocheckpoint/) | Repository Checkpoint create/verify/restore (ADD §19): patches + untracked archive with atomic writes; never mutates the active branch or working tree during capture. |
| [retention](retention/) | ADR-046 tiered telemetry retention (hot window → rollup → gzip archive → verified delete) behind `auspex gc`. See [retention/README.md](retention/README.md). |
| [scheduler](scheduler/) | Durable wake scheduler lease over the `wake_jobs` table (ADD §12.4): claim/renew/complete/fail/retry, correct under concurrent workers. See its `doc.go`. |
| [statecheckpoint](statecheckpoint/) | State Checkpoint manifest (ADD §18.8): the manifest's Go shape, deterministic JSON serialization, and integrity checksum. |
| [storage](storage/) | Storage adapters. `storage/sqlite` is the SQLite runtime (pragmas, transactions) and the forward-only migration engine. See [storage/README.md](storage/README.md). |
| [telemetry](telemetry/) | Provider telemetry normalization and persistence into the frozen `pkg/protocol/v1.Event` envelope; `telemetry/claude` is the sole Claude-payload path. See [telemetry/README.md](telemetry/README.md). |
| [testutil](testutil/) | Test support. `testutil/fakes` holds hand-written fakes for every frozen `app` port, with compile-time interface assertions and fail-loud unconfigured methods. See `testutil/fakes/doc.go`. |
