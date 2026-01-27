package adapters

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/ariguillegp/solo/internal/core"
)

type OSFilesystem struct{}

func NewOSFilesystem() *OSFilesystem {
	return &OSFilesystem{}
}

var ignoreDirs = map[string]bool{
	".git":         true,
	"node_modules": true,
	"vendor":       true,
	".cache":       true,
	"__pycache__":  true,
	".venv":        true,
	"target":       true,
}

func (f *OSFilesystem) ScanDirs(roots []string, maxDepth int) ([]core.DirEntry, error) {
	var dirs []core.DirEntry
	seen := make(map[string]bool)

	for _, root := range roots {
		root = expandPath(root)
		err := scanDir(root, 0, maxDepth, seen, &dirs)
		if err != nil {
			continue
		}
	}

	return dirs, nil
}

func scanDir(path string, depth, maxDepth int, seen map[string]bool, dirs *[]core.DirEntry) error {
	if depth > maxDepth {
		return nil
	}

	entries, err := os.ReadDir(path)
	if err != nil {
		return err
	}

	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		name := entry.Name()
		if strings.HasPrefix(name, ".") || ignoreDirs[name] {
			continue
		}

		fullPath := filepath.Join(path, name)
		if seen[fullPath] {
			continue
		}
		seen[fullPath] = true

		if isGitRepo(fullPath) {
			*dirs = append(*dirs, core.DirEntry{
				Path:   fullPath,
				Name:   name,
				Exists: true,
			})
		}

		scanDir(fullPath, depth+1, maxDepth, seen, dirs)
	}

	return nil
}

func isGitRepo(path string) bool {
	cmd := exec.Command("git", "rev-parse", "--git-dir")
	cmd.Dir = path
	return cmd.Run() == nil
}

func (f *OSFilesystem) MkdirAll(path string) (string, error) {
	path = expandPath(path)
	return path, os.MkdirAll(path, 0755)
}

func expandPath(path string) string {
	if strings.HasPrefix(path, "~/") {
		home, _ := os.UserHomeDir()
		return filepath.Join(home, path[2:])
	}
	return path
}

func (f *OSFilesystem) ListWorktrees(projectPath string) ([]core.Worktree, error) {
	projectPath = expandPath(projectPath)

	checkCmd := exec.Command("git", "rev-parse", "--git-dir")
	checkCmd.Dir = projectPath
	if err := checkCmd.Run(); err != nil {
		return nil, nil
	}

	cmd := exec.Command("git", "worktree", "list", "--porcelain")
	cmd.Dir = projectPath
	output, err := cmd.Output()
	if err != nil {
		return []core.Worktree{{
			Path:   projectPath,
			Name:   filepath.Base(projectPath),
			Branch: "main",
			IsMain: true,
		}}, nil
	}

	var worktrees []core.Worktree
	var current core.Worktree
	isFirst := true

	for _, line := range bytes.Split(output, []byte("\n")) {
		lineStr := string(line)
		switch {
		case strings.HasPrefix(lineStr, "worktree "):
			if current.Path != "" {
				current.IsMain = isFirst
				worktrees = append(worktrees, current)
				isFirst = false
			}
			current = core.Worktree{
				Path: strings.TrimPrefix(lineStr, "worktree "),
			}
		case strings.HasPrefix(lineStr, "branch refs/heads/"):
			current.Branch = strings.TrimPrefix(lineStr, "branch refs/heads/")
			current.Name = current.Branch
		case lineStr == "bare":
			current.Name = "(bare)"
		case lineStr == "detached":
			current.Name = "(detached)"
		}
	}

	if current.Path != "" {
		current.IsMain = isFirst
		worktrees = append(worktrees, current)
	}

	for i := range worktrees {
		worktrees[i].Name = filepath.Base(worktrees[i].Path)
	}

	return worktrees, nil
}

func (f *OSFilesystem) CreateWorktree(projectPath, branchName string) (string, error) {
	projectPath = expandPath(projectPath)

	parentDir := filepath.Dir(projectPath)
	mainName := filepath.Base(projectPath)

	if strings.Contains(mainName, "--") {
		parts := strings.SplitN(mainName, "--", 2)
		mainName = parts[0]
	}

	worktreePath := filepath.Join(parentDir, mainName+"--"+branchName)

	cmd := exec.Command("git", "worktree", "add", "-b", branchName, worktreePath)
	cmd.Dir = projectPath
	output, err := cmd.CombinedOutput()
	if err != nil {
		cmd = exec.Command("git", "worktree", "add", worktreePath, branchName)
		cmd.Dir = projectPath
		output2, err2 := cmd.CombinedOutput()
		if err2 != nil {
			return "", fmt.Errorf("%s: %s", err2, string(output2))
		}
		return worktreePath, nil
	}
	_ = output

	return worktreePath, nil
}
