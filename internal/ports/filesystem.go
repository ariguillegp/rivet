package ports

import "github.com/ariguillegp/solo/internal/core"

type Filesystem interface {
	ScanDirs(roots []string, maxDepth int) ([]core.DirEntry, error)
	MkdirAll(path string) (string, error)
	ListWorktrees(projectPath string) ([]core.Worktree, error)
	CreateWorktree(projectPath, branchName string) (string, error)
}
