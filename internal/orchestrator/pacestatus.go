// pacestatus.go: the hook-side seam for the issue-#90 Phase A "today's
// spend + pace" statusline segment. The aggregation itself lives in
// internal/pace (Store.TodaySpend, pure read over the events table);
// this file adapts it to the two statusline render paths (Claude
// --emit-line, `hook codex status`) with the same nil-is-a-documented-
// degrade convention every other HookDeps seam follows.
package orchestrator

import (
	"context"
	"time"

	"github.com/huaiche94/auspex/internal/evaluation"
	"github.com/huaiche94/auspex/internal/pace"
)

// PaceReader is HookDeps' narrow view of the pace read helper: today's
// aggregated spend for a provider, ok=false when no cost-bearing
// observation exists today (cold start, codex's cost-less telemetry, a
// read error — all indistinguishable by design; the caller omits the
// segment, never prints $0.00 from no data). Implementations must be
// fail-open: an error is an ok=false, never a hook failure. The real
// value is *pace.Store over the same *sqlite.DB every hook writes
// through.
type PaceReader interface {
	TodaySpend(ctx context.Context, provider string) (pace.TodaySpend, bool)
}

// spendPaceStatus resolves the statusline's spend segment input for one
// provider: today's observed spend plus, when an honest rate exists, the
// end-of-day pace extrapolation (pace.ProjectEndOfDay — labeled a pace
// by the renderer, §7). nil when no Pace reader is wired or no cost data
// was observed today — the segment is omitted, exactly like every other
// optional statusline input.
func (d HookDeps) spendPaceStatus(ctx context.Context, provider string) *evaluation.SpendPaceStatus {
	if d.Pace == nil {
		return nil
	}
	spend, ok := d.Pace.TodaySpend(ctx, provider)
	if !ok {
		return nil
	}
	out := &evaluation.SpendPaceStatus{TodayUSD: spend.SpendUSD}
	if projected, okProj := pace.ProjectEndOfDay(spend, d.renderNow()); okProj {
		out.ProjectedEndOfDayUSD = &projected
	}
	return out
}

// renderNow is the statusline render instant: the injected clock when
// wired (every production composition), time.Now otherwise — display
// helpers must not silently freeze to the zero time just because a
// minimal test composition skipped the clock.
func (d HookDeps) renderNow() time.Time {
	if d.Clock != nil {
		return d.Clock.Now()
	}
	return time.Now()
}
