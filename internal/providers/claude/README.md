# internal/providers/claude/ — Claude Code status-line payload parser

> 🌐 English | [繁體中文](README.zh-TW.md)

Parses Claude Code's provider-native status-line JSON snapshots (ADD
§22.5; `Auspex_ADD.md`, now at
[docs/design/Auspex_ADD.md](../../../docs/design/Auspex_ADD.md)) into
intermediate Go structs. There is no doc.go; the package comment at the
top of statusline.go states the contract: this package stops at the
parsing step, and normalizing into the frozen `pkg/protocol/v1.Event`
envelope is [../../telemetry/claude/](../../telemetry/claude/)'s job.

Entry point: `ParseStatusLine(raw []byte) (StatusLineSnapshot, error)`.
The snapshot projects into frozen domain shapes via
`ContextObservation` (context-window usage), `QuotaObservations` (one
`domain.QuotaObservation` per rate-limit window, however many arrive —
issue #21), and `WeeklyLimitUsedPercent` (the statusline's weekly
segment).

Parsing discipline:

- every optional field is a pointer; nil means unknown, never a
  substituted zero (ADD §22.10; CONTRACT_FREEZE.md "Unknown/null
  semantics");
- unknown fields are tolerated at any nesting level;
- an unrecognized encoding of a recognized field degrades that field to
  unknown rather than failing the whole snapshot (`flexTimestamp`; the
  issue #27 incident);
- a syntactically invalid payload returns a `domain.Error` with
  `ErrCodeValidation`, so the hook wrapper can fall back to a safe
  response instead of crashing.

Consumed by internal/orchestrator/hooks.go behind
`auspex hook claude statusline`. The sibling hook-payload parsers
(UserPromptSubmit/Stop/StopFailure) live in
[../../hooks/claude/](../../hooks/claude/).
