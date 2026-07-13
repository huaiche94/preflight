// sessionbootstrap_test.go: issue #17's lazy in-hook session bootstrap —
// a real (temp-file, migrated) SQLite database plus a fake gitx resolver
// prove the upsert chain creates exactly the rows
// evaluation.SQLDataSource.Resolve needs, stays idempotent under repeated
// and concurrent invocations, populates provider_sessions.model only when
// a model was actually observed (unknown is not zero), and no-ops —
// never errors — on every fail-open path (non-git dir, git failure,
// missing cwd, nil receiver/deps). Hook-level tests then prove each of
// the four handlers actually invokes the bootstrap with its payload's own
// directory, before persist/evaluate.
package orchestrator_test

import (
	"context"
	"path/filepath"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/huaiche94/auspex/internal/domain"
	"github.com/huaiche94/auspex/internal/evaluation"
	"github.com/huaiche94/auspex/internal/gitx"
	"github.com/huaiche94/auspex/internal/orchestrator"
	"github.com/huaiche94/auspex/internal/storage/sqlite"
)

// --- harness ------------------------------------------------------------

func openBootstrapDB(t *testing.T) *sqlite.DB {
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

// bootstrapIDs is a concurrency-safe sequential IDGenerator (the
// concurrent test below hits it from several goroutines; hooks_test.go's
// sequentialHookIDs is deliberately not reused for that reason). prefix
// keeps IDs distinct when one test builds several bootstrappers against
// the same DB — production idgen.New() IDs are globally unique, so a
// collision between two generators would be a harness artifact, not a
// scenario worth exercising.
type bootstrapIDs struct {
	mu     sync.Mutex
	prefix string
	n      int
}

func (g *bootstrapIDs) NewID() string {
	g.mu.Lock()
	defer g.mu.Unlock()
	g.n++
	return "boot-id-" + g.prefix + itoaBootstrap(g.n)
}

func itoaBootstrap(n int) string {
	if n == 0 {
		return "0"
	}
	var digits []byte
	for n > 0 {
		digits = append([]byte{byte('0' + n%10)}, digits...)
		n /= 10
	}
	return string(digits)
}

// fakeRepoResolver is the orchestrator.RepoResolver test double:
// configurable RepoInfo/error, with concurrency-safe call recording so
// hook-level tests can assert which directory a handler passed in.
type fakeRepoResolver struct {
	mu    sync.Mutex
	info  gitx.RepoInfo
	err   error
	calls []string
}

func (f *fakeRepoResolver) ResolveRepo(_ context.Context, path string) (gitx.RepoInfo, error) {
	f.mu.Lock()
	f.calls = append(f.calls, path)
	f.mu.Unlock()
	if f.err != nil {
		return gitx.RepoInfo{}, f.err
	}
	return f.info, nil
}

func (f *fakeRepoResolver) callCount() int {
	f.mu.Lock()
	defer f.mu.Unlock()
	return len(f.calls)
}

func mainWorktreeInfo(root string) gitx.RepoInfo {
	return gitx.RepoInfo{
		WorktreeRoot: root,
		GitDir:       filepath.Join(root, ".git"),
		CommonDir:    filepath.Join(root, ".git"),
	}
}

// bootstrapperSeq feeds each newBootstrapper call a distinct ID prefix
// (see bootstrapIDs.prefix).
var bootstrapperSeq atomic.Int64

func newBootstrapper(db *sqlite.DB, git orchestrator.RepoResolver, at time.Time) *orchestrator.SessionBootstrapper {
	return &orchestrator.SessionBootstrapper{
		DB:    db,
		Git:   git,
		Clock: fixedClock{t: at},
		IDs:   &bootstrapIDs{prefix: "g" + itoaBootstrap(int(bootstrapperSeq.Add(1))) + "-"},
	}
}

func countRows(t *testing.T, db *sqlite.DB, query string, args ...any) int {
	t.Helper()
	var n int
	if err := db.Conn().QueryRowContext(context.Background(), query, args...).Scan(&n); err != nil {
		t.Fatalf("count %q: %v", query, err)
	}
	return n
}

func chainRowCounts(t *testing.T, db *sqlite.DB) (repos, worktrees, sessions int) {
	t.Helper()
	return countRows(t, db, `SELECT COUNT(*) FROM repositories`),
		countRows(t, db, `SELECT COUNT(*) FROM worktrees`),
		countRows(t, db, `SELECT COUNT(*) FROM provider_sessions`)
}

var bootstrapTestTime = time.Date(2026, 7, 13, 10, 0, 0, 0, time.UTC)

func claudeBootstrap(sessionID domain.SessionID, dir string, model *string) orchestrator.SessionBootstrap {
	return orchestrator.SessionBootstrap{
		SessionID: sessionID,
		Dir:       dir,
		Provider:  "claude",
		Model:     model,
	}
}

// --- the acceptance-shaped core: rows exist, Resolve succeeds ------------

func TestSessionBootstrapper_CreatesChainResolveFinds(t *testing.T) {
	db := openBootstrapDB(t)
	resolver := &fakeRepoResolver{info: mainWorktreeInfo("/work/repo")}
	b := newBootstrapper(db, resolver, bootstrapTestTime)

	if !b.Bootstrap(context.Background(), claudeBootstrap("sess-boot-1", "/work/repo/sub", nil)) {
		t.Fatal("Bootstrap = false, want true for a resolvable git directory")
	}

	repos, worktrees, sessions := chainRowCounts(t, db)
	if repos != 1 || worktrees != 1 || sessions != 1 {
		t.Fatalf("row counts = %d/%d/%d, want 1/1/1", repos, worktrees, sessions)
	}

	// The whole point of issue #17: the SAME production query the
	// evaluation pipeline runs first (SQLDataSource.Resolve) now finds the
	// session — no test-glue seeding involved.
	resolved, err := evaluation.NewSQLDataSource(db).Resolve(context.Background(), "sess-boot-1")
	if err != nil {
		t.Fatalf("SQLDataSource.Resolve after Bootstrap: %v", err)
	}
	if resolved.RepositoryID == "" {
		t.Error("Resolve returned empty RepositoryID")
	}
	if resolved.TaskID != nil {
		t.Errorf("Resolve TaskID = %v, want nil (no task yet: cold-start, not fabricated)", *resolved.TaskID)
	}

	// Column-level contract: id IS the provider session UUID (what Resolve
	// joins on), provider_session_id mirrors it, invocation_mode is
	// native-hook, and model is NULL because none was observed.
	var id, providerSessionID, invocationMode string
	var model *string
	err = db.Conn().QueryRowContext(context.Background(),
		`SELECT id, provider_session_id, invocation_mode, model FROM provider_sessions WHERE id = ?`,
		"sess-boot-1",
	).Scan(&id, &providerSessionID, &invocationMode, &model)
	if err != nil {
		t.Fatalf("read provider_sessions row: %v", err)
	}
	if providerSessionID != "sess-boot-1" {
		t.Errorf("provider_session_id = %q, want sess-boot-1", providerSessionID)
	}
	if invocationMode != "native-hook" {
		t.Errorf("invocation_mode = %q, want native-hook (issue #17 design)", invocationMode)
	}
	if model != nil {
		t.Errorf("model = %q, want NULL (no model observed; unknown is not zero)", *model)
	}

	// The resolver was handed the directory the caller reported, verbatim.
	if len(resolver.calls) != 1 || resolver.calls[0] != "/work/repo/sub" {
		t.Errorf("resolver calls = %v, want exactly [/work/repo/sub]", resolver.calls)
	}
}

// --- fail-open no-op paths ------------------------------------------------

func TestSessionBootstrapper_FailOpenPaths_WriteNothing(t *testing.T) {
	notGit := &domain.Error{Code: domain.ErrCodeNotFound, Message: "gitx: path is not inside a git repository"}
	gitDown := &domain.Error{Code: domain.ErrCodeUnavailable, Message: "gitx: failed to execute git", Retryable: true}

	cases := []struct {
		name string
		make func(db *sqlite.DB) *orchestrator.SessionBootstrapper
		req  orchestrator.SessionBootstrap
	}{
		{"non-git directory", func(db *sqlite.DB) *orchestrator.SessionBootstrapper {
			return newBootstrapper(db, &fakeRepoResolver{err: notGit}, bootstrapTestTime)
		}, claudeBootstrap("sess-x", "/tmp/plain", nil)},
		{"git unavailable", func(db *sqlite.DB) *orchestrator.SessionBootstrapper {
			return newBootstrapper(db, &fakeRepoResolver{err: gitDown}, bootstrapTestTime)
		}, claudeBootstrap("sess-x", "/work/repo", nil)},
		{"empty dir", func(db *sqlite.DB) *orchestrator.SessionBootstrapper {
			return newBootstrapper(db, &fakeRepoResolver{info: mainWorktreeInfo("/work/repo")}, bootstrapTestTime)
		}, claudeBootstrap("sess-x", "", nil)},
		{"empty session id", func(db *sqlite.DB) *orchestrator.SessionBootstrapper {
			return newBootstrapper(db, &fakeRepoResolver{info: mainWorktreeInfo("/work/repo")}, bootstrapTestTime)
		}, claudeBootstrap("", "/work/repo", nil)},
		{"empty provider", func(db *sqlite.DB) *orchestrator.SessionBootstrapper {
			return newBootstrapper(db, &fakeRepoResolver{info: mainWorktreeInfo("/work/repo")}, bootstrapTestTime)
		}, orchestrator.SessionBootstrap{SessionID: "sess-x", Dir: "/work/repo"}},
		{"nil Git", func(db *sqlite.DB) *orchestrator.SessionBootstrapper {
			return &orchestrator.SessionBootstrapper{DB: db, Clock: fixedClock{t: bootstrapTestTime}, IDs: &bootstrapIDs{}}
		}, claudeBootstrap("sess-x", "/work/repo", nil)},
		{"nil DB", func(_ *sqlite.DB) *orchestrator.SessionBootstrapper {
			return &orchestrator.SessionBootstrapper{Git: &fakeRepoResolver{info: mainWorktreeInfo("/work/repo")}, Clock: fixedClock{t: bootstrapTestTime}, IDs: &bootstrapIDs{}}
		}, claudeBootstrap("sess-x", "/work/repo", nil)},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			db := openBootstrapDB(t)
			b := tc.make(db)
			if b.Bootstrap(context.Background(), tc.req) {
				t.Error("Bootstrap = true, want false (fail-open no-op)")
			}
			repos, worktrees, sessions := chainRowCounts(t, db)
			if repos != 0 || worktrees != 0 || sessions != 0 {
				t.Errorf("row counts = %d/%d/%d, want 0/0/0 (no fabricated rows)", repos, worktrees, sessions)
			}
		})
	}
}

func TestSessionBootstrapper_NilReceiver_IsNoOp(t *testing.T) {
	var b *orchestrator.SessionBootstrapper
	// Must not panic and must report false — the documented nil-receiver
	// contract mirroring EventCorrelator's.
	if b.Bootstrap(context.Background(), claudeBootstrap("sess-x", "/work/repo", nil)) {
		t.Error("nil receiver Bootstrap = true, want false")
	}
}

// --- idempotency / model semantics ----------------------------------------

func TestSessionBootstrapper_RepeatedCalls_Idempotent(t *testing.T) {
	db := openBootstrapDB(t)
	resolver := &fakeRepoResolver{info: mainWorktreeInfo("/work/repo")}

	first := newBootstrapper(db, resolver, bootstrapTestTime)
	if !first.Bootstrap(context.Background(), claudeBootstrap("sess-idem", "/work/repo", nil)) {
		t.Fatal("first Bootstrap = false, want true")
	}
	var repoID string
	if err := db.Conn().QueryRowContext(context.Background(), `SELECT id FROM repositories`).Scan(&repoID); err != nil {
		t.Fatalf("read repository id: %v", err)
	}

	later := bootstrapTestTime.Add(90 * time.Minute)
	second := newBootstrapper(db, resolver, later)
	for i := 0; i < 3; i++ {
		if !second.Bootstrap(context.Background(), claudeBootstrap("sess-idem", "/work/repo", nil)) {
			t.Fatalf("repeat Bootstrap #%d = false, want true", i)
		}
	}

	repos, worktrees, sessions := chainRowCounts(t, db)
	if repos != 1 || worktrees != 1 || sessions != 1 {
		t.Fatalf("row counts after repeats = %d/%d/%d, want 1/1/1", repos, worktrees, sessions)
	}

	// The surviving row keeps its original identity (fresh IDs minted on
	// conflicting calls are discarded) but records the newest sighting.
	var gotID, lastSeen, createdAt string
	if err := db.Conn().QueryRowContext(context.Background(),
		`SELECT id, last_seen_at, created_at FROM repositories`).Scan(&gotID, &lastSeen, &createdAt); err != nil {
		t.Fatalf("re-read repository row: %v", err)
	}
	if gotID != repoID {
		t.Errorf("repository id changed across idempotent upserts: %q -> %q", repoID, gotID)
	}
	if lastSeen != later.UTC().Format(time.RFC3339Nano) {
		t.Errorf("last_seen_at = %q, want refreshed to %q", lastSeen, later.UTC().Format(time.RFC3339Nano))
	}
	if createdAt != bootstrapTestTime.UTC().Format(time.RFC3339Nano) {
		t.Errorf("created_at = %q, want the FIRST observation %q retained", createdAt, bootstrapTestTime.UTC().Format(time.RFC3339Nano))
	}
}

func TestSessionBootstrapper_ModelPopulation_UnknownIsNotZero(t *testing.T) {
	db := openBootstrapDB(t)
	b := newBootstrapper(db, &fakeRepoResolver{info: mainWorktreeInfo("/work/repo")}, bootstrapTestTime)
	ctx := context.Background()

	readModel := func() *string {
		t.Helper()
		var m *string
		if err := db.Conn().QueryRowContext(ctx,
			`SELECT model FROM provider_sessions WHERE id = ?`, "sess-model").Scan(&m); err != nil {
			t.Fatalf("read model: %v", err)
		}
		return m
	}
	opus := "claude-opus-4-1-20250805"
	sonnet := "claude-sonnet-4-5"

	// 1. No model observed (UserPromptSubmit-shaped call): NULL, not "".
	b.Bootstrap(ctx, claudeBootstrap("sess-model", "/work/repo", nil))
	if m := readModel(); m != nil {
		t.Fatalf("model after model-less bootstrap = %q, want NULL", *m)
	}
	// 2. Statusline-shaped call observes a model: populated.
	b.Bootstrap(ctx, claudeBootstrap("sess-model", "/work/repo", &opus))
	if m := readModel(); m == nil || *m != opus {
		t.Fatalf("model after observing %q = %v, want it stored", opus, m)
	}
	// 3. A later model-less call must NOT erase known state (COALESCE):
	// unknown is not zero, and it is not a reset either.
	b.Bootstrap(ctx, claudeBootstrap("sess-model", "/work/repo", nil))
	if m := readModel(); m == nil || *m != opus {
		t.Fatalf("model after later model-less bootstrap = %v, want %q retained", m, opus)
	}
	// 4. A genuinely observed model change is an update, not fabrication.
	b.Bootstrap(ctx, claudeBootstrap("sess-model", "/work/repo", &sonnet))
	if m := readModel(); m == nil || *m != sonnet {
		t.Fatalf("model after observing %q = %v, want it updated", sonnet, m)
	}
}

func TestSessionBootstrapper_LinkedWorktree_SharesRepositoryRow(t *testing.T) {
	db := openBootstrapDB(t)
	commonDir := "/work/repo/.git"
	main := &fakeRepoResolver{info: gitx.RepoInfo{
		WorktreeRoot: "/work/repo", GitDir: commonDir, CommonDir: commonDir,
	}}
	linked := &fakeRepoResolver{info: gitx.RepoInfo{
		WorktreeRoot:     "/work/repo-feature",
		GitDir:           filepath.Join(commonDir, "worktrees", "feature"),
		CommonDir:        commonDir,
		IsLinkedWorktree: true,
	}}
	ctx := context.Background()

	if !newBootstrapper(db, main, bootstrapTestTime).Bootstrap(ctx, claudeBootstrap("sess-main", "/work/repo", nil)) {
		t.Fatal("main-worktree Bootstrap = false, want true")
	}
	if !newBootstrapper(db, linked, bootstrapTestTime).Bootstrap(ctx, claudeBootstrap("sess-linked", "/work/repo-feature", nil)) {
		t.Fatal("linked-worktree Bootstrap = false, want true")
	}

	repos, worktrees, sessions := chainRowCounts(t, db)
	if repos != 1 || worktrees != 2 || sessions != 2 {
		t.Fatalf("row counts = %d/%d/%d, want 1 repository shared by 2 worktrees/2 sessions", repos, worktrees, sessions)
	}
	// Both worktrees FK the one repository, and each session resolves to it.
	ds := evaluation.NewSQLDataSource(db)
	a, err := ds.Resolve(ctx, "sess-main")
	if err != nil {
		t.Fatalf("Resolve(sess-main): %v", err)
	}
	b, err := ds.Resolve(ctx, "sess-linked")
	if err != nil {
		t.Fatalf("Resolve(sess-linked): %v", err)
	}
	if a.RepositoryID != b.RepositoryID {
		t.Errorf("RepositoryID mismatch across worktrees: %q vs %q, want shared (same git_common_dir)", a.RepositoryID, b.RepositoryID)
	}
	// canonical_root for the shared row was derived from the main layout.
	var canonical string
	if err := db.Conn().QueryRowContext(ctx, `SELECT canonical_root FROM repositories`).Scan(&canonical); err != nil {
		t.Fatalf("read canonical_root: %v", err)
	}
	if canonical != "/work/repo" {
		t.Errorf("canonical_root = %q, want /work/repo", canonical)
	}
}

func TestSessionBootstrapper_ConcurrentCalls_SingleChain(t *testing.T) {
	db := openBootstrapDB(t)
	resolver := &fakeRepoResolver{info: mainWorktreeInfo("/work/repo")}
	b := newBootstrapper(db, resolver, bootstrapTestTime)

	const goroutines = 8
	var wg sync.WaitGroup
	results := make([]bool, goroutines)
	for i := 0; i < goroutines; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			results[i] = b.Bootstrap(context.Background(), claudeBootstrap("sess-conc", "/work/repo", nil))
		}(i)
	}
	wg.Wait()

	for i, ok := range results {
		if !ok {
			t.Errorf("goroutine %d: Bootstrap = false, want true (concurrent upserts must converge, not fail)", i)
		}
	}
	repos, worktrees, sessions := chainRowCounts(t, db)
	if repos != 1 || worktrees != 1 || sessions != 1 {
		t.Fatalf("row counts after %d concurrent bootstraps = %d/%d/%d, want 1/1/1", goroutines, repos, worktrees, sessions)
	}
}

// --- hook handlers invoke the bootstrap -----------------------------------

// hookBootstrapDeps returns baseHookDeps plus a real SessionBootstrapper
// over a fake resolver and a real migrated DB — the hook-level harness for
// the four handler tests below.
func hookBootstrapDeps(t *testing.T) (orchestrator.HookDeps, *fakeRepoResolver, *sqlite.DB) {
	t.Helper()
	db := openBootstrapDB(t)
	resolver := &fakeRepoResolver{info: mainWorktreeInfo("/Users/dev/projects/auspex")}
	deps := baseHookDeps()
	deps.Bootstrapper = &orchestrator.SessionBootstrapper{
		DB:    db,
		Git:   resolver,
		Clock: deps.Clock,
		IDs:   &bootstrapIDs{},
	}
	return deps, resolver, db
}

func TestHookHandlers_UserPromptSubmit_BootstrapsSessionFromCWD(t *testing.T) {
	deps, resolver, db := hookBootstrapDeps(t)

	if _, err := orchestrator.HandleUserPromptSubmit(context.Background(), deps, readFixture(t, "userpromptsubmit", "normal.json")); err != nil {
		t.Fatalf("HandleUserPromptSubmit: %v", err)
	}

	if len(resolver.calls) != 1 || resolver.calls[0] != "/Users/dev/projects/auspex" {
		t.Errorf("resolver calls = %v, want the payload's cwd [/Users/dev/projects/auspex]", resolver.calls)
	}
	if n := countRows(t, db, `SELECT COUNT(*) FROM provider_sessions WHERE id = ? AND invocation_mode = 'native-hook'`,
		"sess_01H9X8K7QZ3M4N5P6R7S8T9V0W"); n != 1 {
		t.Errorf("provider_sessions rows for fixture session = %d, want 1", n)
	}
	// UserPromptSubmit carries no model: NULL, never "".
	var model *string
	if err := db.Conn().QueryRowContext(context.Background(),
		`SELECT model FROM provider_sessions WHERE id = ?`, "sess_01H9X8K7QZ3M4N5P6R7S8T9V0W").Scan(&model); err != nil {
		t.Fatalf("read model: %v", err)
	}
	if model != nil {
		t.Errorf("model = %q, want NULL (UserPromptSubmit payload has no model field)", *model)
	}
}

func TestHookHandlers_UserPromptSubmit_MissingCWD_SkipsBootstrap(t *testing.T) {
	deps, resolver, db := hookBootstrapDeps(t)

	// missing_fields.json has no cwd: the handler must not even consult
	// git (nothing to resolve), and the hook result is unchanged.
	result, err := orchestrator.HandleUserPromptSubmit(context.Background(), deps, readFixture(t, "userpromptsubmit", "missing_fields.json"))
	if err != nil {
		t.Fatalf("HandleUserPromptSubmit: %v", err)
	}
	if result.Response.Decision != "allow" {
		t.Errorf("Decision = %q, want allow", result.Response.Decision)
	}
	if resolver.callCount() != 0 {
		t.Errorf("resolver called %d times, want 0 for a payload with no cwd", resolver.callCount())
	}
	if _, _, sessions := chainRowCounts(t, db); sessions != 0 {
		t.Errorf("provider_sessions rows = %d, want 0", sessions)
	}
}

func TestHookHandlers_StatusLine_BootstrapsSessionWithModel(t *testing.T) {
	deps, resolver, db := hookBootstrapDeps(t)

	if _, err := orchestrator.HandleStatusLine(context.Background(), deps, readFixture(t, "statusline", "normal.json")); err != nil {
		t.Fatalf("HandleStatusLine: %v", err)
	}

	if len(resolver.calls) != 1 || resolver.calls[0] != "/Users/dev/projects/auspex" {
		t.Errorf("resolver calls = %v, want the snapshot's workspace.current_dir", resolver.calls)
	}
	// The statusline is the payload that carries a model — issue #17
	// deliverable 3: provider_sessions.model populated when available.
	var model *string
	if err := db.Conn().QueryRowContext(context.Background(),
		`SELECT model FROM provider_sessions WHERE id = ?`, "sess_01H9X8K7QZ3M4N5P6R7S8T9V0W").Scan(&model); err != nil {
		t.Fatalf("read model: %v", err)
	}
	if model == nil || *model != "claude-opus-4-1-20250805" {
		t.Errorf("model = %v, want the fixture's model.id claude-opus-4-1-20250805", model)
	}
}

func TestHookHandlers_StopAndStopFailure_BootstrapFromCWD(t *testing.T) {
	t.Run("stop", func(t *testing.T) {
		deps, _, db := hookBootstrapDeps(t)
		if _, err := orchestrator.HandleStop(context.Background(), deps, readFixture(t, "stop", "normal.json")); err != nil {
			t.Fatalf("HandleStop: %v", err)
		}
		if n := countRows(t, db, `SELECT COUNT(*) FROM provider_sessions WHERE id = ?`, "sess_01H9X8K7QZ3M4N5P6R7S8T9V0W"); n != 1 {
			t.Errorf("provider_sessions rows after Stop = %d, want 1 (Stop payloads carry a cwd)", n)
		}
	})
	t.Run("stop-failure", func(t *testing.T) {
		deps, _, db := hookBootstrapDeps(t)
		if _, err := orchestrator.HandleStopFailure(context.Background(), deps, readFixture(t, "stopfailure", "rate_limit.json")); err != nil {
			t.Fatalf("HandleStopFailure: %v", err)
		}
		if n := countRows(t, db, `SELECT COUNT(*) FROM provider_sessions WHERE id = ?`, "sess_01H9X8K7QZ3M4N5P6R7S8T9V0W"); n != 1 {
			t.Errorf("provider_sessions rows after StopFailure = %d, want 1 (StopFailure payloads carry a cwd)", n)
		}
	})
}

func TestHookHandlers_NilBootstrapper_UnchangedBehavior(t *testing.T) {
	// The pre-issue-#17 composition (no Bootstrapper) must keep working
	// bit-for-bit: hooks succeed, nothing is written.
	deps := baseHookDeps()
	db := openBootstrapDB(t)

	if _, err := orchestrator.HandleUserPromptSubmit(context.Background(), deps, readFixture(t, "userpromptsubmit", "normal.json")); err != nil {
		t.Fatalf("HandleUserPromptSubmit with nil Bootstrapper: %v", err)
	}
	if _, err := orchestrator.HandleStatusLine(context.Background(), deps, readFixture(t, "statusline", "normal.json")); err != nil {
		t.Fatalf("HandleStatusLine with nil Bootstrapper: %v", err)
	}
	if repos, worktrees, sessions := func() (int, int, int) { return chainRowCounts(t, db) }(); repos+worktrees+sessions != 0 {
		t.Errorf("rows written without a Bootstrapper: %d/%d/%d, want none", repos, worktrees, sessions)
	}
}
