# internal/lock/ — single-machine, PID-file-style advisory file lock

> 🌐 English | [繁體中文](README.zh-TW.md)

Package `lock` provides a single-machine, single-daemon advisory file lock. The
package contract is the package comment at the top of `lock.go` (no separate
`doc.go`).

Auspex is a local-first modular monolith (Auspex_ADD.md §1.4; the ADD lives at
[docs/design/Auspex_ADD.md](../../docs/design/Auspex_ADD.md)): exactly one daemon
per machine (or per runtime directory) is meant to own a given SQLite database and
runtime directory at a time. This package gives that daemon — and short-lived CLI
invocations — a crash-safe way to detect and prevent concurrent ownership.

- `Acquire(path) (*FileLock, error)` — takes the lock, returning `ErrLocked` when
  another *live* process already holds it (stale locks from dead PIDs are
  detected and reclaimed).
- `FileLock.Release()` / `FileLock.Path()`.

It is deliberately **not** a distributed lock, not a network lock, and not a
replacement for SQLite's own WAL/busy-timeout concurrency control (which
[`../storage/sqlite`](../storage/sqlite/README.md) owns). Its primary consumer is
[`../daemon`](../daemon/), whose `daemon.lock` in the runtime directory (resolved
by [`../paths`](../paths/README.md)) enforces "one daemon per runtime directory".
