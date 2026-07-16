// Package pace computes the aggregate-first "today's spend + pace"
// observation surface (issue #90 Phase A deliverable 2). Per-turn point
// forecasts are 7–9× off at median (PR #79) and likely irreducibly so;
// cumulative day-level spend curves are smooth — this package aggregates
// the cost actuals the hooks already capture into "today $X" and
// extrapolates the observed burn rate to the end of the local day as an
// explicitly labeled pace, never a forecast promise.
//
// # Honesty rules (Constitution §7 / ADD principle 1)
//
//   - Unknown is not zero: a day with no cost-bearing observations yields
//     ok=false, and callers omit the segment rather than printing $0.00.
//   - The extrapolation is a PACE — a linear stretch of today's observed
//     average rate, labeled "pace"/"~" on every rendered surface — not a
//     calibrated prediction. No probability is ever attached.
//
// # What "today" means
//
// Today is the current LOCAL calendar day of the injected clock's Now()
// (production wires internal/clock, whose Now() is time.Now() in the
// process's local zone). Day boundaries are computed in Now's own
// location, so a user in UTC+8 sees their midnight, not UTC's.
package pace

import "time"

// TodaySpend is one provider's aggregated cost actuals for the current
// local day, as read back from captured provider.usage.observed events
// (see Store.TodaySpend for the exact aggregation).
type TodaySpend struct {
	// Provider is the frozen provider identifier the aggregation was
	// scoped to ("claude", "codex", ...).
	Provider string
	// Day is the local calendar day the figures cover, YYYY-MM-DD.
	Day string
	// SpendUSD is the observed cost accrued today across every session of
	// Provider: per-session cumulative deltas (statusline series) plus
	// per-turn samples (managed runs). It can legitimately be 0.00 when
	// sessions were observed today but their cumulative cost did not move
	// — that is a measurement, distinct from the no-data ok=false case.
	SpendUSD float64
	// Sessions is how many distinct sessions contributed observations.
	Sessions int
	// FirstObservedAt/LastObservedAt bound today's cost-bearing
	// observations — the active window the pace rate is measured over.
	FirstObservedAt time.Time
	LastObservedAt  time.Time
}

// DayBounds returns the [start, end) of now's calendar day in now's own
// location. end is the next day's midnight (exclusive).
func DayBounds(now time.Time) (start, end time.Time) {
	y, m, d := now.Date()
	start = time.Date(y, m, d, 0, 0, 0, 0, now.Location())
	return start, start.AddDate(0, 0, 1)
}

// MinPaceWindow is the minimum observed active window before a pace is
// extrapolated at all. A rate measured over seconds is arithmetic, not a
// pace — one $1.40 sample observed moments ago linearly stretched to
// midnight prints a five-digit absurdity (caught live in this feature's
// own E2E). Below this window the surface shows today's figure alone;
// the number appears once there is a day-scale-honest rate behind it.
const MinPaceWindow = 10 * time.Minute

// ProjectEndOfDay linearly extrapolates today's observed spend to the end
// of now's local day: the average rate over the observed active window
// (FirstObservedAt → now) continued until midnight. This is a pace, not a
// forecast — it assumes the observed rate simply continues, and callers
// must label it so ("pace", "~").
//
// ok=false when no honest rate exists: zero/negative spend (nothing to
// extrapolate — a $0.00 rate stretched to midnight is still $0.00 and
// adds no information), or an observed window shorter than MinPaceWindow
// (including now not after FirstObservedAt at all).
func ProjectEndOfDay(t TodaySpend, now time.Time) (projectedUSD float64, ok bool) {
	elapsed := now.Sub(t.FirstObservedAt)
	if t.SpendUSD <= 0 || elapsed < MinPaceWindow {
		return 0, false
	}
	_, dayEnd := DayBounds(now)
	remaining := dayEnd.Sub(now)
	if remaining < 0 {
		remaining = 0
	}
	rate := t.SpendUSD / elapsed.Hours()
	return t.SpendUSD + rate*remaining.Hours(), true
}
