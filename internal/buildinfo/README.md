# internal/buildinfo/ — build/version metadata behind `auspex version`

> 🌐 English | [繁體中文](README.zh-TW.md)

Package `buildinfo` holds the minimal build/version metadata needed to back the
`auspex version` command. The package contract is the package comment at the top
of `buildinfo.go` (no separate `doc.go`).

- `Version` — currently the fixed development placeholder `"0.0.0-dev"`; wiring
  real ldflags-injected values (release tag, commit SHA, build date) was
  explicitly out of scope for foundation-01 per agents/foundation.md and remains
  pending real release packaging.
- `String()` — the human-readable version string the command prints.

Consumers: [`../cli`](../cli/)'s `auspex version` command prints `String()`, and
`cmd/auspex/wire.go` passes `Version` into [`../daemon`](../daemon/)'s config and
[`../httpapi`](../httpapi/)'s version endpoint.
