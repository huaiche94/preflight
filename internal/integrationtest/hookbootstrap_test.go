// hookbootstrap_test.go: issue #17's literal acceptance, as an integration
// test — "in a real Claude Code session (no test glue): the
// UserPromptSubmit hook's additionalContext carries the forecast card".
// Before issue #17, no production code path ever wrote repositories/
// worktrees/provider_sessions rows, so every E2E test in this package
// (e.g. e2e_highrisk_test.go's qa02SeedChain, evaluate_privacy_test.go's
// seed block) had to seed them by hand — exactly the qa-02-shaped gap the
// issue names ("E2E 測試靠測試 setup 自建資料列所以沒暴露"). This file is
// the first one that deliberately does NOT seed: a REAL temp git
// repository plus the production stack (real gitx.Client resolving it,
// real SessionBootstrapper upserting, real SQLDataSource resolving, real
// evaluation.Service EvaluateTurn/Decide, real EventStore persisting,
// real forecast-card read-back) drive orchestrator.HandleUserPromptSubmit
// end to end, proving the lazy in-hook bootstrap alone makes the pipeline
// live. The non-git companion test proves the fail-open half: outside any
// repository the hook still answers `{}` and fabricates no rows (ADD
// §17.5; Constitution "unknown is not zero").
package integrationtest

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/huaiche94/auspex/internal/domain"
	"github.com/huaiche94/auspex/internal/evaluation"
	"github.com/huaiche94/auspex/internal/gitx"
	claudehooks "github.com/huaiche94/auspex/internal/hooks/claude"
	"github.com/huaiche94/auspex/internal/orchestrator"
	"github.com/huaiche94/auspex/internal/policy"
	"github.com/huaiche94/auspex/internal/predictor/quota"
	"github.com/huaiche94/auspex/internal/predictor/risk"
	"github.com/huaiche94/auspex/internal/predictor/scope"
	"github.com/huaiche94/auspex/internal/predictor/token"
	"github.com/huaiche94/auspex/internal/storage/sqlite"
	claudetelemetry "github.com/huaiche94/auspex/internal/telemetry/claude"
)

// buildIssue17Deps composes the hook handlers' production dependency set
// exactly the way cmd/auspex/wire.go does — real SQLDataSource-backed
// evaluation.Service (mirroring evaluate_privacy_test.go's
// buildEvaluateCLIRoot, including its tokenSourceBridge), real EventStore,
// real EventCorrelator over the same data source, real forecast-card
// source, and (the piece under test) a real SessionBootstrapper over a
// real gitx.Client. No foundation rows are seeded anywhere in this file.
func buildIssue17Deps(t *testing.T) (orchestrator.HookDeps, *evaluation.SQLDataSource, *sqlite.DB) {
	t.Helper()
	db := qa02OpenDB(t)

	clock := fixedClock{t: time.Date(2026, 7, 13, 15, 0, 0, 0, time.UTC)}
	ids := &seqIDs{}
	dataSource := evaluation.NewSQLDataSource(db)
	svc := evaluation.New(
		db,
		dataSource,
		scope.NewRuleScopeEstimator(dataSource),
		token.NewRuleTokenForecaster(tokenSourceBridge{source: dataSource}),
		quota.NewRuleQuotaForecaster(),
		risk.NewRuleRiskCombiner(),
		policy.NewDecider(),
		clock,
		ids,
	)

	deps := orchestrator.HookDeps{
		Clock:      clock,
		IDs:        ids,
		Persister:  claudetelemetry.NewEventStore(db),
		TxRunner:   db,
		Evaluation: svc,
		Correlator: &orchestrator.EventCorrelator{Sessions: dataSource},
		Forecast:   svc,
		Bootstrapper: &orchestrator.SessionBootstrapper{
			DB:    db,
			Git:   gitx.NewClient(gitx.ExecRunner{}),
			Clock: clock,
			IDs:   ids,
		},
	}
	return deps, dataSource, db
}

func issue17PromptPayload(sessionID, cwd string) []byte {
	return []byte(fmt.Sprintf(
		`{"session_id":%q,"cwd":%q,"hook_event_name":"UserPromptSubmit","prompt":"Wire the lazy session bootstrap so the forecast card renders in real sessions."}`,
		sessionID, cwd))
}

func issue17ChainCounts(t *testing.T, db *sqlite.DB, sessionID string) (repos, worktrees, sessions int) {
	t.Helper()
	ctx := context.Background()
	scan := func(q string, args ...any) int {
		var n int
		if err := db.Conn().QueryRowContext(ctx, q, args...).Scan(&n); err != nil {
			t.Fatalf("count %q: %v", q, err)
		}
		return n
	}
	return scan(`SELECT COUNT(*) FROM repositories`),
		scan(`SELECT COUNT(*) FROM worktrees`),
		scan(`SELECT COUNT(*) FROM provider_sessions WHERE id = ?`, sessionID)
}

// TestIssue17_UserPromptSubmit_RealGitRepo_ForecastCardRenders is the
// acceptance path: a real git repository, no seeded rows, one hook
// invocation — afterward the chain exists, Resolve succeeds, the
// evaluation ran, and the hook response's additionalContext carries the
// issue-#14 forecast card.
func TestIssue17_UserPromptSubmit_RealGitRepo_ForecastCardRenders(t *testing.T) {
	repo := newQA02Repo(t) // skips if git is unavailable
	deps, dataSource, db := buildIssue17Deps(t)
	ctx := context.Background()
	const sessionID = "sess-issue17-e2e"

	result, err := orchestrator.HandleUserPromptSubmit(ctx, deps, issue17PromptPayload(sessionID, repo.dir))
	if err != nil {
		t.Fatalf("HandleUserPromptSubmit: %v", err)
	}

	// 1. The bootstrap wrote the full chain — the rows only tests used to
	// seed now exist because production code created them.
	repos, worktrees, sessions := issue17ChainCounts(t, db, sessionID)
	if repos != 1 || worktrees != 1 || sessions != 1 {
		t.Fatalf("row counts = %d/%d/%d, want 1/1/1 from the in-hook bootstrap alone", repos, worktrees, sessions)
	}
	var invocationMode, rootPath string
	if err := db.Conn().QueryRowContext(ctx, `
		SELECT ps.invocation_mode, w.root_path
		FROM provider_sessions ps JOIN worktrees w ON w.id = ps.worktree_id
		WHERE ps.id = ?`, sessionID,
	).Scan(&invocationMode, &rootPath); err != nil {
		t.Fatalf("read bootstrapped chain: %v", err)
	}
	if invocationMode != "native-hook" {
		t.Errorf("invocation_mode = %q, want native-hook", invocationMode)
	}
	if rootPath != repo.dir {
		t.Errorf("worktrees.root_path = %q, want the real repo root %q", rootPath, repo.dir)
	}

	// 2. SQLDataSource.Resolve — the exact query issue #17 reported as
	// permanently not_found in production — now succeeds.
	resolved, err := dataSource.Resolve(ctx, sessionID)
	if err != nil {
		t.Fatalf("SQLDataSource.Resolve after the hook: %v", err)
	}
	if resolved.RepositoryID == "" {
		t.Error("Resolve returned an empty RepositoryID")
	}

	// 3. The evaluation genuinely ran (no fallback-to-plain-allow) and
	// persisted its prediction row.
	if !result.Evaluated {
		t.Fatal("result.Evaluated = false — EvaluateTurn still failing means the bootstrap did not take effect")
	}
	if !result.Persisted {
		t.Error("result.Persisted = false, want the turn.started event persisted")
	}
	var predictions int
	if err := db.Conn().QueryRowContext(ctx, `SELECT COUNT(*) FROM predictions`).Scan(&predictions); err != nil {
		t.Fatalf("count predictions: %v", err)
	}
	if predictions == 0 {
		t.Error("predictions table is empty, want EvaluateTurn's persisted row")
	}

	// 4. Issue #17's literal acceptance: the forecast card (issue #14) is
	// present in the hook response's additionalContext.
	ac := result.Response.AdditionalContext
	if ac == "" {
		t.Fatal("AdditionalContext is empty — issue #17's acceptance is precisely that this card now renders in a real session")
	}
	for _, want := range []string{"Auspex forecast", "uncalibrated estimate", "policy:"} {
		if !strings.Contains(ac, want) {
			t.Errorf("AdditionalContext missing %q:\n%s", want, ac)
		}
	}
	body, err := claudehooks.EncodeUserPromptSubmitResponse(result.Response)
	if err != nil {
		t.Fatalf("EncodeUserPromptSubmitResponse: %v", err)
	}
	if !strings.Contains(string(body), `"additionalContext"`) {
		t.Errorf("wire response missing additionalContext: %s", body)
	}
}

// TestIssue17_StatusLine_BootstrapsAndPopulatesModel covers the second
// producer: the statusline payload is the one that carries a model
// identity, so after one statusline hook the session row exists WITH
// provider_sessions.model populated (issue #17 deliverable 3: "unknown is
// not zero" — the UserPromptSubmit path above proved model stays NULL).
func TestIssue17_StatusLine_BootstrapsAndPopulatesModel(t *testing.T) {
	repo := newQA02Repo(t)
	deps, dataSource, db := buildIssue17Deps(t)
	ctx := context.Background()
	const sessionID = "sess-issue17-statusline"

	payload := []byte(fmt.Sprintf(
		`{"session_id":%q,"model":{"id":"claude-opus-4-1-20250805","display_name":"Opus 4.1"},"workspace":{"current_dir":%q,"project_dir":%q}}`,
		sessionID, repo.dir, repo.dir))

	if _, line, err := orchestrator.HandleStatusLineEmitLine(ctx, deps, payload); err != nil {
		t.Fatalf("HandleStatusLineEmitLine: %v", err)
	} else if line == "" || !strings.Contains(line, "Opus 4.1") {
		t.Errorf("emit-line = %q, want the model segment rendered", line)
	}

	if _, err := dataSource.Resolve(ctx, sessionID); err != nil {
		t.Fatalf("Resolve after statusline hook: %v", err)
	}
	var model *string
	if err := db.Conn().QueryRowContext(ctx,
		`SELECT model FROM provider_sessions WHERE id = ?`, sessionID).Scan(&model); err != nil {
		t.Fatalf("read model: %v", err)
	}
	if model == nil || *model != "claude-opus-4-1-20250805" {
		t.Errorf("provider_sessions.model = %v, want the snapshot's model.id", model)
	}
}

// TestIssue17_UserPromptSubmit_NonGitDir_FailsOpenNoRows proves the
// fail-open half: a session running outside any git repository gets no
// fabricated rows, no evaluation, and the hook still answers the plain
// `{}` allow — bit-for-bit the pre-issue-#17 degrade (ADD §17.5).
func TestIssue17_UserPromptSubmit_NonGitDir_FailsOpenNoRows(t *testing.T) {
	deps, dataSource, db := buildIssue17Deps(t)
	ctx := context.Background()
	const sessionID = "sess-issue17-nongit"
	plainDir := t.TempDir() // no git repository anywhere above a TempDir

	result, err := orchestrator.HandleUserPromptSubmit(ctx, deps, issue17PromptPayload(sessionID, plainDir))
	if err != nil {
		t.Fatalf("HandleUserPromptSubmit in a non-git dir: %v", err)
	}
	if result.Evaluated {
		t.Error("Evaluated = true, want false (Resolve must still honestly fail not_found)")
	}
	if result.Response.Decision != claudehooks.HookDecisionAllow || result.Response.AdditionalContext != "" {
		t.Errorf("Response = %+v, want the plain allow with no card", result.Response)
	}
	body, err := claudehooks.EncodeUserPromptSubmitResponse(result.Response)
	if err != nil {
		t.Fatalf("EncodeUserPromptSubmitResponse: %v", err)
	}
	if string(body) != "{}" {
		t.Errorf("wire response = %s, want {} (the hook contract's plain allow)", body)
	}

	repos, worktrees, sessions := issue17ChainCounts(t, db, sessionID)
	if repos != 0 || worktrees != 0 || sessions != 0 {
		t.Errorf("row counts = %d/%d/%d, want 0/0/0 — a non-git session must not fabricate a chain", repos, worktrees, sessions)
	}
	var derr *domain.Error
	if _, err := dataSource.Resolve(ctx, sessionID); !errors.As(err, &derr) || derr.Code != domain.ErrCodeNotFound {
		t.Errorf("Resolve = %v, want the honest ErrCodeNotFound", err)
	}
}
