# internal/scheduler/ — the durable wake scheduler: lease claim/renew/complete/fail over wake_jobs

> 🌐 English | [繁體中文](README.zh-TW.md)

`Store` operates on the `wake_jobs` table (migration
`internal/storage/sqlite/migrations/0051_wake_jobs.sql`) and implements exactly the
`BEGIN IMMEDIATE` lease-claim transaction ADD §12.4 specifies
([`docs/design/Auspex_ADD.md`](../../docs/design/Auspex_ADD.md)): one worker gets exactly
one due, unleased job at a time, durably, under concurrent claimers. See
[`doc.go`](doc.go) for the package contract.

Job statuses are this package's own vocabulary — `scheduled` → `leased` → `done` | `dead` —
distinct from the frozen `domain.PauseStatus` enum.

Key files:

- [`lease.go`](lease.go) — `NewStore`, `Schedule`, `Get`, `GetByPauseKind`, `Claim`, `Renew`, `Complete`, `Fail` (backoff and re-`scheduled` while attempts remain, `dead` once exhausted), `ReclaimExpired`; `DefaultLeaseDuration` is 60s (ADD §20.7).
- [`restart.go`](restart.go) — lease recovery: `Restart` releases every `leased` row back to `scheduled` at process start (a fresh process categorically cannot be the leaseholder) and reports overdue claimable jobs in a `RestartReport`; claiming them remains `Claim`'s job.
- [`list.go`](list.go) — read-only `List` for the daemon status surfaces; takes no lease, never influences claim order.
- [`cancel.go`](cancel.go) — operator cancellation (FR-163, issue #10): reuses terminal `dead` with `last_error = CancelledByOperator` rather than inventing a fifth status.

Run modes: `auspex scheduler run-once`
([`internal/orchestrator/pauselifecycle.go`](../orchestrator/pauselifecycle.go)) performs a
single claim sweep and deliberately stops at `Claim`; the daemon's resident worker
([`internal/daemon/worker.go`](../daemon/README.md)) drives the full
claim → wake → resume → complete/fail loop continuously. Payload meaning belongs to
[`internal/pause/`](../pause/README.md), not this package.
