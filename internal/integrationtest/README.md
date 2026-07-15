# internal/integrationtest/ — cross-role integration and end-to-end tests

> 🌐 English | [繁體中文](README.zh-TW.md)

Test-only package (no non-test Go files, no doc.go): integration-scope
proofs that separately built and separately unit-tested components
compose correctly end to end. The suites deliberately drive REAL
production code — real on-disk SQLite files (never `:memory:`), real
migrations, real scratch git repositories, real subprocesses — rather
than package-internal fakes. This is one of qa's exclusive paths
(agents/qa.md); package-level unit tests live with each package.

Current suites:

- e2e_highrisk_test.go — the vertical-slice demo (qa-02): one high-risk
  turn driven end to end, from status-line ingestion through evaluation
  block, checkpoint, one-time allow, stop, and pause/wake recovery.
- restart_sameDB_test.go — restart against the same SQLite file across
  multiple roles' storage layers (qa-03).
- duplicate_outoforder_test.go — idempotent event persistence composed
  with duplicate/out-of-order progress handling (qa-04).
- leakage_scanner_test.go — raw-prompt/secret scan over on-disk DB bytes
  (including WAL) and checkpoint artifacts (qa-05).
- malicious_fixture_test.go — path-traversal/symlink/malicious-fixture
  attacks through the frozen checkpoint contract (qa-06).
- scheduler_doubleworker_test.go — scheduler double-worker/lease race at
  integration scope (qa-07).
- hookbootstrap_test.go — issue #17 acceptance: the hook path bootstraps
  repository/worktree/session rows with no test seeding.
- evaluate_privacy_test.go — issue #14: `auspex evaluate` never persists
  raw prompt text (canary scan with a hash-presence negative control).
- forecast_prompt_conditioned_test.go — issue #42: the token forecast
  responds to prompt content.
- managedrun_test.go — issue #8: `auspex run` managed one-shot against a
  compiled fake-provider subprocess; the one suite here that also uses
  the doubles in [../testutil/fakes/](../testutil/fakes/).
