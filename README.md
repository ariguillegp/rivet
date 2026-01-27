# solo

A fast terminal-based project directory switcher with fuzzy filtering and git worktree support.

## Features

- Fuzzy search across project directories (only shows git repositories)
- Git worktree integration: select or create worktrees within projects
- Create new directories on the fly with `ctrl+n`
- Create new git worktrees by typing a branch name
- Keyboard-driven navigation
- Opens a new shell in the selected directory/worktree

## Installation

```bash
go install github.com/ariguillegp/solo/cmd/solo@latest
```

## Usage

```bash
solo [directories...]
```

By default, solo scans `~/Work/tries`. Pass custom directories as arguments:

```bash
solo ~/projects ~/work
```

### Shell Integration

Add to your `.bashrc` or `.zshrc` to enable `ctrl+f` keybinding:

```bash
source /path/to/solo.sh
```

### Workflow

1. Launch solo and type to fuzzy-filter git repositories
2. Press `enter` to select a repository
3. Choose an existing worktree or type a branch name to create a new one
4. Press `enter` to open a shell in the selected worktree

### Keybindings

#### Project Selection

| Key | Action |
|-----|--------|
| `↑` / `ctrl+k` | Previous suggestion |
| `↓` / `ctrl+j` | Next suggestion |
| `enter` | Select project (go to worktree selection) |
| `ctrl+n` | Create new directory |
| `esc` / `ctrl+c` | Quit |

#### Worktree Selection

| Key | Action |
|-----|--------|
| `↑` / `ctrl+k` | Previous worktree |
| `↓` / `ctrl+j` | Next worktree |
| `enter` | Select worktree / create new if typing |
| `ctrl+n` | Create new worktree with typed branch name |
| `esc` | Go back to project selection |
| `ctrl+c` | Quit |

## License

MIT
