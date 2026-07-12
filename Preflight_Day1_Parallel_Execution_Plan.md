# Preflight Day-1 Parallel Implementation Plan

## 1. Decision

Do **not** assign one agent per ADD chapter. Chapters 1–35 mix product constraints, domain contracts, provider adapters, runtime behavior, testing, and delivery concerns. A chapter-per-agent split would create duplicate types, incompatible schemas, and severe merge conflicts.

Use one contract/integration agent and eight bounded-context agents. Every agent works in a dedicated Git worktree and owns a non-overlapping path set.

The day-one objective is a real vertical slice for Claude Code:

```text
Claude status-line / hook event
        ↓
normalized telemetry persisted in SQLite
        ↓
UserPromptSubmit preflight evaluation
        ↓
explainable risk decision
        ↓
ALLOW or BLOCK
        ↓
checkpoint + one-time authorization
        ↓
turn completion/failure outcome persisted
        ↓
pause state and durable wake job can be created/recovered
```

Day-one success is **not** “all ADD chapters implemented.” Success is proving the core loop while preserving the long-term architecture.

## 2. Day-one scope boundary

### Must merge

1. Repository bootstraps and builds on the target OS.
2. Shared domain/event/store contracts are frozen once.
3. Claude status-line and hook payloads are parsed from fixtures.
4. Quota/context/turn telemetry is stored idempotently.
5. A deterministic scope/risk predictor returns P50/P90 estimates, reason codes, confidence, and `calibrated=false` during cold start.
6. `UserPromptSubmit` can allow or block before the main turn.
7. High-risk flow can create State Checkpoint and Repository Checkpoint evidence.
8. Progress node completion is rejected without durable artifact evidence.
9. One-time authorization prevents prompt replay.
10. Pause state machine and durable wake job survive process restart in tests.
11. CLI exposes the vertical slice.
12. One end-to-end fixture test covers the entire flow.

### Explicitly deferred

- Chapter 21 full Codex integration.
- Chapter 25 VS Code extension.
- ML training, personalization, ONNX, and Python research runtime.
- Fully calibrated probability claims.
- Public package-manager release automation beyond basic CI scaffolding.
- OS wake-up from shutdown/suspend.
- Multi-provider routing.
- External provider plugin protocol.
- Production-grade automatic invocation of a sleeping Claude process if provider lifecycle support is not proven by fixture-backed tests.

## 3. Model allocation

Use Fable where correctness and state-machine reasoning dominate:

- Agent 00: contract freeze and integration.
- Agent 05: predictor/policy.
- Agent 06: pause/scheduler.
- Final architecture and race-condition review.

Use a cheaper coding model for deterministic implementation work when available:

- foundation/config/SQLite;
- Claude JSON parsing and hook fixtures;
- Git snapshot/checkpoint plumbing;
- CLI/API wiring;
- CI and test harness.

If every worker must use Fable, do not give the complete 161 KB ADD to every worker. Give each worker only:

1. this common plan;
2. `CONTRACT_FREEZE.md` after Agent 00 produces it;
3. its assigned ADD chapters;
4. its agent packet.

## 4. Parallel topology

```text
                         ┌─────────────────────────────┐
                         │ A00 Contract + Integrator   │
                         │ freeze types / ports / IDs  │
                         └──────────────┬──────────────┘
                                        │
       ┌────────────────┬───────────────┼───────────────┬────────────────┐
       ▼                ▼               ▼               ▼                ▼
 A01 Foundation    A02 Claude      A03 Progress     A04 Repository   A05 Predictor
 Config/SQLite     Telemetry       State CP         Checkpoint       Policy
       │                │               │               │                │
       └────────────────┴───────────────┴───────────────┴────────────────┘
                                        │
                                        ▼
                              A06 Pause + Scheduler
                                        │
                                        ▼
                              A07 CLI/API Orchestration
                                        │
                                        ▼
                              A08 QA/Security/CI/E2E
                                        │
                                        ▼
                              A00 final integration
```

A06 and A07 may start immediately against frozen interfaces and fakes. They must not wait for concrete implementations.

## 5. Chapter 1–35 ownership map

| ADD chapter | Primary owner | Day-one treatment |
|---|---|---|
| 1–8 | A00 | Read as constraints; produce contract freeze and scope guardrails. |
| 9 | A00 | Freeze domain values, IDs, statuses, service ports, provider capability types. |
| 10 | A01, governed by A00 | Bootstrap exact package layout; no speculative packages. |
| 11 | A00 + A02 | A00 freezes envelope; A02 implements Claude normalization/ingestion. |
| 12 | A01 + feature owners | A01 owns DB engine/core migration; agents own allocated migration ranges. |
| 13 | A07 + A05 | A05 evaluates; A07 orchestrates authorization and outcome lifecycle. |
| 14–17 | A05 | Scope, token/quota/runway baseline, risk, policy. |
| 18 | A03 | Progress Tree and State Checkpointing. |
| 19 | A04 | Repository Checkpoint and recovery. |
| 20 | A06 | Graceful Pause and durable wake jobs. |
| 21 | Deferred | No Codex production adapter tomorrow. Fixtures/interfaces only if needed. |
| 22 | A02 | Claude plugin/hooks/status-line/native event normalization. |
| 23–24 | A07 | In-process first; thin daemon/API/CLI surface. |
| 25 | Deferred | No VS Code implementation tomorrow. |
| 26 | A01 | Config model and defaults needed by vertical slice only. |
| 27 | A00 + A08 | Privacy/security constraints and tests. |
| 28 | A07 + A08 | Typed errors, logging, recovery, doctor baseline. |
| 29 | Every agent + A08 | Unit tests owned locally; A08 owns cross-package/E2E tests. |
| 30 | A01 + A08 | Basic OSS files and CI; full release matrix deferred. |
| 31–32 | A00 | Milestone acceptance and final DoD gate. |
| 33 | A00 | ADR compliance; only A00 edits accepted ADRs. |
| 34 | All | Execution and durable-progress contract. |
| 35 | Split by appendix | A02 owns Claude templates; A03 A/B; A04 D; A06 C; A07 F; A08 test/reference validation. |

## 6. Contract-freeze gate

No feature agent should invent a competing domain type. A00 must first commit `docs/implementation/day1/CONTRACT_FREEZE.md` and compileable skeletons for:

- UUIDv7-style ID aliases or wrappers;
- `Session`, `Turn`, `Task`, `ProgressNode`, `ArtifactReference`;
- `StateCheckpoint`, `RepositoryCheckpoint`, `PauseRecord`, `WakeJob`;
- `UsageObservation`, `QuotaObservation`, `ContextObservation`;
- `Evaluation`, `PredictionResult`, `PolicyDecision`, `Authorization`;
- event envelope and event-type constants;
- failure classes and typed error codes;
- service interfaces from ADD §9.9;
- provider capability and hook normalization types from ADD §9.10;
- `Clock` and `IDGenerator` interfaces;
- SQLite transaction boundary conventions;
- JSON/YAML field names and schema-version strings.

### Immutable day-one rules

1. Raw prompts are not persisted by default.
2. Uncalibrated scores are never labeled as probabilities.
3. State Checkpoint and Repository Checkpoint are distinct entities.
4. Progress Tree is canonical task state.
5. Node completion requires durable artifact evidence.
6. Pause full guarantees apply only to managed execution; native-hook behavior is degraded and explicit.
7. All persistence writes are idempotent by stable event/operation ID.
8. Clock, process execution, and ID generation are injectable.
9. Provider wire payloads never leak into domain/storage rows.
10. Operational observation failure may fail open; state-integrity failure fails closed.

## 7. Shared-file and dependency policy

### Files owned only by A00

```text
Preflight_ADD.md
AGENTS.md
internal/domain/**
internal/app/ports.go
pkg/protocol/v1/**
docs/adr/**
docs/implementation/day1/CONTRACT_FREEZE.md
```

### Files owned only by A01

```text
go.mod
go.sum
cmd/preflight/main.go
internal/config/**
internal/paths/**
internal/buildinfo/**
internal/storage/sqlite/db.go
internal/storage/sqlite/migrate.go
internal/storage/sqlite/migrations/0000-0009_*.sql
Makefile
Taskfile.yml
.golangci.yml
```

No other agent edits `go.mod` or `go.sum`. Dependency requests go into its progress file for A01/A00 to apply.

### Migration allocation

```text
0000–0009  A01 core/session/config
0010–0019  A02 telemetry/provider events
0020–0029  A03 progress/state checkpoints
0030–0039  A04 repository checkpoints
0040–0049  A05 evaluations/predictions/authorizations
0050–0059  A06 pause/wake jobs
```

A07 does not add schema unless A00 explicitly assigns a range.

### Unit-test ownership

Each feature agent writes unit tests under its own package. A08 owns only:

- `internal/integrationtest/**`;
- `testdata/e2e/**`;
- cross-component race/restart/security tests;
- CI workflows.

## 8. Git worktree setup

Recommended branches:

```bash
git worktree add ../preflight-a00 -b day1/a00-contract
git worktree add ../preflight-a01 -b day1/a01-foundation
git worktree add ../preflight-a02 -b day1/a02-claude
git worktree add ../preflight-a03 -b day1/a03-progress-state
git worktree add ../preflight-a04 -b day1/a04-repo-checkpoint
git worktree add ../preflight-a05 -b day1/a05-predictor-policy
git worktree add ../preflight-a06 -b day1/a06-pause
git worktree add ../preflight-a07 -b day1/a07-runtime-surface
git worktree add ../preflight-a08 -b day1/a08-qa
```

A00 lands the contract commit first. Every other branch rebases onto that exact commit before writing production code.

## 9. Durable coordination artifacts

Every agent owns exactly one progress artifact:

```text
docs/implementation/day1/A00.md
...
docs/implementation/day1/A08.md
```

After every logical node, the agent must write:

```yaml
node: A05-03
status: completed
artifacts:
  - internal/predictor/heuristic/predictor.go
  - internal/predictor/heuristic/predictor_test.go
validation:
  - go test ./internal/predictor/...
commit: <sha>
next_action: A05-04 implement policy reason codes
assumptions: []
blockers: []
```

Conversation-only progress does not count.

## 10. Merge order

Use this order even when implementation occurs in parallel:

1. A00 contract freeze.
2. A01 foundation/core SQLite.
3. A02, A03, A04, A05 in any order after tests pass.
4. A06 pause/scheduler.
5. A07 CLI/API/orchestration.
6. A08 CI/E2E/security tests.
7. A00 final reconciliation and architecture review.

A00 should cherry-pick or merge **whole reviewed commits**, not copy generated code manually.

## 11. Final day-one demo

The final branch should demonstrate:

```bash
preflight version
preflight init

# Feed a Claude status-line fixture.
preflight hook claude statusline < testdata/provider-events/claude/statusline-high-usage.json

# Evaluate a prompt without persisting raw prompt text.
preflight evaluate \
  --provider claude \
  --prompt-file testdata/e2e/high-risk-prompt.txt \
  --json

# Simulate UserPromptSubmit; high risk returns provider-compatible block output.
preflight hook claude user-prompt-submit \
  < testdata/provider-events/claude/user-prompt-submit-high-risk.json

# Create both state and repository evidence.
preflight checkpoint create --evaluation <id> --json

# Issue a one-time allow decision and consume it once.
preflight decision allow --evaluation <id> --json

# Persist a normal turn completion or rate-limit failure fixture.
preflight hook claude stop < testdata/provider-events/claude/stop.json
preflight hook claude stop-failure < testdata/provider-events/claude/stop-failure-rate-limit.json

# Exercise pause/wake durability without depending on wall-clock sleep.
preflight pause request --session <id> --reason runway --json
preflight scheduler run-once --at <timestamp> --json
preflight status --json
```

## 12. Cut order when integration slips

Cut features in this order:

1. HTTP daemon transport; keep in-process CLI.
2. SSE/live dashboard.
3. Actual provider-process auto-resume; keep durable wake state and fake resumer contract test.
4. Repository restore; keep create/verify.
5. Native status-line composition with pre-existing user command; document manual setup.
6. Full OSS release matrix.

Never cut:

- prompt privacy default;
- authorization replay protection;
- artifact evidence requirement;
- checkpoint atomicity;
- calibrated/uncalibrated distinction;
- pause state durability;
- provider payload fixtures and contract tests.

## 13. Final Fable review prompt

```text
Review the merged Preflight day-one vertical slice against Preflight_ADD.md.

Focus only on:
1. domain/schema contradictions;
2. raw-prompt or secret persistence;
3. authorization replay/staleness;
4. State Checkpoint atomicity and artifact evidence;
5. Repository Checkpoint race/path traversal handling;
6. quota/risk values incorrectly presented as calibrated probabilities;
7. pause/resume state-machine races, duplicate wake, stale lease, and repository conflict;
8. Claude hook output compatibility and unknown-field tolerance;
9. missing restart and idempotency tests.

Do not redesign the project or add future abstractions. Produce a severity-ranked
file-level report, then fix only P0/P1 findings and run the complete test suite.
```


# Agent Packets

The full text of each agent packet lives as a standalone file under
`agent-packets/`, so a single file can be handed to an isolated agent/worktree
without the rest of this plan (see §3 above and `agent-packets/README.md`).

**`agent-packets/0X-*.md` is the single source of truth for packet content.**
This section is an index only — edit packet details there, not here, so the
two never drift out of sync.

| Agent | Packet file | Model | Mission |
|---|---|---|---|
| A00 | [`agent-packets/00-contract-integrator.md`](agent-packets/00-contract-integrator.md) | Fable | Freeze compile-time/persistence contracts; integrate reviewed branches at the end. |
| A01 | [`agent-packets/01-foundation-config-sqlite.md`](agent-packets/01-foundation-config-sqlite.md) | Cheaper model; Fable for migration/recovery review | Buildable Go application foundation and the SQLite runtime every other package depends on. |
| A02 | [`agent-packets/02-claude-telemetry-hooks.md`](agent-packets/02-claude-telemetry-hooks.md) | Fable for hook semantics; cheaper model for parsers/fixtures | Fixture-backed Claude Code hook/status-line normalization into frozen Preflight events. |
| A03 | [`agent-packets/03-progress-state-checkpoint.md`](agent-packets/03-progress-state-checkpoint.md) | Fable | Make Progress Tree the canonical durable task state; no completion without verified artifact evidence. |
| A04 | [`agent-packets/04-repository-checkpoint.md`](agent-packets/04-repository-checkpoint.md) | Cheaper model; Fable for path/race/security review | Capture and verify repository evidence without mutating the active branch. |
| A05 | [`agent-packets/05-predictor-policy.md`](agent-packets/05-predictor-policy.md) | Fable | Deterministic, explainable, cold-start-safe predictor/policy/authorization loop. |
| A06 | [`agent-packets/06-graceful-pause-scheduler.md`](agent-packets/06-graceful-pause-scheduler.md) | Fable | Provider-neutral pause/resume state machine and durable wake scheduling. |
| A07 | [`agent-packets/07-runtime-cli-api.md`](agent-packets/07-runtime-cli-api.md) | Cheaper model; Fable for authorization/pause orchestration review | Wire frozen ports into an in-process-first app; stable CLI/JSON vertical slice. |
| A08 | [`agent-packets/08-qa-security-ci.md`](agent-packets/08-qa-security-ci.md) | Cheaper model for fixtures/CI; Fable for final adversarial review | Objective evidence that the vertical slice is safe, restartable, idempotent, provider-compatible. |

The contract-freeze template A00 fills in when it produces
`docs/implementation/day1/CONTRACT_FREEZE.md` lives at
[`agent-packets/CONTRACT_FREEZE_TEMPLATE.md`](agent-packets/CONTRACT_FREEZE_TEMPLATE.md).
