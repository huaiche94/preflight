# docs/design/ — authoritative design documents

> 🌐 English | [繁體中文](README.zh-TW.md)

The three governing design documents of Auspex. They lived at the
repository root until ADR-049 moved them here (2026-07-14); filenames
are unchanged, so section citations like `Auspex_ADD.md §31` found
throughout code comments and historical documents still refer
unambiguously to these files.

| Document | Role |
|---|---|
| [`Auspex_ADD.md`](Auspex_ADD.md) | **The single authoritative architecture and implementation specification** — product architecture, domain model, functional requirements, roadmap. When code, issues, PRs, or comments conflict with it, this document wins for architecture. Amended only by accepted ADRs under [`../adr/`](../adr/). **Authored in Traditional Chinese** (prose body zh-TW, section labels and code English); the Chinese text is normative and there is no separate `.zh-TW.md` sibling (ADR-049). |
| [`Auspex_Predictor_Design_Supplement.md`](Auspex_Predictor_Design_Supplement.md) | Companion to the ADD (§14–§17): the predictor pipeline in detail — scope estimation, token/quota forecasting, risk combination. Formalized by ADR-041. |
| [`Auspex_Parallel_Execution_Plan.md`](Auspex_Parallel_Execution_Plan.md) | Subordinate execution plan for the first vertical-slice build: the seven-role topology, ownership boundaries, and merge order. Its live execution record is [`../implementation/vertical-slice/`](../implementation/vertical-slice/README.md). |

## Ownership

These files are shared, cross-cutting artifacts owned exclusively by the
`contract-integrator` role (Constitution §4.3). `Auspex_ADD.md` may only
be edited when a genuine contradiction requires it, with the
corresponding ADR landing in the same change (Constitution §3.5).
