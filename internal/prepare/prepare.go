package prepare

import (
	"fmt"
	"os"
	"path/filepath"
)

// CommandType represents the type of grimorio command
type CommandType string

const (
	Spell   CommandType = "spell"
	Cantrip CommandType = "cantrip"
)

// Command holds metadata for a grimorio command
type Command struct {
	Name        string
	Type        CommandType
	Short       string
	Description string
	Usage       string
}

// Commands returns all available grimorio commands
func Commands() []Command {
	return []Command{
		{
			Name:  "augury",
			Type:  Spell,
			Short: "Run a command and analyze errors",
			Description: `Augury runs a command, captures its output, and analyzes any errors using Claude.
Use this when you need to run a shell command and get AI-powered analysis of any failures or errors.`,
			Usage: `grimorio augury "go build"
grimorio augury "npm test"`,
		},
		{
			Name:  "conjure",
			Type:  Cantrip,
			Short: "Conjure a new vertical slice module",
			Description: `Conjure creates a new module with the vertical slice architecture structure.
Use this to scaffold CQRS modules with commands, queries, handlers, and transport layers.`,
			Usage: `grimorio conjure user
grimorio conjure order --transport=http,grpc`,
		},
		{
			Name:  "identify",
			Type:  Spell,
			Short: "Explain code in plain language",
			Description: `Identify reads a file and explains its code using Claude.
Use this when you need to understand what a piece of code does.`,
			Usage: `grimorio identify main.go
grimorio identify handler.go --symbol HandleLogin`,
		},
		{
			Name:  "mending",
			Type:  Cantrip,
			Short: "Format files using LSP",
			Description: `Mending formats files using language server protocol (LSP) formatters.
Use this to format and organize imports in source files.`,
			Usage: `grimorio mending file.go
grimorio mending ./internal/...
grimorio mending --check .`,
		},
		{
			Name:  "modify-memory",
			Type:  Spell,
			Short: "Generate commits from diffs using Claude",
			Description: `Modify-memory analyzes your git changes and generates conventional commit messages using Claude.
Use this when you want to commit changes with an AI-generated message.`,
			Usage: `grimorio modify-memory
grimorio modify-memory -a`,
		},
		{
			Name:  "polymorph",
			Type:  Cantrip,
			Short: "Transform data between formats",
			Description: `Polymorph transforms data between different formats: JSON, YAML, TOML, XML, CSV, Markdown, HTML.
Use this to convert data files from one format to another.`,
			Usage: `grimorio polymorph data.json --to yaml
grimorio polymorph config.xml --to json`,
		},
		{
			Name:  "scrying",
			Type:  Spell,
			Short: "Review staged changes for bugs/issues",
			Description: `Scrying analyzes your staged changes and reviews them for potential issues using Claude.
Use this for AI-powered code review before committing.`,
			Usage: `grimorio scrying
grimorio scrying -a`,
		},
		{
			Name:  "sending",
			Type:  Spell,
			Short: "Generate PR description from branch changes",
			Description: `Sending analyzes your branch commits and generates a pull request description using Claude.
Use this to create PR descriptions with AI assistance.`,
			Usage: `grimorio sending
grimorio sending --base develop`,
		},
		{
			Name:  "summon",
			Type:  Cantrip,
			Short: "Summon a new Go project",
			Description: `Summon creates a new Go project with a standard structure.
Use this to bootstrap new Go projects with proper scaffolding.`,
			Usage: `grimorio summon myapp
grimorio summon myapi --type=api
grimorio summon mysite --type=web`,
		},
		{
			Name:  "stats",
			Type:  Cantrip,
			Short: "View usage statistics",
			Description: `Stats displays usage statistics and insights for grimorio commands.
Use this to see how you've been using grimorio.`,
			Usage: `grimorio stats
grimorio stats --days 7`,
		},
	}
}

// GenerateSkill creates a SKILL.md file for a command
func GenerateSkill(cmd Command) string {
	typeLabel := "Cantrip"
	if cmd.Type == Spell {
		typeLabel = "Spell"
	}

	return fmt.Sprintf(`---
name: grimorio-%s
description: "[%s] %s. Use when the user wants to %s."
---

# Grimorio %s

%s

## Usage

%s

## Instructions

Run the grimorio %s command with the appropriate arguments based on user needs.
`, cmd.Name, typeLabel, cmd.Short, cmd.Short, cmd.Name, cmd.Description, "```bash\n"+cmd.Usage+"\n```", cmd.Name)
}

// Location represents where to install skills
type Location string

const (
	Personal Location = "personal"
	Project  Location = "project"
)

// Install writes skill files to the appropriate location
func Install(loc Location, commands []Command) error {
	var basePath string

	switch loc {
	case Personal:
		home, err := os.UserHomeDir()
		if err != nil {
			return fmt.Errorf("could not get home directory: %w", err)
		}
		basePath = filepath.Join(home, ".claude", "skills")
	case Project:
		basePath = filepath.Join(".claude", "skills")
	default:
		return fmt.Errorf("unknown location: %s", loc)
	}

	for _, cmd := range commands {
		skillDir := filepath.Join(basePath, "grimorio-"+cmd.Name)
		if err := os.MkdirAll(skillDir, 0755); err != nil {
			return fmt.Errorf("could not create directory %s: %w", skillDir, err)
		}

		skillPath := filepath.Join(skillDir, "SKILL.md")
		content := GenerateSkill(cmd)

		if err := os.WriteFile(skillPath, []byte(content), 0644); err != nil {
			return fmt.Errorf("could not write %s: %w", skillPath, err)
		}

		fmt.Printf("Created %s\n", skillPath)
	}

	return nil
}
