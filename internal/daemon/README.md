# internal/daemon/ — the M6 background daemon: lifecycle, auth token, event broker, worker loop

> 🌐 English | [繁體中文](README.zh-TW.md)

The resident process behind `auspex daemon run` (issue #7; ADD §23 —
[`docs/design/Auspex_ADD.md`](../../docs/design/Auspex_ADD.md)). Form and token storage are
decision D-16 in [`docs/DECISION_LOG.md`](../../docs/DECISION_LOG.md): an OS-agnostic
foreground process at the core, with an optional macOS LaunchAgent around it.

Key files:

- [`daemon.go`](daemon.go) — `Daemon.Run` composes: singleton lock (`daemon.lock`, via `internal/lock`) → bearer token → dynamic loopback listener → runtime metadata → serve + work; shutdown runs the reverse order. Signal handling stays with the CLI caller; `Run` only knows contexts.
- [`token.go`](token.go) — `GenerateToken` mints a 256-bit hex token into `<data>/daemon.token` with `0600` permissions, `O_TRUNC`-overwritten (rotated) on every start, so each restart invalidates all previously issued tokens (ADD §23.2, §27.5; D-16). `VerifyToken` compares in constant time; an empty expected token matches nothing.
- [`metadata.go`](metadata.go) — the `auspex.daemon.v1` discovery document at `<runtime>/daemon.json` (PID, address, token file path), written `0600`, removed first on shutdown.
- [`broker.go`](broker.go) — in-memory pub/sub fan-out feeding the SSE stream. Deliberately memory-only: the broker keeps no history, so there is no event replay — a late subscriber reads current state from the status/jobs endpoints instead; slow subscribers drop events rather than block the worker.
- [`worker.go`](worker.go) — the resident scheduler loop (ADD §23.6): reconcile expired leases, claim due wake jobs ([`../scheduler/`](../scheduler/README.md)), execute `pause.Wake` → `GracefulPauseService.Resume` (the real `ValidateResume` runs inside — never bypassed), renew the lease, complete/fail, and publish `pkg/protocol/v1` events through the broker. 5s poll interval by default.

The HTTP handlers themselves live in [`internal/httpapi/`](../httpapi/README.md) — this
package hands them a freshly minted token via `Config.NewHandler`. The launchd install
(`auspex daemon install`, LaunchAgent plist with KeepAlive) is implemented in
[`internal/orchestrator/daemon.go`](../orchestrator/daemon.go) and
[`internal/cli/daemon.go`](../cli/daemon.go), not here. No `doc.go`; each file carries its
own contract comment.
