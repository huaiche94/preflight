# docs/backlog/ — design notes for accepted-but-unscheduled work

> 🌐 English | [繁體中文](README.zh-TW.md)

Each file here is a design note for work the repository owner has
accepted but that is not (fully) scheduled into an active wave. A
backlog note is tied to a tracking GitHub issue, records the audit that
motivated it, and carries its own phased TODO — so scheduling later
never has to reconstruct the reasoning. Notes follow the same grounding
discipline as the wave2 analyses: capture and mechanics may be
designed up front, but numeric decisions (coefficients, thresholds)
wait for real data.

| File | What it covers |
|---|---|
| [`provider-model-effort-features.md`](provider-model-effort-features.md) | Making provider / model / effort prediction inputs (the pipeline was provider-, model-, and effort-blind per the 2026-07-13 audit). Tracking: issue #20, ordering in [`../DECISION_LOG.md`](../DECISION_LOG.md) D-10. Its §4 phases 0 (capture) and 1 (cohort filtering, [ADR-047](../adr/0047-token-cohort-fallback-ladder.md)) landed 2026-07-14; phase 2 (empirical calibration) is blocked on per-cohort data. |

## Neighbors

- When a backlog phase graduates into a contract-level decision, it
  becomes an ADR in [`../adr/`](../adr/README.md) (as ADR-047 did).
- The formulas a note defers to are specified in
  [`../design/Auspex_ADD.md`](../design/Auspex_ADD.md) §15 and the
  [predictor supplement](../design/Auspex_Predictor_Design_Supplement.md).
- Gap analyses that feed the backlog live in
  [`../implementation/vertical-slice/wave2-analysis/`](../implementation/vertical-slice/wave2-analysis/README.md).
