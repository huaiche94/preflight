// pacestatus_test.go: the #90 spend/pace segment on both statusline
// surfaces — the provider each path aggregates for, the pace projection
// under the injected fixed clock, and the omit-on-no-data honesty rule.
package orchestrator_test

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/huaiche94/auspex/internal/orchestrator"
	"github.com/huaiche94/auspex/internal/pace"
	v1 "github.com/huaiche94/auspex/pkg/protocol/v1"
)

type fakePaceReader struct {
	spend       pace.TodaySpend
	ok          bool
	gotProvider string
}

func (f *fakePaceReader) TodaySpend(_ context.Context, provider string) (pace.TodaySpend, bool) {
	f.gotProvider = provider
	return f.spend, f.ok
}

// TestHookHandlers_StatusLineEmitLine_SpendPaceSegment: the Claude
// emit-line aggregates for provider "claude" and renders today's spend
// with the end-of-day pace under the fake clock (2026-01-01T00:00Z →
// 24h remain; $1.40 over the 2h observed window → $0.70/h → ~$18.20 by
// 24:00), labeled as a pace.
func TestHookHandlers_StatusLineEmitLine_SpendPaceSegment(t *testing.T) {
	deps := baseHookDeps()
	reader := &fakePaceReader{
		spend: pace.TodaySpend{
			Provider:        "claude",
			SpendUSD:        1.40,
			Sessions:        1,
			FirstObservedAt: time.Date(2025, 12, 31, 22, 0, 0, 0, time.UTC),
		},
		ok: true,
	}
	deps.Pace = reader

	_, line, err := orchestrator.HandleStatusLineEmitLine(context.Background(), deps, readFixture(t, "statusline", "normal.json"))
	if err != nil {
		t.Fatalf("HandleStatusLineEmitLine: %v", err)
	}
	if reader.gotProvider != "claude" {
		t.Errorf("TodaySpend provider = %q, want claude", reader.gotProvider)
	}
	if want := statusBrand + " Opus 4.1" + statusSep + statusQuota5h + statusSep +
		"today $1.40 · pace → ~$18.20 by 24:00" + statusSep + statusCtx219; line != want {
		t.Errorf("line = %q, want %q", line, want)
	}
}

// TestHookHandlers_StatusLineEmitLine_NoSpendDataOmitsSegment: ok=false
// (no cost observation today) omits the segment — never "$0.00" from no
// data.
func TestHookHandlers_StatusLineEmitLine_NoSpendDataOmitsSegment(t *testing.T) {
	deps := baseHookDeps()
	deps.Pace = &fakePaceReader{ok: false}

	_, line, err := orchestrator.HandleStatusLineEmitLine(context.Background(), deps, readFixture(t, "statusline", "normal.json"))
	if err != nil {
		t.Fatalf("HandleStatusLineEmitLine: %v", err)
	}
	if strings.Contains(line, "today") || strings.Contains(line, "$") {
		t.Errorf("line = %q, must omit the spend segment when no cost data exists today", line)
	}
}

// TestHandleCodexStatus_SpendPaceSegment: the codex status path
// aggregates for provider "codex" through the same seam.
func TestHandleCodexStatus_SpendPaceSegment(t *testing.T) {
	db := openCodexStatusDB(t)
	seedCodexSession(t, db, "codex-pace", "/repo/pace", "gpt-5.2-codex", "2026-07-16T08:00:00Z")
	persistCodexObservation(t, db, codexObservation("ev-pace-q", "codex-pace", v1.EventProviderQuotaObserved,
		time.Date(2026, 7, 16, 9, 0, 0, 0, time.UTC), map[string]any{"limit_id": "secondary", "used_percent": 49.2}))

	deps := baseHookDeps()
	deps.CodexStatus = &orchestrator.CodexStatusStore{DB: db}
	reader := &fakePaceReader{
		spend: pace.TodaySpend{
			Provider:        "codex",
			SpendUSD:        0.42,
			Sessions:        1,
			FirstObservedAt: time.Date(2025, 12, 31, 23, 0, 0, 0, time.UTC),
		},
		ok: true,
	}
	deps.Pace = reader

	line, err := orchestrator.HandleCodexStatus(context.Background(), deps, "/repo/pace")
	if err != nil {
		t.Fatalf("HandleCodexStatus: %v", err)
	}
	if reader.gotProvider != "codex" {
		t.Errorf("TodaySpend provider = %q, want codex", reader.gotProvider)
	}
	if !strings.Contains(line, "◷ weekly ~49%") {
		t.Errorf("line = %q, want the worst quota window", line)
	}
	// $0.42 over 1h observed → $0.42/h; 24h remain at the fake clock's
	// midnight → ~$10.50 by 24:00.
	if !strings.Contains(line, "today $0.42 · pace → ~$10.50 by 24:00") {
		t.Errorf("line = %q, want the codex spend/pace segment", line)
	}
}
