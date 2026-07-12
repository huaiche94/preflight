# claude-provider — Day-1 progress artifact

Role: `claude-provider` (`agents/claude-provider.md`).
Wave: DAG parallel-root nodes `claude-provider-01`, `claude-provider-02`, `claude-provider-03` (per `docs/implementation/day1/EXECUTION_DAG.md`).
Branch: `day1/claude-provider` (local only — not pushed, not merged into `main`).

This file is the durable, validator-checked evidence for this wave per
Constitution §6 rule 7 ("this same discipline applies, by analogy, to the
meta-level progress artifacts"). It is updated, not conversation memory.

---

```yaml
node: claude-provider-01
status: completed
artifacts:
  - internal/providers/claude/statusline.go
  - internal/providers/claude/statusline_test.go
  - testdata/provider-events/claude/statusline/normal.json
  - testdata/provider-events/claude/statusline/missing_fields.json
  - testdata/provider-events/claude/statusline/unknown_fields.json
  - testdata/provider-events/claude/statusline/high_usage.json
  - testdata/provider-events/claude/statusline/malformed.json
validation:
  - "gofmt -l internal/providers/claude internal/hooks/claude  -> clean"
  - "go build ./internal/providers/claude/... ./internal/hooks/claude/...  -> ok"
  - "go vet ./internal/providers/claude/... ./internal/hooks/claude/...  -> ok"
  - "go test ./internal/providers/claude/... -run StatusLine -v  -> PASS (5 fixture subtests + 4 unit tests)"
commit: 69462cce853665262ead9dad5b7f998c00b9bcd4
next_action: claude-provider-04 (blocked - not started this wave; normalizes parser output into pkg/protocol/v1.Event, depends on all of 01/02/03)
assumptions:
  - "No live Claude Code account was available this wave (per task brief). All 5 status-line fixtures were hand-constructed from agents/claude-provider.md and Preflight_ADD.md §22.5/§22.10's described field names (session_id, model.id/display_name, workspace.current_dir/project_dir, context_window.total_input_tokens/total_output_tokens/context_window_size/used_percentage, cost.total_cost_usd/total_duration_ms/total_api_duration_ms/total_lines_added/total_lines_removed, rate_limits.five_hour/seven_day.used_percentage/resets_at) rather than captured from a real session. Field names may not match the live API exactly; ADD Appendix H's official doc URLs (code.claude.com/docs/en/statusline) were not fetched (no external network access this wave) and could not be cross-checked."
  - "The 5 fixture files (normal, missing_fields, unknown_fields, high_usage, malformed) were found already present and untracked in the worktree from a previously interrupted attempt. They were inspected, found to align with ADD §22.5's field list exactly, and kept as-is rather than replaced."
  - "Confidence classification for ContextObservation/QuotaObservation (exact/high/medium/unavailable) is this role's own heuristic - CONTRACT_FREEZE.md freezes the Measurement/Confidence *type*, not per-provider confidence assignment rules."
  - "Five-hour/seven-day quota observations return nil (not a zero-value struct) when both percentage and reset timestamp are absent, per ADD §22.10 (\"Absence => unknown\")."
blockers: []
```

```yaml
node: claude-provider-02
status: completed
artifacts:
  - internal/hooks/claude/userpromptsubmit.go
  - internal/hooks/claude/userpromptsubmit_test.go
  - testdata/provider-events/claude/userpromptsubmit/normal.json
  - testdata/provider-events/claude/userpromptsubmit/missing_fields.json
  - testdata/provider-events/claude/userpromptsubmit/unknown_fields.json
  - testdata/provider-events/claude/userpromptsubmit/empty_prompt.json
  - testdata/provider-events/claude/userpromptsubmit/malformed.json
  - testdata/provider-events/claude/userpromptsubmit/response_allow.golden.json
  - testdata/provider-events/claude/userpromptsubmit/response_block.golden.json
validation:
  - "gofmt -l internal/providers/claude internal/hooks/claude  -> clean"
  - "go build ./internal/providers/claude/... ./internal/hooks/claude/...  -> ok"
  - "go vet ./internal/providers/claude/... ./internal/hooks/claude/...  -> ok"
  - "go test ./internal/hooks/claude/... -run UserPromptSubmit -v  -> PASS (5 fixture subtests + 5 unit/golden tests, including raw-prompt-never-persisted privacy assertion)"
commit: 69462cce853665262ead9dad5b7f998c00b9bcd4
next_action: claude-provider-04 (blocked - not started this wave)
assumptions:
  - "Exact stdin field names for the UserPromptSubmit hook payload (session_id, transcript_path, cwd, hook_event_name, prompt) are constructed from Claude Code hook conventions described in the packet and ADD §22.3/Appendix E.3, not captured from a live payload. ADD Appendix H names code.claude.com/docs/en/hooks as the authoritative source but it was not fetched this wave (no external network access)."
  - "The block-response JSON shape (decision/reason/hookSpecificOutput.hookEventName/additionalContext) is taken verbatim from ADD §22.3's example. The allow-response shape (bare `{}`, omitting the decision key entirely) is this role's own interpretation of 'hook protocol convention of omitting fields that don't apply' since the ADD does not show an explicit allow example - flagged here as a judgment call, not a frozen contract, in case contract-integrator wants to pin it down explicitly before claude-provider-06 (the CLI wrapper) ships."
  - "Approximate token counting (~4 bytes/token) is a coarse, explicitly-approximate heuristic per the packet's Privacy section ('approximate token count only') - not calibrated against Claude's actual tokenizer."
  - "On internal/parse failure, the hook wrapper is expected to call FallbackAllowResponse() to fail open (never block the user due to a Preflight bug) - this fail-open behavior for a hook-path *operational* failure is consistent with CONTRACT_FREEZE.md's fail-open/fail-closed split (operational observation failures may fail open) but the fallback wiring itself belongs to claude-provider-06's CLI wrapper, not this node - only the primitive (FallbackAllowResponse) is delivered here."
blockers: []
```

```yaml
node: claude-provider-03
status: completed
artifacts:
  - internal/hooks/claude/stop.go
  - internal/hooks/claude/stop_test.go
  - testdata/provider-events/claude/stop/normal.json
  - testdata/provider-events/claude/stop/missing_fields.json
  - testdata/provider-events/claude/stop/unknown_fields.json
  - testdata/provider-events/claude/stop/malformed.json
  - testdata/provider-events/claude/stopfailure/rate_limit.json
  - testdata/provider-events/claude/stopfailure/overloaded.json
  - testdata/provider-events/claude/stopfailure/context_length.json
  - testdata/provider-events/claude/stopfailure/network.json
  - testdata/provider-events/claude/stopfailure/unknown_category.json
  - testdata/provider-events/claude/stopfailure/missing_fields.json
  - testdata/provider-events/claude/stopfailure/malformed.json
validation:
  - "gofmt -l internal/providers/claude internal/hooks/claude  -> clean"
  - "go build ./internal/providers/claude/... ./internal/hooks/claude/...  -> ok"
  - "go vet ./internal/providers/claude/... ./internal/hooks/claude/...  -> ok"
  - "go test ./internal/hooks/claude/... -run 'Stop|StopFailure' -v  -> PASS (4 Stop fixture subtests + 7 StopFailure fixture subtests + error-message-not-retained privacy test + classifier unit test)"
commit: 69462cce853665262ead9dad5b7f998c00b9bcd4
next_action: claude-provider-04 (blocked - not started this wave)
assumptions:
  - "Exact StopFailure error-object shape (error.type/message/status_code, modeled loosely on Anthropic API error conventions: rate_limit_error, overloaded_error, invalid_request_error, connection_error) is hand-constructed - no frozen contract or live payload exists for this. ADD Appendix H names code.claude.com/docs/en/hooks as authoritative but it was not fetched this wave."
  - "classifyFailure's mapping from provider error type/status code to domain.FailureClass is this role's own heuristic, not a frozen contract: rate_limit_error/429 -> FailureProviderRateLimit; overloaded_error/529/5xx/api_error -> FailureProviderInternal; message containing 'too long'/context-length phrasing -> FailureContext; connection_error/network keywords -> FailureNetwork; permission_error/403 and authentication_error/401 -> FailurePermission; timeout_error -> FailureTimeout; anything unmatched -> FailureUnknown. This mapping should be revisited once real StopFailure payloads are observed against a live account, per the packet's stretch-goal note."
  - "StopHookActive (stop_hook_active) uses *bool so 'field absent' (nil, unknown) is distinguishable from 'field present and false' - modeled after Claude Code's documented stop-hook-loop-prevention field, but not verified against a live payload."
  - "Raw error message text is deliberately never retained on StopFailureEvent (only ErrorMessageLen, an int) even though the packet's Privacy section is written primarily about prompts - applied the same discipline defensively since provider error messages can echo back request content."
blockers: []
```
