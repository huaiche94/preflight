# internal/telemetry/claude/ — the sole path from Claude Code payloads into the frozen v1.Event envelope, plus idempotent persistence

> 🌐 English | [繁體中文](README.zh-TW.md)

Package `claude` normalizes the intermediate structs produced by
[`../../providers/claude`](../../providers/) and [`../../hooks/claude`](../../hooks/)
(StatusLineSnapshot, UserPromptSubmitEvent, StopEvent, StopFailureEvent) into the
frozen `pkg/protocol/v1.Event` envelope (Auspex_ADD.md §11.1, CONTRACT_FREEZE.md;
the ADD lives at [docs/design/Auspex_ADD.md](../../../docs/design/Auspex_ADD.md)).
The package contract is the package comment at the top of `normalizer.go` (no
separate `doc.go`): this is the only package in the repository that constructs a
`v1.Event` from Claude payloads, and it only ever emits `EventType` values already
defined in `pkg/protocol/v1`'s closed taxonomy.

Two halves:

- **`normalizer.go` / `managedrun.go` — normalization.** `Normalizer`
  (`NewNormalizer(clock, ids)`) exposes `NormalizeStatusLine`,
  `NormalizeUserPromptSubmit`, `NormalizeStop`, `NormalizeStopFailure`, and
  `NormalizeManagedRun`. This package owns the exact idempotency-key digest
  algorithm (CONTRACT_FREEZE.md freezes the field; the owning provider role defines
  the digest). `managedrun.go` handles `auspex run`'s terminal outcome
  ([`../../managed`](../../managed/) parses the stream-json lines into
  `ManagedRunOutcome` and hands it here): one terminal turn event plus, when the
  provider's result line carried usage, a turn-scoped `provider.usage.observed`
  event — exact one-turn attribution, unlike the cumulative statusline snapshot.
- **`store.go` — persistence.** `EventStore` (`NewEventStore(db)`) writes events to
  SQLite durably and idempotently, keyed by `Event.IdempotencyKey`, always through
  [`../../storage/sqlite`](../../storage/sqlite/README.md)'s `WithTx` /
  `app.TxRunner` boundary (`Persist`, `PersistAll`).

Privacy: raw prompt text never reaches this package — upstream parsers retain only
hashes and derived signals, and `Event.Payload` is populated after redaction per
the provider role's own contract.
