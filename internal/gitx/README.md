# internal/gitx/ — argv-only Git plumbing for the Repository Checkpoint layer

> 🌐 English | [繁體中文](README.zh-TW.md)

Every Git invocation goes through `domain.ProcessRunner` as an argv list — this package never builds or
executes a shell command string (Constitution §7 rule 5). `ExecRunner` (`runner.go`) is the
`exec.Command`-backed implementation; a non-zero exit is data (`ProcessResult.ExitCode`), not an error.

`Client` (`client.go`) wraps a runner with the fixed set of Git operations the checkpoint layer needs, with
flags pinned so user Git configuration can never change what gets captured (Auspex_ADD.md §19.4 — the ADD
now lives at [docs/design/Auspex_ADD.md](../../docs/design/Auspex_ADD.md)):

- **`Status`** — `git status --porcelain=v2 -z --branch --untracked-files=all --find-renames`, parsed by
  `porcelain.go` into typed entries plus `BranchInfo`.
- **`DiffNumstat`** (`numstat.go`) — `git diff --numstat -z --no-ext-diff --find-renames`, staged
  (`--cached`) or unstaged.
- **`DiffPatch` / `ApplyCheck` / `Apply` / `ListUntracked`** (`patch.go`) — binary-safe patches
  (`--binary --full-index --no-ext-diff`), dry-run apply checks, and the two mutating `git apply` calls
  restore uses (incapable of moving refs).
- **`ResolveRepo`** (`resolver.go`) — maps any path inside a working tree to `RepoInfo` (worktree root,
  git dir, common dir, linked-worktree detection); requires git >= 2.31.
- **`Fingerprint`** (`fingerprint.go`) — a deterministic SHA-256 digest of repository state
  (schema `auspex.gitx.fingerprint.v1`) covering HEAD, status entries, numstat counts, and the untracked
  enumeration policy, used for checkpoint identity and change detection.

Primary consumer: [`../repocheckpoint/`](../repocheckpoint/), whose capture path is restricted to the
read-only subset (status, diff, ls-files).
