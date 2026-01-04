# Grimorio

A CLI spellbook for developer productivity.

## Installation

```bash
# With nix flake
nix profile install github:emiliopalmerini/grimorio

# Or build locally
go build -o grimorio .
```

## Commands

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

### transmute

Convert data between formats:

```bash
grimorio transmute data.json --to yaml
grimorio transmute config.xml --to json
grimorio transmute users.csv --to markdown
grimorio transmute table.html --to json
cat data.json | grimorio transmute --from json --to xml
```

Supported formats (all read/write):
- JSON, YAML, TOML, XML, CSV, Markdown, HTML

| Flag | Description |
|------|-------------|
| `--to, -t` | Output format (required) |
| `--from, -f` | Input format (auto-detected from extension) |
| `--output, -o` | Output file (default: stdout) |

### mend

Format files using LSP servers (organizes imports + formats):

```bash
grimorio mend file.go
grimorio mend ./internal/...
grimorio mend --check .
grimorio mend --diff file.py
```

Supported: Go, Python, Rust, C#, TypeScript, JavaScript, HTML, JSON, YAML, Nix, Lua

| Flag | Description |
|------|-------------|
| `--check, -c` | Check only, exit 1 if changes needed |
| `--diff, -d` | Show diff of changes |
