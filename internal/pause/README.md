# internal/pause/ — the Graceful Pause / Safe Points state machine and its composing service

> 🌐 English | [繁體中文](README.zh-TW.md)

Implements ADD §20's Graceful Pause lifecycle
([`docs/design/Auspex_ADD.md`](../../docs/design/Auspex_ADD.md)). See [`doc.go`](doc.go) for
the package contract, including how the twelve frozen `domain.PauseStatus` wire strings map
onto the design documents' prose state names.

Key files:

- [`statemachine.go`](statemachine.go) — the pure `(current state, event) → next state` transition table over the frozen enum; `Validate`, `Apply`, `IsTerminal`, `ValidEvents`. No I/O.
- [`observe.go`](observe.go) — debounce/hysteresis around runway observations; [`requestpause.go`](requestpause.go) — idempotent `RequestPause` plus the in-memory `MemStore` reference store.
- [`safepoint.go`](safepoint.go) — safe-boundary detection and the `PersistThenInterrupt` ordering; [`persistphase.go`](persistphase.go) — the five-phase persist orchestrator; [`interrupt.go`](interrupt.go) — `InterruptAndSleep`.
- [`lifecycle.go`](lifecycle.go) — manual `Cancel` / `Resume`; [`wake.go`](wake.go) — scheduler-driven `Wake`, enforcing pause-level exactly-once transitions so even split-brain double leaseholders cannot both advance one pause.
- [`resumevalidation.go`](resumevalidation.go) — the `ValidateResume` checklist (quota safety, repository compatibility, session capability, authorization) plus wake-job rescheduling when quota is still unsafe.
- [`service.go`](service.go) — `Service`, the concrete `app.GracefulPauseService` ([`../app/ports.go`](../app/ports.go)) composing all of the above.
- [`sqlitestore.go`](sqlitestore.go) — the durable `PauseStore` over the `pause_records` table (migration 0050).
- [`contextstore.go`](contextstore.go) — pause context (`QuotaBaseline`, `GitHeadBaseline`, `WorktreeID`, `PausedWorkPaths`) persisted under the existing `pause_records.metadata_json` column's `"context"` key, with a merge-preserving read-modify-write so sibling keys survive; required because the process that requests a pause is never the daemon process that resumes it (decision D-16, [`docs/DECISION_LOG.md`](../../docs/DECISION_LOG.md)).

Sleeping pauses wake through durable wake jobs owned by
[`internal/scheduler/`](../scheduler/README.md); the resident loop that drives
wake → resume unattended is [`internal/daemon/`](../daemon/README.md)'s worker. The CLI
surfaces go through [`internal/orchestrator/pauselifecycle.go`](../orchestrator/pauselifecycle.go).
