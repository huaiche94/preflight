# internal/storage/sqlite/ — SQLite runtime: connection, pragmas, transactions, and the migration engine

> 🌐 English | [繁體中文](README.zh-TW.md)

Package `sqlite` is Auspex's SQLite runtime. The package contract is the package
comment at the top of `db.go` (there is no separate `doc.go`). The driver is
`modernc.org/sqlite` (pure Go, no CGO), per Auspex_ADD.md §1.4's tech-stack decision
(the ADD lives at [docs/design/Auspex_ADD.md](../../../docs/design/Auspex_ADD.md)).

Two halves:

- **`db.go` — connection/pragma/transaction engine.** `Open(ctx, path)` returns a
  `*DB` configured with the journal/durability pragmas every later role's storage
  code depends on. `DB` implements the frozen `app.TxRunner` port
  ([../../app/ports.go](../../app/ports.go)): `WithTx(ctx, fn)` runs `fn` inside one
  transaction, rolling back on a non-nil error. `Querier` /
  `QuerierFromContext(ctx, db)` let store code run the same queries inside or
  outside a `WithTx` callback. Opening an empty file yields a valid, correctly
  configured, completely empty database — `db.go` creates no schema.
- **`migrate.go` — forward-only migration engine (ADD §12.5).** `AllMigrations()`
  loads every `NNNN_name.sql` file embedded (via `go:embed`) from
  [`migrations/`](migrations/README.md); `LoadMigrationsFS` is deliberately strict
  (malformed filenames and duplicate versions are hard errors, never skips).
  `DB.Migrate` applies pending migrations as a **set difference** against the
  applied rows in `schema_migrations` — not as "everything above MAX(version)" —
  inside a single `BEGIN IMMEDIATE` transaction, so backfilled, gap-numbered
  migrations are applied and concurrent migrators cannot race (see
  [migrations/README.md](migrations/README.md) for the range/backfill rules).
  If the database's highest applied version is newer than anything this binary
  knows, `Migrate` returns `ErrSchemaNewerThanBinary` and applies nothing; callers
  must fail closed to read-only diagnostics (ADD §12.5). `CurrentVersion` reports
  the highest applied version for diagnostics.

Neighbors: everything that persists goes through this package —
[`../../telemetry/claude`](../../telemetry/claude/)'s EventStore,
[`../../progress`](../../progress/), [`../../evaluation`](../../evaluation/),
[`../../scheduler`](../../scheduler/), and [`../../retention`](../../retention/)
all write through `DB.WithTx`.
