# internal/idgen/ — the UUIDv7 implementation of domain.IDGenerator

> 🌐 English | [繁體中文](README.zh-TW.md)

Package `idgen` provides the real implementation of
[`../domain`](../domain/)'s `IDGenerator` interface. The package contract is the
package comment at the top of `idgen.go` (no separate `doc.go`).

Per
[CONTRACT_FREEZE.md](../../docs/implementation/vertical-slice/CONTRACT_FREEZE.md)'s
ID rules, all Auspex-owned entity IDs (the opaque string types in
`internal/domain/ids.go`) are UUIDv7 at generation time, generated here and never
parsed for meaning by callers.

- `UUIDv7` — the stateless implementation backed by `github.com/google/uuid`;
  safe for concurrent use.
- `New() domain.IDGenerator` — the constructor production wiring uses.
- `NewID()` returns a lowercase, hyphenated UUIDv7 string. UUIDv7 is time-ordered,
  which keeps primary-key/index locality reasonable for
  [`../storage/sqlite`](../storage/sqlite/README.md).

Production code depends on `domain.IDGenerator`, not this package directly, so
tests can inject deterministic IDs.
