# internal/clock/ — the real wall-clock implementation of domain.Clock

> 🌐 English | [繁體中文](README.zh-TW.md)

Package `clock` provides the real, wall-clock implementation of
[`../domain`](../domain/)'s `Clock` interface. The package contract is the package
comment at the top of `clock.go` (no separate `doc.go`).

- `System` — the stateless implementation backed by `time.Now()`; safe for
  concurrent use.
- `New() domain.Clock` — the constructor production wiring uses.

Production code depends on `domain.Clock`, never on this package directly, so
tests can substitute a fake and stay deterministic — for example
[`../retention`](../retention/README.md) and
[`../telemetry/claude`](../telemetry/claude/README.md) take a `domain.Clock`
rather than calling `time.Now()`.
