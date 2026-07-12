# A03 — Progress Tree and State Checkpointing

## Model

Use Fable because this is a product-defining integrity boundary.

## ADD ownership

§18, Appendix A/B, State Checkpoint scenarios in §29.5, ADR-027 through ADR-030 and ADR-039 as constraints.

## Exclusive paths

```text
internal/progress/**
internal/statecheckpoint/**
internal/artifacts/**
schemas/progress-tree.schema.json
schemas/state-checkpoint.schema.json
testdata/progress-trees/**
testdata/checkpoints/state/**
internal/storage/sqlite/migrations/0020-0029_*.sql
docs/implementation/day1/A03.md
```

## Mission

Make Progress Tree the canonical durable task state and enforce that a node cannot become complete without verified artifact evidence.

## Deliverables

1. Task/node/edge/artifact stores.
2. Node state machine with explicit valid transitions.
3. Artifact validators:
   - file exists;
   - checksum matches;
   - Markdown heading exists;
   - Markdown code fences balanced;
   - optional custom validator interface.
4. `CompleteNode` atomic protocol:
   - stage/verify artifact evidence;
   - update node;
   - create State Checkpoint;
   - commit in one DB transaction where applicable;
   - publish normalized events after commit.
5. State Checkpoint manifest serialization and checksum.
6. Startup reconciliation for staged artifact vs DB crash windows.
7. Completion idempotency key and duplicate provider event handling.
8. Snapshot/load-latest/verify APIs.

## Must reject

- “agent says complete” with no artifact;
- missing or changed artifact;
- completed child with violated dependency policy;
- duplicate completion with conflicting evidence;
- invalid state transition;
- checkpoint manifest referencing uncommitted rows.

## Required tests

- valid Markdown section completes and checkpoints;
- missing heading or unbalanced fence rejected;
- crash injection at each completion phase and reconciliation;
- 100 sequential nodes produce 100 verifiable checkpoints;
- same idempotency key returns same result;
- conflicting idempotency payload rejected;
- concurrent completion race.

## Boundary

Repository file/diff capture belongs to A04. A03 stores references through frozen ports and fakes A04 in tests.
