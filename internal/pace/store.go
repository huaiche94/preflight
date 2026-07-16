// store.go: the SQL read helper behind the pace surface — a small,
// read-only aggregation over the events table (the single source of
// truth every hook already writes), mirroring the fail-open read-store
// conventions of internal/orchestrator's CodexStatusStore/
// RunwayForecastStore: nil receiver/DB, a query error, or no data all
// degrade to ok=false, never an error — the statusline must keep
// rendering when this store cannot answer.
package pace

import (
	"context"
	"encoding/json"
	"time"

	"github.com/huaiche94/auspex/internal/domain"
	"github.com/huaiche94/auspex/internal/storage/sqlite"
	v1 "github.com/huaiche94/auspex/pkg/protocol/v1"
)

// Store reads today's spend for a provider from persisted
// provider.usage.observed events.
type Store struct {
	DB *sqlite.DB
	// Clock supplies "now" (and therefore "today"). nil falls back to
	// time.Now — production injects the same domain.Clock every other
	// hook seam uses.
	Clock domain.Clock
}

// sqlDayBoundary formats a day boundary for lexical comparison against
// the events table's occurred_at column (RFC3339Nano UTC text, written
// by telemetry's formatTime). The fraction is kept at full width rather
// than RFC3339Nano's trimmed form so that a stored value at exactly the
// boundary instant — with or without a fractional part — always compares
// >= the boundary string ("...00Z" and "...00.5Z" both sort after
// "...00.000000000Z"), which a trimmed "...00Z" boundary would not
// guarantee.
const sqlDayBoundary = "2006-01-02T15:04:05.000000000Z07:00"

// TodaySpend aggregates provider's captured cost actuals for the current
// local day. Two cost shapes exist in the events table, distinguished by
// the event's source column, and both are pure aggregation of actuals —
// no estimate, no price table:
//
//   - status_line usage events carry a session-CUMULATIVE total_cost_usd
//     (Claude's statusline cost object). Today's contribution per session
//     is the delta: the day's latest cumulative value minus the session's
//     last cumulative value observed BEFORE today (0 when the session
//     first appeared today). Negative deltas clamp to 0 (a cumulative
//     counter cannot honestly decrease; a decrease is a provider-side
//     reset/correction, not negative spend).
//   - every other source (managed runs' provider_event usage) carries a
//     TURN-EXACT total_cost_usd; today's samples sum directly.
//
// ok=false when no cost-bearing usage event was observed today — the
// caller omits the segment (unknown is not zero, never "$0.00" from no
// data). Codex sessions carry token counts but no cost field, so a
// codex-only day is honestly ok=false rather than a fabricated dollar
// figure derived from a price table.
func (s *Store) TodaySpend(ctx context.Context, provider string) (TodaySpend, bool) {
	if s == nil || s.DB == nil || provider == "" {
		return TodaySpend{}, false
	}
	now := s.now()
	dayStart, _ := DayBounds(now)
	boundary := dayStart.UTC().Format(sqlDayBoundary)

	rows, err := s.DB.Conn().QueryContext(ctx, `
		SELECT session_id, source, occurred_at, payload_json FROM events
		WHERE provider = ? AND event_type = ? AND occurred_at >= ?
		ORDER BY occurred_at ASC, rowid ASC`,
		provider, string(v1.EventProviderUsageObserved), boundary,
	)
	if err != nil {
		return TodaySpend{}, false
	}
	defer func() { _ = rows.Close() }()

	type sessionSpend struct {
		lastCumulative *float64 // newest status_line total_cost_usd today
		turnSum        float64  // sum of turn-exact samples today
	}
	bySession := make(map[string]*sessionSpend)
	out := TodaySpend{Provider: provider, Day: dayStart.Format("2006-01-02")}
	for rows.Next() {
		var sessionID, source, occurredAt, payloadJSON string
		if err := rows.Scan(&sessionID, &source, &occurredAt, &payloadJSON); err != nil {
			return TodaySpend{}, false
		}
		var payload map[string]any
		if json.Unmarshal([]byte(payloadJSON), &payload) != nil {
			continue // undecodable payloads are skipped, not fatal.
		}
		cost, okCost := payload["total_cost_usd"].(float64)
		if !okCost {
			continue // a usage event without a cost measures no spend.
		}
		observedAt, perr := time.Parse(time.RFC3339Nano, occurredAt)
		if perr != nil {
			continue
		}
		ss := bySession[sessionID]
		if ss == nil {
			ss = &sessionSpend{}
			bySession[sessionID] = ss
		}
		if source == string(domain.SourceStatusLine) {
			c := cost
			ss.lastCumulative = &c // rows arrive oldest-first; last write wins.
		} else {
			ss.turnSum += cost
		}
		if out.FirstObservedAt.IsZero() || observedAt.Before(out.FirstObservedAt) {
			out.FirstObservedAt = observedAt
		}
		if observedAt.After(out.LastObservedAt) {
			out.LastObservedAt = observedAt
		}
	}
	if rows.Err() != nil || len(bySession) == 0 {
		return TodaySpend{}, false
	}

	for sessionID, ss := range bySession {
		spend := ss.turnSum
		if ss.lastCumulative != nil {
			delta := *ss.lastCumulative - s.cumulativeBaseline(ctx, sessionID, provider, boundary)
			if delta > 0 {
				spend += delta
			}
		}
		out.SpendUSD += spend
		out.Sessions++
	}
	return out, true
}

// cumulativeBaseline returns the session's newest status_line
// total_cost_usd observed strictly before today, or 0 when none exists
// (a session that first appeared today starts its delta from zero — the
// provider's cumulative counter itself started near zero at session
// start). Fail-open: any query trouble degrades to 0, which at worst
// over-counts one session's day-spanning delta rather than erroring the
// status bar.
func (s *Store) cumulativeBaseline(ctx context.Context, sessionID, provider, boundary string) float64 {
	rows, err := s.DB.Conn().QueryContext(ctx, `
		SELECT payload_json FROM events
		WHERE session_id = ? AND provider = ? AND event_type = ? AND source = ? AND occurred_at < ?
		ORDER BY occurred_at DESC, rowid DESC LIMIT 8`,
		sessionID, provider, string(v1.EventProviderUsageObserved), string(domain.SourceStatusLine), boundary,
	)
	if err != nil {
		return 0
	}
	defer func() { _ = rows.Close() }()
	for rows.Next() {
		var payloadJSON string
		if err := rows.Scan(&payloadJSON); err != nil {
			return 0
		}
		var payload map[string]any
		if json.Unmarshal([]byte(payloadJSON), &payload) != nil {
			continue
		}
		if cost, ok := payload["total_cost_usd"].(float64); ok {
			return cost // newest-first: the first cost-bearing row is the baseline.
		}
	}
	return 0
}

func (s *Store) now() time.Time {
	if s.Clock != nil {
		return s.Clock.Now()
	}
	return time.Now()
}
