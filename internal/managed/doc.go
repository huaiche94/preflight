// Package managed implements the managed one-shot runner behind `auspex
// run` (ADD §8.1; GitHub issue #8, increment 1 — the MVP; extended to
// codex by issue #9 M7 Phase 1, ADD §21.8; extended by issue #122 with the
// M10 Graceful Pause auto-trigger, pausedrive.go). It owns three concerns
// plus that one optional trigger, and deliberately nothing else:
//
//  1. the pre-prompt gate: the same production evaluate/decide path the
//     UserPromptSubmit hook runs (internal/orchestrator's
//     EvaluateManagedPrompt over the shared evaluateSubmittedPrompt core),
//     applied BEFORE the provider process exists — a BLOCK decision means
//     the provider is never spawned at all, which is the one enforcement
//     capability native hook mode can only approximate (ADD §8.3 vs §8.1);
//  2. provider subprocess lifecycle: spawning the provider argv-only
//     (Constitution §7 rule 5 — never a shell command string; internal/
//     gitx.ExecRunner's established discipline) and defensively parsing
//     the resulting event stream — `claude -p <prompt> --output-format
//     stream-json --verbose` (stream.go) or `codex exec --json <prompt>`
//     (codexstream.go), per the provider spec table in provider.go;
//  3. outcome attribution: normalizing the run's terminal outcome into the
//     frozen event envelope via the provider's own telemetry normalizer
//     (internal/telemetry/claude.NormalizeManagedRun, internal/telemetry/
//     codex.NormalizeManagedExec) and best-effort persisting through the
//     same EventPersister/TxRunner seam the hook path uses, so a managed
//     run's telemetry lands in the same events table, keyed to one TurnID
//     from provider.turn.started through provider.usage.observed — ADD
//     §8.7's "exact completed usage: yes" cell for Claude managed
//     stream-json, and its §21.8 analog for codex exec.
//
// # Deliberate exclusions (issues #8/#9's later increments, not this one)
//
//   - `auspex shell` (ADD §8.2 managed shell mode) — no prompt loop exists;
//     the CLI has no shell command yet and this increment adds none.
//   - Codex-specific graceful interrupt (`turn/interrupt` over the app
//     server, ADD §21.2/§21.8 "Managed pause in exec mode") — issue #9
//     Phase 2. The M10 auto-pause trigger (issue #122, pausedrive.go) DOES
//     interrupt a running provider, via the process-level signal path
//     (SIGINT, kill escalation) that both providers' one-shot subprocesses
//     share; a protocol-level Codex interrupt is a later refinement of
//     that same seam, not a missing branch here.
//   - Daemon/event-stream integration — the runner is a plain in-process
//     CLI flow with no daemon dependency; the codex app-server persistent
//     subscription (ADD §21.2) is a different milestone.
//   - Auto-resume / session resume (ADD §8.1 "verified auto-resume";
//     `codex exec resume`) — a run is one shot; the provider session/
//     thread ID the stream reports is captured as attribution data only,
//     never driven.
//   - Per-message live usage modeling from the stream itself (ADD §8.1) —
//     the stream is parsed for the terminal result only; per-message usage
//     (claude) and item.* progress (codex) are skipped-but-counted, not
//     modeled. The #122 auto-pause trigger observes the session's
//     PERSISTED quota telemetry (the same events table the hooks/watcher
//     fill), not the live stream.
package managed
