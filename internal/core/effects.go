package core

type Effect interface {
	isEffect()
}

type EffScanDirs struct {
	Roots []string
}

func (EffScanDirs) isEffect() {}

type EffMkdirAll struct {
	Path string
}

func (EffMkdirAll) isEffect() {}

type EffOpenSession struct {
	Spec SessionSpec
}

func (EffOpenSession) isEffect() {}

type EffQuit struct{}

func (EffQuit) isEffect() {}

type EffLoadWorktrees struct {
	ProjectPath string
}

func (EffLoadWorktrees) isEffect() {}

type EffCreateWorktree struct {
	ProjectPath string
	BranchName  string
}

func (EffCreateWorktree) isEffect() {}
