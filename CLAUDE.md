# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

Dboxed is a container orchestration system written in Go. It manages "boxes" (containerized environments), volumes, workspaces, and networks. The system includes both a CLI client and API server.

## Build & Development Commands

### Building
```bash
# Build the main binary
go build ./cmd/dboxed

# Build with goreleaser (cross-platform builds)
task goreleaser-build-snapshot

# Build for release
task goreleaser-release-nightly  # Nightly release
task goreleaser-release-final    # Final release
```

### Testing
```bash
# Run all tests
go test ./...

# Run specific package tests
go test ./pkg/volume/losetup
```

### Database Migrations
```bash
# Generate new migration (requires MIGRATION_NAME)
task generate-migrations MIGRATION_NAME=my_migration_name

# Generate SQLite migration only
task generate-migrations-sqlite MIGRATION_NAME=my_migration_name

# Generate Postgres migration only
task generate-migrations-postgres MIGRATION_NAME=my_migration_name
```

## Architecture

### CLI Structure (`cmd/dboxed/commands/`)

The CLI uses the Kong framework for command parsing. Commands are organized hierarchically:

- **Top-level commands**: `box`, `volume`, `workspace`, `token`, `sandbox`, `server`
- **Sub-commands**: Commands can have nested sub-commands (e.g., `box compose create`, `box volume attach`)
- **Command pattern**: Each command implements a `Run(g *flags.GlobalFlags) error` method

Example command structure:
```go
type BoxCommands struct {
    Create  CreateCmd          `cmd:"" help:"Create a box"`
    Compose compose.ComposeCmd `cmd:"" help:"Manage compose projects"`
    Volume  volume.VolumeCmd   `cmd:"" help:"Manage volume attachments"`
}
```

### Client Architecture (`pkg/baseclient/` and `pkg/clients/`)

- **baseclient.Client**: Low-level API client handling auth, workspace context, and HTTP requests
- **Domain clients** (BoxClient, VolumeClient, etc.): High-level clients wrapping baseclient for specific resources
- **Authentication**: Supports OAuth2 and static tokens
- **Workspace context**: All API calls are scoped to a workspace (stored in `~/.dboxed/client-auth.yaml`)

### Server Architecture (`pkg/server/`)

The server is a REST API built with Huma (v2) and Gin:

- **Resources** (`pkg/server/resources/`): API handlers for each domain (boxes, volumes, workspaces, etc.)
- **Models** (`pkg/server/models/`): Request/response models and DB model conversion
- **Database** (`pkg/server/db/`): Database abstraction supporting SQLite and PostgreSQL
  - Uses `dmodel` package for database models
  - Migrations in `pkg/server/db/migration/sqlite/` and `postgres/`
  - Uses Atlas for migration generation from Go schemas

### Core Concepts

#### Boxes
A "box" is a containerized environment that can have:
- **Compose projects**: Docker Compose configurations attached to the box
- **Volume attachments**: Persistent volumes mounted into the box
- **Network**: Optional network association (e.g., Netbird for VPN)
- **BoxSpec** (`pkg/boxspec/`): The specification defining what runs in a box

#### Volumes
Persistent storage managed by volume providers:
- Volume providers handle different storage backends
- Volumes can be attached to boxes with specific mount parameters (UID, GID, mode)
- Volume locking prevents concurrent access

#### Workspaces
Multi-tenancy concept - all resources belong to a workspace.

#### Runner System (`pkg/runner/`)
Executes box specifications in sandboxes:
- **box-spec-runner**: Runs BoxSpec configurations
- **sandbox**: Manages container sandboxes
- **run-in-sandbox**: Executes commands inside sandboxes
- **dockercli**: Docker command-line interface wrapper

## Code Patterns

### Adding a New CLI Command

1. Create command file in `cmd/dboxed/commands/<domain>/`
2. Implement struct with Kong tags:
   ```go
   type MyCmd struct {
       Name string `help:"Resource name" required:"" arg:""`
   }

   func (cmd *MyCmd) Run(g *flags.GlobalFlags) error {
       ctx := context.Background()
       c, err := g.BuildClient(ctx)
       // ... implementation
   }
   ```
3. Add to parent command struct
4. Use `slog.Info()` for output (not `fmt.Print`)

### Adding a Client Method

1. Add method to appropriate client in `pkg/clients/`:
   ```go
   func (c *BoxClient) MyMethod(ctx context.Context, req models.MyRequest) (*models.MyResponse, error) {
       p, err := c.Client.BuildApiPath(true, "boxes", "my-endpoint")
       if err != nil {
           return nil, err
       }
       return baseclient.RequestApi[models.MyResponse](ctx, c.Client, "POST", p, req)
   }
   ```

### Adding a Server Endpoint

1. Add handler method to resource server (e.g., `pkg/server/resources/boxes/`)
2. Register in `Init()` method using Huma:
   ```go
   huma.Post(workspacesGroup, "/boxes/{id}/endpoint", s.restMyHandler)
   ```
3. Handler signature: `func (s *Server) restMyHandler(ctx context.Context, input *MyInput) (*MyOutput, error)`

### Database Migrations

1. Define schema changes in `pkg/server/db/migration/templates/`
2. Run `task generate-migrations MIGRATION_NAME=description`
3. This generates migrations in both SQLite and Postgres directories
4. Migrations use Goose format with Atlas for generation

## TUI (Terminal UI)

The `tui` command provides an interactive interface built with:
- **bubbletea**: Main TUI framework (state management, event loop)
- **huh**: Form components (for creating resources)
- **lipgloss**: Styling

TUI architecture:
- Main model with view states (main menu, box list, volume list, etc.)
- Each view has its own update and render logic
- Commands return `tea.Cmd` for async operations
- Messages (types ending in `Msg`) communicate between components

## Important Details

### Debug Logging
Use `--debug` flag to enable debug logging. This sets log level to Debug and logs all API requests/responses.

### Global Flags
Available on all commands:
- `--debug`: Enable debug logging
- `--api-url`: Override API URL
- `--api-token`: Override API token
- `--workspace`: Override workspace
- `--client-auth-file`: Override client auth file location

### Command Aliases
Many commands support aliases (defined via Kong tags):
- `ls` for `list`
- `rm`, `delete` for delete operations

### Helper Utilities
- `commandutils.GetBox()`: Flexible box lookup (by ID, UUID, or name)
- `commandutils.GetVolume()`: Flexible volume lookup
- `commandutils.GetWorkspace()`: Flexible workspace lookup

These helpers try multiple strategies automatically.

## Database

- **SQLite**: Default for local development
- **PostgreSQL**: Production deployments
- **Transactions**: Use `querier.GetQuerier(ctx)` to get DB handle with transaction support
- **Soft deletes**: Resources are soft-deleted (marked as deleted, not removed)
