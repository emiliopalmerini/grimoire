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
