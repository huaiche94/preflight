// context.go: the ADR-043 increment-2 context-window threshold rule
// (DECISION_LOG.md D-08, issue #13). ADR-043 promotes the context window
// from a QuotaForecast sub-field to a first-class exhaustible resource,
// and D-08 decides its factory posture: unlike the cost budget (which
// stays policy-inert until a user declares one — ADR-043 Decision 2's
// "absence of a budget means the resource is simply not policy-active"),
// the context window is NOT user-declared, has an objective hard ceiling,
// and hitting it mid-turn is catastrophic — so conservative default
// thresholds ship ACTIVE out of the box:
//
//	projected P90 context utilization > 85%  -> WARN
//	projected P90 context utilization > 95%  -> CHECKPOINT_AND_RUN
//
// with three honesty guards, all from D-08's own text:
//
//  1. Confidence discipline ("cold-start 信心不足不觸發"): the rule only
//     fires when the projection meets the pipeline's existing
//     confidence/data-quality bar — Confidence at least medium (the same
//     rank ordering internal/predictor/risk.confidenceRank fixes) and no
//     PREDICTION_COLD_START reason on the forecast. Today's
//     RuleQuotaForecaster is unconditionally cold-start
//     (Confidence: low + PREDICTION_COLD_START, per CONTRACT_FREEZE.md's
//     licensed estimate), so in the shipped binary this rule is
//     structurally silent until a calibrated/warmed forecaster (issue
//     #11's data) raises the forecast's confidence — thresholds are
//     active by default, but a projection that hasn't earned trust never
//     triggers them. Absent/low-confidence projection => exactly the
//     pre-increment behavior.
//  2. Never-downgrade ordering: the rule can only strengthen a decision
//     the existing gates picked (RUN -> WARN, WARN -> CHECKPOINT_AND_RUN,
//     ...), never weaken one — see actionStrength and
//     applyContextThresholds.
//  3. Adjustable and disable-able ("config 可關可調"): thresholds live on
//     Config (ContextP90WarnThresholdPercent /
//     ContextP90CheckpointThresholdPercent /
//     DisableContextUtilizationThresholds), reachable programmatically
//     via DecideRequest.Config — the seam evaluation.Service.Policy
//     exposes to the composition root. D-08's revisit clause ("降級為惰性
//     是一行 config 預設值的事") is literally DefaultConfig flipping one
//     boolean.
//
// Constitution principle #2 (an uncalibrated score is never a
// probability) holds by construction: a context-utilization percentage is
// a projected resource level, not a hit probability, so no code path in
// this file ever assigns Decision.Probability a non-nil value — a
// context-driven upgrade explicitly nils it (see applyContextThresholds),
// and an annotation-only overlay never touches it.
package policy

import (
	"github.com/huaiche94/auspex/internal/app"
	"github.com/huaiche94/auspex/internal/domain"
)

// DefaultContextP90WarnThresholdPercent and
// DefaultContextP90CheckpointThresholdPercent are D-08's owner-approved
// factory defaults (DECISION_LOG.md D-08, 2026-07-13; ADR-043 increment
// 2): projected P90 context utilization strictly greater than 85% of the
// window suggests WARN, strictly greater than 95% suggests
// CHECKPOINT_AND_RUN. Strictly-greater (not >=) follows D-08's own ">85%"
// / ">95%" wording. These are deliberately conservative bootstrap values,
// scheduled for re-review once calibration data (issue #11) exists —
// D-08's recorded revisit condition.
const (
	DefaultContextP90WarnThresholdPercent       = 85.0
	DefaultContextP90CheckpointThresholdPercent = 95.0
)

// ReasonContextWarnThreshold / ReasonContextCheckpointThreshold are this
// package's plain-string policy reason codes for the two D-08 tiers —
// the same package-local vocabulary convention ReasonEmergencyThreshold
// established (policy-layer trigger names, distinct from the shared
// domain.ReasonCode enum, which carries the matching
// CONTEXT_*_THRESHOLD_EXCEEDED values for cross-role consumers).
const (
	ReasonContextWarnThreshold       = "context_warn_threshold_exceeded"
	ReasonContextCheckpointThreshold = "context_checkpoint_threshold_exceeded"
)

// contextTier is the outcome of evaluating the D-08 thresholds against
// one QuotaForecast: which tier, if any, the projection crossed.
type contextTier int

const (
	contextTierNone contextTier = iota
	contextTierWarn
	contextTierCheckpoint
)

// policyConfidenceRank orders domain.Confidence from least to most
// trustworthy, mirroring internal/predictor/risk's confidenceRank exactly
// (same values, same "unavailable ranks below low" reasoning; duplicated
// rather than imported for the same documented reason
// internal/predictor/quota keeps nominalTurnTokens independent of
// internal/predictor/token — two packages measuring one shared concept
// from independent vantage points, without a cross-stage import). An
// unknown/empty Confidence ranks alongside unavailable: an input that
// cannot even say how confident it is has not earned the D-08 bar.
var policyConfidenceRank = map[domain.Confidence]int{
	domain.ConfidenceUnavailable: 0,
	domain.ConfidenceLow:         1,
	domain.ConfidenceMedium:      2,
	domain.ConfidenceHigh:        3,
	domain.ConfidenceExact:       4,
}

// contextProjectionMeetsBar implements D-08's confidence-discipline gate
// ("cold-start 信心不足不觸發"): the projected context P90 must exist, the
// forecast's Confidence must rank at least medium, and the forecast must
// not flag PREDICTION_COLD_START. Both checks are kept even though
// today's cold-start forecaster always fails the confidence check alone
// (Confidence: low, per CONTRACT_FREEZE.md's licensed cold-start
// estimate): the explicit reason-code check makes D-08's "cold-start 不觸發"
// literal rather than an accident of today's confidence assignment, so a
// future forecaster claiming medium confidence while still flagging
// cold-start stays silent too.
func contextProjectionMeetsBar(qf domain.QuotaForecast) bool {
	if qf.ProjectedContextUsedP90 == nil {
		return false
	}
	if policyConfidenceRank[qf.Confidence] < policyConfidenceRank[domain.ConfidenceMedium] {
		return false
	}
	for _, rc := range qf.ReasonCodes {
		if rc == domain.ReasonPredictionColdStart {
			return false
		}
	}
	return true
}

// contextThresholdTier evaluates the D-08 thresholds for one forecast
// under cfg (already normalized). Returns contextTierNone when the rule
// is disabled, the projection is absent, the confidence bar is not met,
// or no threshold is crossed — every one of which means "exactly today's
// behavior" for the caller.
func contextThresholdTier(qf domain.QuotaForecast, cfg Config) contextTier {
	if cfg.DisableContextUtilizationThresholds {
		return contextTierNone
	}
	if !contextProjectionMeetsBar(qf) {
		return contextTierNone
	}
	projected := *qf.ProjectedContextUsedP90
	switch {
	case projected > cfg.ContextP90CheckpointThresholdPercent:
		return contextTierCheckpoint
	case projected > cfg.ContextP90WarnThresholdPercent:
		return contextTierWarn
	default:
		return contextTierNone
	}
}

// actionStrength fixes this package's severity ordering over the eight
// frozen app.PolicyAction values (internal/app/ports.go), used by
// applyContextThresholds' never-downgrade rule. The ordering follows ADD
// §17.2/§17.3's escalation ladder: RUN < WARN < REQUIRE_CONFIRMATION <
// CHECKPOINT_AND_RUN < SPLIT < PAUSE < PAUSE_AND_AUTO_RESUME < BLOCK. An
// action this table does not know (including the zero value) ranks above
// everything — an unknown action is never downgraded by this rule, the
// conservative reading.
var actionStrength = map[app.PolicyAction]int{
	app.PolicyRun:                 0,
	app.PolicyWarn:                1,
	app.PolicyRequireConfirmation: 2,
	app.PolicyCheckpointAndRun:    3,
	app.PolicySplit:               4,
	app.PolicyPause:               5,
	app.PolicyPauseAndAutoResume:  6,
	app.PolicyBlock:               7,
}

func strengthOf(action app.PolicyAction) int {
	if s, ok := actionStrength[action]; ok {
		return s
	}
	// Unknown action: treat as maximally strong so the context rule can
	// never replace (i.e. downgrade) something it does not understand.
	return len(actionStrength)
}

// applyContextThresholds overlays the D-08 context-utilization rule onto
// the decision the existing ADD §17.3 gates already produced. Semantics,
// in D-08's own terms:
//
//   - Tier none (disabled / no projection / confidence bar unmet / below
//     both thresholds): base is returned untouched — bit-for-bit today's
//     behavior.
//   - Tier fired, base action already at least as strong: the base
//     decision's action, severity, Calibrated, Confidence, and — the
//     Constitution-#2-load-bearing field — Probability are all left
//     exactly as the stronger gate set them (never downgrade, never
//     muddy a calibrated runway decision's probability); only the reason
//     codes gain the tier's CONTEXT_*_THRESHOLD_EXCEEDED /
//     context_*_threshold_exceeded entries, so the card/CLI can still
//     say the threshold was crossed.
//   - Tier fired, base action strictly weaker: the action is upgraded to
//     the tier's action (WARN or CHECKPOINT_AND_RUN). The rebuilt
//     decision keeps the base's reason codes (they still explain the
//     risk landscape) plus the tier's codes; Probability is nil
//     unconditionally (a context-utilization projection is a resource
//     level, never a hit probability — Constitution principle #2);
//     Calibrated is the AND of the base's and the forecast's (a decision
//     now partly based on an uncalibrated projection must not claim
//     calibration — mirrors Decision.Calibrated's own "every input this
//     decision was based on" rule); Confidence is the most conservative
//     of the two consulted inputs (Decision.Confidence's documented
//     rule); RiskScore is the higher of the base's score and the
//     pipeline's own ContextRisk term (ADD §16.2's
//     sigmoid((projected_context_p90-85)/7) expression of this same
//     projection — the score that most directly drove a context-driven
//     action, and never lower than the base's, consistent with never
//     downgrading).
func applyContextThresholds(base Decision, req DecideRequest, cfg Config) Decision {
	tier := contextThresholdTier(req.Quota, cfg)
	if tier == contextTierNone {
		return base
	}

	var (
		tierAction   app.PolicyAction
		tierDomain   domain.ReasonCode
		tierPolicy   string
		tierSeverity string
	)
	switch tier {
	case contextTierCheckpoint:
		tierAction = app.PolicyCheckpointAndRun
		tierDomain = domain.ReasonContextCheckpointThresholdExceeded
		tierPolicy = ReasonContextCheckpointThreshold
		tierSeverity = "critical" // context ceiling imminent — same label as the critical risk band's CHECKPOINT_AND_RUN
	default: // contextTierWarn
		tierAction = app.PolicyWarn
		tierDomain = domain.ReasonContextWarnThresholdExceeded
		tierPolicy = ReasonContextWarnThreshold
		tierSeverity = "warning" // same label as the medium risk band's WARN
	}

	if strengthOf(base.Action) >= strengthOf(tierAction) {
		// Annotation only: the existing gates already chose an action at
		// least as strong — never downgrade it, never alter its
		// calibration/probability posture; just disclose that the
		// threshold was crossed.
		base.ReasonCodes = appendReasonCodeOnce(base.ReasonCodes, tierDomain)
		base.PolicyReasonCodes = append(base.PolicyReasonCodes, tierPolicy)
		return base
	}

	riskScore := base.RiskScore
	if ctxScore := clamp01Risk(req.Risk.ContextRisk.Score); ctxScore > riskScore {
		riskScore = ctxScore
	}

	return Decision{
		Action:            tierAction,
		Calibrated:        base.Calibrated && req.Quota.Calibrated,
		Confidence:        lowerConfidence(base.Confidence, req.Quota.Confidence),
		RiskScore:         riskScore,
		Probability:       nil, // a utilization projection is never a probability (Constitution principle #2)
		ReasonCodes:       appendReasonCodeOnce(base.ReasonCodes, tierDomain),
		PolicyReasonCodes: append(base.PolicyReasonCodes, tierPolicy),
		// The tier actions (WARN, CHECKPOINT_AND_RUN) never require
		// confirmation — mirrors riskBandDecision's high-blast-radius
		// case, where an upgrade to CHECKPOINT_AND_RUN clears the flag.
		RequiresConfirmation: false,
		Severity:             tierSeverity,
	}
}

// lowerConfidence returns the more conservative (lower-ranked) of two
// Confidence values under policyConfidenceRank — Decision.Confidence's
// documented "most conservative among the inputs actually consulted"
// rule, applied to exactly the two inputs a context-driven upgrade
// consults (the base decision's inputs and the quota forecast).
func lowerConfidence(a, b domain.Confidence) domain.Confidence {
	if policyConfidenceRank[a] <= policyConfidenceRank[b] {
		return a
	}
	return b
}

// appendReasonCodeOnce returns codes plus code unless code is already
// present — decision reason codes are a set-like vocabulary
// (internal/predictor/quota.dedupeReasons applies the same discipline one
// stage up). It always copies before appending: the incoming slice is
// typically the caller-supplied CombineRiskResult's own ReasonCodes
// (propagated by riskBandDecision), and appending in place could mutate
// the caller's input through a shared backing array.
func appendReasonCodeOnce(codes []domain.ReasonCode, code domain.ReasonCode) []domain.ReasonCode {
	for _, c := range codes {
		if c == code {
			return codes
		}
	}
	out := make([]domain.ReasonCode, len(codes), len(codes)+1)
	copy(out, codes)
	return append(out, code)
}
