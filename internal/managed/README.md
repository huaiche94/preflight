# internal/managed/ — the managed one-shot runner behind `auspex run`

> 🌐 English | [繁體中文](README.zh-TW.md)

The CLI-free core of `auspex run` (issue #8's MVP increment; ADD §8.1 —
[`docs/design/Auspex_ADD.md`](../../docs/design/Auspex_ADD.md)). See [`doc.go`](doc.go) for
the package contract and the full list of deliberate exclusions. Three concerns, nothing else:

1. **Pre-prompt gate** — the same production evaluate/decide path the `UserPromptSubmit` hook runs, via [`internal/orchestrator`](../orchestrator/README.md)'s `EvaluateManagedPrompt`, applied before the provider process exists: a BLOCK decision means the provider is never spawned.
2. **Provider subprocess lifecycle** — [`run.go`](run.go)'s `Runner.Run` spawns `claude -p <prompt> --output-format stream-json --verbose` argv-only (never a shell string); [`stream.go`](stream.go) parses the stream-json output defensively and fail-open — an unrecognized line becomes a skip count, never a crash. Result/message content is never retained (length-only), per Constitution §7's privacy rule.
3. **Outcome attribution** — the terminal outcome is normalized through `internal/telemetry/claude` into the frozen event envelope and best-effort persisted through the same seams the hook path uses, keyed to one `TurnID`.

Claude is the only supported provider in this increment (`ProviderClaude`; the Codex managed
adapter is ADD milestone M7). The CLI half lives in
[`internal/cli/run.go`](../cli/run.go). `testdata/` holds stream fixtures.

Not implemented here (issue #8's later increments): managed shell mode — `auspex shell`,
ADD §8.2, scheduled as ADD milestone M11 — plus turn interrupt / safe-point pause during a
run, daemon/event-stream integration, verified auto-resume, and live per-message usage
modeling. The spawned process runs to completion; context cancellation kills it.
