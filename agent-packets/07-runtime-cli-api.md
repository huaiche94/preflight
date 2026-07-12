# A07 — Application Orchestration, CLI, Local API, and Vertical-Slice Wiring

## Model

A cheaper coding model is sufficient; use Fable for authorization/pause orchestration review.

## ADD ownership

§13 pipeline orchestration, §§23–24, operational subset of §28, Appendix F.

## Exclusive paths

```text
internal/orchestrator/**
internal/cli/**
internal/httpapi/**
internal/daemon/**
internal/app/wiring/**
internal/testutil/fakes/** (coordinate with A08)
docs/implementation/day1/A07.md
```

Do not edit `cmd/preflight/main.go`; A00/A01 integrate root wiring. Add command constructors under owned paths.

## Mission

Wire the frozen ports into an in-process-first application and expose the day-one flow through stable CLI/JSON contracts. HTTP daemon is secondary to a working CLI.

## P0 commands

```text
preflight version
preflight init
preflight hook claude statusline
preflight hook claude user-prompt-submit
preflight hook claude stop
preflight hook claude stop-failure
preflight evaluate
preflight decision allow
preflight decision deny
preflight checkpoint create
preflight progress show
preflight state show
preflight pause request
preflight pause cancel
preflight resume
preflight scheduler run-once
preflight status
preflight doctor
```

## Pipeline behavior

1. Receive provider-normalized or CLI input.
2. Resolve repository/worktree/session.
3. Load current Progress Tree and usage observations.
4. Snapshot lightweight Git state.
5. Evaluate through A05.
6. Apply policy.
7. For allow: produce provider-compatible response.
8. For block/checkpoint: persist evaluation and return stable decision ID/instructions.
9. `checkpoint create` calls A03 then A04 according to frozen transaction/orchestration contract.
10. `decision allow` issues one-time authorization.
11. Resubmitted prompt consumes authorization exactly once before allowing.
12. Stop/StopFailure completes outcome labeling.

## JSON and errors

- stable schema-versioned output;
- typed error code, message, retryable, details;
- no raw prompt in logs/errors;
- machine mode never emits decorative text to stdout;
- hook fallback remains syntactically valid when Preflight fails.

## HTTP stretch

Implement authenticated loopback endpoints only after CLI E2E passes. No SSE until the core loop is stable.

## Tests

- CLI golden tests;
- no-TTY behavior;
- malformed stdin;
- high-risk block and allow-once flow;
- second authorization replay rejected;
- checkpoint failure does not issue authorization;
- provider hook always receives valid response;
- process exit codes;
- in-process restart using same SQLite file.
