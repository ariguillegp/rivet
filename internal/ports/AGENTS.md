# Ports (interfaces)

Interfaces to keep core/UI decoupled from I/O.

## Responsibilities
- Define filesystem and session contracts.
- Keep interfaces minimal and stable.

## Key Files
- `filesystem.go`: `Filesystem` interface.
- `sessions.go`: `SessionManager` interface.

## Rules
- Avoid concrete implementations here.
- Changing signatures requires adapter + UI updates.
