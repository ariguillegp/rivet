package core

import "testing"

func TestSessionDisplayLabelUsesProjectBranchTool(t *testing.T) {
	session := SessionInfo{
		Name:    "raw",
		DirPath: "/home/demo/.rivet/worktrees/rivet--feature-1",
		Tool:    "amp",
	}

	label := SessionDisplayLabel(session)
	if label != "rivet/feature-1 - amp" {
		t.Fatalf("expected label to include project/branch/tool, got %q", label)
	}
}

func TestSessionDisplayLabelStripsProjectHash(t *testing.T) {
	session := SessionInfo{
		Name:    "raw",
		DirPath: "/home/demo/.rivet/worktrees/rivet-abcdef--feature-1",
		Tool:    "amp",
	}

	label := SessionDisplayLabel(session)
	if label != "rivet/feature-1 - amp" {
		t.Fatalf("expected label to strip hash, got %q", label)
	}
}

func TestSessionWorktreeProjectBranchSplitsOnLastDelimiter(t *testing.T) {
	project, branch := SessionWorktreeProjectBranch("my-project--feature-x")
	if project != "my-project" || branch != "feature-x" {
		t.Fatalf("unexpected split: project=%q branch=%q", project, branch)
	}
}
