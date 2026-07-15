# internal/predictor/ — deterministic, explainable prediction primitives and the four-stage forecast pipeline

> 🌐 English | [繁體中文](README.zh-TW.md)

This tree holds the Predictor pipeline that turns a prompt into a pre-turn forecast. Every
implementation here is Version 1 (rule-based, deterministic) per the Evolution Roadmap in
[Auspex_Predictor_Design_Supplement.md](../../docs/design/Auspex_Predictor_Design_Supplement.md),
the pipeline's design doc; the frozen pipeline shape itself is
[ADR-041](../../docs/adr/0041-predictor-forecast-layer.md). Code comments cite "ADD §…" —
those sections live in [Auspex_ADD.md](../../docs/design/Auspex_ADD.md).

Pipeline (each stage is a frozen `internal/app` port, wired end-to-end by
[`internal/evaluation`](../evaluation/README.md)'s `EvaluateTurn`):

1. Prompt features and the task classifier — [`internal/features`](../features/README.md)
   (`ExtractPromptFeatures`, `ClassifyTask`), the layer this package sits above.
2. Scope estimator — [`scope/`](scope/README.md) (`app.ScopeEstimator`): files/lines expected.
3. Token forecaster — [`token/`](token/README.md) (`app.TokenForecaster`): turn token cost.
4. Quota forecaster — [`quota/`](quota/README.md) (`app.QuotaForecaster`): projected quota and
   context-window position after the turn.
5. Risk combiner — [`risk/`](risk/README.md) (`app.RiskCombiner`): four risk components + overall.
6. Policy — [`internal/policy`](../policy/README.md): the terminal stage, maps risk + runway to an action.

[`runway/`](runway/README.md) is deliberately outside this chain (ADR-041): it answers "is a quota
window about to run out within the horizon", feeds `GracefulPauseService.Observe`, and is not a
RiskCombiner input.

The root package itself contains one shared primitive: `Quantiles` / `EmpiricalQuantiles`
(`quantile.go`), the frozen P50/P80/P90 triple with an unconditional P50 <= P80 <= P90 guarantee
and no NaN/Inf outputs.

Cold-start contract (see `doc.go` for the package contract): day-one output is a risk score and
quantile estimate, never a calibrated probability — Constitution §7 rule 7, "Uncalibrated risk
scores are never labeled as probabilities."
