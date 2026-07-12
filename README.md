# Preflight

**Preflight is a local-first predictive runtime guard for AI coding agents.**
Before and during each turn with a provider like Codex or Claude Code, it
estimates scope, token/quota consumption, completion likelihood, and
blast-radius risk, then applies policy to run, warn, checkpoint, split,
gracefully pause, or block that turn.

It answers a different question than checkpoint/resume/memory tools do:
not "how do we continue?" but **"should we even start this turn?"**

> **Project status: pre-implementation.** This repository currently contains
> the architecture and Day-1 execution plan only. No Go module, CLI, or
> daemon exists yet — that begins at milestone **M0** (see the roadmap in
> `Preflight_ADD.md` §31).

## Source of truth

| Document | Role |
|---|---|
| [`Preflight_ADD.md`](Preflight_ADD.md) | **The single authoritative architecture and implementation specification.** When code, issues, PRs, or comments conflict with it, this document (and accepted ADRs under `docs/adr/`) wins. |
| [`AGENTS.md`](AGENTS.md) | Contributor/agent instructions — required reading before any implementation work. |
| [`Preflight_Day1_Parallel_Execution_Plan.md`](Preflight_Day1_Parallel_Execution_Plan.md) | Subordinate execution plan for the first vertical-slice build: 9-agent (A00–A08) topology, ownership boundaries, merge order. |
| [`agent-packets/`](agent-packets/) | One canonical hand-off packet per Day-1 agent, linked from the plan above. |
| [`docs/repository_inventory.md`](docs/repository_inventory.md) | Audit of every markdown file in the repo and its authority/status. |
| [`docs/archive/`](docs/archive/) | Superseded documents, kept for historical reference, not for implementation. |

Do not treat any other document, prior draft, or conversation as
authoritative over `Preflight_ADD.md`.

## Two core continuity guarantees

- **State Checkpointing** — no unit of work may be marked complete without
  durable, validator-checked evidence (file, DB record, checksum, Git
  snapshot).
- **Graceful Pause** — when a quota limit is calibrated-likely to hit soon,
  Preflight checkpoints state, interrupts at a safe point, and persists a
  durable wake job in SQLite; auto-resume is opt-in and re-verified before
  it runs.

See `Preflight_ADD.md` §1 for the full executive decision record.

## Tech stack

- **Production runtime:** Go 1.26.x, single static binary, SQLite (WAL)
- **VS Code companion:** TypeScript (strict mode)
- **Research/offline only:** Python 3.12+ — never a runtime dependency
- **License:** Apache-2.0

## Repository layout (target — see `Preflight_ADD.md` §10 for the full tree)

```text
cmd/preflight/       entrypoint (not yet created)
internal/             application, domain, adapters (not yet created)
pkg/protocol/v1/      public wire protocol (not yet created)
vscode/                VS Code extension (not yet created)
research/              Python offline research (not yet created)
docs/adr/               accepted architecture decision records
docs/archive/           superseded documents
agent-packets/          Day-1 per-agent hand-off packets
```

## Contributing

Read `Preflight_ADD.md` and `AGENTS.md` in full before proposing or
implementing changes. Work is milestone-gated (`Preflight_ADD.md` §31);
do not implement ahead of the current milestone or add speculative
abstractions for future providers/features.
