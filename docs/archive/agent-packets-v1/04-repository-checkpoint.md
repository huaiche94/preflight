# A04 — Git Observation and Repository Checkpoint

> 🌐 English | [繁體中文](04-repository-checkpoint.zh-TW.md)

## Model

A cheaper coding model can implement most of this; use Fable for path/race/security review.

## ADD ownership

§19, Git/security aspects of §§27–29, Appendix D, M2 subset needed by day-one flow.

## Exclusive paths

```text
internal/gitx/**
internal/repocheckpoint/**
internal/redact/**
schemas/repository-checkpoint.schema.json
testdata/repositories/**
testdata/checkpoints/repository/**
internal/storage/sqlite/migrations/0030-0039_*.sql
docs/implementation/day1/A04.md
```

## Mission

Capture and verify repository evidence without mutating the active branch. Provide a safe checkpoint primitive to A03/A06/A07.

## P0 deliverables

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

## Security requirements

- reject path traversal and symlink escape;
- never include `.git` internals or configured excluded paths;
- redact/omit likely secrets by default;
- never execute a shell string; use argv process calls;
- cap artifact size and file count;
- verify checksums before restore planning.

## Required tests

Tracked/staged/unstaged/untracked, rename/delete, binary file, spaces/newlines in path where platform permits, nested worktree, concurrent mutation, temp cleanup, path traversal, oversize, and secret-like file exclusion.

## Boundary

Do not update Progress Tree directly. Return `RepositoryCheckpoint` and evidence references through A00 ports.
