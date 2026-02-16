# Usage Guide

This guide covers how to install, configure, and use Picky Claude for quality-enforced development with Claude Code.

## Table of Contents

- [Installation](#installation)
- [Running](#running)
- [Hook System](#hook-system)
- [Console Server](#console-server)
- [Endless Mode](#endless-mode)
- [Worktree Isolation](#worktree-isolation)
- [Persistent Memory](#persistent-memory)
- [Spec-Driven Development](#spec-driven-development)
- [Configuration](#configuration)
- [Troubleshooting](#troubleshooting)

---

## Installation

### Building from Source

```bash
git clone https://github.com/jesperpedersen/picky-claude.git
cd picky-claude
make build
```

This produces `bin/picky`. Add it to your PATH:

```bash
# Option 1: Copy to system path
cp bin/picky /usr/local/bin/

# Option 2: Add bin/ to PATH in your shell config
export PATH="$PATH:/path/to/picky-claude/bin"
```

### Cross-Platform Builds

```bash
make release
```

Produces binaries for macOS (arm64/amd64) and Linux (arm64/amd64) in `bin/`.

### Setting Up a Project

Navigate to your project and run the installer:

```bash
cd your-project
picky install
```

The installer runs these steps in order:

1. **Prerequisites** — Verifies git and Claude Code are installed
2. **Dependencies** — Installs vexor, playwright-cli, mcp-cli (optional tools)
3. **Shell Config** — Adds aliases to `.zshrc`/`.bashrc`
4. **Claude Files** — Creates `.claude/` with rules, commands, agents
5. **Config Files** — Writes `settings.json`, hooks configuration, MCP config
6. **VS Code** — Recommends useful extensions
7. **Finalize** — Verifies setup and prints summary

If any step fails, all previously completed steps are rolled back automatically.

Skip specific steps with flags:

```bash
picky install --skip-prereqs    # Skip prerequisite checks
picky install --skip-deps       # Skip dependency installation
```

---

## Running

### Standard Launch

```bash
picky run
```

This:
1. Generates a unique session ID
2. Starts the console server as a background goroutine (port 41777)
3. Launches Claude Code as a subprocess with hooks and environment configured
4. Forwards signals (SIGINT, SIGTERM) to Claude Code
5. Cleans up when Claude Code exits

All Claude Code arguments are passed through:

```bash
picky run --model opus     # Pass flags to Claude Code
picky run -p "fix the bug" # Pass a prompt
```

### Console Server Only

For debugging or development, run the console server standalone:

```bash
picky serve
```

The server runs on `http://localhost:41777` (configurable via `PICKY_PORT`).

---

## Hook System

Hooks are quality gates that run automatically during Claude Code sessions. They are configured in `.claude/settings.json` (or `hooks.json`) and call back into the `picky` binary.

### How Hooks Work

1. Claude Code triggers a hook event (e.g., a file was written)
2. Claude Code calls `picky hook <name>` with event data on stdin
3. The hook processes the event and returns a result on stdout
4. Claude Code acts on the result (show message, block action, etc.)

### Available Hooks

#### file-checker

**Trigger:** PostToolUse on Write/Edit (blocking)

Detects the language of the changed file and runs the appropriate linter/formatter:

- **Python:** `ruff check --fix`, `ruff format`, `basedpyright`
- **TypeScript:** `prettier --write`, `eslint --fix`, `tsc --noEmit`
- **Go:** `gofmt -w`, `golangci-lint run`

Returns errors and warnings to Claude Code so it can fix issues immediately.

#### tdd-enforcer

**Trigger:** PostToolUse on Write/Edit (non-blocking)

Monitors whether production code is written before a corresponding test. Emits a warning if TDD order (test first, then implementation) is violated.

#### context-monitor

**Trigger:** PostToolUse on most tools (non-blocking)

Reads the context usage percentage from the session cache and emits warnings at thresholds (40%, 60%, 80%, 90%, 95%). At 90%+, instructs Claude to initiate an Endless Mode handoff.

#### tool-redirect

**Trigger:** PreToolUse (blocking)

Intercepts certain tool calls and redirects them:
- Blocks built-in `WebSearch`/`WebFetch` in favor of MCP equivalents
- Blocks `EnterPlanMode`/`ExitPlanMode` (use `/spec` workflow instead)

#### spec-stop-guard

**Trigger:** Stop (blocking)

Prevents Claude from stopping prematurely during a `/spec` workflow if verification hasn't completed.

#### spec-plan-validator

**Trigger:** PostToolUse (blocking)

Validates the structure of plan files (correct headers, task format, status fields).

#### spec-verify-validator

**Trigger:** PostToolUse (blocking)

Validates the results of verification steps in the `/spec` workflow.

#### notify

**Trigger:** Various (non-blocking)

Sends desktop notifications:
- macOS: via `osascript`
- Linux: via `notify-send`

---

## Console Server

The console server is the central coordination point. It runs on `http://localhost:41777` and provides:

### HTTP API

| Endpoint | Method | Description |
|----------|--------|-------------|
| `/health` | GET | Health check |
| `/api/observations` | POST | Create an observation |
| `/api/observations/{id}` | GET | Get a specific observation |
| `/api/observations/search` | GET | Full-text search observations |
| `/api/observations/hybrid-search` | GET | Hybrid FTS + semantic search |
| `/api/observations/timeline/{id}` | GET | Timeline around an observation |
| `/api/sessions` | GET/POST | List/create sessions |
| `/api/sessions/{id}` | GET | Get session details |
| `/api/sessions/{id}/end` | POST | End a session |
| `/api/summaries` | POST | Create session summary |
| `/api/summaries/recent` | GET | Recent summaries |
| `/api/plans` | POST | Register a plan |
| `/api/plans/by-path` | GET | Look up plan by file path |
| `/api/plans/{id}/status` | PATCH | Update plan status |
| `/api/context/inject` | GET | Build context injection for session start |
| `/api/events` | GET | SSE event stream |
| `/api/search/reindex` | POST | Trigger search reindex |

### MCP Server

Mounted at `/mcp`, providing memory tools for Claude Code:

- `search` — Find observations by query
- `timeline` — Chronological context around a result
- `get_observations` — Fetch full details by IDs
- `save_memory` — Store a new observation

### Web Viewer

The embedded web viewer is served at the root (`/`). It shows a real-time stream of observations via Server-Sent Events.

### Database

SQLite database stored at `~/.picky/db/picky.db`. Tables:

- `observations` — Discoveries, changes, decisions
- `sessions` — Session tracking
- `summaries` — Session-end summaries
- `plans` — Plan file metadata
- `prompts` — Stored prompts
- FTS5 virtual tables for full-text search

---

## Endless Mode

Endless Mode enables seamless session continuation when Claude Code's context window fills up.

### How It Works

1. The `context-monitor` hook tracks context usage
2. At 80%, it warns to wrap up current work
3. At 90%, it triggers mandatory handoff:
   - Claude writes a `continuation.md` file with session state
   - Claude calls `picky send-clear <plan.md>` (or `--general`)
4. `send-clear` executes the restart sequence:
   - Waits 10s for memory capture
   - Writes clear signal to session directory
   - Waits 5s for session-end hooks
   - Outputs continuation prompt
5. New session starts with context injected from the console server

### Checking Context

```bash
# Human-readable
picky check-context
# Output: Context: 47.0% (OK)

# JSON
picky check-context --json
# Output: {"status":"OK","percentage":47.0}
```

### Manual Restart

```bash
# With a plan file
picky send-clear docs/plans/2026-02-16-feature.md

# Without a plan
picky send-clear --general
```

---

## Worktree Isolation

Git worktrees let you develop in isolation without affecting the main branch.

### Commands

```bash
# Create an isolated worktree
picky worktree create my-feature
# Creates .worktrees/spec-my-feature-<hash>/ with branch spec/my-feature

# Check if a worktree exists
picky worktree detect my-feature

# List changed files
picky worktree diff my-feature

# Squash merge back to base branch
picky worktree sync my-feature

# Remove worktree and branch
picky worktree cleanup my-feature

# Show active worktree info
picky worktree status
```

All commands support `--json` for structured output.

### Workflow

1. `picky worktree create <slug>` — Creates worktree, auto-stashes any dirty state
2. All work happens in the worktree directory
3. `picky worktree diff <slug>` — Review changes
4. `picky worktree sync <slug>` — Squash merge to base branch
5. `picky worktree cleanup <slug>` — Remove worktree

---

## Persistent Memory

The console server stores observations (discoveries, decisions, changes) in SQLite and makes them searchable across sessions.

### 3-Layer Search

Optimized for token efficiency:

1. **search** — Returns index with IDs (~50-100 tokens per result)
2. **timeline** — Get chronological context around a specific result
3. **get_observations** — Fetch full details only for the IDs you need

### MCP Tools

When running via `picky run`, Claude Code can use MCP tools to interact with memory:

- `search(query, limit, type, project)` — Find observations
- `timeline(anchor, depth_before, depth_after)` — Context around an observation
- `get_observations(ids)` — Full details for specific IDs
- `save_memory(text, title, project)` — Store a new observation

### Hybrid Search

Combines SQLite FTS5 full-text search with optional vector/semantic search using local embeddings. Falls back to FTS-only if semantic search isn't available.

---

## Spec-Driven Development

The `/spec` command in Claude Code triggers a structured development workflow:

1. **Plan** — Explore codebase, design implementation plan, get user approval
2. **Implement** — TDD loop for each task in the plan
3. **Verify** — Run tests, code review, compliance check

Picky Claude provides hooks that enforce this workflow:
- `spec-plan-validator` validates plan file structure
- `spec-verify-validator` validates verification results
- `spec-stop-guard` prevents premature stops during the workflow

Plans are stored as markdown files in `docs/plans/` and tracked in the database.

---

## Configuration

### Environment Variables

| Variable | Default | Description |
|----------|---------|-------------|
| `PICKY_HOME` | `~/.picky` | Base directory for data, database, sessions, logs |
| `PICKY_PORT` | `41777` | Console server HTTP port |
| `PICKY_LOG_LEVEL` | `info` | Log level: debug, info, warn, error |
| `PICKY_SESSION_ID` | auto-generated | Session identifier (set by `picky run`) |
| `PICKY_NO_UPDATE` | — | Set to any value to disable auto-update checks |

### Directory Structure

```
~/.picky/                    # Data directory (PICKY_HOME)
├── db/
│   └── picky.db            # SQLite database
├── sessions/
│   └── <session-id>/       # Per-session state files
└── logs/                    # Log files

your-project/
├── .claude/                 # Created by picky install
│   ├── rules/              # Markdown rule files
│   ├── commands/           # Spec commands
│   ├── agents/             # Agent definitions
│   ├── settings.json       # Claude Code settings (includes hooks)
│   ├── .mcp.json           # MCP server configuration
│   └── .lsp.json           # LSP configuration
└── .worktrees/             # Git worktrees (auto-added to .gitignore)
```

---

## Troubleshooting

### "claude code not found"

`picky run` can't locate the `claude` binary. Make sure Claude Code is installed and on your PATH:

```bash
which claude
```

### Console server won't start

Check if port 41777 is already in use:

```bash
lsof -i :41777
```

Change the port with `PICKY_PORT`:

```bash
PICKY_PORT=42000 picky run
```

### Hooks aren't running

Verify the hooks configuration exists:

```bash
cat .claude/settings.json | grep hooks
```

If missing, re-run the installer:

```bash
picky install --skip-prereqs --skip-deps
```

### Database errors

The SQLite database is at `~/.picky/db/picky.db`. To reset:

```bash
rm ~/.picky/db/picky.db
```

A fresh database is created automatically on next run.

### Build fails with Go version error

Picky Claude requires Go 1.25+. Check your version:

```bash
go version
```

### Tests fail

Run tests with verbose output to identify failures:

```bash
go test -v ./...
```
