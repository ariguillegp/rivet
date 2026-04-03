package main

import (
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/ariguillegp/rivet/internal/adapters"
	"github.com/ariguillegp/rivet/internal/core"
	"github.com/ariguillegp/rivet/internal/ports"
	"github.com/ariguillegp/rivet/internal/ui"
)

func main() {
	var projectFlag string
	var worktreeFlag string
	var toolFlag string
	var doctorFlag bool
	var createProjectFlag bool
	var detachFlag bool
	flag.StringVar(&projectFlag, "project", "", "Project container name or path")
	flag.StringVar(&worktreeFlag, "worktree", "", "Worktree name or path")
	flag.StringVar(&toolFlag, "tool", "", "Tool to run (opencode, amp, claude, codex, or none)")
	flag.BoolVar(&doctorFlag, "doctor", false, "Run environment diagnostics and print fixes")
	flag.BoolVar(&createProjectFlag, "create-project", false, "Create the project container if missing")
	flag.BoolVar(&detachFlag, "detach", false, "Create the tmux session without attaching")
	flag.Parse()

	roots := flag.Args()
	if len(roots) == 0 {
		roots = []string{"~/Projects"}
	}
	roots = expandRoots(roots)

	if doctorFlag {
		healthy := runDoctor(roots, os.Stdout)
		if healthy {
			return
		}
		os.Exit(1)
	}

	fs := adapters.NewOSFilesystem()
	sessions := adapters.NewTmuxSession()
	availableTools := availableToolsForUI()

	if projectFlag != "" || worktreeFlag != "" || toolFlag != "" || createProjectFlag || detachFlag {
		spec, err := resolveSessionSpec(fs, roots, projectFlag, worktreeFlag, toolFlag, createProjectFlag, detachFlag)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
		if spec.Tool != core.ToolNone {
			if _, err := doctorLookPath(spec.Tool); err != nil {
				fmt.Fprintf(os.Stderr, "Error: tool %q is not installed or not in PATH\n", spec.Tool)
				os.Exit(1)
			}
		}
		if err := sessions.OpenSession(spec); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
		return
	}

	m := ui.NewWithTools(roots, availableTools, fs, sessions)
	p := tea.NewProgram(m, tea.WithAltScreen())

	result, err := p.Run()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	final := result.(ui.Model)
	if final.SelectedSpec != nil {
		if err := resetTerminal(); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
		if err := sessions.OpenSession(*final.SelectedSpec); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
		return
	}
	if final.SelectedSessionName != "" {
		if err := resetTerminal(); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
		if err := sessions.AttachSession(final.SelectedSessionName); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
		return
	}
	_ = returnToPreviousSession()
}

type doctorResult struct {
	Name       string
	OK         bool
	Details    string
	Suggestion string
}

func runDoctor(roots []string, out io.Writer) bool {
	results := diagnoseEnvironment(roots)
	failing := 0

	fmt.Fprintln(out, "rivet doctor report")
	fmt.Fprintln(out, "==================")
	for _, result := range results {
		status := "OK"
		if !result.OK {
			status = "FAIL"
			failing++
		}
		fmt.Fprintf(out, "- [%s] %s: %s\n", status, result.Name, result.Details)
		if !result.OK && strings.TrimSpace(result.Suggestion) != "" {
			fmt.Fprintf(out, "  fix: %s\n", result.Suggestion)
		}
	}
	if failing == 0 {
		fmt.Fprintln(out, "\nDoctor summary: all required checks passed.")
		return true
	}
	fmt.Fprintf(out, "\nDoctor summary: %d required check(s) failed.\n", failing)
	return false
}

func resetTerminal() error {
	stty := exec.Command("stty", "sane")
	stty.Stdin = os.Stdin
	stty.Stdout = os.Stdout
	stty.Stderr = os.Stderr
	return stty.Run()
}

func returnToPreviousSession() error {
	if os.Getenv("TMUX") == "" {
		return nil
	}
	cmd := exec.Command("tmux", "switch-client", "-l")
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func resolveSessionSpec(fs ports.Filesystem, roots []string, project, worktree, tool string, createProject, detach bool) (core.SessionSpec, error) {
	if project == "" {
		return core.SessionSpec{}, errors.New("--project is required")
	}
	if worktree == "" {
		return core.SessionSpec{}, errors.New("--worktree is required")
	}
	if tool == "" {
		return core.SessionSpec{}, errors.New("--tool is required")
	}
	if !core.IsSupportedTool(tool) {
		return core.SessionSpec{}, fmt.Errorf("unsupported tool: %s", tool)
	}

	projectPath, err := resolveProjectPath(fs, roots, project, createProject)
	if err != nil {
		return core.SessionSpec{}, err
	}

	worktreePath, err := resolveWorktreePath(fs, projectPath, worktree)
	if err != nil {
		return core.SessionSpec{}, err
	}

	return core.SessionSpec{DirPath: worktreePath, Tool: tool, Detach: detach}, nil
}

func resolveProjectPath(fs ports.Filesystem, roots []string, project string, createProject bool) (string, error) {
	if looksLikePath(project) {
		path := expandPath(project)
		if path == "" {
			return "", fmt.Errorf("invalid project path")
		}
		if !filepath.IsAbs(path) {
			absPath, err := filepath.Abs(path)
			if err != nil {
				return "", fmt.Errorf("cannot resolve path: %w", err)
			}
			path = absPath
		}
		if !exists(path) {
			if createProject {
				createdPath, err := fs.CreateProject(path)
				if err != nil {
					return "", err
				}
				return createdPath, nil
			}
			return "", fmt.Errorf("project not found: %s", project)
		}
		return path, nil
	}

	for _, root := range roots {
		candidate := filepath.Join(expandPath(root), project)
		if exists(candidate) {
			return candidate, nil
		}
	}

	if createProject {
		if len(roots) == 0 {
			return "", fmt.Errorf("no roots available to create project")
		}
		candidate := filepath.Join(expandPath(roots[0]), project)
		createdPath, err := fs.CreateProject(candidate)
		if err != nil {
			return "", err
		}
		return createdPath, nil
	}

	return "", fmt.Errorf("project not found: %s", project)
}

func resolveWorktreePath(fs ports.Filesystem, projectPath, worktree string) (string, error) {
	if looksLikePath(worktree) {
		path := expandPath(worktree)
		if !filepath.IsAbs(path) {
			absPath, err := filepath.Abs(path)
			if err != nil {
				return "", fmt.Errorf("cannot resolve path: %w", err)
			}
			path = absPath
		}
		if exists(path) {
			return path, nil
		}
		return "", fmt.Errorf("worktree not found: %s", worktree)
	}

	listing, err := fs.ListWorktrees(projectPath)
	if err != nil {
		return "", err
	}
	if listing.Warning != "" {
		return "", fmt.Errorf("%s", listing.Warning)
	}

	for _, wt := range listing.Worktrees {
		if wt.Branch == worktree || wt.Name == worktree {
			return wt.Path, nil
		}
	}

	return fs.CreateWorktree(projectPath, worktree)
}

func expandRoots(roots []string) []string {
	if len(roots) == 0 {
		return roots
	}
	expanded := make([]string, 0, len(roots))
	for _, root := range roots {
		expandedRoot := expandPath(root)
		if expandedRoot == "" {
			expandedRoot = root
		}
		expanded = append(expanded, expandedRoot)
	}
	return expanded
}

func expandPath(path string) string {
	if strings.HasPrefix(path, "~/") {
		home, _ := os.UserHomeDir()
		if home == "" {
			return ""
		}
		return filepath.Join(home, path[2:])
	}
	return path
}

func looksLikePath(s string) bool {
	return filepath.IsAbs(s) ||
		strings.HasPrefix(s, "~/") ||
		strings.HasPrefix(s, "./") ||
		strings.HasPrefix(s, "../") ||
		strings.Contains(s, string(filepath.Separator))
}

func exists(path string) bool {
	info, err := os.Stat(path)
	if err != nil {
		return false
	}
	return info.IsDir()
}
