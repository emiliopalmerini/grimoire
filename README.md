# Grimoire

A CLI spellbook for developer productivity.

## Installation

```bash
# With nix flake
nix profile install github:emiliopalmerini/grimoire

# Or build locally
go build -o grimoire .
```

## Commands

### summon

Create a new Go project with standard structure:

```bash
grimoire summon myapp
grimoire summon myapi --type=api
grimoire summon mysite --type=web
grimoire summon myservice --type=grpc
grimoire summon hybrid --type=api --transport=http,grpc
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
grimoire conjure user
grimoire conjure user --transport=http,grpc
grimoire conjure user --api=html --persistence=postgres
grimoire conjure order --transport=http,amqp
```

| Flag | Default | Description |
|------|---------|-------------|
| `--transport, -T` | `http` | Transport layers: `http`, `grpc`, `amqp` |
| `--api, -a` | `json` | API type: `json`, `html` |
| `--persistence, -p` | - | Persistence: `sqlite`, `postgres`, `mongodb` |

### transmute

Convert data between formats:

```bash
grimoire transmute data.json --to yaml
grimoire transmute config.xml --to json
grimoire transmute users.csv --to markdown
grimoire transmute table.html --to json
cat data.json | grimoire transmute --from json --to xml
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
grimoire mend file.go
grimoire mend ./internal/...
grimoire mend --check .
grimoire mend --diff file.py
```

Supported: Go, Python, Rust, C#, TypeScript, JavaScript, HTML, JSON, YAML, Nix, Lua

| Flag | Description |
|------|-------------|
| `--check, -c` | Check only, exit 1 if changes needed |
| `--diff, -d` | Show diff of changes |
