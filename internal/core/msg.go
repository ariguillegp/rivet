package core

type Msg interface {
	isMsg()
}

type MsgScanCompleted struct {
	Dirs []DirEntry
	Err  error
}

func (MsgScanCompleted) isMsg() {}

type MsgCreateDirCompleted struct {
	Path string
	Err  error
}

func (MsgCreateDirCompleted) isMsg() {}

type MsgKeyPress struct {
	Key string
}

func (MsgKeyPress) isMsg() {}

type MsgQueryChanged struct {
	Query string
}

func (MsgQueryChanged) isMsg() {}

type MsgWorktreesLoaded struct {
	Worktrees []Worktree
	Err       error
}

func (MsgWorktreesLoaded) isMsg() {}

type MsgWorktreeCreated struct {
	Path string
	Err  error
}

func (MsgWorktreeCreated) isMsg() {}

type MsgWorktreeQueryChanged struct {
	Query string
}

func (MsgWorktreeQueryChanged) isMsg() {}
