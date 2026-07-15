# Checkpoint

> 🌐 English | [繁體中文](checkpoint.zh-TW.md)

Consolidates what were previously two separate agent packets — **Progress
Tree / State Checkpointing** and **Repository Checkpoint** — into one
bounded context. Both halves are owned by the same role because they are
always consumed together (the runtime role's pause persist-phase writes a
State Checkpoint and a Repository Checkpoint in the same logical step), but
they remain distinct sub-components internally: keep Part A and Part B
implementations, migrations, and tests separate within this role's paths.

## Model

Use Fable for both parts — this role owns the product's core integrity
boundary (Part A) and its Git-safety boundary (Part B).

## ADD ownership

Part A: §18, Appendix A/B, State Checkpoint scenarios in §29.5, ADR-027 through ADR-030 and ADR-039 as constraints.
Part B: §19, Git/security aspects of §§27–29, Appendix D, M2 subset needed by the day-one flow.

## Exclusive paths

```text
# Part A — Progress Tree / State Checkpointing
internal/progress/**
internal/statecheckpoint/**
internal/artifacts/**
schemas/progress-tree.schema.json
schemas/state-checkpoint.schema.json
testdata/progress-trees/**
testdata/checkpoints/state/**
internal/storage/sqlite/migrations/0020-0029_*.sql

# Part B — Repository Checkpoint
internal/gitx/**
internal/repocheckpoint/**
internal/redact/**
schemas/repository-checkpoint.schema.json
testdata/repositories/**
testdata/checkpoints/repository/**
internal/storage/sqlite/migrations/0030-0039_*.sql

docs/implementation/vertical-slice/checkpoint.md
```

---

## Part A — Progress Tree and State Checkpointing

### Mission

Make Progress Tree the canonical durable task state and enforce that a node cannot become complete without verified artifact evidence.

### Deliverables

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

### Must reject

- "agent says complete" with no artifact;
- missing or changed artifact;
- completed child with violated dependency policy;
- duplicate completion with conflicting evidence;
- invalid state transition;
- checkpoint manifest referencing uncommitted rows.

### Required tests

- valid Markdown section completes and checkpoints;
- missing heading or unbalanced fence rejected;
- crash injection at each completion phase and reconciliation;
- 100 sequential nodes produce 100 verifiable checkpoints;
- same idempotency key returns same result;
- conflicting idempotency payload rejected;
- concurrent completion race.

---

## Part B — Repository Checkpoint

### Mission

Capture and verify repository evidence without mutating the active branch. Provide a safe checkpoint primitive to the runtime role and to the contract-integrator's final review.

### P0 deliverables

1. Repository/worktree resolver.
2. `git status --porcelain=v2 -z` parser.
3. Snapshot fingerprint:
   - repository identity;
   - worktree path;
   - branch/HEAD;
   - index/worktree status;
   - changed paths and numstat;
   - untracked policy metadata.
4. Repository Checkpoint create and verify.
5. Binary-safe patch generation or manifest reference per ADD.
6. Safe untracked archive policy with size/path/secret filters.
7. Atomic temp-to-final artifact write and cleanup.
8. Race detection if Git state changes during capture.
9. Restore **dry-run**; actual restore is stretch.

### Security requirements

- reject path traversal and symlink escape;
- never include `.git` internals or configured excluded paths;
- redact/omit likely secrets by default;
- never execute a shell string; use argv process calls;
- cap artifact size and file count;
- verify checksums before restore planning.

### Required tests

Tracked/staged/unstaged/untracked, rename/delete, binary file, spaces/newlines in path where platform permits, nested worktree, concurrent mutation, temp cleanup, path traversal, oversize, and secret-like file exclusion.

---

## Cross-part boundary

Part A stores references to Part B's checkpoints through frozen ports; it
does not reach into Part B's Git plumbing directly, and Part B does not
update the Progress Tree directly. This internal boundary should stay a
real interface seam even though one role owns both sides, so the two halves
stay independently testable and so a future split back into two roles (if
ever needed) is cheap.
