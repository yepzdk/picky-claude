# CLAUDE.md — Developer Guide for Picky Claude

This file is for Claude Code (or any AI assistant) working on this codebase. It describes the project structure, how to build and test, and how to extend the system.

## Project Overview

Picky Claude compiles to a single binary (`picky`) that wraps Claude Code with quality hooks, persistent memory, context management, and spec-driven development support.

## Build and Test

```bash
make build        # Build bin/picky for current platform
make test         # Run all tests (go test ./...)
make lint         # Run golangci-lint
make release      # Cross-compile for macOS + Linux (arm64/amd64)
make all          # Build web viewer then compile binary
```

Version is injected at build time via `-ldflags`. The `dev` version is used for local builds.

## Project Structure

```
cmd/picky/main.go              # Entry point — calls cli.Execute()
internal/
  cli/                         # All CLI commands (cobra)
    root.go                    # Root command, --json global flag
    run.go                     # picky run — launches Claude Code
    serve.go                   # picky serve — standalone console server
    hook.go                    # picky hook <name> — dispatcher entry
    install.go                 # picky install — multi-step installer
    worktree.go                # picky worktree subcommands
    session.go                 # picky session list
    context.go                 # picky check-context
    sendclear.go               # picky send-clear
    registerplan.go            # picky register-plan
    greet.go                   # picky greet
    statusline.go              # picky statusline
  config/
    branding.go                # ★ Product name, env prefix, config dir name
    constants.go               # Version, default port, default log level
    config.go                  # Runtime config from env vars
    paths.go                   # Derived paths (~/.picky/db/, sessions/, logs/)
  hooks/
    dispatcher.go              # Hook registry and dispatch (name → handler)
    protocol.go                # Hook input/output JSON protocol (stdin/stdout)
    context_monitor.go         # Context usage tracking
    file_checker.go            # Language-aware lint/format
    tdd_enforcer.go            # TDD order enforcement
    tool_redirect.go           # Block/redirect tool calls
    spec_stop_guard.go         # Prevent premature stop during /spec
    spec_plan_validator.go     # Plan file structure validation
    spec_verify_validator.go   # Verification result validation
    notify.go                  # Desktop notifications
    checkers/
      checker.go               # Checker interface
      python.go                # ruff + basedpyright
      typescript.go            # prettier + eslint + tsc
      golang.go                # gofmt + golangci-lint
  console/
    server.go                  # HTTP server, route registration, lifecycle
    mcp.go                     # MCP server with memory tools
    sse.go                     # Server-Sent Events broadcaster
    handlers.go                # HTTP handlers (observations, sessions, plans, etc.)
    handlers_search.go         # Search-specific handlers
    handlers_session.go        # Session-specific handlers
    context/
      builder.go               # Build startup context from observations
      token.go                 # Token estimation
  db/
    database.go                # SQLite connection, Open(), Close()
    migrations.go              # Schema migrations (run on Open)
    observations.go            # Observation CRUD
    sessions.go                # Session CRUD
    summaries.go               # Summary CRUD
    plans.go                   # Plan CRUD
    prompts.go                 # Prompt storage
  search/
    orchestrator.go            # Hybrid search coordinator
    vector_search.go           # Embedding-based semantic search
    embedding.go               # Local embedding generation
    retention.go               # Background cleanup scheduler
    formatter.go               # Result formatting
  session/
    manager.go                 # Session lifecycle
    runner.go                  # Claude Code process runner
    client.go                  # HTTP client for console API
    context_check.go           # Read context percentage from session dir
    sendclear.go               # Clear signal + continuation prompt
  worktree/
    manager.go                 # Create, detect, diff, sync, cleanup, status
    stash.go                   # Auto-stash/restore for dirty trees
    merge.go                   # Squash merge logic
  installer/
    installer.go               # Step orchestrator with rollback
    ui.go                      # Terminal UI (banner, progress, summary)
    steps/
      prerequisites.go         # Check git, Claude Code
      dependencies.go          # Install optional tools
      shell_config.go          # Add to .zshrc/.bashrc
      claude_files.go          # Create .claude/ directory
      config_files.go          # Write settings, hooks, MCP config
      vscode.go                # VS Code recommendations
      finalize.go              # Final verification
  statusline/
    formatter.go               # JSON→formatted status line
    tips.go                    # Contextual tips
  updater/
    checker.go                 # Check GitHub releases for updates
    updater.go                 # Self-update binary
  notify/
    notify.go                  # Cross-platform desktop notifications
  assets/
    embed.go                   # embed.FS for rules, commands, agents, viewer
    extract.go                 # Extract embedded files to disk
assets/                        # Source files that get embedded
  rules/                       # Markdown rule files
  commands/                    # Spec commands
  agents/                      # Agent definitions
  viewer/                      # Web viewer static files
docs/
  plan.md                      # Original implementation plan
  usage.md                     # User-facing usage documentation
```

## Key Conventions

### Branding

All product name references derive from `internal/config/branding.go`:

```go
BinaryName    = "picky"         // CLI name
DisplayName   = "Picky Claude"  // Human-readable
EnvPrefix     = "PICKY"         // Env var prefix (PICKY_PORT, etc.)
ConfigDirName = ".picky"        // ~/.picky/
```

To rename, change these four constants, update `go.mod` module path, and `BINARY_NAME` in the Makefile.

### Hook Registration

Hooks self-register via `init()` functions. The dispatcher in `hooks/dispatcher.go` maintains a `map[string]Hook` registry. To add a new hook:

1. Create `internal/hooks/my_hook.go`
2. Implement the `Hook` function signature: `func(input *Input) error`
3. Call `Register("my-hook", myHookFunc)` in an `init()` function
4. The hook is now callable via `picky hook my-hook`

Hooks read JSON from stdin (`ReadInput()`) and write results via `WriteOutput()`, `BlockWithError()`, or `ExitOK()` — all defined in `protocol.go`.

### CLI Commands

Commands use [cobra](https://github.com/spf13/cobra). Each command is in its own file under `internal/cli/`. The pattern:

1. Create a `*cobra.Command` variable
2. Add it to `rootCmd` (or a parent command) in an `init()` function
3. Use `jsonOutput` global flag for `--json` support

### Installer Steps

Each installer step implements the `Step` interface:

```go
type Step interface {
    Name() string
    Run(ctx *Context) error
    Rollback(ctx *Context)
}
```

Steps are in `internal/installer/steps/`. Add new steps there and wire them into `cli/install.go`.

### Console Routes

HTTP routes are registered in `console/server.go` → `registerRoutes()`. Handlers are in `handlers.go`, `handlers_search.go`, and `handlers_session.go`. Add new routes by:

1. Adding the handler method to `*Server`
2. Registering the route in `registerRoutes()`

### Database

SQLite via `modernc.org/sqlite` (pure Go, no CGO). Schema migrations are in `db/migrations.go` and run automatically on `db.Open()`. To add a new table:

1. Add a migration function in `migrations.go`
2. Register it in the migrations slice
3. Create a new file for CRUD operations (e.g., `db/my_table.go`)

### Embedded Assets

Static files in `assets/` are embedded via `embed.FS` in `internal/assets/embed.go`. The `Extract()` function writes them to disk during `picky install`. The web viewer is served directly from the embedded FS.

## Testing

Tests use the standard `testing` package with table-driven patterns. Most packages have `*_test.go` files alongside the source.

```bash
# All tests
go test ./...

# Specific package
go test ./internal/hooks/...

# Verbose (for debugging)
go test -v ./internal/db/...

# With race detector
go test -race ./...

# Coverage
go test -cover ./...
```

Database tests use in-memory SQLite (`:memory:`). Console tests use `httptest.NewRecorder()`.

## Dependencies

| Dependency | Purpose |
|------------|---------|
| `github.com/spf13/cobra` | CLI framework |
| `github.com/go-chi/chi/v5` | HTTP router |
| `modernc.org/sqlite` | Pure Go SQLite (no CGO) |
| `github.com/mark3labs/mcp-go` | MCP protocol server |

Keep dependencies minimal. Use stdlib where possible (`net/http`, `log/slog`, `encoding/json`, `embed`).

## Common Tasks

### Add a new CLI command

1. Create `internal/cli/mycommand.go`
2. Define a `cobra.Command` with `Use`, `Short`, `RunE`
3. In `init()`, call `rootCmd.AddCommand(myCmd)`

### Add a new hook

1. Create `internal/hooks/my_hook.go`
2. Write the handler: `func myHook(input *Input) error { ... }`
3. In `init()`, call `Register("my-hook", myHook)`
4. Add the hook to the hooks.json template in `installer/steps/config_files.go`

### Add a new language checker

1. Create `internal/hooks/checkers/mylang.go`
2. Implement the `Checker` interface (Name, Extensions, Check)
3. Register it in an `init()` function via `RegisterChecker()`

### Add a new API endpoint

1. Add handler method to `*Server` in `console/handlers.go` (or a new handlers file)
2. Register the route in `console/server.go` → `registerRoutes()`
3. Add tests in `console/server_test.go`

### Add a new installer step

1. Create `internal/installer/steps/mystep.go`
2. Implement `Step` interface (Name, Run, Rollback)
3. Add it to the step list in `cli/install.go`

### Add a new database table

1. Add migration in `db/migrations.go`
2. Create `db/mytable.go` with CRUD methods on `*DB`
3. Add tests in `db/database_test.go`
