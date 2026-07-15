# pkg/protocol/ — the versioned public wire protocol

> 🌐 English | [繁體中文](README.zh-TW.md)

Each protocol version gets its own subdirectory so consumers pin an
exact contract by import path. Today there is exactly one:

- [`v1/`](v1/README.md) — the frozen `auspex.*.v1` event envelope,
  event-type taxonomy, and schema-version constants.

A breaking change never edits `v1/` in place — it would be a new
`v2/` package alongside it, sanctioned by an ADR (Constitution §3;
see
[`CONTRACT_FREEZE.md`](../../docs/implementation/vertical-slice/CONTRACT_FREEZE.md)
for what "frozen" commits to). The JSON Schema documents for the
non-Go wire shapes of the same contract live in
[`../../schemas/`](../../schemas/README.md).
