# UI (Bubble Tea)

Presentation layer: renders state and executes effects.

## Responsibilities
- Bubble Tea model lifecycle (init/update/view).
- Run core effects via filesystem/session ports.
- Manage input fields, theme picker, and visual styles.

## Key Files
- `model.go`: UI state + effect runner.
- `view.go`: rendering per mode.
- `styles.go`: lipgloss styles.
- `theme.go`: theme definitions.

## Rules
- Do not embed business rules here; defer to `internal/core`.
- Side effects must be triggered via core effect types.
