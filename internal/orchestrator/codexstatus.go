// codexstatus.go: the SQL-backed CodexStatusReader (issue #9 Phase 1b).
// Lives in this package for OpenTurnStore's reason (openturn.go): it is
// hook-path infrastructure with behavior worth testing against a real
// migrated DB in-package, not a pure DTO translation.
//
// Why a DB read-back exists at all: Codex has no statusline hook, so there
// is no stdin-fed render moment the way Claude's `hook claude statusline
// --emit-line` has — the only way a periodic caller (tmux status-right)
// can show a Codex session's quota/context state is to read what the
// SessionStart/UserPromptSubmit/Stop hooks already persisted. The queries
// below touch only rows this branch's own codex pipeline writes
// (provider_sessions via the issue-#17 bootstrap, events via the codex
// normalizer) — numbers and ids, no content columns exist to leak.
package orchestrator

import (
	"context"
	"encoding/json"
	"sort"
	"time"

	"github.com/huaiche94/auspex/internal/domain"
	"github.com/huaiche94/auspex/internal/evaluation"
	"github.com/huaiche94/auspex/internal/storage/sqlite"
	codextelemetry "github.com/huaiche94/auspex/internal/telemetry/codex"
	v1 "github.com/huaiche94/auspex/pkg/protocol/v1"
)

// CodexStatusStore resolves the latest Codex session for a directory from
// provider_sessions (joined through worktrees so a cwd anywhere inside the
// registered worktree matches), then the session's most recent
// context/quota observations from the events table.
type CodexStatusStore struct {
	DB *sqlite.DB
}

var _ CodexStatusReader = (*CodexStatusStore)(nil)

// LatestCodexStatus implements CodexStatusReader. Fail-open by that
// contract: a nil receiver/DB, query error, or no matching session is
// ok=false — never an error, never a fabricated measurement. cwd == ""
// matches the latest Codex session anywhere (a caller that genuinely has
// no directory still gets the freshest line available).
func (s *CodexStatusStore) LatestCodexStatus(ctx context.Context, cwd string) (CodexStatusSnapshot, bool) {
	if s == nil || s.DB == nil {
		return CodexStatusSnapshot{}, false
	}

	query := `
		SELECT ps.id, COALESCE(ps.model, '')
		FROM provider_sessions ps
		JOIN worktrees w ON w.id = ps.worktree_id
		WHERE ps.provider = ?`
	args := []any{codextelemetry.Provider}
	if cwd != "" {
		// The worktree root itself, or any directory beneath it. root_path
		// is an absolute path the issue-#17 bootstrap wrote; the || '/%'
		// suffix keeps /a/b from matching /a/bc.
		query += ` AND (w.root_path = ? OR ? LIKE w.root_path || '/%')`
		args = append(args, cwd, cwd)
	}
	query += ` ORDER BY ps.started_at DESC, ps.id DESC LIMIT 1`

	var sessionID, model string
	if err := s.DB.Conn().QueryRowContext(ctx, query, args...).Scan(&sessionID, &model); err != nil {
		return CodexStatusSnapshot{}, false
	}

	snap := CodexStatusSnapshot{
		SessionID: domain.SessionID(sessionID),
		Model:     model,
	}
	snap.ContextUsedPercent = s.latestContextPercent(ctx, sessionID)
	snap.QuotaWindows = s.latestQuotaWindows(ctx, sessionID)
	return snap, true
}

// latestContextPercent derives the context fill percentage from the
// session's newest provider.context.observed event. nil whenever either
// measurement is missing — a percentage synthesized from one known and one
// unknown half would be a fabrication (unknown is not zero).
func (s *CodexStatusStore) latestContextPercent(ctx context.Context, sessionID string) *float64 {
	payload, ok := s.latestPayload(ctx, sessionID, v1.EventProviderContextObserved, 1)
	if !ok || len(payload) == 0 {
		return nil
	}
	used, okUsed := payloadNumber(payload[0], "used_tokens")
	window, okWindow := payloadNumber(payload[0], "window_tokens")
	if !okUsed || !okWindow || window <= 0 {
		return nil
	}
	pct := used / window * 100
	return &pct
}

// latestQuotaWindows reads the newest provider.quota.observed measurement
// per limit window (primary/secondary today; the set stays open — any
// limit_id that arrives is surfaced). The per-window grouping runs
// Go-side over the last few quota rows rather than via a JSON SQL
// expression, keeping the query portable across SQLite builds. Rows come
// back newest-first, so the FIRST payload seen per limit_id is that
// window's latest sample. Windows are returned sorted by LimitID for a
// deterministic render; nil when nothing was observed (unknown is not
// zero).
func (s *CodexStatusStore) latestQuotaWindows(ctx context.Context, sessionID string) []evaluation.QuotaWindowStatus {
	payloads, ok := s.latestPayload(ctx, sessionID, v1.EventProviderQuotaObserved, 16)
	if !ok {
		return nil
	}
	var out []evaluation.QuotaWindowStatus
	seen := map[string]bool{}
	for _, p := range payloads {
		id, _ := p["limit_id"].(string)
		if id == "" || seen[id] {
			continue
		}
		seen[id] = true
		w := evaluation.QuotaWindowStatus{LimitID: id}
		if pct, okPct := payloadNumber(p, "used_percent"); okPct {
			w.UsedPercent = &pct
		}
		if raw, okRaw := p["resets_at"].(string); okRaw && raw != "" {
			if t, err := time.Parse(time.RFC3339Nano, raw); err == nil {
				w.ResetsAt = &t
			}
		}
		if w.UsedPercent == nil && w.ResetsAt == nil {
			continue // a window that measured nothing observes nothing.
		}
		out = append(out, w)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].LimitID < out[j].LimitID })
	return out
}

// latestPayload returns up to limit decoded payloads of the session's
// newest events of eventType, newest first. ok=false on any query/decode
// trouble (individual undecodable rows are skipped, not fatal).
func (s *CodexStatusStore) latestPayload(ctx context.Context, sessionID string, eventType v1.EventType, limit int) ([]map[string]any, bool) {
	rows, err := s.DB.Conn().QueryContext(ctx, `
		SELECT payload_json FROM events
		WHERE session_id = ? AND provider = ? AND event_type = ?
		ORDER BY observed_at DESC, rowid DESC LIMIT ?`,
		sessionID, codextelemetry.Provider, string(eventType), limit,
	)
	if err != nil {
		return nil, false
	}
	defer func() { _ = rows.Close() }()

	var out []map[string]any
	for rows.Next() {
		var raw string
		if err := rows.Scan(&raw); err != nil {
			return nil, false
		}
		var payload map[string]any
		if json.Unmarshal([]byte(raw), &payload) != nil {
			continue
		}
		out = append(out, payload)
	}
	if rows.Err() != nil {
		return nil, false
	}
	return out, true
}

// payloadNumber reads a numeric payload field. JSON round-trips numbers as
// float64; ok=false for absent or non-numeric values.
func payloadNumber(payload map[string]any, key string) (float64, bool) {
	v, ok := payload[key].(float64)
	return v, ok
}
