package adapters

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/ariguillegp/rivet/internal/core"
)

func TestKillSessionIgnoresNoServerRunning(t *testing.T) {
	tmpDir := t.TempDir()
	tmuxPath := filepath.Join(tmpDir, "tmux")

	script := "#!/bin/sh\n" +
		"echo \"no server running on /tmp/tmux-0/default\" 1>&2\n" +
		"exit 1\n"

	if err := os.WriteFile(tmuxPath, []byte(script), 0755); err != nil {
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

func TestKillSessionIgnoresMissingSession(t *testing.T) {
	tmpDir := t.TempDir()
	tmuxPath := filepath.Join(tmpDir, "tmux")

	script := "#!/bin/sh\n" +
		"echo \"can't find session\" 1>&2\n" +
		"exit 1\n"

	if err := os.WriteFile(tmuxPath, []byte(script), 0755); err != nil {
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

	if err := os.WriteFile(tmuxPath, []byte(script), 0755); err != nil {
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

func TestListSessionsIncludesLastActiveWhenAvailable(t *testing.T) {
	tmpDir := t.TempDir()
	tmuxPath := filepath.Join(tmpDir, "tmux")

	script := "#!/bin/sh\n" +
		"echo \"demo\t/tmp/projects/rivet/main\t1735689600\"\n" +
		"exit 0\n"

	if err := os.WriteFile(tmuxPath, []byte(script), 0755); err != nil {
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
