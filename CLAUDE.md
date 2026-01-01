# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Build and Run

```bash
go build -o grimoire .
./grimoire [command]
```

## Architecture

Grimoire is a multi-tool CLI for developer productivity, built with Cobra.

### Structure

- `cmd/` - Cobra command definitions
  - `root.go` - Root command setup, registers all subcommands
  - `<command>/` - Each subcommand in its own package (e.g., `conjure/`)
- `internal/` - Implementation logic for each command

### Adding New Commands

1. Create a new package under `cmd/<command-name>/`
2. Define a `Cmd` variable of type `*cobra.Command`
3. Register it in `cmd/root.go` via `rootCmd.AddCommand()`
4. Place implementation logic in `internal/<feature>/`

### Conjure Module (Scaffolding)

The `conjure` command generates vertical slice modules with CQRS pattern:

```bash
grimoire conjure user --transport=http,amqp --api=html --crud
grimoire conjure order --transport=http --api=json
```

Generated structure includes:
- `commands/` and `queries/` with handler structs
- `transport/http/` with chi router setup, `transport/amqp/` for consumers
- `persistence/` with repository interface
- `views/` with HTMX templates when `--api=html`

Must be run from a Go project root containing `internal/` and `go.mod`.
