// posttooluse.go: parsing + tool classification for the Claude Code
// PostToolUse hook (issue #67 slice 3a, approved by ADR-052). Claude Code
// fires PostToolUse once per tool call with the tool's name, input, and
// response; the file-touching tools are the repeated-file-operation
// signal (research doc §7.2):
//
//	view   — Read
//	modify — Edit, Write, MultiEdit, NotebookEdit
//
// # Privacy (ADR-052's binding invariant; Constitution §7 rule 2)
//
// A PostToolUse payload is one of the most sensitive inputs Auspex
// reads: tool_input carries raw file paths AND file content (Write's
// content, Edit's old_string/new_string), and tool_response can echo the
// file back. This parser therefore retains NONE of it:
//
//   - tool_input is decoded through a typed two-field projection that
//     reads ONLY file_path/notebook_path — content fields are never
//     extracted into any Go value beyond the transient raw bytes every
//     parser necessarily holds;
//   - even the file path itself is reduced to a presence BIT
//     (HasFileTarget) before ParsePostToolUse returns. The parsed struct
//     carries no path field at all, so no later code — normalizer,
//     scratch store, logger — can leak what it never received. Path
//     identity for the per-turn distinct/repeat aggregates is resolved
//     separately, in a single process's memory at Stop time
//     (internal/telemetry/claude/toolops.go), and discarded.
package claude

import (
	"encoding/json"
	"fmt"

	"github.com/huaiche94/auspex/internal/domain"
)

// ToolOpClass classifies a provider tool name for the issue-#67 per-turn
// file-operation aggregate (research doc §7.2, restated as binding by
// ADR-052).
type ToolOpClass string

const (
	// ToolOpView is a file-reading operation (Read).
	ToolOpView ToolOpClass = "view"
	// ToolOpModify is a file-writing operation (Edit, Write, MultiEdit,
	// NotebookEdit).
	ToolOpModify ToolOpClass = "modify"
	// ToolOpIgnored is every other tool: not part of the file-operation
	// aggregate (Bash, Glob, Grep, Task, WebFetch, unknown future tools —
	// deliberately an open set that defaults to "not counted").
	ToolOpIgnored ToolOpClass = "ignored"
)

// ClassifyToolOp maps a Claude Code tool name onto the frozen §7.2/ADR-052
// classification. Unknown names are ToolOpIgnored: a tool this build has
// never heard of must not be guessed into the aggregate.
func ClassifyToolOp(toolName string) ToolOpClass {
	switch toolName {
	case "Read":
		return ToolOpView
	case "Edit", "Write", "MultiEdit", "NotebookEdit":
		return ToolOpModify
	default:
		return ToolOpIgnored
	}
}

// PostToolUseEvent is the parsed, already privacy-reduced representation
// of a Claude Code PostToolUse hook payload. Note what is absent: no file
// path, no tool input, no tool response — see the file doc comment.
type PostToolUseEvent struct {
	SessionID domain.SessionID

	// TurnID is the turn identifier when the payload carries one. Claude
	// Code's documented PostToolUse payload has no turn field today; this
	// is parsed defensively so a provider that starts sending one is
	// honored the day it appears. nil means absent (the caller resolves
	// the session's open turn instead — unknown is not fabricated).
	TurnID *string

	// ToolName is the provider's own tool identifier ("Read", "Edit",
	// ...). Empty when the payload carried none — such an event is
	// ToolOpIgnored by construction.
	ToolName string

	// Class is ClassifyToolOp(ToolName), precomputed so callers share one
	// classification point.
	Class ToolOpClass

	// HasFileTarget reports whether tool_input named a file
	// (file_path/notebook_path non-empty). The path itself is discarded
	// during parsing and is not carried here.
	HasFileTarget bool
}

// FileOp reports whether this event is a countable file operation for the
// §7.3 per-turn aggregate: a view/modify tool that actually named a file.
func (e PostToolUseEvent) FileOp() bool {
	return e.Class != ToolOpIgnored && e.HasFileTarget
}

// rawPostToolUse mirrors the documented PostToolUse stdin envelope
// (session_id / transcript_path / cwd / hook_event_name / tool_name /
// tool_input / tool_response), tolerating unknown fields. tool_response
// is deliberately not declared: nothing in it is consumed, so its bytes
// are never decoded into a Go value at all.
type rawPostToolUse struct {
	SessionID     string          `json:"session_id"`
	TurnID        *string         `json:"turn_id"`
	HookEventName string          `json:"hook_event_name"`
	ToolName      string          `json:"tool_name"`
	ToolInput     json.RawMessage `json:"tool_input"`
}

// rawToolInputTarget is the typed two-field projection of tool_input:
// file_path (Read/Edit/Write/MultiEdit) or notebook_path (NotebookEdit).
// Content-bearing fields (content, old_string, new_string, edits, ...)
// have no struct field and are never extracted.
type rawToolInputTarget struct {
	FilePath     string `json:"file_path"`
	NotebookPath string `json:"notebook_path"`
}

// ParsePostToolUse parses a Claude Code PostToolUse hook stdin payload,
// tolerating unknown fields, and reduces it to the privacy-safe
// PostToolUseEvent (see the file doc comment). Like the sibling parsers,
// invalid JSON and a missing session_id are validation errors — the hook
// command's caller fails OPEN on them (never blocking the provider's
// turn), but they must be distinguishable from a parsed no-op.
func ParsePostToolUse(raw []byte) (PostToolUseEvent, error) {
	var r rawPostToolUse
	if err := json.Unmarshal(raw, &r); err != nil {
		return PostToolUseEvent{}, &domain.Error{
			Code:      domain.ErrCodeValidation,
			Message:   fmt.Sprintf("claude posttooluse: invalid JSON: %v", err),
			Retryable: false,
		}
	}

	if r.SessionID == "" {
		return PostToolUseEvent{}, &domain.Error{
			Code:      domain.ErrCodeValidation,
			Message:   "claude posttooluse: missing session_id",
			Retryable: false,
		}
	}

	ev := PostToolUseEvent{
		SessionID: domain.SessionID(r.SessionID),
		TurnID:    r.TurnID,
		ToolName:  r.ToolName,
		Class:     ClassifyToolOp(r.ToolName),
	}
	if len(r.ToolInput) > 0 {
		// The one place path bytes exist in this package: a transient
		// decode for the presence bit. A malformed tool_input is tolerated
		// (HasFileTarget stays false — unknown is not a counted op).
		var target rawToolInputTarget
		if json.Unmarshal(r.ToolInput, &target) == nil {
			ev.HasFileTarget = target.FilePath != "" || target.NotebookPath != ""
		}
	}
	return ev, nil
}
