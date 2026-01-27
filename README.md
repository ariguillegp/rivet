# solo

A fast terminal-based project directory switcher with fuzzy filtering.

## Features

- Fuzzy search across project directories
- Create new directories on the fly with `ctrl+n`
- Keyboard-driven navigation
- Opens a new shell in the selected directory

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

### Keybindings

| Key | Action |
|-----|--------|
| `↑` / `ctrl+k` | Previous suggestion |
| `↓` / `ctrl+j` | Next suggestion |
| `enter` | Select directory |
| `ctrl+n` | Create new directory |
| `esc` / `ctrl+c` | Quit |

### Shell Integration

Source `solo.sh` in your shell config to enable `ctrl+f` to launch solo:

```bash
# .bashrc or .zshrc
source /path/to/solo.sh
```

## License

MIT
