# Agent Packets

This directory holds one **hand-off packet per Day-1 agent** (`A00`–`A08`).
Each file is meant to be given, on its own, to an isolated agent/worktree so
it does not need the full architecture document or the full execution plan
in context.

Per `Preflight_Day1_Parallel_Execution_Plan.md` §3, a worker should receive
only:

1. `Preflight_Day1_Parallel_Execution_Plan.md` (the common plan);
2. `docs/implementation/day1/CONTRACT_FREEZE.md` (once A00 produces it);
3. its assigned `Preflight_ADD.md` chapters;
4. its packet file from this directory.

## Files

| File | Agent |
|---|---|
| `00-contract-integrator.md` | A00 — Contract Freeze and Integration |
| `01-foundation-config-sqlite.md` | A01 — Foundation, Configuration, Paths, Core SQLite |
| `02-claude-telemetry-hooks.md` | A02 — Claude Telemetry, Hooks, Provider Normalization |
| `03-progress-state-checkpoint.md` | A03 — Progress Tree and State Checkpointing |
| `04-repository-checkpoint.md` | A04 — Git Observation and Repository Checkpoint |
| `05-predictor-policy.md` | A05 — Scope Estimator, Predictor, Risk, Policy, Authorization |
| `06-graceful-pause-scheduler.md` | A06 — Graceful Pause, Safe Points, Durable Scheduler |
| `07-runtime-cli-api.md` | A07 — Application Orchestration, CLI, Local API |
| `08-qa-security-ci.md` | A08 — Cross-component QA, Security, Reliability, CI |
| `CONTRACT_FREEZE_TEMPLATE.md` | Scaffold for A00's `docs/implementation/day1/CONTRACT_FREEZE.md` deliverable |

## Source of truth

These packet files are the **canonical, editable copy** of each agent's
mission, ownership paths, deliverables, and required tests. The Day-1 plan's
"Agent Packets" section is only a summary index that links back here — if
you change a packet's scope, edit it here, not in the plan doc.

For overall context (vision, scope boundary, topology, merge order, cut
list), see `Preflight_Day1_Parallel_Execution_Plan.md` at the repository
root. For architecture, see `Preflight_ADD.md`.
