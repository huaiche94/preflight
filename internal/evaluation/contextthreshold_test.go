// contextthreshold_test.go: ADR-043 increment-2 (DECISION_LOG.md D-08)
// service-level tests — the context projection is persisted by
// EvaluateTurn and read back onto the card; today's REAL cold-start
// pipeline keeps the D-08 thresholds silent even at high projected
// utilization (the confidence gate, end to end); a warmed (medium-
// confidence, non-cold-start) forecaster makes the thresholds fire
// through the full evaluate -> persist -> Decide -> card chain; and the
// Service.Policy seam disables them.
package evaluation_test

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/huaiche94/auspex/internal/app"
	"github.com/huaiche94/auspex/internal/domain"
	"github.com/huaiche94/auspex/internal/evaluation"
	"github.com/huaiche94/auspex/internal/policy"
)

// warmedQuotaForecaster is a stub Stage-3 forecaster representing the
// shape a calibration-fed (issue #11) implementation will produce: a
// context projection with medium confidence and no PREDICTION_COLD_START
// flag — the D-08 eligibility bar today's RuleQuotaForecaster never meets
// (it is unconditionally cold-start; see internal/predictor/quota's doc).
// Calibrated stays false: eligibility is a confidence question, not a
// calibration claim, and the resulting decisions must still carry
// probability: null (Constitution principle #2).
type warmedQuotaForecaster struct {
	projectedContextP90 float64
}

func (w warmedQuotaForecaster) ForecastQuota(_ context.Context, _ app.ForecastQuotaRequest) (domain.QuotaForecast, error) {
	p := w.projectedContextP90
	return domain.QuotaForecast{
		ProjectedContextUsedP90: &p,
		Calibrated:              false,
		Confidence:              domain.ConfidenceMedium,
	}, nil
}

// TestEvaluateTurn_HighContext_ColdStartProjectionPersistedButSilent is
// the high-context end-to-end scenario against the REAL cold-start
// pipeline: a session already at 88% context produces a projection well
// above the 85% warn threshold, EvaluateTurn persists it (migration
// 0045), the card renders it — and yet the D-08 thresholds stay silent,
// because the real RuleQuotaForecaster is cold-start/low-confidence and
// D-08's confidence gate ("cold-start 信心不足不觸發") must keep exactly
// today's decision behavior until a projection earns trust.
func TestEvaluateTurn_HighContext_ColdStartProjectionPersistedButSilent(t *testing.T) {
	clk := newFakeClock(time.Date(2026, 7, 13, 10, 0, 0, 0, time.UTC))
	source := newFakeDataSource()
	source.contextObs = domain.ContextObservation{
		UsedPercent: ptrF64(88),
		Confidence:  domain.ConfidenceHigh,
		ObservedAt:  clk.Now(),
	}
	svc, _ := newTestService(t, clk, &sequentialIDs{prefix: "id"}, source)
	ctx := context.Background()

	eval, err := svc.EvaluateTurn(ctx, app.EvaluateTurnRequest{
		SessionID: "sess-highctx", TurnID: "turn-1", Provider: "claude",
	})
	if err != nil {
		t.Fatalf("EvaluateTurn: %v", err)
	}

	card, err := svc.ForecastCard(ctx, eval.ID)
	if err != nil {
		t.Fatalf("ForecastCard: %v", err)
	}

	// The projection was persisted and read back: current 88% plus the
	// cold-start growth delta, never below the observed current usage.
	if card.ContextProjectedP90 == nil {
		t.Fatal("ContextProjectedP90 = nil, want the persisted projection (migration 0045)")
	}
	if *card.ContextProjectedP90 < 88 || *card.ContextProjectedP90 > 100 {
		t.Errorf("ContextProjectedP90 = %v, want within [88, 100] (current + cold-start growth, capped)", *card.ContextProjectedP90)
	}

	// D-08 confidence gate, proven through the whole service: the
	// projection is far above both thresholds, yet no threshold state was
	// recorded on the decision — the cold-start forecast has not earned
	// the bar. (The card flags are a faithful read-back of the persisted
	// policy_decisions.reason_codes_json, so flags-false means the codes
	// were never emitted.)
	if card.ContextWarnThresholdExceeded || card.ContextCheckpointThresholdExceeded {
		t.Errorf("threshold state = warn:%v checkpoint:%v, want both false — cold-start projections must stay silent (D-08)",
			card.ContextWarnThresholdExceeded, card.ContextCheckpointThresholdExceeded)
	}

	// The rendered surfaces still SHOW the projection (with the card's
	// uncalibrated labeling), just without any threshold claim.
	ac := card.AdditionalContext()
	if !strings.Contains(ac, "context: P90 ~") {
		t.Errorf("AdditionalContext missing the context projection line:\n%s", ac)
	}
	if strings.Contains(ac, "threshold exceeded") {
		t.Errorf("AdditionalContext claims a threshold the policy engine never recorded:\n%s", ac)
	}
}

// TestEvaluateTurn_WarmedForecaster_CheckpointThresholdFires proves the
// D-08 thresholds are genuinely ACTIVE by default through the full
// service path: with a Stage-3 forecast that meets the confidence bar and
// projects 97%, EvaluateTurn's persisted decision is upgraded to
// CHECKPOINT_AND_RUN (the base decision for this input is the high risk
// band's REQUIRE_CONFIRMATION — context risk sigmoid((97-85)/7) ≈ 0.85⁻ —
// so the upgrade is attributable to the threshold rule, not the bands),
// and the card reads the threshold state back for every surface.
func TestEvaluateTurn_WarmedForecaster_CheckpointThresholdFires(t *testing.T) {
	clk := newFakeClock(time.Date(2026, 7, 13, 10, 0, 0, 0, time.UTC))
	source := newFakeDataSource()
	stages := realStages(source)
	stages.Quota = warmedQuotaForecaster{projectedContextP90: 97}
	svc, _ := newTestServiceWithStages(t, clk, &sequentialIDs{prefix: "id"}, source, stages)
	ctx := context.Background()

	eval, err := svc.EvaluateTurn(ctx, app.EvaluateTurnRequest{
		SessionID: "sess-warm", TurnID: "turn-1", Provider: "claude",
	})
	if err != nil {
		t.Fatalf("EvaluateTurn: %v", err)
	}

	decision, err := svc.Decide(ctx, app.DecideRequest{EvaluationID: eval.ID})
	if err != nil {
		t.Fatalf("Decide: %v", err)
	}
	if decision.Action != app.PolicyCheckpointAndRun {
		t.Fatalf("Action = %v, want CHECKPOINT_AND_RUN (projected 97%% > 95%% with the confidence bar met)", decision.Action)
	}

	card, err := svc.ForecastCard(ctx, eval.ID)
	if err != nil {
		t.Fatalf("ForecastCard: %v", err)
	}
	if !card.ContextCheckpointThresholdExceeded {
		t.Error("ContextCheckpointThresholdExceeded = false, want true (read back from the persisted decision reason codes)")
	}
	if card.Probability != nil {
		t.Errorf("Probability = %v, want nil — a context-driven decision is never a probability claim (Constitution principle #2)", *card.Probability)
	}
	if ac := card.AdditionalContext(); !strings.Contains(ac, "context: P90 ~97% of window (projected) — CHECKPOINT threshold exceeded") {
		t.Errorf("AdditionalContext missing the threshold state:\n%s", ac)
	}
	if line := evaluation.StatusLineText("Opus 4.1", &card, nil); !strings.Contains(line, "context worst-case ~97% (checkpoint)") {
		t.Errorf("StatusLineText missing the context segment: %q", line)
	}
}

// TestEvaluateTurn_PolicySeam_DisablesContextThresholds proves the D-08
// "config 可關可調" seam at the service level: the identical warmed
// high-context input with Service.Policy disabling the rule persists the
// un-upgraded base decision and no threshold state — while the projection
// itself is still persisted and rendered (disabling the POLICY rule does
// not hide the DATA).
func TestEvaluateTurn_PolicySeam_DisablesContextThresholds(t *testing.T) {
	clk := newFakeClock(time.Date(2026, 7, 13, 10, 0, 0, 0, time.UTC))
	source := newFakeDataSource()
	stages := realStages(source)
	stages.Quota = warmedQuotaForecaster{projectedContextP90: 97}
	svc, _ := newTestServiceWithStages(t, clk, &sequentialIDs{prefix: "id"}, source, stages)
	svc.Policy = policy.Config{DisableContextUtilizationThresholds: true}
	ctx := context.Background()

	eval, err := svc.EvaluateTurn(ctx, app.EvaluateTurnRequest{
		SessionID: "sess-disabled", TurnID: "turn-1", Provider: "claude",
	})
	if err != nil {
		t.Fatalf("EvaluateTurn: %v", err)
	}

	decision, err := svc.Decide(ctx, app.DecideRequest{EvaluationID: eval.ID})
	if err != nil {
		t.Fatalf("Decide: %v", err)
	}
	if decision.Action == app.PolicyCheckpointAndRun {
		t.Errorf("Action = %v, want the un-upgraded base decision with the rule disabled", decision.Action)
	}

	card, err := svc.ForecastCard(ctx, eval.ID)
	if err != nil {
		t.Fatalf("ForecastCard: %v", err)
	}
	if card.ContextWarnThresholdExceeded || card.ContextCheckpointThresholdExceeded {
		t.Error("threshold state recorded despite the rule being disabled")
	}
	if card.ContextProjectedP90 == nil || *card.ContextProjectedP90 != 97 {
		t.Errorf("ContextProjectedP90 = %v, want 97 still persisted/rendered — disabling the rule never hides the data", card.ContextProjectedP90)
	}
}
