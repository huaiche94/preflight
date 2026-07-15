# cmd/auspex/ — the `auspex` CLI binary (thin main + composition root)

> 🌐 English | [繁體中文](README.zh-TW.md)

Per [`Auspex_ADD.md`](../../docs/design/Auspex_ADD.md) §10.1 (cited in
`main.go`'s package comment), this package only does wiring and
process exit — no business logic lives here or in Cobra command
handlers. The frozen command tree itself (`evaluate`, `decision`,
`checkpoint`, `pause`/`resume`/`scheduler`, `status`, `doctor`,
`hook claude ...`) is built by
[`internal/app/wiring`](../../internal/app/wiring/); this package
composes real service implementations into that container.

## Files

- `main.go` — entrypoint. `main` calls `run()` and passes its return
  code to `os.Exit`, so deferred cleanup (DB close) always runs before
  exit. Also keeps `newRootCmd`, a minimal version-only fallback
  exercised directly by `main_test.go`.
- `wire.go` — the composition root: opens and migrates the SQLite
  database under the OS user-data directory, constructs one real
  implementation of every frozen `app.*` service (progress tree,
  state/repository checkpoint, evaluation pipeline, graceful pause,
  daemon, retention), and assembles them into
  `internal/app/wiring.Services`; `wiring.New(services).RootCmd()`
  returns the fully wired Cobra tree.
- `adapters.go` — interface-bridging glue only: DTO-shape translation
  and read-only SQL adapters for package-local seams (e.g.
  `pause.SessionContextResolver`), plus documented fail-closed stubs
  for capabilities that do not exist yet (managed turn interrupt).
- `main_test.go` — exercises the version-only fallback command.

## Relations

- [`../../internal/`](../../internal/) — all service implementations.
- [`../../pkg/protocol/v1/`](../../pkg/protocol/v1/README.md) — the
  frozen wire protocol the composed services speak.
