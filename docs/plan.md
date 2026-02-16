# Picky Claude — Implementation Plan

> **Working title.** The name "picky-claude" / `picky` is a placeholder. All references
> to the product name are concentrated in `internal/config/branding.go` so renaming
> requires changing a single file (binary name, env prefix, config dirs, display name).
> The Makefile `BINARY_NAME` variable controls build output.

A free, open-source quality layer for Claude Code. Single Go binary providing quality-enforced, context-managed, spec-driven development. Licensed under MIT.

## 1. Architecture Overview

### Architecture (picky-claude)

| Component | Implementation | Purpose |
|-----------|---------------|---------|
| **CLI** | Single Go binary (`picky`) | All commands: run, session, worktree, install |
| **Console Server** | Goroutine within `picky serve` | Memory, MCP, HTTP API, SSE — replaces the Bun console |
| **Hooks** | Go subcommands (`picky hook <name>`) | All hooks as compiled code, no Python/Bun dependency |
| **Web Viewer** | Embedded static files (`embed.FS`) | React SPA served by the console HTTP server |
| **Installer** | `picky install` subcommand | Self-contained setup |

**Key advantage:** Single binary distribution. No runtime dependencies beyond the tools it orchestrates (Claude Code, git, language-specific linters).

---

## 2. Module Structure

```
picky-claude/                              # repo name (rename freely)
├── cmd/
│   └── picky/                             # binary entry point (matches branding.go)
│       └── main.go
├── internal/
│   ├── cli/                           # CLI command definitions
│   │   ├── root.go                    # Root command, global flags
│   │   ├── run.go                     # `picky run` — launch Claude Code with Endless Mode
│   │   ├── install.go                 # `picky install` — multi-step installer
│   │   ├── serve.go                   # `picky serve` — start console server
│   │   ├── hook.go                    # `picky hook <name>` — run a specific hook
│   │   ├── session.go                 # `picky session *` — session commands
│   │   ├── worktree.go               # `picky worktree *` — git worktree management
│   │   ├── context.go                # `picky check-context` — context usage check
│   │   ├── sendclear.go              # `picky send-clear` — trigger Endless Mode restart
│   │   ├── registerplan.go           # `picky register-plan` — plan-session association
│   │   ├── greet.go                  # `picky greet` — welcome banner
│   │   └── statusline.go            # `picky statusline` — status bar formatter
│   │
│   ├── hooks/                         # Hook implementations
│   │   ├── dispatcher.go             # Route hook events to handlers
│   │   ├── context_monitor.go        # Context usage tracking + handoff enforcement
│   │   ├── file_checker.go           # Language-specific lint/format orchestrator
│   │   ├── tdd_enforcer.go           # TDD red-green-refactor enforcement
│   │   ├── tool_redirect.go          # Tool routing (block WebSearch, redirect, etc.)
│   │   ├── spec_stop_guard.go        # Prevent premature stop during /spec
│   │   ├── spec_plan_validator.go    # Plan file structure validation
│   │   ├── spec_verify_validator.go  # Verify phase validation
│   │   ├── session_end.go            # Session cleanup + memory save
│   │   ├── notify.go                 # Desktop notifications
│   │   └── checkers/                 # Language-specific checkers
│   │       ├── checker.go            # Checker interface
│   │       ├── python.go             # ruff + basedpyright
│   │       ├── typescript.go         # prettier + eslint + tsc
│   │       └── golang.go             # gofmt + golangci-lint
│   │
│   ├── console/                       # Console server (replaces Bun console)
│   │   ├── server.go                 # HTTP server setup, middleware, routes
│   │   ├── mcp.go                    # MCP server (memory search tools)
│   │   ├── sse.go                    # Server-Sent Events broadcaster
│   │   ├── routes/                   # HTTP route handlers
│   │   │   ├── memory.go            # Search, timeline, observations
│   │   │   ├── session.go
│   │   │   ├── plan.go
│   │   │   ├── worktree.go
│   │   │   ├── metrics.go
│   │   │   ├── search.go

│   │   │   └── viewer.go            # Serve embedded web UI
│   │   └── context/                  # Context injection builder
│   │       ├── builder.go           # Build startup context from observations
│   │       ├── compiler.go          # Observation compiler
│   │       └── token.go             # Token estimation
│   │
│   ├── db/                            # SQLite persistence
│   │   ├── database.go               # Connection management, migrations
│   │   ├── migrations.go             # Schema migrations
│   │   ├── observations.go           # CRUD for observations
│   │   ├── sessions.go              # Session tracking
│   │   ├── summaries.go             # Session summaries
│   │   ├── plans.go                 # Plan file tracking
│   │   ├── prompts.go               # Prompt storage
│   │   └── timeline.go             # Timeline queries
│   │
│   ├── search/                        # Search subsystem
│   │   ├── orchestrator.go           # Hybrid search coordinator
│   │   ├── sqlite_search.go         # FTS5 full-text search
│   │   ├── vector_search.go         # Embedding-based semantic search
│   │   ├── embedding.go             # Embedding generation (local model)
│   │   ├── filters.go               # Date, project, type filters
│   │   └── formatter.go            # Result formatting
│   │
│   ├── session/                       # Session management
│   │   ├── manager.go               # Session lifecycle
│   │   ├── continuation.go          # Endless Mode continuation logic
│   │   ├── tracker.go               # Message counting, context tracking
│   │   └── cleanup.go              # Session cleanup
│   │
│   ├── worktree/                      # Git worktree management
│   │   ├── manager.go               # Create, detect, diff, sync, cleanup
│   │   ├── stash.go                 # Auto-stash/restore for dirty trees
│   │   └── merge.go                # Squash merge logic
│   │
│   ├── installer/                     # Installation subsystem
│   │   ├── installer.go             # Orchestrator with rollback
│   │   ├── steps/                   # Individual install steps
│   │   │   ├── prerequisites.go     # Check brew, git, etc.
│   │   │   ├── dependencies.go      # Install vexor, playwright-cli, mcp-cli, etc.
│   │   │   ├── shell_config.go      # Add aliases to .zshrc/.bashrc
│   │   │   ├── claude_files.go      # Set up .claude/ directory structure
│   │   │   ├── config_files.go      # Write settings, MCP, LSP configs
│   │   │   ├── vscode.go           # VS Code extension recommendations
│   │   │   └── finalize.go         # Final verification
│   │   └── ui.go                   # Terminal UI for installer (progress, prompts)
│   │
│   ├── updater/                       # Auto-update
│   │   ├── checker.go               # Check for new versions
│   │   └── updater.go              # Self-update binary
│   │
│   ├── statusline/                    # Status bar formatting
│   │   ├── formatter.go             # JSON→formatted status line
│   │   ├── providers.go             # Data providers (plan, session, worktree, etc.)
│   │   ├── widgets.go               # Individual status widgets
│   │   └── tips.go                 # Contextual tips
│   │
│   └── config/                        # Configuration
│       ├── branding.go              # ★ SINGLE SOURCE for product name, env prefix, paths
│       ├── paths.go                  # Standard paths (~/.picky/, .claude/, etc.)
│       ├── config.go                # Config file loading
│       └── constants.go             # Version, defaults
│
├── assets/                            # Static assets
│   ├── rules/                        # Markdown rule files (embedded)
│   ├── commands/                     # Spec commands (embedded)
│   ├── agents/                       # Agent definitions (embedded)
│   ├── hooks.json                   # Hook configuration template
│   ├── settings.json                # Claude Code settings template
│   └── viewer/                      # Web viewer SPA (embedded)
│       ├── index.html
│       ├── viewer.js
│       └── viewer.css
│
├── web/                               # Web viewer source (React + Tailwind)
│   ├── src/
│   ├── package.json
│   └── vite.config.ts
│
├── go.mod
├── go.sum
├── Makefile
└── README.md
```

---

## 3. Feature Parity Breakdown

### 3.1 CLI Commands

| Command | Notes |
|---------|-------|
| `picky run` | Launch Claude Code with hooks + Endless Mode |
| `picky check-context --json` | Read context percentage from cache |
| `picky send-clear <plan>` | Trigger Endless Mode restart |
| `picky send-clear --general` | Restart without plan context |
| `picky register-plan <path> <status>` | Associate plan with session |
| `picky greet` | Welcome banner |
| `picky statusline` | Read JSON from stdin, format status bar |
| `picky worktree *` | All worktree subcommands (create, detect, diff, sync, cleanup, status) |
| `picky session list` | List active sessions |
| `picky install` | Self-contained installer |
| `picky serve` | Start console server standalone |

All commands support `--json` for structured output.

### 3.2 Hooks System

Claude Code hooks are configured in `.claude/hooks.json` (or `settings.json`). Each hook calls back into the `picky` binary:

```json
{
  "hooks": {
    "PostToolUse": [
      {
        "matcher": "Write|Edit|MultiEdit",
        "command": "picky hook file-checker",
        "blocking": true,
        "timeout": 15000
      },
      {
        "matcher": "Write|Edit|MultiEdit",
        "command": "picky hook tdd-enforcer",
        "blocking": false,
        "timeout": 15000
      },
      {
        "matcher": "Read|Write|Edit|MultiEdit|Bash|Task|Skill|Grep|Glob",
        "command": "picky hook context-monitor",
        "blocking": false,
        "timeout": 15000
      }
    ],
    "PreToolUse": [
      {
        "matcher": "Bash|WebSearch|WebFetch|Grep|Task|EnterPlanMode|ExitPlanMode",
        "command": "picky hook tool-redirect",
        "blocking": true,
        "timeout": 15000
      }
    ],
    "SessionStart": [
      {
        "command": "picky hook session-start",
        "blocking": true,
        "timeout": 15000
      }
    ],
    "Stop": [
      {
        "command": "picky hook spec-stop-guard",
        "blocking": true,
        "timeout": 15000
      }
    ],
    "SessionEnd": [
      {
        "command": "picky hook session-end",
        "blocking": true,
        "timeout": 15000
      }
    ]
  }
}
```

Hooks are compiled Go code invoked as subcommands of the same binary. Sub-5ms hook execution.

### 3.3 Hook Implementations

| Hook | Trigger | Blocking | Behavior |
|------|---------|----------|----------|
| **file-checker** | PostToolUse (Write/Edit) | Yes | Detect language from changed file, run appropriate linter/formatter, return errors |
| **tdd-enforcer** | PostToolUse (Write/Edit) | No | Check if production code was written before a test. Warn if TDD order violated |
| **context-monitor** | PostToolUse (most tools) | No | Read context-pct.json, emit warnings at 40/60/80/90/95% thresholds |
| **tool-redirect** | PreToolUse | Yes | Block built-in WebSearch/WebFetch, redirect to MCP equivalents. Block EnterPlanMode/ExitPlanMode |
| **spec-stop-guard** | Stop | Yes | Prevent Claude from stopping during /spec if verification isn't complete |
| **session-start** | SessionStart | Yes | Load context from console, inject observations |
| **session-end** | SessionEnd | Yes | Save session summary to console, cleanup |
| **spec-plan-validator** | PostToolUse | Yes | Validate plan file structure |
| **spec-verify-validator** | PostToolUse | Yes | Validate verification results |
| **notify** | Various | No | Desktop notification via `osascript` (macOS) or `notify-send` (Linux) |

### 3.4 Language Checkers

Each checker follows the same interface:

```go
type Checker interface {
    Name() string
    Extensions() []string
    Check(ctx context.Context, filePath string) (*CheckResult, error)
}

type CheckResult struct {
    Errors   []Diagnostic
    Warnings []Diagnostic
    Fixed    bool // Whether auto-fix was applied
}
```

| Language | Tools | Auto-fix |
|----------|-------|----------|
| Python | `ruff check --fix`, `ruff format`, `basedpyright` | Yes (ruff) |
| TypeScript | `prettier --write`, `eslint --fix`, `tsc --noEmit` | Yes (prettier, eslint) |
| Go | `gofmt -w`, `golangci-lint run` | Yes (gofmt) |

### 3.5 Console Server

The console server runs on `localhost:41777` (or configurable port).

**Components:**

1. **HTTP API** — REST endpoints for memory, sessions, plans, worktrees, search, metrics
2. **MCP Server** — Model Context Protocol server exposing `search`, `timeline`, `get_observations`, `save_memory` tools
3. **SSE Broadcaster** — Real-time event stream for the web viewer
4. **Context Builder** — Generates context injection from recent observations at session start
5. **Background Worker** — Retention cleanup, embedding indexing, health monitoring

**Database:** SQLite via `modernc.org/sqlite` (pure Go, no CGO). Schema:

- `observations` — Individual discoveries/changes with metadata
- `sessions` — Session tracking with start/end/message count
- `summaries` — Session-end summaries
- `plans` — Plan file metadata and status
- `prompts` — Stored prompts for context injection
- FTS5 virtual tables for full-text search

### 3.6 Search System

Three-layer search:

1. **search(query)** → Returns index with IDs (~50-100 tokens/result)
2. **timeline(anchor=ID)** → Chronological context around results
3. **get_observations([IDs])** → Full details for filtered IDs

**Implementation:**
- SQLite FTS5 for keyword search
- Optional vector search using local embeddings (via `github.com/nicholasgasior/gopher-transformers` or similar Go bindings)
- Hybrid scoring that combines FTS5 rank + cosine similarity

### 3.7 Session Management & Endless Mode

```
Session Lifecycle:
1. `picky run` starts
2. Assigns PICKY_SESSION_ID (PID-based or UUID)
3. Launches Claude Code with environment + hooks
4. Hooks report to console server via HTTP
5. Context monitor watches usage
6. At 90%: writes continuation.md → calls `picky send-clear`
7. `picky send-clear`:
   a. Waits for memory capture (10s)
   b. Sends /clear to Claude Code
   c. Sends continuation prompt
8. New session starts with context injection from console
9. Repeat until task complete
```

### 3.8 Worktree Management

Full git worktree lifecycle:

```go
// Manager interface
type WorktreeManager interface {
    Detect(slug string) (*WorktreeInfo, error)
    Create(slug string) (*WorktreeInfo, error)   // Auto-stash, create branch + worktree
    Diff(slug string) (*DiffResult, error)        // Changed files in worktree
    Sync(slug string) (*SyncResult, error)        // Squash merge to base branch
    Cleanup(slug string) error                    // Remove worktree + branch
    Status() (*StatusInfo, error)                 // Active worktree info
}
```

Worktrees are created at `.worktrees/spec-<slug>-<hash>/` with branch `spec/<slug>`. Auto-stash handles dirty working trees.

### 3.9 Rules, Commands, Agents

These are markdown files embedded into the binary and extracted during `picky install`:

| Type | Location | Count (approx) |
|------|----------|-----------------|
| Rules | `.claude/rules/*.md` | ~30 files |
| Commands | `.claude/commands/*.md` | 6 files (spec, spec-plan, spec-implement, spec-verify, learn, sync) |
| Agents | `.claude/agents/*.md` | 4 files (plan-verifier, plan-challenger, spec-reviewer-compliance, spec-reviewer-quality) |
| Settings | `.claude/settings.json` | 1 file |
| LSP config | `.claude/.lsp.json` | 1 file |
| MCP config | `.claude/.mcp.json` | 1 file |

### 3.10 Installer

`picky install` runs a multi-step setup:

| Step | Actions |
|------|---------|
| 1. Prerequisites | Verify git, Claude Code installed. Check OS/arch support |
| 2. Dependencies | Install vexor, playwright-cli, mcp-cli (via npm/brew) |
| 3. Shell config | Add `picky` alias/PATH to .zshrc/.bashrc |
| 4. Claude files | Create .claude/ directory with rules, commands, agents, hooks |
| 5. Config files | Write settings.json, .lsp.json, .mcp.json, hooks.json |
| 6. VS Code | Recommend extensions |
| 7. Finalize | Verify setup, print summary |

Each step supports rollback on failure. Progress shown with terminal UI.

### 3.11 Status Line

Reads JSON from stdin (piped from Claude Code), formats a status bar with widgets:

- Session info (ID, duration)
- Context usage percentage
- Active plan + status
- Worktree status
- Contextual tips

### 3.12 Auto-Updater

- Check GitHub releases for latest version on startup
- Compare semver with current binary version
- Download + replace binary (with platform/arch detection)
- Fallback: print update instructions

---

## 4. Go Library Choices

| Need | Library | Why |
|------|---------|-----|
| CLI framework | `github.com/spf13/cobra` | Standard for Go CLIs, subcommand support |
| HTTP server | `net/http` (stdlib) + `github.com/go-chi/chi/v5` | Lightweight router, middleware support |
| SQLite | `modernc.org/sqlite` | Pure Go, no CGO, cross-compile friendly |
| MCP protocol | `github.com/mark3labs/mcp-go` | Go MCP SDK |
| JSON | `encoding/json` (stdlib) | Standard |
| YAML | `gopkg.in/yaml.v3` | For config files |
| Logging | `log/slog` (stdlib) | Structured logging, Go 1.21+ |
| Embed | `embed` (stdlib) | Embed rules, assets, viewer into binary |
| Terminal UI | `github.com/charmbracelet/bubbletea` | Installer TUI, progress display |
| Git operations | `os/exec` calling `git` | Direct git CLI, reliable |
| Process management | `os/exec` | Launch Claude Code, linters |
| SSE | Hand-rolled (trivial in Go) | `text/event-stream` with `http.Flusher` |
| Semver | `golang.org/x/mod/semver` | Version comparison for updater |
| Testing | `testing` (stdlib) + `github.com/stretchr/testify` | Assertions, mocking |

---

## 5. Data Flow Diagrams

### 5.1 Normal Session Flow

```
User runs `picky run`
        │
        ▼
┌─────────────────┐     HTTP      ┌──────────────────┐
│  Claude Code    │◄────────────►│  Console Server   │
│  (subprocess)   │               │  (goroutine)      │
└───────┬─────────┘               │  - MCP server     │
        │                         │  - SQLite DB       │
        │ hooks                   │  - SSE broadcast   │
        ▼                         │  - Web viewer      │
┌─────────────────┐               └──────────────────┘
│  picky hook *   │                       ▲
│  (same binary)  │───────────────────────┘
│  - file-checker │     POST observations
│  - tdd-enforcer │     GET context
│  - ctx-monitor  │
└─────────────────┘
```

### 5.2 Endless Mode Flow

```
Context at 90%
      │
      ▼
context-monitor hook fires
      │
      ▼
Returns instruction to Claude:
"Write continuation.md, then call picky send-clear"
      │
      ▼
Claude writes continuation.md
      │
      ▼
Claude calls: picky send-clear <plan.md>
      │
      ├─ Wait 10s for memory capture
      ├─ Send /clear to Claude Code stdin
      ├─ Wait 5s for session end hooks
      ├─ Send continuation prompt
      │
      ▼
New session starts
      │
      ▼
session-start hook → injects context from console
      │
      ▼
Claude reads continuation.md + injected context
      │
      ▼
Continues where it left off
```

---

## 6. Implementation Phases

### Phase 1: Foundation
- Project scaffolding (go mod, Makefile, CI)
- CLI framework with cobra (root + placeholder subcommands)
- Configuration system (paths, defaults, config loading)
- SQLite database with migrations
- Basic console HTTP server skeleton

### Phase 2: Core Hooks
- Hook dispatcher (`picky hook <name>` routing)
- file-checker with Python, TypeScript, Go checkers
- tdd-enforcer
- context-monitor (read context cache, threshold logic)
- tool-redirect

### Phase 3: Console Server
- Observation CRUD (store, search, get)
- Session management (create, track, cleanup)
- FTS5 search implementation
- Timeline queries
- Context builder (observation → startup injection)
- MCP server with memory tools

### Phase 4: Session Management
- `picky run` — launch Claude Code with env + hooks
- Session ID management
- `picky send-clear` — Endless Mode restart
- `picky check-context` — context percentage
- `picky register-plan` — plan association
- Continuation file protocol

### Phase 5: Worktree & Git
- Worktree create/detect/diff/sync/cleanup/status
- Auto-stash logic
- Squash merge implementation

### Phase 6: Spec Workflow Support
- spec-stop-guard hook
- spec-plan-validator hook
- spec-verify-validator hook
- Rules, commands, agents embedding + extraction

### Phase 7: Installer
- Multi-step installer with TUI
- Prerequisites check
- Dependency installation
- Shell config setup
- Claude files extraction from embedded assets
- Rollback on failure

### Phase 8: Polish
- Status line formatter
- Auto-updater (GitHub Releases, self-replace binary)
- Web viewer (embed pre-built SPA)
- Greet/banner command
- Desktop notifications

### Phase 9: Search Enhancement
- Vector search with local embeddings (optional)
- Hybrid search scoring
- Retention/cleanup scheduler

---

## 7. Key Design Decisions

### 7.1 Single Binary
Everything compiles to one binary. Rules, commands, agents, and web viewer assets are embedded via `embed.FS`. The `picky install` command extracts these to `.claude/` in the project directory.

### 7.2 No CGO
Using `modernc.org/sqlite` (pure Go) instead of `mattn/go-sqlite3` (CGO). This enables easy cross-compilation for macOS (arm64/amd64), Linux (arm64/amd64), and Windows.

### 7.3 Console as Goroutine
The console server runs as a goroutine when `picky run` is called, not as a separate process. This simplifies lifecycle management. Alternatively, `picky serve` can run it standalone for debugging.

### 7.4 Hooks as Subcommands
Hooks are Go subcommands (`picky hook file-checker`). Claude Code's hooks.json calls back into the same binary. Sub-5ms hook execution.

### 7.5 MCP Server
The MCP server runs within the console goroutine, exposing memory tools (search, timeline, get_observations, save_memory) over stdio or HTTP transport as needed by Claude Code's MCP config.

### 7.6 Web Viewer
The React web viewer is a separate build step (npm/vite in `web/`). The built assets are embedded into the Go binary at compile time. Served at `http://localhost:41777`.

### 7.7 Centralized Branding (Easy Rename)

All name-dependent values live in `internal/config/branding.go`:

```go
package config

const (
    // ★ Change these to rename the entire product
    BinaryName  = "picky"          // CLI executable name
    DisplayName = "Picky Claude"   // Human-readable name (banner, docs)
    EnvPrefix   = "PICKY"          // Environment variable prefix (PICKY_HOME, etc.)
    ConfigDir   = ".picky"         // ~/.picky/
)
```

Everything else derives from these constants:
- `paths.go` builds `~/.picky/sessions/`, `~/.picky/db/` etc. from `ConfigDir`
- `hooks.json` template uses `BinaryName` for hook commands (`picky hook file-checker`)
- `statusline` uses `DisplayName` in the banner
- Env vars use `EnvPrefix + "_SESSION_ID"`, etc.
- `Makefile` reads `BINARY_NAME ?= picky` and passes it via `-ldflags`

To rename: change the four constants, update `go.mod` module path, and `BINARY_NAME` in the Makefile. Everything else propagates automatically.

### 7.8 No License System

This is free, open-source software (MIT license). There is no license server, activation,
trial, fingerprinting, or analytics. The console server has no auth/license routes.
All features are available unconditionally.

---

## 8. Environment Variables

The `PICKY_` prefix is derived from `branding.go` and changes automatically with a rename.

| Variable | Purpose |
|----------|---------|
| `PICKY_SESSION_ID` | Current session identifier |
| `PICKY_HOME` | Override default config dir (~/.picky) |
| `PICKY_PORT` | Console server port (default: 41777) |
| `PICKY_LOG_LEVEL` | Logging level (debug, info, warn, error) |
| `PICKY_NO_UPDATE` | Disable auto-update check |
| `CLAUDE_CODE_TASK_LIST_ID` | Task list isolation per session |

---

## 9. Testing Strategy

| Layer | Approach |
|-------|----------|
| Unit tests | Table-driven tests for all business logic |
| Hook tests | Test each hook's stdin→stdout behavior with mock tool events |
| Database tests | In-memory SQLite for fast CRUD tests |
| Integration tests | Launch `picky run` with a mock Claude Code, verify hook flows |
| E2E tests | Script that runs a full session and checks continuation works |

---

## 10. Build & Distribution

```makefile
BINARY_NAME ?= picky

# Build for current platform
build:
	go build -o bin/$(BINARY_NAME) ./cmd/picky

# Build for all platforms
release:
	GOOS=darwin GOARCH=arm64 go build -o bin/$(BINARY_NAME)-darwin-arm64 ./cmd/picky
	GOOS=darwin GOARCH=amd64 go build -o bin/$(BINARY_NAME)-darwin-amd64 ./cmd/picky
	GOOS=linux GOARCH=arm64 go build -o bin/$(BINARY_NAME)-linux-arm64 ./cmd/picky
	GOOS=linux GOARCH=amd64 go build -o bin/$(BINARY_NAME)-linux-amd64 ./cmd/picky

# Build web viewer then embed
viewer:
	cd web && npm run build
	cp -r web/dist/* assets/viewer/

# Full release
all: viewer build
```

Distribute via GitHub Releases with platform-specific binaries. Install script downloads the right binary for the platform.

To release under a different name: `make release BINARY_NAME=newname`.
