# cignore

CLI tool to manage Claude Code related files in `.gitignore` with an interactive fuzzy finder interface.

## Features

- Interactive fuzzy finder to select Claude Code patterns (CLAUDE.md, .claude/, etc.)
- Toggle patterns on/off with Tab, apply with Enter
- Automatically manages a dedicated section in `.gitignore`
- Preserves existing `.gitignore` content

## Installation

```bash
go install github.com/rnakamine/cignore@latest
```

## Usage

Run inside a git repository:

```bash
cignore
```

1. A fuzzy finder displays Claude Code related patterns
2. Use Tab to toggle patterns for `.gitignore` inclusion
3. Press Enter to apply changes, Esc to cancel
4. Changes are written to `.gitignore` automatically

## Default Patterns

| Pattern | Description |
|---------|-------------|
| `CLAUDE.md` | Claude Code project information file |
| `.claude/` | Claude Code configuration directory |
| `.claude/*` | All files in Claude Code config directory |
| `claude.json` | Claude Code configuration file |
