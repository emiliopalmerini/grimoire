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

### Conjure (Scaffolding)

Generates vertical slice modules with CQRS pattern:

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
