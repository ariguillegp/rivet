package adapters

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/ariguillegp/rivet/internal/core"
)

type TmuxSession struct{}

const lazygitWindowName = "lazygit"

func NewTmuxSession() *TmuxSession {
	return &TmuxSession{}
}

func (t *TmuxSession) OpenSession(spec core.SessionSpec) error {
	sessionName, err := sessionNameFor(spec)
	if err != nil {
		return err
	}

	if _, err := ensureWorkspaceLayout(sessionName, spec.DirPath, spec.Tool); err != nil {
		return err
	}

	if err := selectWindow(sessionName, spec.Tool); err != nil {
		return err
	}

	if spec.Detach {
		return nil
	}

	if os.Getenv("TMUX") != "" {
		return switchClient(sessionName)
	}

	cmd := exec.Command("tmux", "attach-session", "-t", tmuxSessionTarget(sessionName))
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func (t *TmuxSession) PrewarmSession(spec core.SessionSpec) (bool, error) {
	sessionName, err := sessionNameFor(spec)
	if err != nil {
		return false, err
	}
	return ensureWorkspaceLayout(sessionName, spec.DirPath, spec.Tool)
}

func (t *TmuxSession) KillSession(spec core.SessionSpec) error {
	sessionName, err := sessionNameFor(spec)
	if err != nil {
		return err
	}

	cmd := exec.Command("tmux", "kill-session", "-t", tmuxSessionTarget(sessionName))
	if output, err := cmd.CombinedOutput(); err != nil {
		outputText := string(output)
		if strings.Contains(outputText, "can't find session") || isTmuxServerUnavailable(outputText) {
			return nil
		}
		return fmt.Errorf("failed to kill tmux session: %w (output: %s)", err, string(output))
	}
	return nil
}

func (t *TmuxSession) ListSessions() ([]core.SessionInfo, error) {
	cmd := exec.Command("tmux", "list-sessions", "-F", "#{session_name}\t#{session_path}\t#{session_last_attached}")
	output, err := cmd.CombinedOutput()
	if err != nil {
		if isTmuxServerUnavailable(string(output)) {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to list tmux sessions: %w (output: %s)", err, strings.TrimSpace(string(output)))
	}

	lines := strings.Split(strings.TrimSpace(string(output)), "\n")
	if len(lines) == 1 && strings.TrimSpace(lines[0]) == "" {
		return nil, nil
	}

	var sessions []core.SessionInfo
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		parts := strings.SplitN(line, "\t", 3)
		name := strings.TrimSpace(parts[0])
		if name == "" {
			continue
		}
		if name == "rv-launcher" || name == "rivet-launcher" {
			continue
		}
		sessionPath := ""
		if len(parts) > 1 {
			sessionPath = strings.TrimSpace(parts[1])
		}
		lastActive := parseTmuxUnixTime(parts, 2)
		info, ok := parseSessionName(name)
		if ok {
			info.Name = name
			if sessionPath != "" {
				info.DirPath = sessionPath
			}
		} else {
			info = core.SessionInfo{Name: name, DirPath: sessionPath}
		}
		info.LastActive = lastActive
		projectName, branch := sessionMetadata(sessionPath)
		if strings.TrimSpace(projectName) != "" {
			info.Project = projectName
		}
		if strings.TrimSpace(branch) != "" {
			info.Branch = branch
		}
		sessions = append(sessions, info)
	}

	return sessions, nil
}

func isTmuxServerUnavailable(output string) bool {
	output = strings.ToLower(strings.TrimSpace(output))
	if output == "" {
		return false
	}
	return strings.Contains(output, "no server running") ||
		strings.Contains(output, "failed to connect to server") ||
		(strings.Contains(output, "error connecting to") && strings.Contains(output, "no such file or directory"))
}

func parseTmuxUnixTime(parts []string, idx int) time.Time {
	if len(parts) <= idx {
		return time.Time{}
	}
	unixText := strings.TrimSpace(parts[idx])
	if unixText == "" || unixText == "0" {
		return time.Time{}
	}
	unixValue, err := strconv.ParseInt(unixText, 10, 64)
	if err != nil || unixValue <= 0 {
		return time.Time{}
	}
	return time.Unix(unixValue, 0)
}

func sessionMetadata(dirPath string) (projectName, branch string) {
	dirPath = strings.TrimSpace(dirPath)
	if dirPath == "" {
		return "", ""
	}
	projectName = sessionProjectName(dirPath)
	branch = gitOutput("-C", dirPath, "branch", "--show-current")
	return strings.TrimSpace(projectName), strings.TrimSpace(branch)
}

func sessionProjectPath(dirPath string) string {
	commonDir := gitOutput("-C", dirPath, "rev-parse", "--git-common-dir")
	if commonDir != "" {
		if !filepath.IsAbs(commonDir) {
			commonDir = filepath.Clean(filepath.Join(dirPath, commonDir))
		}
		if strings.EqualFold(filepath.Base(commonDir), ".git") {
			projectPath := filepath.Dir(commonDir)
			if projectPath != "" && projectPath != "." && projectPath != string(filepath.Separator) {
				return projectPath
			}
		}
	}

	topLevel := gitOutput("-C", dirPath, "rev-parse", "--show-toplevel")
	if topLevel == "" {
		return ""
	}
	return topLevel
}

func sessionProjectName(dirPath string) string {
	topLevel := sessionProjectPath(dirPath)
	if topLevel == "" {
		topLevel = dirPath
	}
	name := strings.TrimSpace(filepath.Base(filepath.Clean(topLevel)))
	if name == "" || name == "." || name == string(filepath.Separator) {
		return ""
	}
	return name
}

func gitOutput(args ...string) string {
	cmd := exec.Command("git", args...)
	output, err := cmd.Output()
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(output))
}

func (t *TmuxSession) AttachSession(name string) error {
	name = strings.TrimSpace(name)
	if name == "" {
		return fmt.Errorf("session name is required")
	}
	if os.Getenv("TMUX") != "" {
		return switchClient(name)
	}
	cmd := exec.Command("tmux", "attach-session", "-t", tmuxSessionTarget(name))
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

var sessionNamePattern = regexp.MustCompile(`[^a-zA-Z0-9_-]+`)

func sanitizeSessionPart(name, fallback string) string {
	name = strings.TrimSpace(name)
	if name == "" {
		return fallback
	}
	name = sessionNamePattern.ReplaceAllString(name, "-")
	if name == "" {
		return fallback
	}
	return name
}

func sessionNameFor(spec core.SessionSpec) (string, error) {
	if strings.TrimSpace(spec.DirPath) == "" {
		return "", fmt.Errorf("session directory is required")
	}

	cleanPath := filepath.Clean(spec.DirPath)
	name := sessionProjectName(cleanPath)
	if strings.TrimSpace(name) == "" {
		name = filepath.Base(cleanPath)
	}
	if name == "." || name == string(filepath.Separator) || strings.TrimSpace(name) == "" {
		name = "worktree"
	}
	name = strings.Trim(sanitizeSessionPart(name, "worktree"), "-")
	if name == "" {
		name = "worktree"
	}
	return name + "-" + sessionPathHash(cleanPath), nil
}

func ensureWorkspaceLayout(sessionName, dirPath, selectedTool string) (bool, error) {
	createdLazygit, err := ensureSessionWindow(sessionName, dirPath, lazygitWindowName)
	if err != nil {
		return false, err
	}
	createdTool, err := ensureSessionWindow(sessionName, dirPath, selectedTool)
	if err != nil {
		return false, err
	}
	if createdLazygit && createdTool {
		if err := initializeWorkspaceWindows(sessionName, selectedTool); err != nil {
			return false, err
		}
	}
	return createdTool, nil
}

func ensureSessionWindow(sessionName, dirPath, windowName string) (bool, error) {
	windowName = strings.TrimSpace(windowName)
	if windowName == "" {
		return false, fmt.Errorf("session window is required")
	}

	sessionExists := hasSession(sessionName)

	if !sessionExists {
		if err := createSessionWithWindow(sessionName, dirPath, windowName); err != nil {
			if !isTmuxDuplicateSessionError(err) {
				return false, err
			}
		} else {
			return true, nil
		}
	}

	if hasWindow(sessionName, windowName) {
		return false, nil
	}

	if err := createWindow(sessionName, dirPath, windowName); err != nil {
		if isTmuxDuplicateWindowError(err) {
			return false, nil
		}
		return false, err
	}

	return true, nil
}

func hasSession(sessionName string) bool {
	check := exec.Command("tmux", "has-session", "-t", tmuxSessionTarget(sessionName))
	return check.Run() == nil
}

func isTmuxDuplicateSessionError(err error) bool {
	return strings.Contains(strings.ToLower(err.Error()), "duplicate session")
}

func isTmuxDuplicateWindowError(err error) bool {
	return strings.Contains(strings.ToLower(err.Error()), "duplicate window")
}

func createSessionWithWindow(sessionName, dirPath, windowName string) error {
	shell, commandArgs := toolCommand(windowName)
	args := []string{"new-session", "-d", "-s", sessionName}
	args = append(args, tmuxEnvArgs(windowName)...)
	args = append(args, "-n", windowName, "-c", dirPath, shell)
	args = append(args, commandArgs...)
	cmd := exec.Command("tmux", args...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to create tmux session: %w (output: %s)", err, string(output))
	}
	return nil
}

func hasWindow(sessionName, windowName string) bool {
	check := exec.Command("tmux", "list-windows", "-t", tmuxSessionTarget(sessionName), "-F", "#{window_name}")
	output, err := check.Output()
	if err != nil {
		return false
	}
	for line := range strings.SplitSeq(string(output), "\n") {
		if strings.TrimSpace(line) == windowName {
			return true
		}
	}
	return false
}

func createWindow(sessionName, dirPath, windowName string) error {
	shell, commandArgs := toolCommand(windowName)
	args := []string{"new-window", "-d", "-t", tmuxSessionTarget(sessionName), "-n", windowName}
	args = append(args, tmuxEnvArgs(windowName)...)
	args = append(args, "-c", dirPath, shell)
	args = append(args, commandArgs...)
	cmd := exec.Command("tmux", args...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to create tmux window: %w (output: %s)", err, string(output))
	}
	return nil
}

func selectWindow(sessionName, tool string) error {
	cmd := exec.Command("tmux", "select-window", "-t", tmuxSessionTarget(sessionName)+":"+tool)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to select tmux window: %w (output: %s)", err, string(output))
	}
	return nil
}

// initializeWorkspaceWindows moves lazygit and the selected tool to indices 1
// and 2. Tmux disallows moves that collide with existing indices, so both
// windows are parked at high temporary indices first to free the low slots.
func initializeWorkspaceWindows(sessionName, selectedTool string) error {
	if err := moveWindow(sessionName, lazygitWindowName, 900); err != nil {
		return err
	}
	if err := moveWindow(sessionName, selectedTool, 901); err != nil {
		return err
	}
	if err := moveWindow(sessionName, lazygitWindowName, 1); err != nil {
		return err
	}
	return moveWindow(sessionName, selectedTool, 2)
}

func moveWindow(sessionName, windowName string, index int) error {
	cmd := exec.Command(
		"tmux",
		"move-window",
		"-s", tmuxSessionTarget(sessionName)+":"+windowName,
		"-t", tmuxSessionTarget(sessionName)+":"+strconv.Itoa(index),
	)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to move tmux window: %w (output: %s)", err, string(output))
	}
	return nil
}

func switchClient(sessionName string) error {
	args := []string{"switch-client", "-t", tmuxSessionTarget(sessionName)}
	client := currentClientTTY()
	if client != "" {
		args = []string{"switch-client", "-c", client, "-t", tmuxSessionTarget(sessionName)}
	}
	cmd := exec.Command("tmux", args...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil && client != "" {
		fallback := exec.Command("tmux", "switch-client", "-t", tmuxSessionTarget(sessionName))
		fallback.Stdin = os.Stdin
		fallback.Stdout = os.Stdout
		fallback.Stderr = os.Stderr
		return fallback.Run()
	}
	return nil
}

func tmuxSessionTarget(sessionName string) string {
	return "=" + sessionName
}

func currentClientTTY() string {
	pane := os.Getenv("TMUX_PANE")
	if pane == "" {
		return ""
	}
	cmd := exec.Command("tmux", "display-message", "-p", "-t", pane, "#{client_tty}")
	output, err := cmd.Output()
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(output))
}

func toolCommand(tool string) (shell string, args []string) {
	shell = os.Getenv("SHELL")
	if strings.TrimSpace(shell) == "" {
		shell = "/bin/sh"
	}
	if tool != lazygitWindowName && !core.ToolNeedsWarmup(tool) {
		return shell, nil
	}
	return shell, []string{"-c", `"$1"; exec "$0"`, shell, tool}
}

func tmuxEnvArgs(tool string) []string {
	keys := []string{
		"COLORFGBG",
		"COLORTERM",
		"TERM",
		"TERM_PROGRAM",
		"TERM_PROGRAM_VERSION",
	}
	args := make([]string, 0, len(keys)*2+2)

	for _, env := range core.ToolEnv(tool) {
		args = append(args, "-e", env)
	}

	for _, key := range keys {
		val := strings.TrimSpace(os.Getenv(key))
		if val == "" {
			continue
		}
		args = append(args, "-e", key+"="+val)
	}

	return args
}

func parseSessionName(name string) (core.SessionInfo, bool) {
	return core.SessionInfo{}, false
}
