# internal/policy/ — terminal pipeline stage: risk + runway → one frozen policy action

> 🌐 English | [繁體中文](README.zh-TW.md)

`Decider.Decide` (`decide.go`) turns the combined risk result from
[`internal/predictor/risk`](../predictor/risk/README.md) and the independent runway forecast from
[`internal/predictor/runway`](../predictor/runway/README.md) — the one place both signals
legitimately meet (ADR-041) — into one of the eight frozen `app.PolicyAction` values
(`internal/app/ports.go`, ADD §17.2):

`RUN`, `WARN`, `REQUIRE_CONFIRMATION`, `CHECKPOINT_AND_RUN`, `SPLIT`, `PAUSE`,
`PAUSE_AND_AUTO_RESUME`, `BLOCK`. (This Decider currently emits six of them; `SPLIT` and
`PAUSE_AND_AUTO_RESUME` are frozen enum values no code path here produces yet, though
`context.go`'s action-strength ladder already ranks them.)

Gates run in the fixed ADD §17.3 priority order, first match wins: explicit deny → integrity
failure (both caller-supplied booleans; both fail closed to `BLOCK`) → runway pause → mandatory
checkpoint boundary → then the ADD §16.5 risk bands (<0.45 RUN, 0.45–0.65 WARN, 0.65–0.85
REQUIRE_CONFIRMATION or CHECKPOINT_AND_RUN when blast radius is also high, >=0.85
CHECKPOINT_AND_RUN). Two overlay rules can only strengthen, never weaken, the chosen action:
the D-08 context-utilization thresholds (`context.go`; projected P90 context >85% → WARN, >95%
→ CHECKPOINT_AND_RUN; active by default but confidence-gated, so today's cold-start forecaster
never trips them) and the opt-in per-turn cost budget (`costbudget.go`, ADR-043 increment 3,
priced by [`internal/pricing`](../pricing/README.md)).

Runway PAUSE has two legs: a calibrated hit-probability >= 0.80 observed with the §17.6
double-sample debounce (`PriorRunwayHitConfirmed`, caller-owned state), or an uncalibrated
emergency (limit reached, used >= 98%, or time-to-limit P50 <= 60s) that skips the debounce and
always carries reason `emergency_threshold` — never a probability claim.

Cold-start policy (the load-bearing invariant, Constitution §7 rule 7): whenever any upstream
input is `Calibrated == false`, `Decision.Probability` is nil unconditionally, whatever action is
chosen. Exactly one code path sets a non-nil probability, and it checks
`RunwayForecast.Calibrated == true` directly; a risk score is never copied into Probability.
`coldstart.go`'s `ColdStartExample` pins that contract shape in a testable literal. `Decide`
never returns an error — every gap degrades to the most conservative applicable decision. See
`doc.go` for the package contract; ADD sections cited above live in
[Auspex_ADD.md](../../docs/design/Auspex_ADD.md).
