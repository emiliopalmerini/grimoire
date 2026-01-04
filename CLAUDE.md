# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Build and Run

```bash
go build -o grimoire .
./grimoire [command]
```

## Setup

Enable git hooks for automatic vendorHash updates:

```bash
git config core.hooksPath .githooks
```

## Architecture

Grimoire is a multi-tool CLI for developer productivity, built with Cobra.

### Structure

- `cmd/` - Cobra command definitions
  - `root.go` - Root command setup, registers all subcommands
  - `<command>/` - Each subcommand in its own package (e.g., `conjure/`)
- `internal/` - Implementation logic for each command
- `.githooks/` - Git hooks (pre-commit updates vendorHash when go.mod changes)

### Adding New Commands

1. Create a new package under `cmd/<command-name>/`
2. Define a `Cmd` variable of type `*cobra.Command`
3. Register it in `cmd/root.go` via `rootCmd.AddCommand()`
4. Place implementation logic in `internal/<feature>/`

## Commands

### Summon (Project Initialization)

Creates a new Go project with standard structure:

```bash
grimoire summon myapp
grimoire summon myapi --type=api
grimoire summon mysite --type=web
grimoire summon myservice --type=grpc
grimoire summon hybrid --type=api --transport=http,grpc
```

Project types: `api` (default), `web` (sessions, CSRF, templ), `grpc`.

### Conjure (Module Scaffolding)

Generates vertical slice modules with CQRS pattern:

```bash
grimoire conjure user
grimoire conjure user --transport=http,grpc
grimoire conjure user --api=html --persistence=postgres
grimoire conjure order --transport=http,amqp
```

Generated structure includes:
- `commands/` and `queries/` with handler structs
- `transport/http/` with chi router, `transport/grpc/` for gRPC, `transport/amqp/` for consumers
- `persistence/` with repository interface
- `views/` with templ templates when `--api=html`

Must be run from a Go project root containing `internal/` and `go.mod`.

### Mend (LSP Formatting)

Formats files using LSP servers. Organizes imports and formats code.

```bash
grimoire mend file.go
grimoire mend ./internal/...
grimoire mend --check .
grimoire mend --diff file.py
```

Supports: Go, Python, Rust, C#, TypeScript, JavaScript, HTML, JSON, YAML, Nix, Lua.

### Transmute (Format Conversion)

Converts data between formats: JSON, YAML, TOML, XML, CSV, Markdown, HTML.

```bash
grimoire transmute data.json --to yaml
grimoire transmute config.xml --to json
grimoire transmute users.json --to markdown
cat data.json | grimoire transmute --from json --to xml
```

All formats support both reading and writing.
