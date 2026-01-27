package core

func Update(m Model, msg Msg) (Model, []Effect) {
	switch msg := msg.(type) {
	case MsgScanCompleted:
		if msg.Err != nil {
			m.Mode = ModeError
			m.Err = msg.Err
			return m, nil
		}
		m.Mode = ModeBrowsing
		m.Dirs = msg.Dirs
		m.Filtered = FilterDirs(m.Dirs, m.Query)
		m.SelectedIdx = 0
		return m, nil

	case MsgQueryChanged:
		m.Query = msg.Query
		m.Filtered = FilterDirs(m.Dirs, m.Query)
		m.SelectedIdx = 0
		return m, nil

	case MsgKeyPress:
		return handleKey(m, msg.Key)

	case MsgCreateDirCompleted:
		if msg.Err != nil {
			m.Mode = ModeError
			m.Err = msg.Err
			return m, nil
		}
		m.SelectedProject = msg.Path
		m.Mode = ModeWorktree
		m.WorktreeQuery = ""
		m.WorktreeIdx = 0
		return m, []Effect{EffLoadWorktrees{ProjectPath: msg.Path}}

	case MsgWorktreesLoaded:
		if msg.Err != nil {
			m.Mode = ModeError
			m.Err = msg.Err
			return m, nil
		}
		if len(msg.Worktrees) == 0 {
			spec := SessionSpec{
				Backend: m.Backend,
				DirPath: m.SelectedProject,
			}
			return m, []Effect{EffOpenSession{Spec: spec}}
		}
		m.Worktrees = msg.Worktrees
		m.FilteredWT = FilterWorktrees(m.Worktrees, m.WorktreeQuery)
		m.WorktreeIdx = 0
		return m, nil

	case MsgWorktreeQueryChanged:
		m.WorktreeQuery = msg.Query
		m.FilteredWT = FilterWorktrees(m.Worktrees, m.WorktreeQuery)
		m.WorktreeIdx = 0
		return m, nil

	case MsgWorktreeCreated:
		if msg.Err != nil {
			m.Mode = ModeError
			m.Err = msg.Err
			return m, nil
		}
		spec := SessionSpec{
			Backend: m.Backend,
			DirPath: msg.Path,
		}
		return m, []Effect{EffOpenSession{Spec: spec}}
	}

	return m, nil
}

func handleKey(m Model, key string) (Model, []Effect) {
	switch m.Mode {
	case ModeBrowsing:
		return handleBrowsingKey(m, key)
	case ModeCreateDir:
		return handleCreateDirKey(m, key)
	case ModeWorktree:
		return handleWorktreeKey(m, key)
	}
	return m, nil
}

func handleBrowsingKey(m Model, key string) (Model, []Effect) {
	switch key {
	case "up", "ctrl+k":
		if m.SelectedIdx > 0 {
			m.SelectedIdx--
		}
	case "down", "ctrl+j":
		if m.SelectedIdx < len(m.Filtered)-1 {
			m.SelectedIdx++
		}
	case "enter":
		if dir, ok := m.SelectedDir(); ok {
			m.SelectedProject = dir.Path
			m.Mode = ModeWorktree
			m.WorktreeQuery = ""
			m.WorktreeIdx = 0
			return m, []Effect{EffLoadWorktrees{ProjectPath: dir.Path}}
		}
		if m.Query != "" && len(m.RootPaths) > 0 {
			path := m.RootPaths[0] + "/" + m.Query
			return m, []Effect{EffMkdirAll{Path: path}}
		}
	case "ctrl+n":
		if m.Query != "" && len(m.RootPaths) > 0 {
			path := m.RootPaths[0] + "/" + m.Query
			return m, []Effect{EffMkdirAll{Path: path}}
		}
	case "esc", "ctrl+c":
		return m, []Effect{EffQuit{}}
	}
	return m, nil
}

func handleCreateDirKey(m Model, key string) (Model, []Effect) {
	switch key {
	case "enter":
		if m.Query != "" {
			return m, []Effect{EffMkdirAll{Path: m.Query}}
		}
	case "esc":
		m.Mode = ModeBrowsing
	}
	return m, nil
}

func handleWorktreeKey(m Model, key string) (Model, []Effect) {
	switch key {
	case "up", "ctrl+k":
		if m.WorktreeIdx > 0 {
			m.WorktreeIdx--
		}
	case "down", "ctrl+j":
		if m.WorktreeIdx < len(m.FilteredWT)-1 {
			m.WorktreeIdx++
		}
	case "enter":
		if wt, ok := m.SelectedWorktree(); ok {
			if wt.IsNew {
				return m, []Effect{EffCreateWorktree{
					ProjectPath: m.SelectedProject,
					BranchName:  m.WorktreeQuery,
				}}
			}
			spec := SessionSpec{
				Backend: m.Backend,
				DirPath: wt.Path,
			}
			return m, []Effect{EffOpenSession{Spec: spec}}
		}
		if m.WorktreeQuery != "" {
			return m, []Effect{EffCreateWorktree{
				ProjectPath: m.SelectedProject,
				BranchName:  m.WorktreeQuery,
			}}
		}
	case "ctrl+n":
		if m.WorktreeQuery != "" {
			return m, []Effect{EffCreateWorktree{
				ProjectPath: m.SelectedProject,
				BranchName:  m.WorktreeQuery,
			}}
		}
	case "esc":
		m.Mode = ModeBrowsing
		m.WorktreeQuery = ""
		m.Worktrees = nil
		m.FilteredWT = nil
		m.WorktreeIdx = 0
	case "ctrl+c":
		return m, []Effect{EffQuit{}}
	}
	return m, nil
}

func Init(m Model) (Model, []Effect) {
	return m, []Effect{EffScanDirs{Roots: m.RootPaths}}
}
