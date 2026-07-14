# internal/artifacts/ — artifact evidence validators behind node completion

> 🌐 English | [繁體中文](README.zh-TW.md)

The concrete checks that turn a claimed piece of evidence into a verified one. "Completed means evidenced"
(Constitution §6.2) is enforced here: every validator inspects the actual filesystem or file content, never
the caller's assertion about it. This package is the pure validation seam that
[`../progress/`](../progress/)'s `CompleteNode` protocol calls into; it does not orchestrate transactions
or persistence itself (that stays in `internal/progress`).

Key entry points (`validator.go`):

- **`Validator`** — the narrow interface: `Kind() string` plus `Validate(ctx, Candidate) (Result, error)`.
- **`Candidate`** — the evidence under test (path, expected SHA-256, expected heading, ...).
- **`Result`** — `Passed` plus human-readable `Reasons`; a failing result always carries at least one reason.

Built-in validators, keyed by the `Kind()` string stored as `artifacts.validator_id` and used by
acceptance criteria (Auspex_ADD.md §18.5 — the ADD now lives at
[docs/design/Auspex_ADD.md](../../docs/design/Auspex_ADD.md)):

- `file_exists` (`file_exists.go`) — the path is an existing regular file.
- `checksum_matches` (`checksum.go`) — the file's actual SHA-256 equals the recorded evidence digest; this
  is the check that makes "agent says complete" insufficient on its own.
- `heading_exists` (`heading.go`) — a Markdown file contains the claimed heading as a real ATX heading
  line, ignoring matches inside fenced code blocks.
- `fence_balance` (`fence_balance.go`) — every opened Markdown code fence is closed (CommonMark rules).

**`Registry`** (`registry.go`) dispatches by kind, comes pre-populated with the four built-ins, and accepts
custom `Validator` registrations without any schema change. Validators are independently callable — each
does its own stat/open rather than assuming another ran first.
