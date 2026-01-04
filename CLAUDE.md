# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Build and Run

```bash
go build -o grimorio .
./grimorio [command]
```

## Setup

Enable git hooks for automatic vendorHash updates:

```bash
git config core.hooksPath .githooks
```

## Architecture

Grimorio is a multi-tool CLI for developer productivity, built with Cobra.

- **Cantrips**: Deterministic, code-only commands
- **Spells**: AI-powered commands using Claude Code

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

## Cantrips

### Summon (Project Initialization)

Creates a new Go project with standard structure:

```bash
grimorio summon myapp
grimorio summon myapi --type=api
grimorio summon mysite --type=web
grimorio summon myservice --type=grpc
grimorio summon hybrid --type=api --transport=http,grpc
```

Project types: `api` (default), `web` (sessions, CSRF, templ), `grpc`.

### Conjure (Module Scaffolding)

Generates vertical slice modules with CQRS pattern:

```bash
grimorio conjure user
grimorio conjure user --transport=http,grpc
grimorio conjure user --api=html --persistence=postgres
grimorio conjure order --transport=http,amqp
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
grimorio mend file.go
grimorio mend ./internal/...
grimorio mend --check .
grimorio mend --diff file.py
```

Supports: Go, Python, Rust, C#, TypeScript, JavaScript, HTML, JSON, YAML, Nix, Lua.

### Transmute (Format Conversion)

Converts data between formats: JSON, YAML, TOML, XML, CSV, Markdown, HTML.

```bash
grimorio transmute data.json --to yaml
grimorio transmute config.xml --to json
grimorio transmute users.json --to markdown
cat data.json | grimorio transmute --from json --to xml
```

All formats support both reading and writing.

## Spells

All spells use headless Claude Code (`claude -p`).

### Modify-Memory (AI Commit Generation)

```bash
grimorio modify-memory
grimorio modify-memory -a
```

Analyzes staged changes, fetches recent commit history for style matching, asks for motivation, and generates conventional commit messages.

### Divine (Code Explanation)

```bash
grimorio divine main.go
grimorio divine handler.go --symbol HandleLogin
```

Explains code in plain language. Optionally focus on a specific function/type.

### Scry (Code Review)

```bash
grimorio scry
grimorio scry -a
```

Reviews staged changes for bugs, security issues, and code problems.

### Augur (Error Analysis)

```bash
grimorio augur "go build"
grimorio augur "dotnet build"
```

Runs a command, captures output, and analyzes any errors with suggested fixes.
