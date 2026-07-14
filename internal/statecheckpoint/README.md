# internal/statecheckpoint/ — State Checkpoint manifests: build, seal, verify, persist

> 🌐 English | [繁體中文](README.zh-TW.md)

A State Checkpoint is the durable, replayable snapshot of a task's Progress Tree at a semantic boundary
(Auspex_ADD.md §18.8 and Appendix B — the ADD now lives at
[docs/design/Auspex_ADD.md](../../docs/design/Auspex_ADD.md); wire schema `auspex.state-checkpoint.v1`).
This package owns the manifest's Go shape, its deterministic JSON serialization, its integrity checksum,
and the `state_checkpoints` store — not the decision of *when* to checkpoint, which belongs to
[`../progress/`](../progress/)'s `CompleteNode` protocol (Constitution §6.3: every node completion creates
a State Checkpoint in the same atomic operation).

Key pieces:

- **`Manifest`** (`manifest.go`) — the Appendix B document; `IntegritySHA256` is declared last and excluded
  from its own digest.
- **`Build`** (`build.go`) — assembles an unsealed manifest from a `BuildInput` snapshot the caller already
  holds; deliberately does not import `internal/progress` (the dependency points the other way).
- **`Digest` / `Seal` / `Marshal` / `Verify`** (`serialize.go`) — the digest is SHA-256 over the canonical
  JSON encoding with `IntegritySHA256` zeroed; field order is the fixed struct declaration order, so the
  encoding is reproducible. `Verify` always recomputes the digest from the manifest's own content and
  compares it to the stored `integrity_sha256` — the stored value is checked, never trusted.
- **`Store`** (`store.go`) — CRUD over `state_checkpoints` (`migrations/0023_state_checkpoints.sql`); the
  row duplicates a queryable subset of `manifest_json`.
- **`Service`** (`service.go`) — implements the frozen `app.StateCheckpointService` port: `Create` (a
  standalone, ad hoc snapshot outside any node completion), `LoadLatest`, `Snapshot`, `Verify`.
- **`Reconciler`** (`reconcile.go`) — startup reconciliation (ADD §18.9): a read-only integrity scan that
  recomputes every row's digest, checks the schema version, and cross-checks manifest IDs against row
  columns. It repairs nothing by construction — the only durable write (`Store.Insert`) is a single atomic
  SQL statement, so no half-written row state is reachable.
