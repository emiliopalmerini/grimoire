# Grimorio

A CLI spellbook for developer productivity.

## Cantrips vs Spells

- **Cantrips**: Deterministic, code-only commands (conjure, summon, mending, polymorph)
- **Spells**: AI-powered commands using Claude Code (modify-memory, sending, identify, scrying, augury)

## Installation

```bash
# With nix flake
nix profile install github:emiliopalmerini/grimorio

# Or build locally
go build -o grimorio .
```

## Cantrips

### summon

Create a new Go project with standard structure:

```bash
grimorio summon myapp
grimorio summon myapi --type=api
grimorio summon mysite --type=web
grimorio summon myservice --type=grpc
grimorio summon hybrid --type=api --transport=http,grpc
```

| Flag | Default | Description |
|------|---------|-------------|
| `--module, -m` | project name | Go module path |
| `--go-version, -g` | `1.25` | Go version for flake |
| `--type, -t` | `api` | Project type: `api`, `web`, `grpc` |
| `--transport, -T` | based on type | Transports: `http`, `grpc`, `amqp` |

### conjure

Scaffold vertical slice modules with CQRS pattern:

```bash
grimorio conjure user
grimorio conjure user --transport=http,grpc
grimorio conjure user --api=html --persistence=postgres
grimorio conjure order --transport=http,amqp
```

| Flag | Default | Description |
|------|---------|-------------|
| `--transport, -T` | `http` | Transport layers: `http`, `grpc`, `amqp` |
| `--api, -a` | `json` | API type: `json`, `html` |
| `--persistence, -p` | - | Persistence: `sqlite`, `postgres`, `mongodb` |

### polymorph

Transform data between formats:

```bash
grimorio polymorph data.json --to yaml
grimorio polymorph config.xml --to json
grimorio polymorph users.csv --to markdown
grimorio polymorph table.html --to json
cat data.json | grimorio polymorph --from json --to xml
```

Supported formats (all read/write):
- JSON, YAML, TOML, XML, CSV, Markdown, HTML

| Flag | Description |
|------|-------------|
| `--to, -t` | Output format (required) |
| `--from, -f` | Input format (auto-detected from extension) |
| `--output, -o` | Output file (default: stdout) |

### mending

Format files using LSP servers (organizes imports + formats):

```bash
grimorio mending file.go
grimorio mending ./internal/...
grimorio mending --check .
grimorio mending --diff file.py
```

Supported: Go, Python, Rust, C#, TypeScript, JavaScript, HTML, JSON, YAML, Nix, Lua

| Flag | Description |
|------|-------------|
| `--check, -c` | Check only, exit 1 if changes needed |
| `--diff, -d` | Show diff of changes |

## Spells

All spells require `claude` CLI to be installed and available in PATH.

### modify-memory

Generate commit messages from diffs using Claude Code:

```bash
grimorio modify-memory
grimorio modify-memory -a
grimorio modify-memory -n
```

| Flag | Description |
|------|-------------|
| `--all, -a` | Include all changes, not just staged |
| `--dry-run, -n` | Output message only, don't commit |

### sending

Generate PR description from branch changes:

```bash
grimorio sending
grimorio sending -m "Added authentication"
grimorio sending -n
grimorio sending --base develop
```

| Flag | Description |
|------|-------------|
| `--description, -m` | Additional context for the PR |
| `--dry-run, -n` | Output description without creating PR |
| `--base, -b` | Base branch to compare against |

### identify

Explain code in plain language:

```bash
grimorio identify main.go
grimorio identify internal/auth/auth.go
grimorio identify handler.go --symbol HandleLogin
```

| Flag | Description |
|------|-------------|
| `--symbol, -s` | Focus on a specific function/type |

### scrying

Review staged changes for bugs and issues:

```bash
grimorio scrying
grimorio scrying -a
```

| Flag | Description |
|------|-------------|
| `--all, -a` | Include all changes, not just staged |

### augury

Run a command and analyze errors:

```bash
grimorio augury "go build"
grimorio augury "npm test"
grimorio augury "dotnet build"
grimorio augury "cargo check"
```
