package claude

import (
	"bytes"
	"context"
	"fmt"
	"os/exec"
	"strings"
	"time"

	"github.com/emiliopalmerini/grimorio/internal/metrics"
)

type Model string

const (
	Haiku  Model = "haiku"
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

var currentCommand string

func SetCommand(cmd string) {
	currentCommand = cmd
}

func Run(model Model, prompt string) (string, error) {
	start := time.Now()
	cmd := exec.Command("claude", "-p", "--no-session-persistence", "--model", string(model), prompt)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()
	latency := time.Since(start).Milliseconds()
	response := strings.TrimSpace(stdout.String())

	cmdName := currentCommand
	if cmdName == "" {
		cmdName = "unknown"
	}

	if err != nil {
		metrics.Default.RecordAI(context.Background(), cmdName, string(model), len(prompt), 0, latency, false, err.Error())
		return "", fmt.Errorf("claude failed: %w\n%s", err, stderr.String())
	}

	metrics.Default.RecordAI(context.Background(), cmdName, string(model), len(prompt), len(response), latency, true, "")
	return response, nil
}
