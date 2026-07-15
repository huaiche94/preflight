# cmd/ — entrypoints for the Go binaries

> 🌐 English | [繁體中文](README.zh-TW.md)

Standard Go `cmd/` layout: one subdirectory per executable. Auspex
ships exactly one binary today:

- [`auspex/`](auspex/README.md) — the `auspex` CLI (`go build
  ./cmd/auspex`). A thin `main` plus the composition root; every
  service implementation it wires up lives under
  [`../internal/`](../internal/), and the frozen command tree it
  returns is built by `internal/app/wiring`.

Nothing under `cmd/` contains business logic (per
[`Auspex_ADD.md`](../docs/design/Auspex_ADD.md) §10.1) — if a change
here is more than wiring, path resolution, or DTO-shape translation,
it belongs in an `internal/` package instead.
