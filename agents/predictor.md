# Predictor

Scope Estimator, Predictor, Risk, Policy, and Authorization.

## Model

Use Fable.

## ADD ownership

§§13–17, prediction portions of §29.9–29.11, ADR-018/019/020/026/032/033.

## Exclusive paths

```text
internal/features/**
internal/predictor/**
internal/policy/**
internal/evaluation/**
testdata/models/**
internal/storage/sqlite/migrations/0040-0049_*.sql
schemas/model.schema.json
docs/implementation/vertical-slice/predictor.md
```

If `internal/evaluation` is absent from the frozen layout, use the exact path assigned by the contract-integrator; do not create a competing package.

## Mission

Implement a deterministic, explainable, cold-start-safe predictor/policy loop. Day-one output is a risk score and quantile estimate, not a calibrated probability.

## Deliverables

1. Prompt feature extractor without storing raw prompt.
2. Repository/session/progress feature DTOs.
3. Simple task classifier with explicit `unknown`.
4. Empirical P50/P80/P90 utilities with monotonic guarantees.
5. Scope estimates for files read/changed and LOC.
6. Quota-delta and context-growth estimate from recent turns.
7. Ten-minute runway **score** and forecast record.
8. Risk components: quota, context, completion, blast radius.
9. Reason codes and confidence.
10. Policy actions: ALLOW, WARN, CHECKPOINT, SPLIT, PAUSE, ABORT.
11. Evaluation persistence.
12. One-time authorization issuance/consumption with prompt/session/evaluation binding, expiry, and replay rejection.

## Cold-start contract

When sample/calibration gates are not met:

```json
{
  "calibrated": false,
  "confidence": "low",
  "risk_score": 0.84,
  "probability": null,
  "reason_codes": ["insufficient_history", "quota_headroom_low"]
}
```

Never output "84% probability" from this value.

## Initial policy suggestion

- low quota pressure and no integrity issue: ALLOW;
- moderate pressure: WARN;
- predicted P90 exceeds available headroom or high blast radius: CHECKPOINT;
- calibrated ten-minute hit probability >= configured threshold twice: PAUSE;
- uncalibrated emergency condition: PAUSE with reason `emergency_threshold`, not probability claim.

## Required tests

- quantile monotonicity property tests;
- missing values/unknown behavior;
- no divide-by-zero/NaN/Inf;
- deterministic output for same inputs;
- reason-code golden tests;
- policy priority and fail-open/fail-closed;
- authorization consume exactly once;
- stale/wrong prompt/wrong session authorization rejected;
- clock-bound expiry tests;
- benchmark fast path.

## Boundary

No provider JSON parsing, Git commands, checkpoint creation, or process interruption. Return decisions through frozen ports.
