package adapters

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/ariguillegp/rivet/internal/core"
)

func TestSessionNameForUsesWorktreeBasenameAndPathHash(t *testing.T) {
	dirPath := "/home/demo/.rivet/worktrees/rivet-d5b447--tmux-session-name"
	name, err := sessionNameFor(core.SessionSpec{
		DirPath: dirPath,
		Tool:    "codex",
	})
	if err != nil {
		t.Fatalf("unexpected session name error: %v", err)
	}
	expected := expectedPathSessionName(dirPath)
	if name != expected {
		t.Fatalf("expected path-derived session name %q, got %q", expected, name)
	}
}

func TestSessionNameForSanitizesBasenameAndAppendsPathHash(t *testing.T) {
	dirPath := "/tmp/feature branch!"
	name, err := sessionNameFor(core.SessionSpec{
		DirPath: dirPath,
		Tool:    "amp",
	})
	if err != nil {
		t.Fatalf("unexpected session name error: %v", err)
	}
	expected := expectedPathSessionName(dirPath)
	if name != expected {
		t.Fatalf("expected sanitized path-derived session name %q, got %q", expected, name)
	}
}

func TestSessionNameForRootWorktreeUsesStablePathName(t *testing.T) {
	projectPath := filepath.Join(t.TempDir(), "api")
	initRepo(t, projectPath)

	name, err := sessionNameFor(core.SessionSpec{
		DirPath: projectPath,
		Tool:    "codex",
	})
	if err != nil {
		t.Fatalf("unexpected session name error: %v", err)
	}

	expected := expectedPathSessionName(projectPath)
	if name != expected {
		t.Fatalf("expected root worktree session name %q, got %q", expected, name)
	}
}

func TestSessionNameForRootWorktreesWithSameBasenameDoNotCollide(t *testing.T) {
	root := t.TempDir()
	projectA := filepath.Join(root, "client", "api")
	projectB := filepath.Join(root, "scratch", "api")
	initRepo(t, projectA)
	initRepo(t, projectB)

	nameA, err := sessionNameFor(core.SessionSpec{DirPath: projectA, Tool: "codex"})
	if err != nil {
		t.Fatalf("unexpected session name error for project A: %v", err)
	}
	nameB, err := sessionNameFor(core.SessionSpec{DirPath: projectB, Tool: "codex"})
	if err != nil {
		t.Fatalf("unexpected session name error for project B: %v", err)
	}
	if nameA == nameB {
		t.Fatalf("expected distinct session names for same-basename roots, got %q", nameA)
	}
}

func TestSessionNameForAvoidsKnownShortHashCollision(t *testing.T) {
	pathA := "/tmp/collide/1900/api"
	pathB := "/tmp/collide/6785/api"
	if shortPathHash(pathA) != shortPathHash(pathB) {
		t.Fatalf("expected test paths to demonstrate a short hash collision")
	}

	nameA, err := sessionNameFor(core.SessionSpec{DirPath: pathA, Tool: "codex"})
	if err != nil {
		t.Fatalf("unexpected session name error for path A: %v", err)
	}
	nameB, err := sessionNameFor(core.SessionSpec{DirPath: pathB, Tool: "codex"})
	if err != nil {
		t.Fatalf("unexpected session name error for path B: %v", err)
	}
	if nameA == nameB {
		t.Fatalf("expected distinct session names for known short-hash collision, got %q", nameA)
	}
}

func TestSessionNameForIsStableAcrossBranchChanges(t *testing.T) {
	projectPath := filepath.Join(t.TempDir(), "api")
	initRepo(t, projectPath)

	nameBefore, err := sessionNameFor(core.SessionSpec{
		DirPath: projectPath,
		Tool:    "codex",
	})
	if err != nil {
		t.Fatalf("unexpected session name error: %v", err)
	}

	checkoutNewBranch(t, projectPath, "feature/session-name")
	nameAfter, err := sessionNameFor(core.SessionSpec{
		DirPath: projectPath,
		Tool:    "codex",
	})
	if err != nil {
		t.Fatalf("unexpected session name error after branch change: %v", err)
	}
	if nameAfter != nameBefore {
		t.Fatalf("expected session name to stay stable across branch changes, got before=%q after=%q", nameBefore, nameAfter)
	}
}

func TestSessionNameForDetachedWorktreesInSameRepositoryDoNotCollide(t *testing.T) {
	root := t.TempDir()
	projectPath := filepath.Join(root, "api")
	initRepo(t, projectPath)
	worktreeA := filepath.Join(root, "detached-a")
	worktreeB := filepath.Join(root, "detached-b")
	addDetachedWorktree(t, projectPath, worktreeA)
	addDetachedWorktree(t, projectPath, worktreeB)

	nameA, err := sessionNameFor(core.SessionSpec{DirPath: worktreeA, Tool: "codex"})
	if err != nil {
		t.Fatalf("unexpected session name error for worktree A: %v", err)
	}
	nameB, err := sessionNameFor(core.SessionSpec{DirPath: worktreeB, Tool: "codex"})
	if err != nil {
		t.Fatalf("unexpected session name error for worktree B: %v", err)
	}
	if nameA == nameB {
		t.Fatalf("expected distinct detached session names for same-repository worktrees, got %q", nameA)
	}

	expectedA := expectedPathSessionName(worktreeA)
	expectedB := expectedPathSessionName(worktreeB)
	if nameA != expectedA {
		t.Fatalf("expected detached worktree A session name %q, got %q", expectedA, nameA)
	}
	if nameB != expectedB {
		t.Fatalf("expected detached worktree B session name %q, got %q", expectedB, nameB)
	}
}

func TestSessionNameForDetachedRootWorktreesWithSameBasenameDoNotCollide(t *testing.T) {
	root := t.TempDir()
	projectA := filepath.Join(root, "client", "api")
	projectB := filepath.Join(root, "scratch", "api")
	initRepo(t, projectA)
	initRepo(t, projectB)
	detachHead(t, projectA)
	detachHead(t, projectB)

	nameA, err := sessionNameFor(core.SessionSpec{DirPath: projectA, Tool: "codex"})
	if err != nil {
		t.Fatalf("unexpected session name error for project A: %v", err)
	}
	nameB, err := sessionNameFor(core.SessionSpec{DirPath: projectB, Tool: "codex"})
	if err != nil {
		t.Fatalf("unexpected session name error for project B: %v", err)
	}
	if nameA == nameB {
		t.Fatalf("expected distinct detached session names for same-basename roots, got %q", nameA)
	}
}

func TestSessionNameForManagedWorktreeUsesUniformGitWorktreeSchema(t *testing.T) {
	projectPath := filepath.Join(t.TempDir(), "rivet")
	initRepo(t, projectPath)

	fs := &OSFilesystem{}
	worktreePath, err := fs.CreateWorktree(projectPath, "feature/test")
	if err != nil {
		t.Fatalf("unexpected worktree creation error: %v", err)
	}
	t.Cleanup(func() {
		_ = fs.DeleteWorktree(projectPath, worktreePath)
	})

	name, err := sessionNameFor(core.SessionSpec{
		DirPath: worktreePath,
		Tool:    "codex",
	})
	if err != nil {
		t.Fatalf("unexpected session name error: %v", err)
	}
	expected := expectedPathSessionName(worktreePath)
	if name != expected {
		t.Fatalf("expected managed worktree session name %q, got %q", expected, name)
	}
	if strings.Contains(name, "feature-test") {
		t.Fatalf("expected managed worktree session name to exclude branch text, got %q", name)
	}
}

func expectedPathSessionName(dirPath string) string {
	cleanPath := filepath.Clean(dirPath)
	name := sessionProjectName(cleanPath)
	if strings.TrimSpace(name) == "" {
		name = filepath.Base(cleanPath)
	}
	if name == "." || name == string(filepath.Separator) || name == "" {
		name = "worktree"
	}
	name = strings.Trim(sanitizeSessionPart(name, "worktree"), "-")
	if name == "" {
		name = "worktree"
	}
	return name + "-" + sessionPathHash(cleanPath)
}

func detachHead(t *testing.T, dir string) {
	t.Helper()
	cmd := exec.Command("git", "-C", dir, "checkout", "--detach", "HEAD")
	if output, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("git checkout --detach failed: %v: %s", err, string(output))
	}
}

func checkoutNewBranch(t *testing.T, dir, branch string) {
	t.Helper()
	cmd := exec.Command("git", "-C", dir, "checkout", "-b", branch)
	if output, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("git checkout -b failed: %v: %s", err, string(output))
	}
}

func addDetachedWorktree(t *testing.T, projectPath, worktreePath string) {
	t.Helper()
	cmd := exec.Command("git", "-C", projectPath, "worktree", "add", "--detach", worktreePath, "HEAD")
	if output, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("git worktree add --detach failed: %v: %s", err, string(output))
	}
	t.Cleanup(func() {
		cmd := exec.Command("git", "-C", projectPath, "worktree", "remove", "--force", worktreePath)
		_ = cmd.Run()
	})
}

func TestKillSessionIgnoresNoServerRunning(t *testing.T) {
	tmpDir := t.TempDir()
	tmuxPath := filepath.Join(tmpDir, "tmux")

	script := "#!/bin/sh\n" +
		"echo \"no server running on /tmp/tmux-0/default\" 1>&2\n" +
		"exit 1\n"

	if err := os.WriteFile(tmuxPath, []byte(script), 0o755); err != nil {
		t.Fatalf("failed to write tmux stub: %v", err)
	}

	pathEnv := os.Getenv("PATH")
	pathSep := string(os.PathListSeparator)
	t.Setenv("PATH", tmpDir+pathSep+pathEnv)

	session := &TmuxSession{}
	spec := core.SessionSpec{DirPath: "/tmp/project", Tool: "amp"}
	if err := session.KillSession(spec); err != nil {
		t.Fatalf("expected no error when tmux server is missing: %v", err)
	}
}

func TestKillSessionIgnoresMissingTmuxSocket(t *testing.T) {
	tmpDir := t.TempDir()
	tmuxPath := filepath.Join(tmpDir, "tmux")

	script := "#!/bin/sh\n" +
		"echo \"error connecting to /tmp/tmux-1000/default (No such file or directory)\" 1>&2\n" +
		"exit 1\n"

	if err := os.WriteFile(tmuxPath, []byte(script), 0o755); err != nil {
		t.Fatalf("failed to write tmux stub: %v", err)
	}

	pathEnv := os.Getenv("PATH")
	pathSep := string(os.PathListSeparator)
	t.Setenv("PATH", tmpDir+pathSep+pathEnv)

	session := &TmuxSession{}
	spec := core.SessionSpec{DirPath: "/tmp/project", Tool: "amp"}
	if err := session.KillSession(spec); err != nil {
		t.Fatalf("expected no error when tmux socket is missing: %v", err)
	}
}

func TestKillSessionIgnoresMissingSession(t *testing.T) {
	tmpDir := t.TempDir()
	tmuxPath := filepath.Join(tmpDir, "tmux")

	script := "#!/bin/sh\n" +
		"echo \"can't find session\" 1>&2\n" +
		"exit 1\n"

	if err := os.WriteFile(tmuxPath, []byte(script), 0o755); err != nil {
		t.Fatalf("failed to write tmux stub: %v", err)
	}

	pathEnv := os.Getenv("PATH")
	pathSep := string(os.PathListSeparator)
	t.Setenv("PATH", tmpDir+pathSep+pathEnv)

	session := &TmuxSession{}
	spec := core.SessionSpec{DirPath: "/tmp/project", Tool: "amp"}
	if err := session.KillSession(spec); err != nil {
		t.Fatalf("expected no error when tmux session is missing: %v", err)
	}
}

func TestListSessionsIgnoresNoServerRunning(t *testing.T) {
	tmpDir := t.TempDir()
	tmuxPath := filepath.Join(tmpDir, "tmux")

	script := "#!/bin/sh\n" +
		"echo \"no server running on /tmp/tmux-0/default\" 1>&2\n" +
		"exit 1\n"

	if err := os.WriteFile(tmuxPath, []byte(script), 0o755); err != nil {
		t.Fatalf("failed to write tmux stub: %v", err)
	}

	pathEnv := os.Getenv("PATH")
	pathSep := string(os.PathListSeparator)
	t.Setenv("PATH", tmpDir+pathSep+pathEnv)

	session := &TmuxSession{}
	sessions, err := session.ListSessions()
	if err != nil {
		t.Fatalf("expected no error when tmux server is missing: %v", err)
	}
	if len(sessions) != 0 {
		t.Fatalf("expected no sessions when tmux server is missing")
	}
}

func TestListSessionsIgnoresMissingTmuxSocket(t *testing.T) {
	tmpDir := t.TempDir()
	tmuxPath := filepath.Join(tmpDir, "tmux")

	script := "#!/bin/sh\n" +
		"echo \"error connecting to /tmp/tmux-1000/default (No such file or directory)\" 1>&2\n" +
		"exit 1\n"

	if err := os.WriteFile(tmuxPath, []byte(script), 0o755); err != nil {
		t.Fatalf("failed to write tmux stub: %v", err)
	}

	pathEnv := os.Getenv("PATH")
	pathSep := string(os.PathListSeparator)
	t.Setenv("PATH", tmpDir+pathSep+pathEnv)

	session := &TmuxSession{}
	sessions, err := session.ListSessions()
	if err != nil {
		t.Fatalf("expected no error when tmux socket is missing: %v", err)
	}
	if len(sessions) != 0 {
		t.Fatalf("expected no sessions when tmux socket is missing")
	}
}

func TestListSessionsIncludesLastActiveWhenAvailable(t *testing.T) {
	tmpDir := t.TempDir()
	tmuxPath := filepath.Join(tmpDir, "tmux")

	script := "#!/bin/sh\n" +
		"echo \"demo\t/tmp/projects/rivet/main\t1735689600\"\n" +
		"exit 0\n"

	if err := os.WriteFile(tmuxPath, []byte(script), 0o755); err != nil {
		t.Fatalf("failed to write tmux stub: %v", err)
	}

	pathEnv := os.Getenv("PATH")
	pathSep := string(os.PathListSeparator)
	t.Setenv("PATH", tmpDir+pathSep+pathEnv)

	session := &TmuxSession{}
	sessions, err := session.ListSessions()
	if err != nil {
		t.Fatalf("expected no error listing sessions: %v", err)
	}
	if len(sessions) != 1 {
		t.Fatalf("expected one session, got %d", len(sessions))
	}
	if sessions[0].Name != "demo" {
		t.Fatalf("expected session name demo, got %q", sessions[0].Name)
	}
	if sessions[0].LastActive.IsZero() {
		t.Fatalf("expected last active timestamp to be parsed")
	}
}

func TestListSessionsIncludesProjectAndBranchMetadata(t *testing.T) {
	tmpDir := t.TempDir()
	tmuxPath := filepath.Join(tmpDir, "tmux")
	gitPath := filepath.Join(tmpDir, "git")

	tmuxScript := `#!/bin/sh
echo "demo	/home/demo/Projects/rivet/rbac-sentinel	1735689600"
exit 0
`

	gitScript := `#!/bin/sh
if [ "$3" = "rev-parse" ] && [ "$4" = "--git-common-dir" ]; then
  echo "/home/demo/Projects/rivet/.git"
  exit 0
fi
if [ "$3" = "branch" ]; then
  echo "rbac-sentinel"
  exit 0
fi
exit 1
`

	if err := os.WriteFile(tmuxPath, []byte(tmuxScript), 0o755); err != nil {
		t.Fatalf("failed to write tmux stub: %v", err)
	}
	if err := os.WriteFile(gitPath, []byte(gitScript), 0o755); err != nil {
		t.Fatalf("failed to write git stub: %v", err)
	}

	pathEnv := os.Getenv("PATH")
	pathSep := string(os.PathListSeparator)
	t.Setenv("PATH", tmpDir+pathSep+pathEnv)

	session := &TmuxSession{}
	sessions, err := session.ListSessions()
	if err != nil {
		t.Fatalf("expected no error listing sessions: %v", err)
	}
	if len(sessions) != 1 {
		t.Fatalf("expected one session, got %d", len(sessions))
	}
	if sessions[0].Project != "rivet" {
		t.Fatalf("expected project name from git metadata, got %q", sessions[0].Project)
	}
	if sessions[0].Branch != "rbac-sentinel" {
		t.Fatalf("expected branch from git metadata, got %q", sessions[0].Branch)
	}
}
