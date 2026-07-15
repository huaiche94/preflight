# internal/cli/ — the Cobra command tree, schema-versioned JSON output, typed error contract

> 🌐 English | [繁體中文](README.zh-TW.md)

Exposes Cobra command *constructors* (`NewRootCmd` and friends), never a package-level
command instance and never `os.Exit`, so the tree is fully testable (ADD §10.1 —
[`docs/design/Auspex_ADD.md`](../../docs/design/Auspex_ADD.md)). See [`doc.go`](doc.go) for
the package contract, including why hook subcommands are kebab-case.

Command tree ([`root.go`](root.go)): `version`, `init`, `hook`, `evaluate`, `decision`,
`checkpoint`, `progress`, `state`, `pause`, `resume`, `scheduler`, `daemon`, `status`,
`doctor`, `gc`, `export`, `run`. Real handlers are swapped in by
[`internal/app/wiring/`](../app/wiring/README.md); a command whose service is not wired
returns a typed `unavailable` error instead of pretending to work.

Output contracts:

- Success output is schema-versioned JSON per command — e.g. `auspex.evaluate.v1` ([`evaluate.go`](evaluate.go)), `auspex.checkpoint-create.v1` ([`checkpoint.go`](checkpoint.go)), `auspex.daemon-install.v1` ([`daemon.go`](daemon.go)).
- Every error path renders the one shared `auspex.error.v1` envelope ([`errors.go`](errors.go): `SchemaVersionError`, `RenderErrorJSON`, `WithJSONErrorRendering`) wrapping the frozen `domain.Error` fields (`code`, `message`, `retryable`, `details`); non-`domain.Error` values render as `internal`/non-retryable rather than emitting no JSON. `SilenceErrors` keeps Cobra from appending a second plain-text error line.

Tests: [`golden_test.go`](golden_test.go) checks full success-output shapes structurally
against fixtures in `testdata/golden/*.golden.json`, so a silently added/removed/renamed
field fails the build; [`errorcontract_test.go`](errorcontract_test.go) gates the error
envelope across every command.

Neighbors: business logic lives behind [`internal/orchestrator/`](../orchestrator/README.md)
and the frozen ports in [`internal/app/ports.go`](../app/ports.go); the managed `run`
command's core is [`internal/managed/`](../managed/README.md); the daemon commands drive
[`internal/daemon/`](../daemon/README.md).
