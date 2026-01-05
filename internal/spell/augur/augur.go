package augur

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"

	"github.com/emiliopalmerini/grimorio/internal/claude"
	"github.com/emiliopalmerini/grimorio/internal/git"
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

	history, _ := git.GetRecentCommits(5, "%h %s")
	if history != "" {
		prompt += "Recent commits:\n" + history + "\n\n"
	}

	diff, _ := git.GetDiff(git.DiffOptions{All: true})
	if diff != "" {
		prompt += "Current changes:\n" + diff
	}

	return claude.Run(claude.Sonnet, prompt)
}
