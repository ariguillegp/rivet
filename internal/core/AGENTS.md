# Core (state machine)

Pure business logic and state transitions. No I/O.

## Responsibilities
- Modes, messages, and effects for the TUI flow.
- Fuzzy filtering for projects, worktrees, and tools.
- Name sanitization and tool validation.

## Key Files
- `model.go`: core model shape and default state.
- `update.go`: state transitions and effect emission.
- `effects.go`: effect types only (no execution).
- `filter.go`: fuzzy matching helpers.
- `types.go`: shared domain types.

## Rules
- Keep logic deterministic and testable.
- Effects describe side effects; execution belongs in `internal/ui`.
