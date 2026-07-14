# internal/hooks/ — provider lifecycle-hook payload handling

> 🌐 English | [繁體中文](README.zh-TW.md)

Per-provider packages that parse native lifecycle-hook stdin payloads and
encode the provider-compatible stdout responses those hooks expect back.
[claude/](claude/) is the only subpackage today; this directory holds no
Go files of its own and no doc.go.

The hook entry path around this directory:

1. The provider invokes `auspex hook claude <event>` (command tree in
   internal/cli/hook.go), which reads the full raw payload from stdin and
   never logs or echoes it.
2. The command calls the matching `orchestrator.Handle*` function
   (internal/orchestrator/hooks.go), which parses via this directory's
   packages (status-line parsing lives in
   [../providers/claude/](../providers/claude/)), normalizes into
   `pkg/protocol/v1.Event`s via
   [../telemetry/claude/](../telemetry/claude/), optionally persists
   them, and runs an evaluation where the hook semantics call for a
   decision.
3. The command writes a syntactically valid provider-compatible JSON
   response to stdout and exits 0 in every case except a genuine
   command-usage error (e.g. unreadable stdin).

Hooks fail open: an Auspex failure never blocks the provider session. A
malformed payload or internal error yields the safe fallback response
(e.g. `claude.FallbackAllowResponse()`), never a crash or missing
response, and a semantic "block" decision is content in the response
body, never a non-zero process exit — so the provider's hook runner
cannot mistake an ordinary block for Auspex crashing (handler contract in
internal/orchestrator/hooks.go; ADD §17.5's telemetry-unavailable rule —
`Auspex_ADD.md`, now at
[docs/design/Auspex_ADD.md](../../docs/design/Auspex_ADD.md)).
