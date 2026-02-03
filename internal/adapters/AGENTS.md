# Adapters (I/O)

External integrations for filesystem, git, and tmux.

## Responsibilities
- Implement `ports.Filesystem` and `ports.SessionManager`.
- All I/O lives here (filesystem, git, tmux).
- Keep behavior consistent with core expectations.

## Key Files
- `filesystem_os.go`: scan dirs, manage git worktrees, create projects.
- `sessions_tmux.go`: prewarm/open tmux sessions.

## Rules
- Do not import `internal/ui` or `internal/core` beyond types/interfaces.
- Prefer small helpers over long command chains.
