// store_test.go: Store.TodaySpend against a real migrated scratch DB —
// the cumulative-vs-turn-exact aggregation, local-day scoping under a
// non-UTC clock, provider scoping, and the honest no-data ok=false.
package pace_test

import (
	"context"
	"path/filepath"
	"testing"
	"time"

	"github.com/huaiche94/auspex/internal/domain"
	"github.com/huaiche94/auspex/internal/pace"
	"github.com/huaiche94/auspex/internal/storage/sqlite"
	claudetelemetry "github.com/huaiche94/auspex/internal/telemetry/claude"
	v1 "github.com/huaiche94/auspex/pkg/protocol/v1"
)

type fixedClock struct{ t time.Time }

func (c fixedClock) Now() time.Time { return c.t }

func openPaceDB(t *testing.T) *sqlite.DB {
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

func persistUsage(t *testing.T, db *sqlite.DB, id, provider, sessionID string, source domain.MeasurementSource, occurredAt time.Time, payload map[string]any) {
	t.Helper()
	store := claudetelemetry.NewEventStore(db)
	err := store.PersistAll(context.Background(), db, []v1.Event{{
		SchemaVersion:  v1.SchemaVersionEvent,
		EventID:        id,
		EventType:      v1.EventProviderUsageObserved,
		OccurredAt:     occurredAt,
		ObservedAt:     occurredAt,
		IdempotencyKey: "key-" + id,
		Source:         string(source),
		Provider:       provider,
		SessionID:      sessionID,
		Payload:        payload,
	}})
	if err != nil {
		t.Fatalf("PersistAll(%s): %v", id, err)
	}
}

// TestStoreTodaySpend_AggregatesCumulativeDeltasAndTurnSamples is the
// core aggregation case, under a UTC+8 clock so the day boundary is the
// user's midnight (16:00 UTC the previous calendar day):
//
//   - session A (statusline, cumulative): $1.00 yesterday, $1.20 then
//     $1.45 today → today's delta $0.45 (the pre-midnight baseline, not
//     zero, and the NEWEST sample, not the first).
//   - session B (managed, turn-exact): $0.10 + $0.05 today → $0.15.
//   - a cost-less usage event (codex-shaped token payload) contributes
//     nothing and does not count as a session.
//   - a different provider's cost event today is out of scope.
func TestStoreTodaySpend_AggregatesCumulativeDeltasAndTurnSamples(t *testing.T) {
	db := openPaceDB(t)
	tz := time.FixedZone("UTC+8", 8*3600)
	now := time.Date(2026, 7, 16, 10, 0, 0, 0, tz) // day start = 2026-07-15T16:00Z

	// Session A: cumulative statusline series spanning midnight.
	persistUsage(t, db, "a-yday", "claude", "sess-a", domain.SourceStatusLine,
		time.Date(2026, 7, 15, 10, 0, 0, 0, time.UTC), map[string]any{"total_cost_usd": 1.00})
	persistUsage(t, db, "a-1", "claude", "sess-a", domain.SourceStatusLine,
		time.Date(2026, 7, 15, 17, 0, 0, 0, time.UTC), map[string]any{"total_cost_usd": 1.20})
	persistUsage(t, db, "a-2", "claude", "sess-a", domain.SourceStatusLine,
		time.Date(2026, 7, 16, 1, 0, 0, 0, time.UTC), map[string]any{"total_cost_usd": 1.45})
	// Session B: two managed turn-exact samples today.
	persistUsage(t, db, "b-1", "claude", "sess-b", domain.SourceProviderEvent,
		time.Date(2026, 7, 16, 0, 30, 0, 0, time.UTC), map[string]any{"total_cost_usd": 0.10})
	persistUsage(t, db, "b-2", "claude", "sess-b", domain.SourceProviderEvent,
		time.Date(2026, 7, 16, 1, 30, 0, 0, time.UTC), map[string]any{"total_cost_usd": 0.05})
	// Cost-less usage (codex rollout shape) and a foreign provider.
	persistUsage(t, db, "c-1", "claude", "sess-c", domain.SourceProviderEvent,
		time.Date(2026, 7, 16, 1, 0, 0, 0, time.UTC), map[string]any{"input_tokens": 1200, "output_tokens": 300})
	persistUsage(t, db, "x-1", "codex", "sess-x", domain.SourceProviderEvent,
		time.Date(2026, 7, 16, 1, 0, 0, 0, time.UTC), map[string]any{"total_cost_usd": 9.99})

	store := &pace.Store{DB: db, Clock: fixedClock{t: now}}
	spend, ok := store.TodaySpend(context.Background(), "claude")
	if !ok {
		t.Fatal("ok = false, want today's aggregation")
	}
	if diff := spend.SpendUSD - 0.60; diff > 1e-9 || diff < -1e-9 {
		t.Errorf("SpendUSD = %v, want 0.60 (0.45 cumulative delta + 0.15 turn samples)", spend.SpendUSD)
	}
	if spend.Sessions != 2 {
		t.Errorf("Sessions = %d, want 2 (the cost-less session does not count)", spend.Sessions)
	}
	if spend.Day != "2026-07-16" {
		t.Errorf("Day = %q, want the local day 2026-07-16", spend.Day)
	}
	if want := time.Date(2026, 7, 15, 17, 0, 0, 0, time.UTC); !spend.FirstObservedAt.Equal(want) {
		t.Errorf("FirstObservedAt = %v, want %v (the day's first cost observation)", spend.FirstObservedAt, want)
	}
	if want := time.Date(2026, 7, 16, 1, 30, 0, 0, time.UTC); !spend.LastObservedAt.Equal(want) {
		t.Errorf("LastObservedAt = %v, want %v", spend.LastObservedAt, want)
	}

	// The other provider sees only its own event.
	codexSpend, ok := store.TodaySpend(context.Background(), "codex")
	if !ok || codexSpend.SpendUSD != 9.99 || codexSpend.Sessions != 1 {
		t.Errorf("codex spend = (%+v, %v), want its own 9.99", codexSpend, ok)
	}
}

// TestStoreTodaySpend_NoDataIsNotZero: a provider with no cost-bearing
// usage today — including one whose only usage events carry token counts
// but no cost (codex) — is ok=false, never a fabricated $0.00.
func TestStoreTodaySpend_NoDataIsNotZero(t *testing.T) {
	db := openPaceDB(t)
	now := time.Date(2026, 7, 16, 10, 0, 0, 0, time.UTC)

	// Yesterday-only cost data: nothing today.
	persistUsage(t, db, "old-1", "claude", "sess-old", domain.SourceStatusLine,
		time.Date(2026, 7, 15, 10, 0, 0, 0, time.UTC), map[string]any{"total_cost_usd": 4.00})
	// Today, token-only usage (no cost field).
	persistUsage(t, db, "tok-1", "codex", "sess-tok", domain.SourceProviderEvent,
		time.Date(2026, 7, 16, 9, 0, 0, 0, time.UTC), map[string]any{"input_tokens": 100})

	store := &pace.Store{DB: db, Clock: fixedClock{t: now}}
	if _, ok := store.TodaySpend(context.Background(), "claude"); ok {
		t.Error("claude: ok = true with no cost observations today, want false")
	}
	if _, ok := store.TodaySpend(context.Background(), "codex"); ok {
		t.Error("codex: ok = true with token-only telemetry, want false (no price-table fabrication)")
	}
}

// TestStoreTodaySpend_CumulativeResetClampsToZero: a cumulative counter
// that reads BELOW its pre-midnight baseline (provider-side reset or
// correction) contributes zero, never negative spend.
func TestStoreTodaySpend_CumulativeResetClampsToZero(t *testing.T) {
	db := openPaceDB(t)
	now := time.Date(2026, 7, 16, 10, 0, 0, 0, time.UTC)
	persistUsage(t, db, "r-yday", "claude", "sess-r", domain.SourceStatusLine,
		time.Date(2026, 7, 15, 23, 0, 0, 0, time.UTC), map[string]any{"total_cost_usd": 5.00})
	persistUsage(t, db, "r-today", "claude", "sess-r", domain.SourceStatusLine,
		time.Date(2026, 7, 16, 9, 0, 0, 0, time.UTC), map[string]any{"total_cost_usd": 0.20})

	store := &pace.Store{DB: db, Clock: fixedClock{t: now}}
	spend, ok := store.TodaySpend(context.Background(), "claude")
	if !ok {
		t.Fatal("ok = false, want an aggregation (data WAS observed today)")
	}
	if spend.SpendUSD != 0 {
		t.Errorf("SpendUSD = %v, want 0 (negative delta clamps)", spend.SpendUSD)
	}
	if spend.Sessions != 1 {
		t.Errorf("Sessions = %d, want 1", spend.Sessions)
	}
}

// TestStoreTodaySpend_FailOpen: nil receiver/DB/provider never error and
// never panic — the statusline keeps rendering.
func TestStoreTodaySpend_FailOpen(t *testing.T) {
	var nilStore *pace.Store
	if _, ok := nilStore.TodaySpend(context.Background(), "claude"); ok {
		t.Error("nil receiver must be ok=false")
	}
	if _, ok := (&pace.Store{}).TodaySpend(context.Background(), "claude"); ok {
		t.Error("nil DB must be ok=false")
	}
	if _, ok := (&pace.Store{DB: openPaceDB(t)}).TodaySpend(context.Background(), ""); ok {
		t.Error("empty provider must be ok=false")
	}
}
