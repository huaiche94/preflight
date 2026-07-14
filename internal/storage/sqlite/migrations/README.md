# internal/storage/sqlite/migrations/ — embedded, forward-only schema migrations with per-role number ranges

> 🌐 English | [繁體中文](README.zh-TW.md)

Every `NNNN_name.sql` file directly here is embedded into the binary (`go:embed`
in [../migrate.go](../migrate.go)) and applied by `DB.Migrate` ([../README.md](../README.md)).
Malformed filenames and duplicate versions are load-time errors, never skips.
Migrations are forward-only — no down migrations (Auspex_ADD.md §12.5; the ADD
lives at [docs/design/Auspex_ADD.md](../../../../docs/design/Auspex_ADD.md)).

## Number ranges (per [CONTRACT_FREEZE.md](../../../../docs/implementation/vertical-slice/CONTRACT_FREEZE.md) "Migration ranges" — assigned per role, not chronologically)

| Range | Owner |
|---|---|
| 0000–0009 | foundation |
| 0010–0019 | claude-provider |
| 0020–0029 | checkpoint Part A (progress/state) |
| 0030–0039 | checkpoint Part B (repository) |
| 0040–0049 | predictor |
| 0050–0059 | runtime Part A (pause/scheduler) |
| 0060–0069 | retention/gc (cross-cutting; assigned by [ADR-046](../../../../docs/adr/0046-tiered-telemetry-retention.md)) |

runtime Part B has no range; it adds no schema unless contract-integrator assigns one.

## Set-difference application and gap-numbered backfills

Because versions are assigned in ranges, a migration can land in git *after*
higher-numbered ranges were already applied to real databases (issue #22: `0045`
landed after `0050–0052` shipped). The #22 fix: `Migrate` computes pending work as
a **set difference** against all applied `schema_migrations` rows — never as
"everything above MAX(version)", which silently skipped such backfills forever.
As documented on `Migration.Version` in [../migrate.go](../migrate.go):
a backfilled migration's SQL executes on existing databases *after* higher-numbered
migrations have already run, so it must not depend on ordering relative to ranges
above its own (additive statements against its own domain's tables satisfy this
trivially). The fail-closed `ErrSchemaNewerThanBinary` check keys on the *maximum*
applied version only — an applied version below the binary's own maximum is a
backfill the binary predates, already applied and correctly ignored. One `Migrate`
call applies all pending migrations in a single `BEGIN IMMEDIATE` transaction.
