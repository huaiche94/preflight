# pkg/protocol/v1/ — the frozen public wire protocol (`auspex.*.v1`)

> 🌐 English | [繁體中文](README.zh-TW.md)

Package `v1` is the frozen public wire protocol (its own package
comment). It is a compatibility commitment, not an implementation
detail: the baseline is recorded in
[`CONTRACT_FREEZE.md`](../../../docs/implementation/vertical-slice/CONTRACT_FREEZE.md),
and changes to anything here require an ADR (Constitution §3).

## Files

- `event.go` —
  - the six `SchemaVersion*` constants (`auspex.event.v1`,
    `auspex.progress-tree.v1`, `auspex.state-checkpoint.v1`,
    `auspex.repository-checkpoint.v1`, `auspex.pause.v1`,
    `auspex.api.v1`);
  - `EventType`, a closed, versioned taxonomy
    ([`Auspex_ADD.md`](../../../docs/design/Auspex_ADD.md) §11.3) —
    new event types go through the contract-integrator, never ad hoc
    strings from feature code;
  - `Event`, the normalized envelope every provider payload is
    translated into before reaching domain/storage code (ADD §11.1).
- `event_test.go` — pins the schema-version strings byte-for-byte.

## Contract highlights

- Provider wire payloads MUST NOT leak into these types unnormalized,
  and `Event.Payload` is populated only after redaction
  (Constitution §7). Raw prompt text is never a field here.
- `Event.IdempotencyKey` is deterministic per provider event identity
  (stable provider ID where one exists, else a content digest).
- `nil`/absent means unknown, never a substituted zero.

The JSON Schema mirrors of the checkpoint/progress wire shapes live
in [`../../../schemas/`](../../../schemas/README.md).
