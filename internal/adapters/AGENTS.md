# Adapters (I/O)

External integrations for filesystem, git, and tmux.

## Responsibilities
- Implement `ports.Filesystem` and `ports.SessionManager`.
- All I/O lives here (filesystem, git, tmux).
- Keep behavior consistent with core expectations.

## Key Files
- `filesystem_os.go`: scan dirs, manage git worktrees, create projects.
- `sessions_tmux.go`: manage tmux workspace sessions and window layout.

## Rules
- Do not import `internal/ui`; keep `internal/core` usage limited to shared domain types/helpers.
- Prefer small helpers over long command chains.
