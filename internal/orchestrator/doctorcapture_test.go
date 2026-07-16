// doctorcapture_test.go: the issue-#90 capture-health doctor checks
// against a real migrated scratch DB — per-provider last-capture lines,
// the token-actual coverage rate, the 0%-coverage FAIL (the silent-
// breakage guard), and the runway-rows presence check.
package orchestrator_test

import (
	"context"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/huaiche94/auspex/internal/orchestrator"
	"github.com/huaiche94/auspex/internal/storage/sqlite"
	v1 "github.com/huaiche94/auspex/pkg/protocol/v1"
)

// seedTurnCompleted persists n provider.turn.completed events for
// provider, the first withActuals of them carrying total_tokens.
func seedTurnCompleted(t *testing.T, db *sqlite.DB, provider string, n, withActuals int, base time.Time) {
	t.Helper()
	for i := 0; i < n; i++ {
		payload := map[string]any{"stop_hook_active": false}
		if i < withActuals {
			payload["total_tokens"] = 1200 + i
			payload["input_tokens"] = 1000
			payload["output_tokens"] = 200 + i
		}
		persistCodexObservation(t, db, codexObservation(
			fmt.Sprintf("ev-%s-turn-%d", provider, i), "sess-"+provider,
			v1.EventProviderTurnCompleted, base.Add(time.Duration(i)*time.Minute), payload))
	}
	// codexObservation stamps provider "codex"; re-stamp when seeding
	// another provider's rows.
	if provider != "codex" {
		if _, err := db.Conn().ExecContext(context.Background(),
			`UPDATE events SET provider = ? WHERE session_id = ?`, provider, "sess-"+provider); err != nil {
			t.Fatalf("re-stamp provider: %v", err)
		}
	}
}

func TestDoctor_CaptureHealth_CoverageAndTimestamps(t *testing.T) {
	db := openCodexStatusDB(t)
	base := time.Date(2026, 7, 16, 9, 0, 0, 0, time.UTC)

	// claude: 5 completed turns, 3 with token actuals → ok at 60%.
	seedTurnCompleted(t, db, "claude", 5, 3, base)

	result := orchestrator.Doctor(context.Background(), orchestrator.DoctorDeps{DB: db})

	events := findCheck(t, result, "capture:events:claude")
	if events.Status != orchestrator.CheckOK {
		t.Errorf("capture:events:claude Status = %q, want ok (detail: %s)", events.Status, events.Detail)
	}
	if !strings.Contains(events.Detail, "5 events captured, latest 2026-07-16T09:04:00Z") {
		t.Errorf("capture:events:claude Detail = %q, want count + last-capture timestamp", events.Detail)
	}

	actuals := findCheck(t, result, "capture:turn-actuals:claude")
	if actuals.Status != orchestrator.CheckOK {
		t.Errorf("turn-actuals Status = %q, want ok (partial coverage is expected)", actuals.Status)
	}
	if !strings.Contains(actuals.Detail, "3 of the last 5 turn.completed events carry token actuals (60%)") {
		t.Errorf("turn-actuals Detail = %q, want the observed coverage rate", actuals.Detail)
	}
	if !result.Healthy {
		t.Error("Healthy = false, want true — partial coverage is not a failure")
	}
}

// TestDoctor_CaptureHealth_ZeroCoverageFailsLoudly is the silent-breakage
// guard: turns completing with ZERO token actuals is a FAIL that flips
// the overall report unhealthy — the one state that must never pass
// quietly (#90: every aggregate surface would keep rendering from
// nothing).
func TestDoctor_CaptureHealth_ZeroCoverageFailsLoudly(t *testing.T) {
	db := openCodexStatusDB(t)
	base := time.Date(2026, 7, 16, 9, 0, 0, 0, time.UTC)
	seedTurnCompleted(t, db, "codex", 4, 0, base)

	result := orchestrator.Doctor(context.Background(), orchestrator.DoctorDeps{DB: db})
	actuals := findCheck(t, result, "capture:turn-actuals:codex")
	if actuals.Status != orchestrator.CheckFail {
		t.Fatalf("turn-actuals Status = %q, want FAIL at 0%% coverage (detail: %s)", actuals.Status, actuals.Detail)
	}
	if !strings.Contains(actuals.Detail, "0 of the last 4") || !strings.Contains(actuals.Detail, "token capture appears broken") {
		t.Errorf("turn-actuals Detail = %q, want the loud broken-capture wording", actuals.Detail)
	}
	if result.Healthy {
		t.Error("Healthy = true, want false — zero token-actual coverage must fail the report")
	}
}

func TestDoctor_CaptureHealth_EmptyDBWarnsButStaysHealthy(t *testing.T) {
	db := openCodexStatusDB(t)
	result := orchestrator.Doctor(context.Background(), orchestrator.DoctorDeps{DB: db})

	events := findCheck(t, result, "capture:events")
	if events.Status != orchestrator.CheckWarn {
		t.Errorf("capture:events Status = %q, want warn on a fresh install", events.Status)
	}
	runway := findCheck(t, result, "capture:runway")
	if runway.Status != orchestrator.CheckWarn || !strings.Contains(runway.Detail, "no quota telemetry") {
		t.Errorf("capture:runway = %+v, want the no-telemetry warn", runway)
	}
	if !result.Healthy {
		t.Error("Healthy = false, want true — a fresh install warns, it does not fail")
	}
}

func TestDoctor_CaptureHealth_RunwayStates(t *testing.T) {
	db := openCodexStatusDB(t)
	base := time.Date(2026, 7, 16, 9, 0, 0, 0, time.UTC)
	seedCodexSession(t, db, "sess-rw", "/repo/rw", "gpt-5.2-codex", "2026-07-16T08:00:00Z")
	persistCodexObservation(t, db, codexObservation("ev-rw-q", "sess-rw", v1.EventProviderQuotaObserved, base, map[string]any{
		"limit_id": "secondary", "used_percent": 40.0,
	}))

	// Quota telemetry present, zero runway rows → the driver-not-wired warn.
	result := orchestrator.Doctor(context.Background(), orchestrator.DoctorDeps{DB: db})
	runway := findCheck(t, result, "capture:runway")
	if runway.Status != orchestrator.CheckWarn || !strings.Contains(runway.Detail, "no runway forecasts") {
		t.Errorf("capture:runway = %+v, want the driver-not-wired warn", runway)
	}

	// Drive a real forecast through the production driver; the check
	// flips to ok with the row count.
	store := &orchestrator.RunwayForecastStore{DB: db, Clock: fixedClock{t: base.Add(time.Minute)}}
	persistCodexObservation(t, db, codexObservation("ev-rw-q2", "sess-rw", v1.EventProviderQuotaObserved, base.Add(30*time.Second), map[string]any{
		"limit_id": "secondary", "used_percent": 41.0,
	}))
	store.DriveRunway(context.Background(), "sess-rw")

	result = orchestrator.Doctor(context.Background(), orchestrator.DoctorDeps{DB: db})
	runway = findCheck(t, result, "capture:runway")
	if runway.Status != orchestrator.CheckOK || !strings.Contains(runway.Detail, "runway forecasts persisted") {
		t.Errorf("capture:runway = %+v, want ok once rows exist", runway)
	}
}
