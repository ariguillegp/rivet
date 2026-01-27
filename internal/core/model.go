package core

type Mode int

const (
	ModeLoading Mode = iota
	ModeBrowsing
	ModeCreateDir
	ModeWorktree
	ModeError
)

type Model struct {
	Mode            Mode
	Query           string
	Dirs            []DirEntry
	Filtered        []DirEntry
	SelectedIdx     int
	RootPaths       []string
	Backend         SessionBackend
	Err             error
	SelectedProject string
	Worktrees       []Worktree
	FilteredWT      []Worktree
	WorktreeIdx     int
	WorktreeQuery   string
}

func NewModel(roots []string) Model {
	return Model{
		Mode:      ModeLoading,
		RootPaths: roots,
		Backend:   BackendNative,
	}
}

func (m Model) SelectedDir() (DirEntry, bool) {
	if len(m.Filtered) == 0 || m.SelectedIdx >= len(m.Filtered) {
		return DirEntry{}, false
	}
	return m.Filtered[m.SelectedIdx], true
}

func (m Model) SelectedWorktree() (Worktree, bool) {
	if len(m.FilteredWT) == 0 || m.WorktreeIdx >= len(m.FilteredWT) {
		return Worktree{}, false
	}
	return m.FilteredWT[m.WorktreeIdx], true
}
