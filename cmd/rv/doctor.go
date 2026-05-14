package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/ariguillegp/rivet/internal/core"
)

var doctorLookPath = exec.LookPath

func diagnoseEnvironment(roots []string) []doctorResult {
	results := make([]doctorResult, 0, 16)

	results = append(results,
		checkRequiredCommand("git"),
		checkRequiredCommand("tmux"),
		checkRequiredCommand("lazygit"),
	)

	tools := core.SupportedTools()
	installedAgents := make([]string, 0, len(tools))
	for _, tool := range tools {
		if tool == core.ToolNone {
			continue
		}
		result, installed := checkOptionalTool(tool)
		results = append(results, result)
		if installed {
			installedAgents = append(installedAgents, tool)
		}
	}
	results = append(results, checkAtLeastOneAgentTool(installedAgents))

	if len(roots) == 0 {
		roots = []string{"~/Projects"}
	}

	for _, root := range roots {
		results = append(results, checkRoot(root))
	}

	worktreeRoot := expandPath("~/.rivet/worktrees")
	results = append(results, checkWorktreeRoot(worktreeRoot))

	return results
}

func checkAtLeastOneAgentTool(installed []string) doctorResult {
	if len(installed) > 0 {
		return doctorResult{
			Name:    "tools:agents",
			OK:      true,
			Details: fmt.Sprintf("at least one agent tool installed (%s)", strings.Join(installed, ", ")),
		}
	}
	return doctorResult{
		Name:       "tools:agents",
		OK:         false,
		Details:    "no agent tools found in PATH",
		Suggestion: "install at least one agent tool: opencode, amp, claude, or codex",
	}
}

func checkRequiredCommand(name string) doctorResult {
	path, err := doctorLookPath(name)
	if err == nil {
		return doctorResult{
			Name:    fmt.Sprintf("command:%s", name),
			OK:      true,
			Details: fmt.Sprintf("found at %s", path),
		}
	}
	return doctorResult{
		Name:       fmt.Sprintf("command:%s", name),
		OK:         false,
		Details:    "not found in PATH",
		Suggestion: fmt.Sprintf("install %s and ensure it is in PATH before running rv", name),
	}
}

func checkOptionalTool(name string) (doctorResult, bool) {
	path, err := doctorLookPath(name)
	if err == nil {
		return doctorResult{
			Name:    fmt.Sprintf("tool:%s", name),
			OK:      true,
			Details: fmt.Sprintf("found at %s", path),
		}, true
	}
	return doctorResult{
		Name:    fmt.Sprintf("tool:%s", name),
		OK:      true,
		Details: "optional tool not found (sessions with this tool cannot be launched)",
		Suggestion: fmt.Sprintf(
			"install %s if you want rv to launch it, or keep using another supported tool",
			name,
		),
	}, false
}

func checkRoot(root string) doctorResult {
	expanded := expandPath(root)
	name := fmt.Sprintf("root:%s", root)
	info, err := os.Stat(expanded)
	if err != nil {
		return doctorResult{
			Name:       name,
			OK:         false,
			Details:    "directory does not exist",
			Suggestion: fmt.Sprintf("create it with: mkdir -p %q", expanded),
		}
	}
	if !info.IsDir() {
		return doctorResult{
			Name:       name,
			OK:         false,
			Details:    "path exists but is not a directory",
			Suggestion: fmt.Sprintf("replace it with a directory: rm -f %q && mkdir -p %q", expanded, expanded),
		}
	}

	if err := ensureDirPermissions(expanded, true); err != nil {
		return doctorResult{
			Name:       name,
			OK:         false,
			Details:    err.Error(),
			Suggestion: fmt.Sprintf("grant read/write/execute permissions (example: chmod u+rwx %q)", expanded),
		}
	}

	return doctorResult{
		Name:    name,
		OK:      true,
		Details: "exists and is readable/writable/searchable",
	}
}

func checkWorktreeRoot(path string) doctorResult {
	name := "worktrees:~/.rivet/worktrees"
	info, err := os.Stat(path)
	if err != nil {
		if os.IsNotExist(err) {
			return doctorResult{
				Name:    name,
				OK:      true,
				Details: "directory does not exist yet (rv will create it when needed)",
				Suggestion: fmt.Sprintf(
					"optional pre-create: mkdir -p %q",
					path,
				),
			}
		}
		return doctorResult{
			Name:       name,
			OK:         false,
			Details:    fmt.Sprintf("cannot stat directory: %v", err),
			Suggestion: fmt.Sprintf("ensure parent directory is accessible: %q", filepath.Dir(path)),
		}
	}
	if !info.IsDir() {
		return doctorResult{
			Name:       name,
			OK:         false,
			Details:    "path exists but is not a directory",
			Suggestion: fmt.Sprintf("replace it with a directory: rm -f %q && mkdir -p %q", path, path),
		}
	}

	if err := ensureDirPermissions(path, false); err != nil {
		return doctorResult{
			Name:       name,
			OK:         false,
			Details:    err.Error(),
			Suggestion: fmt.Sprintf("grant write and execute permissions (example: chmod u+rwx %q)", path),
		}
	}

	return doctorResult{
		Name:    name,
		OK:      true,
		Details: "exists and is writable/searchable",
	}
}

func ensureDirPermissions(path string, requireRead bool) error {
	info, err := os.Stat(path)
	if err != nil {
		return fmt.Errorf("cannot inspect permissions: %w", err)
	}

	mode := info.Mode().Perm()
	hasRead := mode&0o444 != 0
	hasWrite := mode&0o222 != 0
	hasExec := mode&0o111 != 0

	missing := make([]string, 0, 3)
	if requireRead && !hasRead {
		missing = append(missing, "read")
	}
	if !hasWrite {
		missing = append(missing, "write")
	}
	if !hasExec {
		missing = append(missing, "execute")
	}
	if len(missing) > 0 {
		return fmt.Errorf("missing %s permission", strings.Join(missing, "/"))
	}
	return nil
}
