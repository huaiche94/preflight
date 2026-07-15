# testdata/ — cross-package test fixtures

> 🌐 English | [繁體中文](README.zh-TW.md)

Fixtures shared by tests in several `internal/` packages. Per
[ADR-049](../docs/adr/0049-docs-reorg-bilingual.md) §Decision 4, the
leaf fixture directories deliberately get **no** `README.md` of their
own (`repositories/`'s predates the policy) — fixture contents are
load-bearing for tests, so this file documents the whole subtree. Do
not add, rename, or edit fixture files casually: e.g.
`internal/telemetry/claude/fixture_suite_test.go`'s privacy gate embeds
needle strings copied verbatim from the provider-event fixtures and
fails loudly when they drift.

## Subtree

- `checkpoints/state/` — `sample-manifest.json` is read by
  `internal/statecheckpoint/serialize_test.go`; regenerate it via
  `AUSPEX_GENERATE_FIXTURES=1` (`fixture_gen_test.go`). The
  `add-section-18-*.md` files are **test fixtures, not documentation**
  (never translated — ADR-049 §Decision 5): artifact-validator inputs
  for `internal/artifacts`' heading/fence checks.
- `checkpoints/repository/` — `sample-manifest.json`, a Repository
  Checkpoint manifest generated from an actual Capture run against a
  real temp repository, validated against its schema in
  [`../schemas/`](../schemas/README.md).
- `progress-trees/` — `sample-task.json`, transcribed from
  `Auspex_ADD.md` Appendix A (the file lives at
  [`../docs/design/Auspex_ADD.md`](../docs/design/Auspex_ADD.md)),
  validated against `../schemas/progress-tree.schema.json`.
- `provider-events/claude/` — raw Claude Code hook payloads
  (`statusline/`, `stop/`, `stopfailure/`, `userpromptsubmit/`) in
  normal / malformed / missing-field / unknown-field variants plus
  `.golden.json` responses; consumed by the `internal/hooks/claude`,
  `internal/telemetry/claude`, `internal/providers/claude`, and
  `internal/cli` test suites.
- `repositories/` — repository-*content* fixtures for
  `internal/repocheckpoint`/`internal/gitx`; see
  [`repositories/README.md`](repositories/README.md) (no real `.git`
  directory is checked in; tests build temp repos instead).
