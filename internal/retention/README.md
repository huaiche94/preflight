# internal/retention/ — ADR-046 tiered telemetry retention: hot window → rollup → gzip archive → verified delete

> 🌐 English | [繁體中文](README.zh-TW.md)

Package `retention` implements
[ADR-046](../../docs/adr/0046-tiered-telemetry-retention.md)'s three-tier retention
and is the engine behind `auspex gc` (schema-versioned output `auspex.gc.v1`). The
package contract is the package comment at the top of `engine.go` (no separate
`doc.go`).

The three tiers:

1. **Hot raw window** (`policy.go`) — raw rows younger than the window are never
   touched. Default `DefaultRetentionDays = 90`; `Policy.Cutoff` uses strict
   "older than" so a row exactly at the cutoff is retained.
2. **Rollup summary tables** (`rollup.go`, migration `0060_retention.sql` in the
   ADR-046-assigned 0060–0069 range — see
   [../storage/sqlite/migrations/README.md](../storage/sqlite/migrations/README.md)) —
   before raw rows leave the hot tier, `usage_rollups_daily` and
   `calibration_samples` (the prediction-vs-actual pairs M13 calibration needs)
   are written in the same delete transaction.
3. **Gzip JSONL archive, then delete — fail-closed** (`archive.go`) — expired rows
   are written one JSON object per row, full column fidelity, to
   `<data-dir>/archive/<table>/<YYYY-MM>/…jsonl.gz` with the temp-file → fsync →
   rename discipline from [`../repocheckpoint`](../repocheckpoint/)'s
   `atomicwrite.go`, then re-opened and re-verified (row count + SHA-256) before
   any delete runs.

Entry point: `Engine.Run(ctx, RunRequest) (RunResult, error)`. The pass is strictly
ordered — select, archive, verify, and only then delete all classes plus rollups in
one `app.TxRunner.WithTx` transaction with affected-row counts checked; any earlier
failure leaves every raw row untouched and records a failed run in
`retention_runs`. A dry run performs selection only. Dependencies are the frozen
`domain.Clock`/`domain.IDGenerator` ports plus `*sqlite.DB`
([../storage/sqlite/README.md](../storage/sqlite/README.md)), so tests are
deterministic.

Retention is cross-cutting (it archives and deletes rows across every role's
tables), owned by no vertical-slice role, and adds no frozen port — gc is an
internal maintenance concern behind the CLI.
