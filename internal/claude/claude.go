package claude

import (
	"bytes"
	"fmt"
	"os/exec"
	"strings"
)

type Model string

const (
	Sonnet Model = "sonnet"
	Opus   Model = "opus"
)

// Runner defines the interface for running Claude prompts.
// This allows for mocking in tests.
type Runner interface {
	Run(model Model, prompt string) (string, error)
}

// ExecRunner implements Runner using the actual Claude CLI.
type ExecRunner struct{}

func (r ExecRunner) Run(model Model, prompt string) (string, error) {
	return Run(model, prompt)
}

// DefaultRunner is the default Runner implementation.
var DefaultRunner Runner = ExecRunner{}

func Run(model Model, prompt string) (string, error) {
	cmd := exec.Command("claude", "-p", "--no-session-persistence", "--model", string(model), prompt)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("claude failed: %w\n%s", err, stderr.String())
	}

	return strings.TrimSpace(stdout.String()), nil
}
