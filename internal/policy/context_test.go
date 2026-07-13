// context_test.go: ADR-043 increment-2 / DECISION_LOG.md D-08 context-
// utilization threshold tests — both tiers fire at their documented
// defaults, boundary values follow D-08's strictly-greater wording, the
// confidence bar keeps cold-start/low-confidence projections silent
// (bit-for-bit today's behavior), the never-downgrade ordering holds
// against every stronger existing gate (including preserving a calibrated
// runway decision's probability untouched), thresholds are adjustable and
// disable-able via Config, and no context-driven code path ever emits a
// probability (Constitution principle #2).
package policy

import (
	"reflect"
	"testing"

	"github.com/huaiche94/auspex/internal/app"
	"github.com/huaiche94/auspex/internal/domain"
)

// eligibleContextForecast builds a QuotaForecast whose context projection
// meets the D-08 confidence bar (Confidence medium, no cold-start flag):
// the shape a warmed/calibration-fed forecaster (issue #11) is expected
// to produce, which today's always-cold-start RuleQuotaForecaster never
// does — exactly why these tests construct it by hand.
func eligibleContextForecast(projectedP90 float64) domain.QuotaForecast {
	return domain.QuotaForecast{
		ProjectedContextUsedP90: ptrF(projectedP90),
		Calibrated:              false, // an eligible projection is still not a calibrated one
		Confidence:              domain.ConfidenceMedium,
	}
}

// coldStartContextForecast mirrors what RuleQuotaForecaster actually
// emits today (forecaster.go: Confidence low + PREDICTION_COLD_START,
// unconditionally): the projection exists but has not earned the D-08
// bar.
func coldStartContextForecast(projectedP90 float64) domain.QuotaForecast {
	return domain.QuotaForecast{
		ProjectedContextUsedP90: ptrF(projectedP90),
		Calibrated:              false,
		Confidence:              domain.ConfidenceLow,
		ReasonCodes:             []domain.ReasonCode{domain.ReasonPredictionColdStart},
	}
}

func hasReason(codes []domain.ReasonCode, want domain.ReasonCode) bool {
	for _, c := range codes {
		if c == want {
			return true
		}
	}
	return false
}

func hasPolicyReason(codes []string, want string) bool {
	for _, c := range codes {
		if c == want {
			return true
		}
	}
	return false
}

// --- Threshold tiers and boundaries -----------------------------------------

// TestContextThresholds_TiersAndBoundaries sweeps projections across both
// D-08 default thresholds with a low-risk base (ALLOW), asserting the
// exact tier boundaries: strictly greater than 85 warns, strictly greater
// than 95 checkpoints — 85.0 and 95.0 themselves do not cross (D-08's own
// ">85%" / ">95%" wording).
func TestContextThresholds_TiersAndBoundaries(t *testing.T) {
	d := NewDecider()

	cases := []struct {
		name       string
		projected  float64
		wantAction app.PolicyAction
		wantReason domain.ReasonCode // "" means no threshold reason expected
	}{
		{"well below warn", 50, app.PolicyRun, ""},
		{"exactly warn threshold does not fire", 85.0, app.PolicyRun, ""},
		{"just above warn", 85.01, app.PolicyWarn, domain.ReasonContextWarnThresholdExceeded},
		{"mid warn tier", 91, app.PolicyWarn, domain.ReasonContextWarnThresholdExceeded},
		{"exactly checkpoint threshold stays warn tier", 95.0, app.PolicyWarn, domain.ReasonContextWarnThresholdExceeded},
		{"just above checkpoint", 95.01, app.PolicyCheckpointAndRun, domain.ReasonContextCheckpointThresholdExceeded},
		{"saturated window", 100, app.PolicyCheckpointAndRun, domain.ReasonContextCheckpointThresholdExceeded},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := d.Decide(DecideRequest{
				Risk:   lowRiskAllInputs(0.01),
				Runway: uncalibratedRunway(0.01),
				Quota:  eligibleContextForecast(tc.projected),
			})
			if got.Action != tc.wantAction {
				t.Errorf("projected=%v: Action = %v, want %v", tc.projected, got.Action, tc.wantAction)
			}
			if tc.wantReason != "" && !hasReason(got.ReasonCodes, tc.wantReason) {
				t.Errorf("projected=%v: ReasonCodes = %v, want %v included", tc.projected, got.ReasonCodes, tc.wantReason)
			}
			if tc.wantReason == "" &&
				(hasReason(got.ReasonCodes, domain.ReasonContextWarnThresholdExceeded) ||
					hasReason(got.ReasonCodes, domain.ReasonContextCheckpointThresholdExceeded)) {
				t.Errorf("projected=%v: unexpected threshold reason in %v", tc.projected, got.ReasonCodes)
			}
			// Constitution principle #2: a context-utilization projection
			// is never a probability, for ANY tier outcome.
			if got.Probability != nil {
				t.Errorf("projected=%v: Probability = %v, want nil — a utilization projection is never a probability", tc.projected, *got.Probability)
			}
		})
	}
}

// TestContextThresholds_WarnUpgrade_Fields pins the full shape of a
// context-driven WARN upgrade from a clean ALLOW base: severity, policy
// reason string, uncalibrated posture, and a nil probability.
func TestContextThresholds_WarnUpgrade_Fields(t *testing.T) {
	d := NewDecider()
	got := d.Decide(DecideRequest{
		Risk:   lowRiskAllInputs(0.01),
		Runway: uncalibratedRunway(0.01),
		Quota:  eligibleContextForecast(91),
	})

	if got.Action != app.PolicyWarn {
		t.Fatalf("Action = %v, want PolicyWarn", got.Action)
	}
	if got.Severity != "warning" {
		t.Errorf("Severity = %q, want warning (mirrors the medium risk band's WARN)", got.Severity)
	}
	if !hasPolicyReason(got.PolicyReasonCodes, ReasonContextWarnThreshold) {
		t.Errorf("PolicyReasonCodes = %v, want %q included", got.PolicyReasonCodes, ReasonContextWarnThreshold)
	}
	if got.Calibrated {
		t.Error("Calibrated = true, want false — the context projection consulted is uncalibrated")
	}
	if got.Probability != nil {
		t.Errorf("Probability = %v, want nil (Constitution principle #2)", *got.Probability)
	}
	// Decision.Confidence: most conservative among the consulted inputs
	// (base high-confidence risk vs the forecast's medium).
	if got.Confidence != domain.ConfidenceMedium {
		t.Errorf("Confidence = %v, want medium (the forecast's, lower than the base's high)", got.Confidence)
	}
}

// TestContextThresholds_CheckpointUpgradesWeakerBases: the checkpoint
// tier upgrades every strictly-weaker base action, including
// REQUIRE_CONFIRMATION from the high risk band.
func TestContextThresholds_CheckpointUpgradesWeakerBases(t *testing.T) {
	d := NewDecider()

	cases := []struct {
		name      string
		riskScore float64 // drives the base band decision
		wantBase  app.PolicyAction
	}{
		{"ALLOW base", 0.01, app.PolicyRun},
		{"WARN base (medium band)", 0.50, app.PolicyWarn},
		{"REQUIRE_CONFIRMATION base (high band)", 0.70, app.PolicyRequireConfirmation},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			req := DecideRequest{
				Risk:   lowRiskAllInputs(tc.riskScore),
				Runway: uncalibratedRunway(0.01),
			}
			if base := d.Decide(req); base.Action != tc.wantBase {
				t.Fatalf("precondition: base Action = %v, want %v", base.Action, tc.wantBase)
			}
			req.Quota = eligibleContextForecast(97)
			got := d.Decide(req)
			if got.Action != app.PolicyCheckpointAndRun {
				t.Errorf("Action = %v, want PolicyCheckpointAndRun", got.Action)
			}
			if got.Severity != "critical" {
				t.Errorf("Severity = %q, want critical", got.Severity)
			}
			if got.RequiresConfirmation {
				t.Error("RequiresConfirmation = true, want false on a CHECKPOINT_AND_RUN upgrade")
			}
			if !hasReason(got.ReasonCodes, domain.ReasonContextCheckpointThresholdExceeded) {
				t.Errorf("ReasonCodes = %v, want CONTEXT_CHECKPOINT_THRESHOLD_EXCEEDED included", got.ReasonCodes)
			}
			if got.Probability != nil {
				t.Errorf("Probability = %v, want nil (Constitution principle #2)", *got.Probability)
			}
		})
	}
}

// --- Confidence gating: cold start / low confidence stay silent -------------

// TestContextThresholds_SilentOnColdStartOrLowConfidence is D-08's
// confidence-discipline gate: for every ineligible forecast shape — nil
// projection, low confidence (today's actual RuleQuotaForecaster output),
// unavailable/empty confidence, or a medium-confidence forecast still
// flagging PREDICTION_COLD_START — the decision is DEEPLY equal to the
// decision made with no forecast at all: exactly today's behavior, not
// merely the same action.
func TestContextThresholds_SilentOnColdStartOrLowConfidence(t *testing.T) {
	d := NewDecider()
	baseReq := DecideRequest{
		Risk:   lowRiskAllInputs(0.01),
		Runway: uncalibratedRunway(0.01),
	}
	want := d.Decide(baseReq)

	coldButMedium := domain.QuotaForecast{
		ProjectedContextUsedP90: ptrF(99),
		Confidence:              domain.ConfidenceMedium,
		ReasonCodes:             []domain.ReasonCode{domain.ReasonPredictionColdStart},
	}

	cases := []struct {
		name  string
		quota domain.QuotaForecast
	}{
		{"zero-value forecast (no projection)", domain.QuotaForecast{}},
		{"projection missing, confidence fine", domain.QuotaForecast{Confidence: domain.ConfidenceHigh}},
		{"cold start at 99% (today's real forecaster shape)", coldStartContextForecast(99)},
		{"low confidence without cold-start flag", domain.QuotaForecast{ProjectedContextUsedP90: ptrF(99), Confidence: domain.ConfidenceLow}},
		{"unavailable confidence", domain.QuotaForecast{ProjectedContextUsedP90: ptrF(99), Confidence: domain.ConfidenceUnavailable}},
		{"empty confidence", domain.QuotaForecast{ProjectedContextUsedP90: ptrF(99)}},
		{"medium confidence but still flagged cold start", coldButMedium},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			req := baseReq
			req.Quota = tc.quota
			got := d.Decide(req)
			if !reflect.DeepEqual(got, want) {
				t.Errorf("Decide with ineligible forecast diverged from today's behavior:\n got %+v\nwant %+v", got, want)
			}
		})
	}
}

// --- Never downgrade ---------------------------------------------------------

// TestContextThresholds_NeverDowngrades: every existing gate whose action
// is at least as strong as the tier's keeps its action, severity,
// calibration, and probability untouched — the threshold only ANNOTATES
// via reason codes. Covers the Constitution-#2-critical case: a
// calibrated runway PAUSE's non-nil probability must survive the overlay
// bit-for-bit.
func TestContextThresholds_NeverDowngrades(t *testing.T) {
	d := NewDecider()

	t.Run("calibrated runway PAUSE keeps action and probability", func(t *testing.T) {
		req := DecideRequest{
			Risk: lowRiskAllInputs(0.01),
			Runway: domain.RunwayForecast{
				Calibrated:     true,
				Confidence:     domain.ConfidenceHigh,
				HitProbability: ptrF(0.85),
				RiskScore:      0.85,
			},
			PriorRunwayHitConfirmed: true,
			Quota:                   eligibleContextForecast(97),
		}
		got := d.Decide(req)
		if got.Action != app.PolicyPause {
			t.Fatalf("Action = %v, want PolicyPause (PAUSE outranks CHECKPOINT_AND_RUN — never downgrade)", got.Action)
		}
		if got.Probability == nil || *got.Probability != 0.85 {
			t.Errorf("Probability = %v, want the calibrated 0.85 preserved untouched", got.Probability)
		}
		if !got.Calibrated {
			t.Error("Calibrated = false, want true — annotation must not alter the runway decision's calibration posture")
		}
		if got.Severity != "critical" {
			t.Errorf("Severity = %q, want the runway decision's own critical", got.Severity)
		}
		if !hasReason(got.ReasonCodes, domain.ReasonContextCheckpointThresholdExceeded) {
			t.Errorf("ReasonCodes = %v, want the crossed threshold still disclosed", got.ReasonCodes)
		}
	})

	t.Run("emergency PAUSE keeps action", func(t *testing.T) {
		got := d.Decide(DecideRequest{
			Risk: lowRiskAllInputs(0.01),
			Runway: domain.RunwayForecast{
				CurrentUsedPercent: ptrF(99),
				RiskScore:          0.95,
				Calibrated:         false,
				Confidence:         domain.ConfidenceLow,
			},
			Quota: eligibleContextForecast(97),
		})
		if got.Action != app.PolicyPause {
			t.Errorf("Action = %v, want PolicyPause preserved", got.Action)
		}
		if !hasPolicyReason(got.PolicyReasonCodes, ReasonEmergencyThreshold) {
			t.Errorf("PolicyReasonCodes = %v, want emergency_threshold preserved", got.PolicyReasonCodes)
		}
	})

	t.Run("critical band CHECKPOINT_AND_RUN absorbs checkpoint tier, gains only the reason code", func(t *testing.T) {
		req := DecideRequest{
			Risk:   lowRiskAllInputs(0.90),
			Runway: uncalibratedRunway(0.01),
		}
		want := d.Decide(req)
		req.Quota = eligibleContextForecast(97)
		got := d.Decide(req)
		if got.Action != want.Action || got.Severity != want.Severity || got.Calibrated != want.Calibrated || got.RiskScore != want.RiskScore {
			t.Errorf("equal-strength overlay altered the base decision: got %+v, want %+v (plus reason codes)", got, want)
		}
		if !hasReason(got.ReasonCodes, domain.ReasonContextCheckpointThresholdExceeded) {
			t.Errorf("ReasonCodes = %v, want CONTEXT_CHECKPOINT_THRESHOLD_EXCEEDED appended", got.ReasonCodes)
		}
	})

	t.Run("medium band WARN absorbs warn tier", func(t *testing.T) {
		got := d.Decide(DecideRequest{
			Risk:   lowRiskAllInputs(0.50),
			Runway: uncalibratedRunway(0.01),
			Quota:  eligibleContextForecast(91),
		})
		if got.Action != app.PolicyWarn {
			t.Fatalf("Action = %v, want PolicyWarn", got.Action)
		}
		if !hasReason(got.ReasonCodes, domain.ReasonContextWarnThresholdExceeded) {
			t.Errorf("ReasonCodes = %v, want CONTEXT_WARN_THRESHOLD_EXCEEDED appended", got.ReasonCodes)
		}
		if !hasPolicyReason(got.PolicyReasonCodes, "medium_risk_band") {
			t.Errorf("PolicyReasonCodes = %v, want the base band's own reason preserved", got.PolicyReasonCodes)
		}
	})

	t.Run("runway-armed calibrated WARN upgraded by checkpoint tier suppresses the probability", func(t *testing.T) {
		// The one probability-bearing base the checkpoint tier CAN
		// upgrade (WARN < CHECKPOINT_AND_RUN). The upgraded decision is
		// context-driven, and a utilization projection is never a
		// probability — so the armed WARN's calibrated probability must
		// NOT survive onto the stronger action (Constitution #2's
		// conservative reading: suppress, never re-attribute).
		got := d.Decide(DecideRequest{
			Risk: lowRiskAllInputs(0.01),
			Runway: domain.RunwayForecast{
				Calibrated:     true,
				Confidence:     domain.ConfidenceHigh,
				HitProbability: ptrF(0.85),
				RiskScore:      0.85,
			},
			PriorRunwayHitConfirmed: false, // armed, not confirmed -> WARN base
			Quota:                   eligibleContextForecast(97),
		})
		if got.Action != app.PolicyCheckpointAndRun {
			t.Fatalf("Action = %v, want PolicyCheckpointAndRun (upgrade over the armed WARN)", got.Action)
		}
		if got.Probability != nil {
			t.Errorf("Probability = %v, want nil — the upgraded action is justified by an uncalibrated projection", *got.Probability)
		}
		if got.Calibrated {
			t.Error("Calibrated = true, want false after consulting the uncalibrated projection")
		}
		if !hasPolicyReason(got.PolicyReasonCodes, "runway_hit_probability_armed_pending_confirmation") {
			t.Errorf("PolicyReasonCodes = %v, want the base's armed reason preserved for the audit trail", got.PolicyReasonCodes)
		}
	})

	t.Run("explicit deny BLOCK is untouched even at 100% context", func(t *testing.T) {
		req := DecideRequest{
			ExplicitDeny: true,
			Risk:         lowRiskAllInputs(0.01),
			Runway:       uncalibratedRunway(0.01),
		}
		want := d.Decide(req)
		req.Quota = eligibleContextForecast(100)
		got := d.Decide(req)
		if !reflect.DeepEqual(got, want) {
			t.Errorf("explicit-deny BLOCK altered by the context rule: got %+v, want %+v", got, want)
		}
	})
}

// --- Config: adjustable and disable-able -------------------------------------

// TestContextThresholds_ConfigOverrides covers D-08's "config 可關可調":
// custom thresholds move the boundaries, the disable flag restores
// today's behavior exactly, and zero/negative thresholds normalize to the
// documented defaults rather than firing on everything.
func TestContextThresholds_ConfigOverrides(t *testing.T) {
	d := NewDecider()
	base := DecideRequest{
		Risk:   lowRiskAllInputs(0.01),
		Runway: uncalibratedRunway(0.01),
	}

	t.Run("custom warn threshold is respected", func(t *testing.T) {
		req := base
		req.Quota = eligibleContextForecast(60)
		req.Config = Config{ContextP90WarnThresholdPercent: 50, ContextP90CheckpointThresholdPercent: 99}
		if got := d.Decide(req); got.Action != app.PolicyWarn {
			t.Errorf("Action = %v, want PolicyWarn with the warn threshold lowered to 50", got.Action)
		}
	})

	t.Run("custom checkpoint threshold is respected", func(t *testing.T) {
		req := base
		req.Quota = eligibleContextForecast(91)
		req.Config = Config{ContextP90WarnThresholdPercent: 50, ContextP90CheckpointThresholdPercent: 90}
		if got := d.Decide(req); got.Action != app.PolicyCheckpointAndRun {
			t.Errorf("Action = %v, want PolicyCheckpointAndRun with the checkpoint threshold lowered to 90", got.Action)
		}
	})

	t.Run("thresholds raised above 100 never fire", func(t *testing.T) {
		req := base
		req.Quota = eligibleContextForecast(100)
		req.Config = Config{ContextP90WarnThresholdPercent: 150, ContextP90CheckpointThresholdPercent: 151}
		want := d.Decide(base)
		if got := d.Decide(req); !reflect.DeepEqual(got, want) {
			t.Errorf("out-of-reach thresholds still altered the decision: got %+v, want %+v", got, want)
		}
	})

	t.Run("disable flag restores today's behavior exactly", func(t *testing.T) {
		req := base
		req.Quota = eligibleContextForecast(100)
		req.Config = Config{DisableContextUtilizationThresholds: true}
		want := d.Decide(base)
		if got := d.Decide(req); !reflect.DeepEqual(got, want) {
			t.Errorf("disabled rule still altered the decision: got %+v, want %+v", got, want)
		}
	})

	t.Run("zero and negative thresholds normalize to defaults, not fire-on-everything", func(t *testing.T) {
		req := base
		req.Quota = eligibleContextForecast(50) // below the real 85 default
		req.Config = Config{ContextP90WarnThresholdPercent: -1, ContextP90CheckpointThresholdPercent: 0}
		want := d.Decide(base)
		if got := d.Decide(req); !reflect.DeepEqual(got, want) {
			t.Errorf("degenerate threshold config must fall back to the documented defaults: got %+v, want %+v", got, want)
		}
	})

	t.Run("DefaultConfig carries the D-08 literals", func(t *testing.T) {
		cfg := DefaultConfig()
		if cfg.ContextP90WarnThresholdPercent != 85.0 || cfg.ContextP90CheckpointThresholdPercent != 95.0 {
			t.Errorf("DefaultConfig thresholds = %v/%v, want 85/95 (D-08)", cfg.ContextP90WarnThresholdPercent, cfg.ContextP90CheckpointThresholdPercent)
		}
		if cfg.DisableContextUtilizationThresholds {
			t.Error("DisableContextUtilizationThresholds = true in DefaultConfig — D-08 ships the thresholds ACTIVE")
		}
	})
}

// TestContextThresholds_DoesNotMutateCallerReasonCodes: appending the
// threshold code must never write through into the caller-supplied
// CombineRiskResult's own ReasonCodes slice (the overlay copies before
// appending).
func TestContextThresholds_DoesNotMutateCallerReasonCodes(t *testing.T) {
	d := NewDecider()
	callerCodes := make([]domain.ReasonCode, 1, 8) // spare capacity invites in-place append bugs
	callerCodes[0] = domain.ReasonQuotaNearLimit

	risk := lowRiskAllInputs(0.90)
	risk.OverallRisk.ReasonCodes = callerCodes

	got := d.Decide(DecideRequest{
		Risk:   risk,
		Runway: uncalibratedRunway(0.01),
		Quota:  eligibleContextForecast(97),
	})
	if !hasReason(got.ReasonCodes, domain.ReasonContextCheckpointThresholdExceeded) {
		t.Fatalf("ReasonCodes = %v, want the threshold code appended", got.ReasonCodes)
	}
	if extended := callerCodes[:cap(callerCodes)]; hasReason(extended, domain.ReasonContextCheckpointThresholdExceeded) {
		t.Error("the overlay wrote the threshold code into the caller's slice backing array")
	}
}
