package claude

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/huaiche94/auspex/internal/domain"
)

func postToolUseFixture(t *testing.T, name string) []byte {
	t.Helper()
	b, err := os.ReadFile(filepath.Join("..", "..", "..", "testdata", "provider-events", "claude", "posttooluse", name))
	if err != nil {
		t.Fatalf("reading fixture posttooluse/%s: %v", name, err)
	}
	return b
}

func TestParsePostToolUse_Fixtures(t *testing.T) {
	tests := []struct {
		name string

		fixture string

		wantErr           bool
		wantSessionID     domain.SessionID
		wantToolName      string
		wantClass         ToolOpClass
		wantHasFileTarget bool
		wantFileOp        bool
	}{
		{
			name:              "normal",
			fixture:           "normal.json",
			wantSessionID:     "sess_01H9X8K7QZ3M4N5P6R7S8T9V0W",
			wantToolName:      "Edit",
			wantClass:         ToolOpModify,
			wantHasFileTarget: true,
			wantFileOp:        true,
		},
		{
			// No tool_input at all: the classification still runs, but a
			// view/modify tool with no file target is not a countable op.
			name:          "missing_fields",
			fixture:       "missing_fields.json",
			wantSessionID: "sess_01H9X8K7QZ3M4N5P6R7S8T9V0X",
			wantToolName:  "Read",
			wantClass:     ToolOpView,
			wantFileOp:    false,
		},
		{
			// Unknown envelope/input/response fields must be tolerated
			// (§21.7's unknown-fields rule); NotebookEdit's notebook_path is
			// the second file-target spelling.
			name:              "unknown_fields",
			fixture:           "unknown_fields.json",
			wantSessionID:     "sess_01H9X8K7QZ3M4N5P6R7S8T9V0Z",
			wantToolName:      "NotebookEdit",
			wantClass:         ToolOpModify,
			wantHasFileTarget: true,
			wantFileOp:        true,
		},
		{
			name:    "malformed",
			fixture: "malformed.json",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ev, err := ParsePostToolUse(postToolUseFixture(t, tt.fixture))
			if tt.wantErr {
				if err == nil {
					t.Fatal("expected a parse error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("ParsePostToolUse: %v", err)
			}
			if ev.SessionID != tt.wantSessionID {
				t.Errorf("SessionID = %q, want %q", ev.SessionID, tt.wantSessionID)
			}
			if ev.ToolName != tt.wantToolName {
				t.Errorf("ToolName = %q, want %q", ev.ToolName, tt.wantToolName)
			}
			if ev.Class != tt.wantClass {
				t.Errorf("Class = %q, want %q", ev.Class, tt.wantClass)
			}
			if ev.HasFileTarget != tt.wantHasFileTarget {
				t.Errorf("HasFileTarget = %v, want %v", ev.HasFileTarget, tt.wantHasFileTarget)
			}
			if ev.FileOp() != tt.wantFileOp {
				t.Errorf("FileOp() = %v, want %v", ev.FileOp(), tt.wantFileOp)
			}
			if ev.TurnID != nil {
				t.Errorf("TurnID = %q, want nil (no fixture carries one)", *ev.TurnID)
			}
		})
	}
}

func TestParsePostToolUse_MissingSessionID(t *testing.T) {
	_, err := ParsePostToolUse([]byte(`{"hook_event_name":"PostToolUse","tool_name":"Read","tool_input":{"file_path":"/tmp/x.go"}}`))
	if err == nil {
		t.Fatal("expected a missing-session_id validation error, got nil")
	}
	var derr *domain.Error
	if !errors.As(err, &derr) || derr.Code != domain.ErrCodeValidation {
		t.Fatalf("error = %v, want a domain.Error with ErrCodeValidation", err)
	}
}

// TestClassifyToolOp pins the frozen §7.2/ADR-052 classification: view =
// Read; modify = Edit, Write, MultiEdit, NotebookEdit; everything else —
// including tools this build has never heard of — ignored.
func TestClassifyToolOp(t *testing.T) {
	tests := []struct {
		tool string
		want ToolOpClass
	}{
		{"Read", ToolOpView},
		{"Edit", ToolOpModify},
		{"Write", ToolOpModify},
		{"MultiEdit", ToolOpModify},
		{"NotebookEdit", ToolOpModify},
		{"Bash", ToolOpIgnored},
		{"Glob", ToolOpIgnored},
		{"Grep", ToolOpIgnored},
		{"Task", ToolOpIgnored},
		{"WebFetch", ToolOpIgnored},
		{"read", ToolOpIgnored},           // case-sensitive: not the provider's tool
		{"FutureFileTool", ToolOpIgnored}, // unknown tools are never guessed in
		{"", ToolOpIgnored},
	}
	for _, tt := range tests {
		if got := ClassifyToolOp(tt.tool); got != tt.want {
			t.Errorf("ClassifyToolOp(%q) = %q, want %q", tt.tool, got, tt.want)
		}
	}
}

// TestParsePostToolUse_NoPathOrContentEverRetained is the parser-level
// privacy gate (ADR-052's binding invariant): the parsed struct must not
// carry the payload's file path, file content (old_string/new_string),
// or tool response — verified via a full Go %#v dump, the same technique
// the fixture suite's raw-text gate uses. The needles are copied verbatim
// from the fixture files; the self-check keeps the two from drifting.
func TestParsePostToolUse_NoPathOrContentEverRetained(t *testing.T) {
	needles := map[string][]string{
		"normal.json": {
			"/Users/dev/projects/auspex/internal/rotation/keyring.go",
			"secret canary: keyring rotation cadence",
		},
		"unknown_fields.json": {
			"/Users/dev/projects/auspex/research/calibration/experiments.ipynb",
			"notebook canary: calibration sweep",
		},
	}
	for file, fileNeedles := range needles {
		raw := postToolUseFixture(t, file)
		for _, needle := range fileNeedles {
			if !strings.Contains(string(raw), needle) {
				t.Fatalf("needle table is stale: posttooluse/%s no longer contains %q — update the table", file, needle)
			}
		}

		ev, err := ParsePostToolUse(raw)
		if err != nil {
			t.Fatalf("ParsePostToolUse(%s): %v", file, err)
		}
		dump := fmt.Sprintf("%#v", ev)
		for _, needle := range fileNeedles {
			if strings.Contains(dump, needle) {
				t.Errorf("%s: sensitive payload text %q leaked into parsed PostToolUseEvent: %s", file, needle, dump)
			}
		}
	}
}
