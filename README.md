# cexclude

CLI tool to manage Claude Code related files in `.git/info/exclude` with an interactive fuzzy finder interface.

## Features

- Interactive fuzzy finder to select Claude Code patterns (CLAUDE.md, .claude/, etc.)
- Toggle patterns on/off with Tab, apply with Enter
- Automatically manages a dedicated section in `.git/info/exclude`
- Preserves existing `.git/info/exclude` content
- Dynamically scans `.claude/` directory for individual files

## Installation

```bash
go install github.com/rnakamine/cignore@latest
```

## Usage

Run inside a git repository:

```bash
cexclude
```

1. A fuzzy finder displays Claude Code related patterns (only existing files)
2. Use Tab to toggle patterns for `.git/info/exclude` inclusion
3. Press Enter to apply changes, Esc to cancel
4. Changes are written to `.git/info/exclude` automatically

## Default Patterns

| Pattern | Description |
|---------|-------------|
| `CLAUDE.md` | Claude Code project information file |
| `.claude/` | Claude Code configuration directory |
| `.claude/*` | Individual files discovered under .claude/ |
