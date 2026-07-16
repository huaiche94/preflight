// doctorcapture.go: `auspex doctor`'s capture-health checks (issue #90
// Phase A deliverable 3). The uncertainty report's operating premise is
// that Auspex's value now lives in OBSERVED aggregates (runway, pacing,
// spend) — which makes silent capture breakage the worst failure mode:
// every downstream surface keeps rendering, just from stale or absent
// data. These checks make that loud: per-provider last-capture
// timestamps, the token-actual coverage of recent completed turns (a 0%
// rate is a FAIL, not a warn — the "silent-breakage guard"), and whether
// the runway driver is actually producing rows from the quota telemetry.
//
// All checks are read-only SELECTs over the same DBPinger handle the
// existing database check pings — no new wiring, no writes, matching
// diagnostics.go's read-only contract.
package orchestrator

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"

	v1 "github.com/huaiche94/auspex/pkg/protocol/v1"
)

// turnActualsWindow is how many recent provider.turn.completed events per
// provider the coverage check inspects. 20 keeps the check cheap and
// recent enough that a capture regression (a provider CLI update changing
// its transcript/rollout shape) surfaces within a session or two.
const turnActualsWindow = 20

// captureHealthChecks runs the issue-#90 capture-health suite. Fail-open
// on infrastructure trouble: a nil DB yields one skipped check, and an
// unreadable events table yields one warn — doctor itself must run even
// against a half-initialized installation (the same posture checkDB has).
func captureHealthChecks(ctx context.Context, db DBPinger) []CheckResult {
	if db == nil {
		return []CheckResult{{Name: "capture", Status: CheckSkipped, Detail: "no database configured"}}
	}

	providers, err := captureProviders(ctx, db.Conn())
	if err != nil {
		return []CheckResult{{Name: "capture", Status: CheckWarn, Detail: "events not readable: " + err.Error()}}
	}
	if len(providers) == 0 {
		return []CheckResult{
			{Name: "capture:events", Status: CheckWarn, Detail: "no provider events captured yet — hooks may not be installed (run `auspex init` / check hook wiring)"},
			runwayCheck(ctx, db.Conn()),
		}
	}

	var checks []CheckResult
	for _, p := range providers {
		checks = append(checks, CheckResult{
			Name:   "capture:events:" + p.provider,
			Status: CheckOK,
			Detail: fmt.Sprintf("%d events captured, latest %s", p.count, p.latestObservedAt),
		})
	}
	for _, p := range providers {
		checks = append(checks, turnActualsCheck(ctx, db.Conn(), p.provider))
	}
	checks = append(checks, runwayCheck(ctx, db.Conn()))
	return checks
}

type providerCapture struct {
	provider         string
	count            int64
	latestObservedAt string
}

// captureProviders lists every provider with captured provider.* events,
// with counts and last-capture timestamps — the "is anything arriving at
// all, and how stale is it" table.
func captureProviders(ctx context.Context, conn *sql.DB) ([]providerCapture, error) {
	rows, err := conn.QueryContext(ctx, `
		SELECT provider, COUNT(*), MAX(observed_at) FROM events
		WHERE event_type LIKE 'provider.%' AND provider IS NOT NULL AND provider != ''
		GROUP BY provider ORDER BY provider`)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()

	var out []providerCapture
	for rows.Next() {
		var p providerCapture
		if err := rows.Scan(&p.provider, &p.count, &p.latestObservedAt); err != nil {
			return nil, err
		}
		out = append(out, p)
	}
	return out, rows.Err()
}

// turnActualsCheck measures token-actual coverage over the provider's
// last turnActualsWindow provider.turn.completed events: the fraction
// whose payload carries any of the shared token-actual vocabulary
// (total_tokens / input_tokens / output_tokens — the #72/#9 capture
// paths). Statuses:
//
//   - no completed turns yet → warn (fresh install is normal; still worth
//     a line so "zero turns forever" is visible)
//   - 0% coverage with turns present → FAIL, loudly: turns are completing
//     but the token capture path (transcript read / rollout read) is
//     producing nothing — the exact silent breakage the uncertainty
//     report warned about, and it fails the overall report
//   - partial/full coverage → ok, with the observed rate (partial
//     coverage is expected: e.g. Claude turns whose transcript_path was
//     missing degrade honestly to no actuals)
func turnActualsCheck(ctx context.Context, conn *sql.DB, provider string) CheckResult {
	name := "capture:turn-actuals:" + provider
	rows, err := conn.QueryContext(ctx, `
		SELECT payload_json FROM events
		WHERE provider = ? AND event_type = ?
		ORDER BY observed_at DESC, rowid DESC LIMIT ?`,
		provider, string(v1.EventProviderTurnCompleted), turnActualsWindow)
	if err != nil {
		return CheckResult{Name: name, Status: CheckWarn, Detail: "events not readable: " + err.Error()}
	}
	defer func() { _ = rows.Close() }()

	total, withActuals := 0, 0
	for rows.Next() {
		var payloadJSON string
		if err := rows.Scan(&payloadJSON); err != nil {
			return CheckResult{Name: name, Status: CheckWarn, Detail: "events not readable: " + err.Error()}
		}
		total++
		if payloadHasTokenActuals(payloadJSON) {
			withActuals++
		}
	}
	if err := rows.Err(); err != nil {
		return CheckResult{Name: name, Status: CheckWarn, Detail: "events not readable: " + err.Error()}
	}

	switch {
	case total == 0:
		return CheckResult{Name: name, Status: CheckWarn, Detail: "no turn.completed events captured yet"}
	case withActuals == 0:
		return CheckResult{
			Name:   name,
			Status: CheckFail,
			Detail: fmt.Sprintf("0 of the last %d turn.completed events carry token actuals — token capture appears broken (transcript/rollout read path); aggregates are running blind", total),
		}
	default:
		return CheckResult{
			Name:   name,
			Status: CheckOK,
			Detail: fmt.Sprintf("%d of the last %d turn.completed events carry token actuals (%d%%)", withActuals, total, withActuals*100/total),
		}
	}
}

// payloadHasTokenActuals reports whether a turn.completed payload carries
// any numeric token-actual field. Decoding failures count as "no actuals"
// rather than erroring — one undecodable row is a data point about
// capture health, not a reason to abort the check.
func payloadHasTokenActuals(payloadJSON string) bool {
	var payload map[string]any
	if json.Unmarshal([]byte(payloadJSON), &payload) != nil {
		return false
	}
	for _, key := range []string{"total_tokens", "input_tokens", "output_tokens"} {
		if _, isNum := payload[key].(float64); isNum {
			return true
		}
	}
	return false
}

// runwayCheck reports whether the M10 runway driver is actually
// producing forecasts. Quota telemetry present but zero runway rows means
// the drive seam is not wired (or silently failing) — the statusline's
// runway hint and the policy's runway gate would both be running on the
// cold-start zero forecast without anyone noticing.
func runwayCheck(ctx context.Context, conn *sql.DB) CheckResult {
	const name = "capture:runway"
	var count int64
	var latest sql.NullString
	if err := conn.QueryRowContext(ctx, `SELECT COUNT(*), MAX(created_at) FROM runway_forecasts`).Scan(&count, &latest); err != nil {
		return CheckResult{Name: name, Status: CheckWarn, Detail: "runway_forecasts not readable: " + err.Error()}
	}
	if count > 0 {
		return CheckResult{Name: name, Status: CheckOK, Detail: fmt.Sprintf("%d runway forecasts persisted, latest %s", count, latest.String)}
	}

	var quotaCount int64
	if err := conn.QueryRowContext(ctx, `SELECT COUNT(*) FROM events WHERE event_type = ?`, string(v1.EventProviderQuotaObserved)).Scan(&quotaCount); err != nil {
		return CheckResult{Name: name, Status: CheckWarn, Detail: "events not readable: " + err.Error()}
	}
	if quotaCount > 0 {
		return CheckResult{Name: name, Status: CheckWarn, Detail: fmt.Sprintf("quota telemetry present (%d observations) but no runway forecasts — the runway driver may not be wired", quotaCount)}
	}
	return CheckResult{Name: name, Status: CheckWarn, Detail: "no runway forecasts yet (no quota telemetry captured)"}
}
