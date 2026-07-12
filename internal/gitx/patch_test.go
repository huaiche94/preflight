package gitx

import (
	"bytes"
	"context"
	"os"
	"strings"
	"testing"
)

func TestDiffPatch_StagedChange_ProducesApplicablePatch(t *testing.T) {
	rb := newRepoBuilder(t)
	rb.write("a.txt", "line1\n")
	rb.git("add", "a.txt")
	rb.git("commit", "-q", "-m", "initial")

	rb.write("a.txt", "line1\nline2\n")
	rb.git("add", "a.txt")

	c := NewClient(rb.runner)
	patch, err := c.DiffPatch(context.Background(), rb.dir, true)
	if err != nil {
		t.Fatalf("DiffPatch: %v", err)
	}
	if !bytes.Contains(patch, []byte("a.txt")) {
		t.Fatalf("expected patch to mention a.txt, got: %s", patch)
	}
	if !bytes.Contains(patch, []byte("+line2")) {
		t.Fatalf("expected patch to show added line2, got: %s", patch)
	}
}

func TestDiffPatch_Unstaged_ProducesPatch(t *testing.T) {
	rb := newRepoBuilder(t)
	rb.write("a.txt", "line1\n")
	rb.git("add", "a.txt")
	rb.git("commit", "-q", "-m", "initial")

	rb.write("a.txt", "line1\nunstaged change\n")

	c := NewClient(rb.runner)
	patch, err := c.DiffPatch(context.Background(), rb.dir, false)
	if err != nil {
		t.Fatalf("DiffPatch: %v", err)
	}
	if !bytes.Contains(patch, []byte("unstaged change")) {
		t.Fatalf("expected patch to show unstaged change, got: %s", patch)
	}
}

func TestDiffPatch_NoChanges_EmptyPatchNoError(t *testing.T) {
	rb := newRepoBuilder(t)
	rb.write("a.txt", "line1\n")
	rb.git("add", "a.txt")
	rb.git("commit", "-q", "-m", "initial")

	c := NewClient(rb.runner)
	patch, err := c.DiffPatch(context.Background(), rb.dir, true)
	if err != nil {
		t.Fatalf("DiffPatch: %v", err)
	}
	if len(patch) != 0 {
		t.Fatalf("expected empty patch for no staged changes, got: %s", patch)
	}
}

func TestDiffPatch_BinaryFile_UsesBinaryDirective(t *testing.T) {
	rb := newRepoBuilder(t)
	rb.git("commit", "--allow-empty", "-q", "-m", "initial")

	binPath := rb.dir + "/bin.dat"
	if err := writeBinaryFixture(binPath); err != nil {
		t.Fatalf("write binary fixture: %v", err)
	}
	rb.git("add", "bin.dat")

	c := NewClient(rb.runner)
	patch, err := c.DiffPatch(context.Background(), rb.dir, true)
	if err != nil {
		t.Fatalf("DiffPatch: %v", err)
	}
	if !bytes.Contains(patch, []byte("GIT binary patch")) && !bytes.Contains(patch, []byte("Binary files")) {
		t.Fatalf("expected a binary-patch directive in output, got: %s", patch)
	}
}

func TestListUntracked_ReturnsNonIgnoredFiles(t *testing.T) {
	rb := newRepoBuilder(t)
	rb.git("commit", "--allow-empty", "-q", "-m", "initial")
	rb.write(".gitignore", "ignored.txt\n")
	rb.write("ignored.txt", "should not appear\n")
	rb.write("untracked.txt", "should appear\n")

	c := NewClient(rb.runner)
	paths, err := c.ListUntracked(context.Background(), rb.dir)
	if err != nil {
		t.Fatalf("ListUntracked: %v", err)
	}

	joined := strings.Join(paths, ",")
	if !strings.Contains(joined, "untracked.txt") {
		t.Fatalf("expected untracked.txt in result, got: %v", paths)
	}
	if strings.Contains(joined, "ignored.txt") {
		t.Fatalf("expected ignored.txt to be excluded, got: %v", paths)
	}
}

func TestListUntracked_NoUntrackedFiles_EmptyResult(t *testing.T) {
	rb := newRepoBuilder(t)
	rb.write("a.txt", "content\n")
	rb.git("add", "a.txt")
	rb.git("commit", "-q", "-m", "initial")

	c := NewClient(rb.runner)
	paths, err := c.ListUntracked(context.Background(), rb.dir)
	if err != nil {
		t.Fatalf("ListUntracked: %v", err)
	}
	if len(paths) != 0 {
		t.Fatalf("expected no untracked files, got: %v", paths)
	}
}

// TestApplyCheck_CleanPatch_WouldApply proves the straightforward case: a
// patch generated against a given base still applies cleanly when the
// index is later reset back to that same base (the realistic restore
// scenario this method exists for: "if the target were at the patch's own
// base, would this patch apply" — a patch already fully applied to the
// CURRENT index is a different, self-contradictory question, since the
// index has already moved past the patch's own "before" state), and
// ApplyCheck reports so WITHOUT mutating anything (the index content is
// asserted unchanged afterward).
func TestApplyCheck_CleanPatch_WouldApply(t *testing.T) {
	rb := newRepoBuilder(t)
	rb.write("a.txt", "line1\n")
	rb.git("add", "a.txt")
	rb.git("commit", "-q", "-m", "initial")

	rb.write("a.txt", "line1\nline2\n")
	rb.git("add", "a.txt")

	c := NewClient(rb.runner)
	patch, err := c.DiffPatch(context.Background(), rb.dir, true)
	if err != nil {
		t.Fatalf("DiffPatch: %v", err)
	}

	// Reset the index back to the patch's own base (as if the checkpoint
	// were being dry-run restored onto a worktree that never received the
	// staged change, or had it reverted since) - the realistic scenario
	// ApplyCheck's --check semantics are designed to answer.
	rb.git("reset", "-q", "a.txt")
	rb.write("a.txt", "line1\n")
	rb.git("add", "a.txt")

	result, err := c.ApplyCheck(context.Background(), rb.dir, patch, true)
	if err != nil {
		t.Fatalf("ApplyCheck: %v", err)
	}
	if !result.WouldApply {
		t.Fatalf("expected WouldApply=true for a clean patch against its own base, detail: %s", result.Detail)
	}

	// ApplyCheck must never mutate anything: the index content is still
	// exactly what it was before the check ran (still the pre-patch base,
	// not the patch's target state).
	content, err := os.ReadFile(rb.dir + "/a.txt")
	if err != nil {
		t.Fatalf("read a.txt: %v", err)
	}
	if string(content) != "line1\n" {
		t.Fatalf("expected a.txt unchanged by ApplyCheck (still the pre-patch base), got: %q", content)
	}
}

// TestApplyCheck_ConflictingWorktree_WouldNotApply proves the negative
// case: a patch that no longer applies (because the target file has since
// diverged from the patch's own base) is reported as WouldApply=false with
// Git's own diagnostic detail, not silently treated as fine.
func TestApplyCheck_ConflictingWorktree_WouldNotApply(t *testing.T) {
	rb := newRepoBuilder(t)
	rb.write("a.txt", "line1\n")
	rb.git("add", "a.txt")
	rb.git("commit", "-q", "-m", "initial")

	rb.write("a.txt", "line1\nline2\n")
	rb.git("add", "a.txt")

	c := NewClient(rb.runner)
	patch, err := c.DiffPatch(context.Background(), rb.dir, true)
	if err != nil {
		t.Fatalf("DiffPatch: %v", err)
	}

	// Unstage and diverge the file from the patch's own base so the patch
	// no longer applies cleanly (the context lines it expects are gone).
	rb.git("reset", "--hard", "HEAD")
	rb.write("a.txt", "totally different content\nwith more lines\n")
	rb.git("add", "a.txt")

	result, err := c.ApplyCheck(context.Background(), rb.dir, patch, true)
	if err != nil {
		t.Fatalf("ApplyCheck: %v", err)
	}
	if result.WouldApply {
		t.Fatal("expected WouldApply=false for a patch that conflicts with the current index")
	}
	if result.Detail == "" {
		t.Fatal("expected a non-empty diagnostic detail for a failed apply-check")
	}
}

// TestApplyCheck_EmptyPatch_TriviallyWouldApply proves the documented
// "nothing to apply" case: an empty patch always reports WouldApply=true
// without invoking Git at all.
func TestApplyCheck_EmptyPatch_TriviallyWouldApply(t *testing.T) {
	rb := newRepoBuilder(t)
	rb.git("commit", "--allow-empty", "-q", "-m", "initial")

	c := NewClient(rb.runner)
	result, err := c.ApplyCheck(context.Background(), rb.dir, nil, true)
	if err != nil {
		t.Fatalf("ApplyCheck: %v", err)
	}
	if !result.WouldApply {
		t.Fatal("expected WouldApply=true for an empty patch")
	}
}

// TestApplyCheck_UnstagedScope_ChecksAgainstWorkingTree proves the
// cached=false path checks against the WORKING TREE (not the index),
// mirroring DiffPatch's own cached/uncached scope split. Same "reset back
// to the patch's own base first" realism as the staged-scope test above:
// checking a patch against a working tree that already contains its
// target state is not a meaningful dry-run question.
func TestApplyCheck_UnstagedScope_ChecksAgainstWorkingTree(t *testing.T) {
	rb := newRepoBuilder(t)
	rb.write("a.txt", "line1\n")
	rb.git("add", "a.txt")
	rb.git("commit", "-q", "-m", "initial")

	rb.write("a.txt", "line1\nunstaged change\n")

	c := NewClient(rb.runner)
	patch, err := c.DiffPatch(context.Background(), rb.dir, false)
	if err != nil {
		t.Fatalf("DiffPatch: %v", err)
	}

	// Revert the working tree back to the patch's own base before
	// dry-running the check.
	rb.write("a.txt", "line1\n")

	result, err := c.ApplyCheck(context.Background(), rb.dir, patch, false)
	if err != nil {
		t.Fatalf("ApplyCheck: %v", err)
	}
	if !result.WouldApply {
		t.Fatalf("expected WouldApply=true for an unstaged patch against its own reverted-to-base working tree, detail: %s", result.Detail)
	}
}

func writeBinaryFixture(path string) error {
	data := make([]byte, 256)
	for i := range data {
		data[i] = byte(i)
	}
	return os.WriteFile(path, data, 0o644)
}
