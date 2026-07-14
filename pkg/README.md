# pkg/ — publicly importable Go packages

> 🌐 English | [繁體中文](README.zh-TW.md)

The only part of the `github.com/huaiche94/auspex` module that
external Go code may import. Everything else lives under
[`../internal/`](../internal/) and is compiler-enforced private.

Contents:

- [`protocol/`](protocol/README.md) — the versioned public wire
  protocol; today only [`protocol/v1/`](protocol/v1/README.md), the
  frozen `auspex.*.v1` contract.

Keeping this tree minimal is deliberate: a type placed here is a
public compatibility commitment (see
[`CONTRACT_FREEZE.md`](../docs/implementation/vertical-slice/CONTRACT_FREEZE.md)),
so shapes stay in `internal/` until they must be shared.
