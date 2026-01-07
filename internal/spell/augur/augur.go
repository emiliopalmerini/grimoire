package augur

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/emiliopalmerini/grimorio/internal/claude"
	"github.com/emiliopalmerini/grimorio/internal/git"
)

const (
	maxOutputLines = 100
	maxDiffLines   = 200
)

type Result struct {
	Command  string
	ExitCode int
	Stdout   string
	Stderr   string
}

func truncateTail(s string, maxLines int) string {
	if s == "" || maxLines <= 0 {
		return s
	}
	lines := strings.Split(s, "\n")
	if len(lines) <= maxLines {
		return s
	}
	omitted := len(lines) - maxLines
	kept := lines[omitted:]
	return fmt.Sprintf("[... %d lines omitted, showing last %d ...]\n", omitted, maxLines) + strings.Join(kept, "\n")
}

func RunCommand(command string) (*Result, error) {
	cmd := exec.Command("sh", "-c", command)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	cmd.Env = os.Environ()

	err := cmd.Run()
	exitCode := 0
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			exitCode = exitErr.ExitCode()
		}
	}

	return &Result{
		Command:  command,
		ExitCode: exitCode,
		Stdout:   stdout.String(),
		Stderr:   stderr.String(),
	}, nil
}

func Analyze(result *Result) (string, error) {
	if result.ExitCode == 0 && result.Stderr == "" {
		return "Command succeeded with no errors.", nil
	}

	prompt := fmt.Sprintf("Analyze errors/warnings and suggest fixes.\n\nCommand: %s\nExit code: %d\n\n", result.Command, result.ExitCode)

	if result.Stderr != "" {
		prompt += "Stderr:\n" + truncateTail(result.Stderr, maxOutputLines) + "\n\n"
	}
	if result.Stdout != "" {
		prompt += "Stdout:\n" + truncateTail(result.Stdout, maxOutputLines) + "\n\n"
	}

	errorOutput := result.Stderr + result.Stdout
	if looksCodeRelated(errorOutput) {
		diff, _ := git.GetDiff(git.DiffOptions{All: true, MaxLines: maxDiffLines})
		if diff != "" {
			prompt += "Recent changes:\n" + diff
		}
	}

	claude.SetCommand("augur")
	return claude.Run(claude.Sonnet, prompt)
}

func looksCodeRelated(output string) bool {
	codePatterns := []string{
		"undefined", "not found", "cannot find", "no such file",
		"syntax error", "unexpected", "expected", "undeclared",
		"import", "package", "module", "type", "func", "class",
		".go:", ".py:", ".ts:", ".js:", ".rs:", ".java:",
	}
	lower := strings.ToLower(output)
	for _, p := range codePatterns {
		if strings.Contains(lower, p) {
			return true
		}
	}
	return false
}
