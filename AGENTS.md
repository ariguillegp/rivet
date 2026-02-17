# rivet AGENTS

Use progressive guidance: start here for orientation, then check `internal/*/AGENTS.md` for details.

## Entry Points
- CLI/TUI entry: `cmd/rv/main.go`
- Interactive mode: `rv [directories...]`
- Non-interactive: `rv --project NAME --worktree BRANCH --tool TOOL [--detach]`

## Architecture Map
- `internal/core`: pure state machine, effects, filtering logic
- `internal/ui`: Bubble Tea model, rendering, effect runner
- `internal/ports`: interfaces for filesystem + sessions
- `internal/adapters`: OS/git/tmux implementations

## Build & Verify
- **After every code change**, run `make -j validate` (fast gate: fmt → compile ∥ lint → test-short ∥ build).
- Always use `-j` so independent targets run concurrently while respecting dependency edges defined in the Makefile.
- Do NOT skip `make -j validate` or run individual targets in isolation — the full pipeline is designed to fail fast.
- If `make -j validate` fails, fix the failure and re-run before making further changes.
- `make -j validate-full` for thorough checks (race detector, coverage threshold, vulncheck, module tidiness) — run before finalizing a PR.
- `make deploy` to install the binary to `~/.local/bin/rv`.

## Design Guidance
- Keep business logic in `internal/core` (no I/O)
- Keep I/O in `internal/adapters` behind `internal/ports`
- UI should orchestrate effects, not implement core logic
