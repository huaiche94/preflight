# schemas/ — JSON Schemas for the frozen wire shapes

> 🌐 English | [繁體中文](README.zh-TW.md)

Machine-checkable JSON Schema documents for the `auspex.*.v1` wire
shapes that are serialized as JSON files (checkpoint manifests,
progress-tree exports). Each schema's own `description` field states
what it mirrors; summarized here:

| Schema | Pins | Mirrors |
|---|---|---|
| `progress-tree.schema.json` | `auspex.progress-tree.v1` | `internal/progress`'s Node/Edge Go types and the `progress_nodes`/`progress_edges` tables (migrations 0020–0021). Shape from `Auspex_ADD.md` Appendix A, §18. |
| `state-checkpoint.schema.json` | `auspex.state-checkpoint.v1` | `internal/statecheckpoint.Manifest` and the `state_checkpoints` table (migration 0023). Shape from `Auspex_ADD.md` §18.8, Appendix B. `integrity_sha256` MUST be independently recomputed before trusting a stored manifest. |
| `repository-checkpoint.schema.json` | `auspex.repository-checkpoint.v1` | `internal/repocheckpoint.Manifest` (a checkpoint's `manifest.json`). Shape from `Auspex_ADD.md` §19, Appendix D. |

The schema-version strings are defined as Go constants in
[`../pkg/protocol/v1/`](../pkg/protocol/v1/README.md) and frozen by
[`CONTRACT_FREEZE.md`](../docs/implementation/vertical-slice/CONTRACT_FREEZE.md).
The `Auspex_ADD.md` citations inside the schema `description` strings
are kept verbatim per ADR-049 §Decision 3; the document itself lives at
[`../docs/design/Auspex_ADD.md`](../docs/design/Auspex_ADD.md).

Example instances live under [`../testdata/`](../testdata/README.md)
(`progress-trees/sample-task.json`, both `sample-manifest.json`
fixtures) and were schema-validated when generated (recorded in
[`../docs/implementation/vertical-slice/checkpoint.md`](../docs/implementation/vertical-slice/checkpoint.md)).
CI has no JSON Schema validation job yet — deliberately deferred, per
the header comment in
[`../.github/workflows/ci.yml`](../.github/workflows/ci.yml).
