# Project: Picky Claude

**Last Updated:** 2026-02-16

## Overview

A free, open-source quality layer for Claude Code. Compiles to a single Go binary (`picky`) with quality hooks, persistent memory, context management, and spec-driven development support.

## Technology Stack

- **Language:** Go 1.25
- **CLI Framework:** Cobra (github.com/spf13/cobra)
- **HTTP Router:** Chi (github.com/go-chi/chi/v5)
- **Database:** SQLite (modernc.org/sqlite) - pure Go, no CGO
- **MCP Server:** mark3labs/mcp-go
- **Build Tool:** Make
- **Testing:** Standard Go testing with table-driven patterns

## Directory Structure

```
cmd/picky/main.go              # Entry point
internal/
  cli/                         # CLI commands (cobra)
  config/                      # Configuration and branding
  hooks/                       # Hook implementations
  console/                     # HTTP server, MCP, SSE
  db/                          # SQLite CRUD operations
  search/                      # Vector + full-text search
  session/                     # Session management
  worktree/                    # Git worktree operations
  installer/                   # Project setup wizard
  statusline/                  # Status bar formatting
  updater/                     # Self-update logic
  notify/                      # Desktop notifications
  assets/                      # Embedded files
assets/                        # Source files (rules, commands, viewer)
docs/                          # Documentation
```

## Key Files

- **Configuration:** `internal/config/branding.go` (product name, env prefix)
- **Entry Point:** `cmd/picky/main.go`
- **Tests:** `*_test.go` files alongside source

## Development Commands

```bash
# Build for current platform
make build

# Run tests
make test

# Run linter
make lint

# Cross-compile for all platforms
make release

# Build web viewer then compile
make all

# Clean build artifacts
make clean
```

## Testing

Uses standard Go testing with table-driven patterns. Database tests use in-memory SQLite (`:memory:`). Console tests use `httptest.NewRecorder()`.

Run tests with minimal output to avoid context bloat:
```bash
go test ./...              # All tests
go test -v ./...           # Verbose (only when debugging)
go test -race ./...        # With race detector
go test -cover ./...       # With coverage
```

## Architecture Notes

- Single binary with embedded assets (via `embed.FS`)
- Hooks are subcommands of the same binary
- Console server runs as a goroutine within `picky run`
- Pure Go SQLite - no CGO required
- Cross-platform: macOS + Linux (arm64 + amd64)

## API Response Format

HTTP handlers use simple map-based JSON responses:
```go
writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid JSON"})
writeJSON(w, http.StatusCreated, map[string]int64{"id": id})
```

## Environment Variables

| Variable | Default | Description |
|----------|---------|-------------|
| `PICKY_HOME` | `~/.picky` | Data directory |
| `PICKY_PORT` | `41777` | Console server port |
| `PICKY_LOG_LEVEL` | `info` | Log level (debug, info, warn, error) |
| `PICKY_SESSION_ID` | auto | Session identifier |
| `PICKY_NO_UPDATE` | — | Disable auto-update check |

## Extending the System

- **Add CLI command:** Create `internal/cli/mycommand.go`, add to `rootCmd`
- **Add hook:** Create `internal/hooks/my_hook.go`, register in `init()`
- **Add language checker:** Create `internal/hooks/checkers/mylang.go`, implement `Checker` interface
- **Add API endpoint:** Add handler to `console/handlers.go`, register in `server.go`
- **Add installer step:** Create `internal/installer/steps/mystep.go`, implement `Step` interface

## Renaming

All branding is in `internal/config/branding.go`. To rename:
1. Change `BinaryName`, `DisplayName`, `EnvPrefix`, `ConfigDirName`
2. Update module path in `go.mod`
3. Update `BINARY_NAME` in `Makefile`
