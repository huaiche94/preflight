# internal/hooks/claude/ — Claude Code hook payload parsing and response encoding

> 🌐 English | [繁體中文](README.zh-TW.md)

Parses Claude Code's native lifecycle-hook stdin payloads and encodes the
provider-compatible stdout responses. There is no doc.go; the package
comment at the top of userpromptsubmit.go states the contract.

Auspex handles four Claude Code hook events, exposed as
`auspex hook claude ...` subcommands (internal/cli/hook.go):

- `user-prompt-submit` — `ParseUserPromptSubmit` → `UserPromptSubmitEvent`.
  Privacy-safe by construction: the raw prompt is reduced to a SHA-256
  hash, size signals, and derived `features.PromptFeatures` inside the
  parse call; raw text never survives the stack frame (Constitution §7
  rule 2). `EncodeUserPromptSubmitResponse` renders the allow/block
  response (ADD §22.3); `FallbackAllowResponse` is the fail-open body
  used when Auspex itself fails.
- `stop` — `ParseStop` → `StopEvent` (a clean turn/session stop).
- `stop-failure` — `ParseStopFailure` → `StopFailureEvent`, classifying
  the provider error into the frozen `domain.FailureClass` enum (the
  mapping is an Auspex heuristic, not a frozen contract).
- `statusline` — parsed by the sibling package
  [../../providers/claude/](../../providers/claude/), not here.

Session bootstrap (issue #17): before events are persisted or an
evaluation runs, the hook handlers idempotently register the session's
repositories/worktrees/provider_sessions rows from the payload's reported
directory (internal/orchestrator/sessionbootstrap.go — upserts onto
existing frozen unique constraints; fail-open, so a non-git directory or
SQL error is a silent no-op, never a hook failure).

This package only parses and encodes; persistence, evaluation, and the
forecast card are internal/orchestrator/hooks.go's layer. (ADD citations
refer to `Auspex_ADD.md`, now at
[docs/design/Auspex_ADD.md](../../../docs/design/Auspex_ADD.md).)
