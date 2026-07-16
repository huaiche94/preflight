// pace_test.go: the pure pace math under a fixed clock — day boundaries
// in the clock's own zone, and the end-of-day extrapolation's honesty
// rules (no rate from no spend, no rate from no elapsed time).
package pace_test

import (
	"math"
	"testing"
	"time"

	"github.com/huaiche94/auspex/internal/pace"
)

func TestDayBounds_UsesTheClocksOwnZone(t *testing.T) {
	tz := time.FixedZone("UTC+8", 8*3600)
	now := time.Date(2026, 7, 16, 10, 30, 0, 0, tz)

	start, end := pace.DayBounds(now)
	if want := time.Date(2026, 7, 16, 0, 0, 0, 0, tz); !start.Equal(want) {
		t.Errorf("start = %v, want local midnight %v", start, want)
	}
	if want := time.Date(2026, 7, 17, 0, 0, 0, 0, tz); !end.Equal(want) {
		t.Errorf("end = %v, want next local midnight %v", end, want)
	}
	// The local day starts 16:00 UTC the previous calendar day — the
	// boundary is the USER's midnight, not UTC's.
	if want := time.Date(2026, 7, 15, 16, 0, 0, 0, time.UTC); !start.UTC().Equal(want) {
		t.Errorf("start UTC = %v, want %v", start.UTC(), want)
	}
}

func TestProjectEndOfDay_LinearStretchOfObservedRate(t *testing.T) {
	// $2.00 observed over 09:00→12:00 (3h) → $0.667/h; 12h remain until
	// midnight → $2.00 + $8.00 = $10.00.
	spend := pace.TodaySpend{
		SpendUSD:        2.0,
		FirstObservedAt: time.Date(2026, 7, 16, 9, 0, 0, 0, time.UTC),
	}
	now := time.Date(2026, 7, 16, 12, 0, 0, 0, time.UTC)

	projected, ok := pace.ProjectEndOfDay(spend, now)
	if !ok {
		t.Fatal("ok = false, want a projection")
	}
	if math.Abs(projected-10.0) > 1e-9 {
		t.Errorf("projected = %v, want 10.0", projected)
	}
}

func TestProjectEndOfDay_NoHonestRate(t *testing.T) {
	first := time.Date(2026, 7, 16, 9, 0, 0, 0, time.UTC)
	cases := map[string]struct {
		spend pace.TodaySpend
		now   time.Time
	}{
		"zero spend": {
			spend: pace.TodaySpend{SpendUSD: 0, FirstObservedAt: first},
			now:   first.Add(2 * time.Hour),
		},
		"no elapsed time": {
			spend: pace.TodaySpend{SpendUSD: 1.5, FirstObservedAt: first},
			now:   first,
		},
		"clock behind first observation": {
			spend: pace.TodaySpend{SpendUSD: 1.5, FirstObservedAt: first},
			now:   first.Add(-time.Minute),
		},
		// A rate observed over seconds is not a day-scale pace — the
		// MinPaceWindow guard (its absence printed a five-digit "pace"
		// from a 1-second window in this feature's E2E).
		"window shorter than MinPaceWindow": {
			spend: pace.TodaySpend{SpendUSD: 1.5, FirstObservedAt: first},
			now:   first.Add(pace.MinPaceWindow - time.Second),
		},
	}
	for name, tc := range cases {
		if _, ok := pace.ProjectEndOfDay(tc.spend, tc.now); ok {
			t.Errorf("%s: ok = true, want false — no fabricated pace", name)
		}
	}
}

func TestProjectEndOfDay_AtMidnightBoundaryClampsRemaining(t *testing.T) {
	// now exactly at the NEXT day's midnight relative to the observation
	// window never yields a negative remaining span.
	spend := pace.TodaySpend{
		SpendUSD:        3.0,
		FirstObservedAt: time.Date(2026, 7, 16, 23, 0, 0, 0, time.UTC),
	}
	now := time.Date(2026, 7, 17, 0, 0, 0, 0, time.UTC)
	projected, ok := pace.ProjectEndOfDay(spend, now)
	if !ok {
		t.Fatal("ok = false, want a projection")
	}
	// At the fresh day's midnight a full 24h remain of the NEW day; the
	// projection stays finite and ≥ the observed spend.
	if projected < 3.0 {
		t.Errorf("projected = %v, want >= observed spend", projected)
	}
}
