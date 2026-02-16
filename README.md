# Picky Claude

A free, open-source quality layer for Claude Code. Single binary, zero runtime dependencies — quality-enforced, context-managed, spec-driven development.

## What It Does

Picky Claude wraps Claude Code with:

- **Quality hooks** — automatic linting, formatting, and TDD enforcement on every file write
- **Context management** — monitors context usage and seamlessly continues sessions when limits are reached (Endless Mode)
- **Persistent memory** — SQLite-backed observation store with full-text and semantic search, accessible via MCP
- **Spec-driven development** — structured plan → implement → verify workflow
- **Git worktree isolation** — develop in isolated worktrees, squash-merge when done
- **Web viewer** — real-time observation stream at `http://localhost:41777`

### Why Picky Claude?

Picky Claude compiles to a **single Go binary** with all assets embedded. No interpreters, no `npm install`, no virtual environments. Hooks are compiled Go code — sub-5ms execution.

## Prerequisites

- [Claude Code](https://docs.anthropic.com/en/docs/claude-code) installed and configured
- [Git](https://git-scm.com/) 2.20+
- Go 1.25+ (only for building from source)

## Installation

### From Source

```bash
git clone https://github.com/jesperpedersen/picky-claude.git
cd picky-claude
make build
```

The binary is at `bin/picky`. Add it to your PATH or copy it somewhere in your PATH:

```bash
cp bin/picky /usr/local/bin/
```

### Project Setup

In any project directory, run the installer:

```bash
picky install
```

This sets up:
- `.claude/` directory with rules, commands, agents, and hooks configuration
- Shell aliases in your `.zshrc`/`.bashrc`
- VS Code extension recommendations
- MCP server configuration for persistent memory

Use `--skip-prereqs` or `--skip-deps` to skip specific steps.

## Quick Start

```bash
# 1. Install into your project
cd your-project
picky install

# 2. Launch Claude Code with hooks and Endless Mode
picky run

# 3. Or start the console server standalone (for debugging)
picky serve
```

When launched via `picky run`, Claude Code gets:
- All quality hooks active (file checking, TDD enforcement, context monitoring)
- A console server running on port 41777 with persistent memory
- Automatic session management and Endless Mode continuation

## CLI Commands

| Command | Description |
|---------|-------------|
| `picky run` | Launch Claude Code with hooks and Endless Mode |
| `picky install` | Set up project with rules, hooks, and configuration |
| `picky serve` | Start the console server standalone |
| `picky hook <name>` | Run a specific hook (called by Claude Code, not directly) |
| `picky greet` | Print the welcome banner |
| `picky check-context` | Get current context usage percentage |
| `picky send-clear [plan]` | Trigger Endless Mode session restart |
| `picky register-plan <path> <status>` | Associate a plan file with the current session |
| `picky session list` | List active sessions |
| `picky statusline` | Format the status bar (reads JSON from stdin) |
| `picky worktree <subcommand>` | Git worktree management (create, detect, diff, sync, cleanup, status) |

All commands support `--json` for structured output.

## Hooks

Hooks run as subcommands of the same binary, invoked automatically by Claude Code's hook system:

| Hook | Trigger | Description |
|------|---------|-------------|
| `file-checker` | PostToolUse (Write/Edit) | Runs language-specific linter/formatter |
| `tdd-enforcer` | PostToolUse (Write/Edit) | Warns if production code is written before tests |
| `context-monitor` | PostToolUse (most tools) | Tracks context usage, triggers handoff at thresholds |
| `tool-redirect` | PreToolUse | Blocks/redirects certain tool calls (e.g., WebSearch → MCP) |
| `spec-stop-guard` | Stop | Prevents premature stop during /spec workflow |
| `spec-plan-validator` | PostToolUse | Validates plan file structure |
| `spec-verify-validator` | PostToolUse | Validates verification results |
| `notify` | Various | Desktop notifications (macOS/Linux) |

### Supported Languages

| Language | Tools | Auto-fix |
|----------|-------|----------|
| Python | ruff, basedpyright | Yes (ruff) |
| TypeScript | prettier, eslint, tsc | Yes (prettier, eslint) |
| Go | gofmt, golangci-lint | Yes (gofmt) |

## Architecture

```
picky run
    |
    v
+------------------+     HTTP      +-------------------+
|  Claude Code     |<------------>|  Console Server    |
|  (subprocess)    |               |  (goroutine)       |
+--------+---------+               |  - MCP server      |
         |                         |  - SQLite DB        |
         | hooks                   |  - SSE broadcast    |
         v                         |  - Web viewer       |
+------------------+               +-------------------+
|  picky hook *    |                       ^
|  (same binary)   |───────────────────────┘
|  - file-checker  |     POST observations
|  - tdd-enforcer  |     GET context
|  - ctx-monitor   |
+------------------+
```

Everything runs in a single process. The console server is a goroutine within `picky run`. Hooks are compiled Go code invoked as subcommands of the same binary.

## Configuration

Configuration is via environment variables:

| Variable | Default | Description |
|----------|---------|-------------|
| `PICKY_HOME` | `~/.picky` | Data directory for database, sessions, logs |
| `PICKY_PORT` | `41777` | Console server port |
| `PICKY_LOG_LEVEL` | `info` | Log level (debug, info, warn, error) |
| `PICKY_SESSION_ID` | auto-generated | Session identifier |
| `PICKY_NO_UPDATE` | — | Disable auto-update check |

## Building

```bash
# Build for current platform
make build

# Build for all platforms (macOS + Linux, arm64 + amd64)
make release

# Run tests
make test

# Run linter
make lint

# Build web viewer then compile
make all
```

## Cross-Platform

Picky Claude uses pure Go SQLite (`modernc.org/sqlite`) — no CGO required. The release builds target:

- macOS arm64 (Apple Silicon)
- macOS amd64 (Intel)
- Linux arm64
- Linux amd64

## Renaming

All branding is in `internal/config/branding.go`. To rename:

1. Change `BinaryName`, `DisplayName`, `EnvPrefix`, `ConfigDirName` in `branding.go`
2. Update the module path in `go.mod`
3. Update `BINARY_NAME` in the `Makefile`

Everything else propagates automatically.

## License

MIT
