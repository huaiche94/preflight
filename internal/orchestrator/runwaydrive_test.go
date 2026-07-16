// runwaydrive_test.go: M10 — the hook-side driver that fills
// runway_forecasts from persisted quota telemetry, and the read-back the
// evaluation/policy path consumes (ADR-041). Tests run against a real
// migrated DB (openturn_test.go / codexstatus_test.go precedent), package
// orchestrator_test, exported surfaces only.
package orchestrator_test

import (
	"context"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/huaiche94/auspex/internal/domain"
	"github.com/huaiche94/auspex/internal/evaluation"
	"github.com/huaiche94/auspex/internal/orchestrator"
	"github.com/huaiche94/auspex/internal/storage/sqlite"
	claudetelemetry "github.com/huaiche94/auspex/internal/telemetry/claude"
	v1 "github.com/huaiche94/auspex/pkg/protocol/v1"
)

func openRunwayDB(t *testing.T) *sqlite.DB {
	t.Helper()
	db, err := sqlite.Open(context.Background(), filepath.Join(t.TempDir(), "auspex.db"))
	if err != nil {
		t.Fatalf("sqlite.Open: %v", err)
	}
	t.Cleanup(func() { _ = db.Close() })
	migrations, err := sqlite.AllMigrations()
	if err != nil {
		t.Fatalf("AllMigrations: %v", err)
	}
	if err := db.Migrate(context.Background(), migrations); err != nil {
		t.Fatalf("Migrate: %v", err)
	}
	return db
}

// seedRunwaySession writes the repositories -> worktrees -> provider_sessions
// chain the issue-#17 bootstrap creates, so runway_forecasts' session_id FK
// resolves.
func seedRunwaySession(t *testing.T, db *sqlite.DB, sessionID, provider string) {
	t.Helper()
	ctx := context.Background()
	exec := func(q string, args ...any) {
		if _, err := db.Conn().ExecContext(ctx, q, args...); err != nil {
			t.Fatalf("seed %q: %v", q, err)
		}
	}
	root := "/repo/" + sessionID
	at := "2026-07-14T10:00:00Z"
	exec(`INSERT OR IGNORE INTO repositories (id, canonical_root, git_common_dir, created_at, last_seen_at)
	      VALUES (?, ?, ?, ?, ?)`, "repo-"+sessionID, root, root+"/.git", at, at)
	exec(`INSERT OR IGNORE INTO worktrees (id, repository_id, root_path, git_dir, created_at, last_seen_at)
	      VALUES (?, ?, ?, ?, ?, ?)`, "wt-"+sessionID, "repo-"+sessionID, root, root+"/.git", at, at)
	exec(`INSERT INTO provider_sessions (id, worktree_id, provider, provider_session_id, invocation_mode, started_at)
	      VALUES (?, ?, ?, ?, 'native-hook', ?)`, sessionID, "wt-"+sessionID, provider, sessionID, at)
}

// persistQuota writes one provider.quota.observed event at observedAt with
// the given limit_id/used_percent (and optional resets_at), through the real
// EventStore — exactly the row normalizer.go's quotaEvent produces.
func persistQuota(t *testing.T, db *sqlite.DB, id, sessionID, provider, limitID string, usedPercent float64, observedAt time.Time, resetsAt *time.Time) {
	t.Helper()
	payload := map[string]any{"limit_id": limitID, "used_percent": usedPercent}
	if resetsAt != nil {
		payload["resets_at"] = resetsAt.UTC().Format(time.RFC3339Nano)
	}
	ev := v1.Event{
		SchemaVersion:  v1.SchemaVersionEvent,
		EventID:        id,
		EventType:      v1.EventProviderQuotaObserved,
		OccurredAt:     observedAt,
		ObservedAt:     observedAt,
		IdempotencyKey: "key-" + id,
		Source:         "status_line",
		Provider:       provider,
		SessionID:      sessionID,
		Payload:        payload,
	}
	store := claudetelemetry.NewEventStore(db)
	if err := store.PersistAll(context.Background(), db, []v1.Event{ev}); err != nil {
		t.Fatalf("PersistAll quota %s: %v", id, err)
	}
}

// countForecasts returns how many runway_forecasts rows exist for a session.
func countForecasts(t *testing.T, db *sqlite.DB, sessionID string) int {
	t.Helper()
	var n int
	if err := db.Conn().QueryRowContext(context.Background(),
		`SELECT COUNT(*) FROM runway_forecasts WHERE session_id = ?`, sessionID,
	).Scan(&n); err != nil {
		t.Fatalf("count runway_forecasts: %v", err)
	}
	return n
}

func newRunwayStore(db *sqlite.DB, now time.Time) *orchestrator.RunwayForecastStore {
	return &orchestrator.RunwayForecastStore{DB: db, Clock: fixedClock{t: now}}
}

// --- DriveRunway: cold start -----------------------------------------------

func TestRunwayForecastStore_DriveRunway_ColdStartPersistsForecast(t *testing.T) {
	db := openRunwayDB(t)
	base := time.Date(2026, 7, 14, 12, 0, 0, 0, time.UTC)
	seedRunwaySession(t, db, "sess-cold", "claude")

	// A single quota sample: no burn rate is computable yet (cold start),
	// but a forecast row must still be persisted from the very first sample.
	persistQuota(t, db, "q1", "sess-cold", "claude", "seven_day", 40.0, base, nil)

	store := newRunwayStore(db, base)
	store.DriveRunway(context.Background(), "sess-cold")

	if got := countForecasts(t, db, "sess-cold"); got != 1 {
		t.Fatalf("runway_forecasts rows = %d, want 1 after first drive", got)
	}

	ds := evaluation.NewSQLDataSource(db)
	f, ok, err := ds.RunwayForecast(context.Background(), "sess-cold")
	if err != nil || !ok {
		t.Fatalf("RunwayForecast read = (ok=%v, err=%v), want a row", ok, err)
	}
	if f.Calibrated {
		t.Error("cold-start forecast must be uncalibrated")
	}
	if f.BurnRateP50 != nil {
		t.Errorf("cold start must have no burn rate, got %v", *f.BurnRateP50)
	}
	if !containsReason(f.ReasonCodes, "prediction_cold_start") {
		t.Errorf("reason codes = %v, want prediction_cold_start", f.ReasonCodes)
	}
	if f.CurrentUsedPercent == nil || *f.CurrentUsedPercent != 40.0 {
		t.Errorf("CurrentUsedPercent = %v, want 40", f.CurrentUsedPercent)
	}
}

// --- DriveRunway: two samples => burn rate, read back by evaluation --------

func TestRunwayForecastStore_DriveRunway_BurnRatePersistedAndReadBack(t *testing.T) {
	db := openRunwayDB(t)
	base := time.Date(2026, 7, 14, 12, 0, 0, 0, time.UTC)
	seedRunwaySession(t, db, "sess-burn", "claude")

	// Two samples, 5 minutes apart, 40% -> 50% => 2%/min burn rate.
	persistQuota(t, db, "q1", "sess-burn", "claude", "seven_day", 40.0, base.Add(-5*time.Minute), nil)
	persistQuota(t, db, "q2", "sess-burn", "claude", "seven_day", 50.0, base, nil)

	store := newRunwayStore(db, base)
	store.DriveRunway(context.Background(), "sess-burn")

	// Read back through the SAME query the policy pipeline uses (ADR-041):
	// this is the policy-consumption proof — the table now returns real data.
	ds := evaluation.NewSQLDataSource(db)
	f, ok, err := ds.RunwayForecast(context.Background(), "sess-burn")
	if err != nil || !ok {
		t.Fatalf("RunwayForecast read = (ok=%v, err=%v), want a row", ok, err)
	}
	if f.BurnRateP50 == nil {
		t.Fatal("BurnRateP50 = nil, want a computed burn rate from two samples")
	}
	if got := *f.BurnRateP50; got < 1.99 || got > 2.01 {
		t.Errorf("BurnRateP50 = %v, want ~2.0 %%/min", got)
	}
	if f.EstimatedTimeToLimitP50Seconds == nil {
		t.Fatal("EstimatedTimeToLimitP50Seconds = nil, want an ETA once a burn rate exists")
	}
	// remaining 50% / 2%min = 25 min = 1500s.
	if got := *f.EstimatedTimeToLimitP50Seconds; got < 1490 || got > 1510 {
		t.Errorf("ETA P50 = %ds, want ~1500", got)
	}
	if f.Confidence != domain.ConfidenceMedium {
		t.Errorf("Confidence = %q, want medium (two fresh samples)", f.Confidence)
	}
	if f.HorizonSeconds != 600 {
		t.Errorf("HorizonSeconds = %d, want 600 (default)", f.HorizonSeconds)
	}
}

// --- DriveRunway: projected exceeds within horizon => risk + hint ----------

func TestRunwayForecastStore_DriveRunway_ProjectedExceedsWithinHorizon(t *testing.T) {
	db := openRunwayDB(t)
	base := time.Date(2026, 7, 14, 12, 0, 0, 0, time.UTC)
	seedRunwaySession(t, db, "sess-hot", "claude")

	// 90% -> 92% over 1 min = 2%/min. Over the 10-min horizon that projects
	// to 92 + 20 = 112% (>100), i.e. exhaustion within the horizon.
	persistQuota(t, db, "q1", "sess-hot", "claude", "seven_day", 90.0, base.Add(-time.Minute), nil)
	persistQuota(t, db, "q2", "sess-hot", "claude", "seven_day", 92.0, base, nil)

	store := newRunwayStore(db, base)
	store.DriveRunway(context.Background(), "sess-hot")

	ds := evaluation.NewSQLDataSource(db)
	f, ok, _ := ds.RunwayForecast(context.Background(), "sess-hot")
	if !ok {
		t.Fatal("want a forecast row")
	}
	if !containsReason(f.ReasonCodes, "quota_projected_exceeds_limit_within_horizon") {
		t.Errorf("reason codes = %v, want quota_projected_exceeds_limit_within_horizon", f.ReasonCodes)
	}
	if f.RiskScore < 0.8 {
		t.Errorf("RiskScore = %v, want >= 0.8 for a within-horizon projected exceed", f.RiskScore)
	}

	// The statusline hint must reflect the within-horizon exhaustion and
	// therefore contribute an ETA segment.
	hint, hok := store.LatestRunwayHint(context.Background(), "sess-hot")
	if !hok {
		t.Fatal("LatestRunwayHint ok=false, want a hint")
	}
	if !hint.ProjectedExceedsWithinHorizon {
		t.Error("hint.ProjectedExceedsWithinHorizon = false, want true")
	}
	if hint.TimeToLimitP50Seconds == nil {
		t.Error("hint.TimeToLimitP50Seconds = nil, want an ETA")
	}
}

// --- DriveRunway: identical-% spam still finds the last real change --------

func TestRunwayForecastStore_DriveRunway_SkipsIdenticalPercentSamples(t *testing.T) {
	db := openRunwayDB(t)
	base := time.Date(2026, 7, 14, 12, 0, 0, 0, time.UTC)
	seedRunwaySession(t, db, "sess-spam", "claude")

	// A real move (40 -> 50 over 5 min), then TWO more identical-percent
	// renders (the statusline re-emitting the same 50% snapshot). The burn
	// rate must be measured against the last DIFFERENT sample (40%), not the
	// zero-delta neighbor — otherwise identical-% spam would flatten it to 0.
	persistQuota(t, db, "q1", "sess-spam", "claude", "seven_day", 40.0, base.Add(-5*time.Minute), nil)
	persistQuota(t, db, "q2", "sess-spam", "claude", "seven_day", 50.0, base.Add(-2*time.Second), nil)
	persistQuota(t, db, "q3", "sess-spam", "claude", "seven_day", 50.0, base.Add(-1*time.Second), nil)
	persistQuota(t, db, "q4", "sess-spam", "claude", "seven_day", 50.0, base, nil)

	store := newRunwayStore(db, base)
	store.DriveRunway(context.Background(), "sess-spam")

	ds := evaluation.NewSQLDataSource(db)
	f, ok, _ := ds.RunwayForecast(context.Background(), "sess-spam")
	if !ok {
		t.Fatal("want a forecast row")
	}
	if f.BurnRateP50 == nil {
		t.Fatal("BurnRateP50 = nil — identical-% spam flattened the burn rate (regression)")
	}
	// 10% over 5 min = 2%/min, measured across the q1->q4 span.
	if got := *f.BurnRateP50; got < 1.99 || got > 2.01 {
		t.Errorf("BurnRateP50 = %v, want ~2.0 %%/min against the last real change", got)
	}
}

// --- DriveRunway: idempotent on the same sample ----------------------------

func TestRunwayForecastStore_DriveRunway_IdempotentOnSameSample(t *testing.T) {
	db := openRunwayDB(t)
	base := time.Date(2026, 7, 14, 12, 0, 0, 0, time.UTC)
	seedRunwaySession(t, db, "sess-idem", "codex")

	persistQuota(t, db, "q1", "sess-idem", "codex", "secondary", 55.0, base, nil)

	store := newRunwayStore(db, base)
	store.DriveRunway(context.Background(), "sess-idem")
	store.DriveRunway(context.Background(), "sess-idem") // re-entrant Stop, same sample

	if got := countForecasts(t, db, "sess-idem"); got != 1 {
		t.Errorf("runway_forecasts rows = %d, want 1 (idempotent on the same sample)", got)
	}
}

// --- DriveRunway: combine windows takes the worst --------------------------

func TestRunwayForecastStore_DriveRunway_CombinesWindowsTakesWorst(t *testing.T) {
	db := openRunwayDB(t)
	base := time.Date(2026, 7, 14, 12, 0, 0, 0, time.UTC)
	seedRunwaySession(t, db, "sess-multi", "codex")

	// A calm primary window and a critical secondary window observed at the
	// same time — the persisted forecast must be the worst (secondary).
	persistQuota(t, db, "p1", "sess-multi", "codex", "primary", 10.0, base, nil)
	persistQuota(t, db, "s1", "sess-multi", "codex", "secondary", 99.0, base, nil)

	store := newRunwayStore(db, base)
	store.DriveRunway(context.Background(), "sess-multi")

	ds := evaluation.NewSQLDataSource(db)
	f, ok, _ := ds.RunwayForecast(context.Background(), "sess-multi")
	if !ok {
		t.Fatal("want a forecast row")
	}
	if f.LimitID != "secondary" {
		t.Errorf("LimitID = %q, want secondary (the worst window)", f.LimitID)
	}
	if f.RiskScore < 0.95 {
		t.Errorf("RiskScore = %v, want ~1.0 (secondary at 99%%)", f.RiskScore)
	}
}

// --- DriveRunway: no quota => no row (honest cold start) --------------------

func TestRunwayForecastStore_DriveRunway_NoQuotaNoRow(t *testing.T) {
	db := openRunwayDB(t)
	seedRunwaySession(t, db, "sess-empty", "claude")

	store := newRunwayStore(db, time.Now())
	store.DriveRunway(context.Background(), "sess-empty")

	if got := countForecasts(t, db, "sess-empty"); got != 0 {
		t.Errorf("runway_forecasts rows = %d, want 0 with no quota telemetry", got)
	}
}

// --- fail-open: nil receiver, nil DB, missing session FK -------------------

func TestRunwayForecastStore_FailOpen(t *testing.T) {
	// nil receiver and nil DB must not panic and must be no-op / ok=false.
	var nilStore *orchestrator.RunwayForecastStore
	nilStore.DriveRunway(context.Background(), "x")
	if _, ok := nilStore.LatestRunwayHint(context.Background(), "x"); ok {
		t.Error("nil receiver LatestRunwayHint must be ok=false")
	}
	empty := &orchestrator.RunwayForecastStore{}
	empty.DriveRunway(context.Background(), "x")
	if _, ok := empty.LatestRunwayHint(context.Background(), "x"); ok {
		t.Error("nil DB LatestRunwayHint must be ok=false")
	}

	// A session with quota telemetry but NO provider_sessions row: the
	// FK-violating insert is swallowed (no row, no panic, no error surfaced).
	db := openRunwayDB(t)
	base := time.Date(2026, 7, 14, 12, 0, 0, 0, time.UTC)
	// Persisting the event itself needs no session FK (events.session_id is
	// unconstrained), so this exercises the runway insert's fail-open path.
	persistQuota(t, db, "q1", "ghost", "claude", "seven_day", 40.0, base, nil)
	store := newRunwayStore(db, base)
	store.DriveRunway(context.Background(), "ghost") // must not panic
	if got := countForecasts(t, db, "ghost"); got != 0 {
		t.Errorf("runway_forecasts rows = %d, want 0 (FK insert fails open)", got)
	}
}

// --- LatestRunwayHint gating: headroom stays quiet -------------------------

func TestRunwayStatusHint_HeadroomProducesNoETA(t *testing.T) {
	db := openRunwayDB(t)
	base := time.Date(2026, 7, 14, 12, 0, 0, 0, time.UTC)
	seedRunwaySession(t, db, "sess-calm", "claude")

	// 10% -> 11% over 1 min: burn rate exists, but projected well under the
	// limit within the horizon => no within-horizon warning => quiet line.
	persistQuota(t, db, "q1", "sess-calm", "claude", "seven_day", 10.0, base.Add(-time.Minute), nil)
	persistQuota(t, db, "q2", "sess-calm", "claude", "seven_day", 11.0, base, nil)

	store := newRunwayStore(db, base)
	store.DriveRunway(context.Background(), "sess-calm")

	hint, ok := store.LatestRunwayHint(context.Background(), "sess-calm")
	if !ok {
		t.Fatal("want a hint")
	}
	if hint.ProjectedExceedsWithinHorizon {
		t.Error("calm session must not project a within-horizon exceed")
	}
}

// --- statusline render (evaluation.StatusLineText) -------------------------

func TestStatusLineText_RunwaySegment(t *testing.T) {
	secs := int64(150)
	withRunway := evaluation.StatusLineText(evaluation.StatusLineInput{
		Model:                    "opus",
		RunwayTimeToLimitSeconds: &secs,
	})
	if !strings.Contains(withRunway, "runway") {
		t.Errorf("line = %q, want a runway segment", withRunway)
	}
	if !strings.Contains(withRunway, "~2m") {
		t.Errorf("line = %q, want a ~2m ETA (150s rounded to minutes)", withRunway)
	}

	// nil (the common headroom case) renders no runway segment.
	without := evaluation.StatusLineText(evaluation.StatusLineInput{Model: "opus"})
	if strings.Contains(without, "runway") {
		t.Errorf("line = %q, want no runway segment when nil", without)
	}
}

func containsReason(reasons []string, want string) bool {
	for _, r := range reasons {
		if r == want {
			return true
		}
	}
	return false
}

// --- hook wiring: the Stop/statusline handlers invoke the driver ----------

// recordingRunway records DriveRunway's session ids and answers hints from a
// fixed table — a test double for the RunwayDriver seam HookDeps.Runway
// carries, proving the hook handlers actually drive the forecaster.
type recordingRunway struct {
	driven []domain.SessionID
	hints  map[domain.SessionID]orchestrator.RunwayHint
}

func (r *recordingRunway) DriveRunway(_ context.Context, sessionID domain.SessionID) {
	r.driven = append(r.driven, sessionID)
}

func (r *recordingRunway) LatestRunwayHint(_ context.Context, sessionID domain.SessionID) (orchestrator.RunwayHint, bool) {
	h, ok := r.hints[sessionID]
	return h, ok
}

func TestHandleStop_DrivesRunway(t *testing.T) {
	deps := baseHookDeps()
	rec := &recordingRunway{}
	deps.Runway = rec

	if _, err := orchestrator.HandleStop(context.Background(), deps, readFixture(t, "stop", "normal.json")); err != nil {
		t.Fatalf("HandleStop: %v", err)
	}
	if len(rec.driven) != 1 || rec.driven[0] != "sess_01H9X8K7QZ3M4N5P6R7S8T9V0W" {
		t.Errorf("driven = %v, want one drive for the Stop session", rec.driven)
	}
}

func TestHandleStatusLine_DrivesRunway(t *testing.T) {
	deps := baseHookDeps()
	rec := &recordingRunway{}
	deps.Runway = rec

	// Claude's quota lands at the statusline, so this is the primary drive
	// point for Claude — the ingest path must invoke the driver.
	if _, err := orchestrator.HandleStatusLine(context.Background(), deps, readFixture(t, "statusline", "normal.json")); err != nil {
		t.Fatalf("HandleStatusLine: %v", err)
	}
	if len(rec.driven) != 1 || rec.driven[0] != "sess_01H9X8K7QZ3M4N5P6R7S8T9V0W" {
		t.Errorf("driven = %v, want one drive for the statusline session", rec.driven)
	}
}

func TestHandleStop_NilRunwayFailsOpen(t *testing.T) {
	deps := baseHookDeps() // Runway left nil
	if _, err := orchestrator.HandleStop(context.Background(), deps, readFixture(t, "stop", "normal.json")); err != nil {
		t.Fatalf("HandleStop with nil Runway must not error: %v", err)
	}
}
