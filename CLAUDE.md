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

### Mending (LSP Formatting)

Formats files using LSP servers. Organizes imports and formats code.

```bash
grimorio mending file.go
grimorio mending ./internal/...
grimorio mending --check .
grimorio mending --diff file.py
```

Supports: Go, Python, Rust, C#, TypeScript, JavaScript, HTML, JSON, YAML, Nix, Lua.

### Polymorph (Format Conversion)

Transforms data between formats: JSON, YAML, TOML, XML, CSV, Markdown, HTML.

```bash
grimorio polymorph data.json --to yaml
grimorio polymorph config.xml --to json
grimorio polymorph users.json --to markdown
cat data.json | grimorio polymorph --from json --to xml
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

### Sending (AI PR Description)

```bash
grimorio sending
grimorio sending -m "Added authentication"
grimorio sending -n
grimorio sending --base develop
```

Compares current branch against main/master, analyzes commits and diff, generates PR title and description. Options: create PR via `gh`, edit in `$EDITOR`, or copy to clipboard.

### Identify (Code Explanation)

```bash
grimorio identify main.go
grimorio identify handler.go --symbol HandleLogin
```

Explains code in plain language. Optionally focus on a specific function/type.

### Scrying (Code Review)

```bash
grimorio scrying
grimorio scrying -a
```

Reviews staged changes for bugs, security issues, and code problems.

### Augury (Error Analysis)

```bash
grimorio augury "go build"
grimorio augury "dotnet build"
```

Runs a command, captures output, and analyzes any errors with suggested fixes.
