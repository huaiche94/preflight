# internal/telemetry/ — provider telemetry normalization and durable event persistence

> 🌐 English | [繁體中文](README.zh-TW.md)

This directory namespaces the per-provider telemetry pipelines that turn raw
provider payloads into Auspex's frozen wire event envelope,
`pkg/protocol/v1.Event` (Auspex_ADD.md §11.1; the ADD lives at
[docs/design/Auspex_ADD.md](../../docs/design/Auspex_ADD.md)), and persist them
idempotently. There is no Go code at this level.

Its single subpackage is [`claude/`](claude/README.md), which by its own package
contract is the **sole** path from Claude Code provider payloads into a
`v1.Event` — no other package in the repository constructs one from Claude
payloads.

The division of labor with its neighbors:

- [`../providers/claude`](../providers/) and [`../hooks/claude`](../hooks/) parse
  raw provider JSON into intermediate, privacy-safe Go structs (parse step only);
- `telemetry/claude` normalizes those structs into `v1.Event` values (including
  idempotency-key construction) and writes them to SQLite through
  [`../storage/sqlite`](../storage/sqlite/README.md)'s `WithTx` boundary;
- consumers read the resulting `events` rows (migration `0010_events.sql`) — they
  never re-parse provider payloads.
