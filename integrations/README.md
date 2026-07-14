# integrations/ — provider integration configuration

> 🌐 English | [繁體中文](README.zh-TW.md)

Shipped configuration examples that wire an AI coding agent's own
extension points to the `auspex` binary. One subdirectory per
provider; today only one exists:

- [`claude/`](claude/README.md) — Claude Code hook and plugin wiring
  (`hooks.json`, `plugin.json`): routes UserPromptSubmit / Stop /
  StopFailure / statusline events through `auspex hook claude
  <event>`. Its README documents the file shapes, a recorded CLI
  subcommand-naming discrepancy, and the `--emit-line` status-line
  behavior.

The root [`README.md`](../README.md) Quick start points here for
wiring Auspex into Claude Code. The Go-side counterparts of these
files are `internal/hooks/claude` and `internal/telemetry/claude`;
the raw payload fixtures those packages test against live under
[`../testdata/`](../testdata/README.md) `provider-events/claude/`.
A future provider adapter (e.g. Codex, M7/M8, issue #9) would add a
sibling directory here for its shipped configuration.
