# internal/repocheckpoint/ — Repository Checkpoint: working-tree evidence capture, verify, restore

> 🌐 English | [繁體中文](README.zh-TW.md)

Captures exact working-tree evidence before a pause or high-risk turn without ever mutating the repository
(Auspex_ADD.md §19 — the ADD now lives at [docs/design/Auspex_ADD.md](../../docs/design/Auspex_ADD.md);
wire schema `auspex.repository-checkpoint.v1`, Appendix D). Constitution §7 rule 6: repository checkpoints
are atomic and never silently commit the active branch — every Git operation in capture is read-only
plumbing via [`../gitx/`](../gitx/).

Key pieces:

- **`Capture`** (`capture.go`) — writes a checkpoint artifact directory: `staged.patch.gz` /
  `unstaged.patch.gz` (from `gitx.Client.DiffPatch`), `untracked.zip` (`archive.go`), and `manifest.json`.
- **Security policy** (`security.go`) — every archived path passes `validateUntrackedPath`: path-traversal,
  symlink, and `.git`-internal rejection, plus size caps (5 MiB/file, 100 MiB total, 10 000 files); skipped
  candidates are recorded with a `SkipReason` ledger, feeding recoverability warnings.
- **Secret redaction** — untracked files are scanned with [`../redact/`](../redact/) and skipped on a match
  (`archive.go`); patch content gets in-place span redaction of `+`/`-` line bodies (`patchredact.go`).
  Filenames in diff headers and binary-diff header/payload lines are deliberately not rewritten — the
  accepted residual surface recorded in
  [ADR-042](../../docs/adr/0042-patch-redaction-residual-surface.md).
- **Atomicity** (`atomicwrite.go`) — artifacts are staged in a temp directory, fsynced, then atomically
  renamed into place; a checkpoint directory is either complete or absent. `orphanscan.go` removes temp
  directories left by a killed process at startup.
- **`Verify`** (`verify.go`) — re-reads `manifest.json` and every artifact file from disk and recomputes
  sizes and SHA-256 digests; the DB row's own fields are never taken at face value.
- **Restore** — `restoredryrun.go` is the report-only dry run (ADD §19.6 checks, `git apply --check`);
  the real mutating restore (`restore.go`, `RestoreApply`) landed with issue #6 / ADR-048: replays staged
  and unstaged patches via `git apply`, extracts untracked files with a strict no-clobber rule, and never
  touches refs (no checkout, reset, commit, or stash).
- **`Service`** (`service.go`) — implements the frozen `app.RepositoryCheckpointService` port
  (`Create` / `Verify` / `Restore`); rows persist via `store.go` (`migrations/0030_repository_checkpoints.sql`).
