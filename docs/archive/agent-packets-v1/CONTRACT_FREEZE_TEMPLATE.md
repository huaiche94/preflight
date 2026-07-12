# Preflight Day-1 Contract Freeze

Status: DRAFT — A00 must replace every placeholder before feature branches rebase.
Contract commit: `<sha>`
Go module: `<module>`
Schema baseline: `<version>`

## Import paths

| Concern | Package |
|---|---|
| Domain entities | `<path>` |
| Cross-component ports | `<path>` |
| Event protocol | `<path>` |
| SQLite runtime | `<path>` |

## Schema-version strings

```text
preflight.event.v1
preflight.progress-tree.v1
preflight.state-checkpoint.v1
preflight.repository-checkpoint.v1
preflight.pause.v1
preflight.api.v1
```

## ID and idempotency rules

Document each entity ID, operation/event idempotency key, and replay behavior.

## Unknown/null semantics

Document unknown usage, quota, context, probability, provider capability, and reset timestamp handling.

## Transaction boundaries

Document `CompleteNode`, checkpoint creation, authorization consumption, pause persist, and wake lease transactions.

## Error contract

Document stable error codes and fail-open/fail-closed classification.

## Privacy contract

Raw prompts, transcripts, secrets, and repository artifacts policy.

## Migration ranges

- 0000–0009 A01
- 0010–0019 A02
- 0020–0029 A03
- 0030–0039 A04
- 0040–0049 A05
- 0050–0059 A06

## Frozen state transitions

Insert turn, progress node, checkpoint, pause, wake job, and authorization state tables.
