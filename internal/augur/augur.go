package augur

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"strings"
)

type Result struct {
	Command  string
	ExitCode int
	Stdout   string
	Stderr   string
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

func GetRecentCommits(n int) string {
	cmd := exec.Command("git", "log", fmt.Sprintf("-%d", n), "--pretty=format:%h %s")
	out, err := cmd.Output()
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(out))
}

func GetDiff() string {
	cmd := exec.Command("git", "diff", "HEAD")
	out, err := cmd.Output()
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(out))
}

func Analyze(result *Result) (string, error) {
	if result.ExitCode == 0 && result.Stderr == "" {
		return "Command succeeded with no errors.", nil
	}

	prompt := fmt.Sprintf(`Analyze this command output and explain any errors or warnings.
Suggest fixes if possible.

Command: %s
Exit code: %d

`, result.Command, result.ExitCode)

	if result.Stderr != "" {
		prompt += "Stderr:\n" + result.Stderr + "\n\n"
	}
	if result.Stdout != "" {
		prompt += "Stdout:\n" + result.Stdout + "\n\n"
	}

	history := GetRecentCommits(5)
	if history != "" {
		prompt += "Recent commits:\n" + history + "\n\n"
	}

	diff := GetDiff()
	if diff != "" {
		prompt += "Current changes:\n" + diff
	}

	cmd := exec.Command("claude", "-p", prompt)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("claude failed: %w\n%s", err, stderr.String())
	}

	return strings.TrimSpace(stdout.String()), nil
}
