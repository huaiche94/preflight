# internal/progress/ — the Progress Tree: canonical durable task state

> 🌐 English | [繁體中文](README.zh-TW.md)

Implements the Progress Tree domain service (Constitution §6; Auspex_ADD.md §18 — the ADD now lives at
[docs/design/Auspex_ADD.md](../../docs/design/Auspex_ADD.md)). The Progress Tree is the canonical durable
task state: conversation context and an agent's own claim of "done" are never the source of truth.

Key pieces:

- **Stores** — `TaskStore` (`tasks`), `NodeStore` (`progress_nodes`), `EdgeStore` (`progress_edges`),
  `ArtifactStore` (`artifacts` evidence rows with their `validation_status`). `NodeStore` never persists a
  status change without calling `ValidateTransition` first.
- **State machine** (`statemachine.go`) — the fixed `domain.ProgressNodeStatus` enum (`pending`, `ready`,
  `in_progress`, `checkpointing`, `paused`, `completed`, `failed`, `skipped`, `blocked`;
  `internal/domain/status.go`, Constitution §6.4) with a frozen transition table; no ad hoc statuses.
- **`CompleteNode`** (`complete_node.go`) — the evidence-gated atomic completion protocol: idempotency check
  against the `node_completions` ledger (`idempotency.go`; same key + same payload replays the prior result,
  a different payload under the same key is a conflict, never a silent merge — Constitution §6.6), staging of
  evidence to a content-addressed copy (`stager.go`), verification via [`../artifacts/`](../artifacts/)
  validators, then one SQLite transaction that transitions the node, commits artifact rows, and seals and
  inserts a State Checkpoint manifest via [`../statecheckpoint/`](../statecheckpoint/) (Constitution §6.3).
  Completion never trusts the agent's own claim: evidence must exist and pass validators (§6.2). Events are
  published only after the transaction commits.
- **`Reconciler`** (`reconcile.go`) — startup reconciliation for the staged-artifact-vs-DB crash window
  (ADD §18.9); a read-only scan that surfaces orphaned staged evidence rather than mutating state.
- **`Service`** (`service.go`) — the concrete implementation of the frozen `app.ProgressTreeService` port
  (`internal/app/ports.go`), composing the pieces above; DTO translation only, no new logic.

Crash recovery is proven by phase-level crash injection (`complete_node_crash_test.go` stops the protocol
after each named `Phase` and asserts reconciliation holds — Constitution §6.5).
