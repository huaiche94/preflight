# internal/app/wiring/ — in-process composition root for the frozen ports

> 🌐 English | [繁體中文](README.zh-TW.md)

One source file, [`wiring.go`](wiring.go). It owns exactly one concern: collecting one
implementation of each frozen service interface ([`../ports.go`](../ports.go)) into a
validated container the rest of the runtime consumes without knowing which concrete
implementation it got (ADD §13 — [`docs/design/Auspex_ADD.md`](../../../docs/design/Auspex_ADD.md)).

Key types and entry points:

- `Services` — five required interface fields (`Evaluation`, `ProgressTree`, `StateCheckpoint`, `GracefulPause`, `RepositoryCheckpoint`) plus optional support bundles: `Hooks` (hook handlers / event persistence), `Diagnostics` (`auspex doctor` checks — omitted means every check reports skipped, not an error), and pause-lifecycle support for `pause` / `resume` / `scheduler run-once`.
- `New(Services) (*App, error)` — rejects a partially populated struct; real implementations, `internal/testutil/fakes` doubles, and future composites are equally valid values.
- `App.RootCmd()` — builds the Cobra tree by taking [`internal/cli/`](../../cli/README.md)'s stub tree and swapping in real handlers (`replaceSubcommand`) for whichever optional bundles were configured.

Explicit non-goal: root wiring in `cmd/auspex/main.go` is not this package's job — the
`contract-integrator` / `foundation` roles own composing this container into the binary.

Neighbors: consumes [`internal/cli/`](../../cli/README.md) command constructors and
[`internal/orchestrator/`](../../orchestrator/README.md) deps bundles; swap tests
(`evaluate_swap_test.go`, `gc_swap_test.go`, `restart_test.go`) prove a real implementation
replaces a fake with no signature change. No `doc.go`; the package comment is at the top of
`wiring.go`.
