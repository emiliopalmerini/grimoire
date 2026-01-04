package scry

import (
	"bytes"
	"fmt"
	"os/exec"
	"strings"
)

func GetDiff(all bool) (string, error) {
	var cmd *exec.Cmd
	if all {
		cmd = exec.Command("git", "diff", "HEAD")
	} else {
		cmd = exec.Command("git", "diff", "--cached")
	}

	out, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("failed to get diff: %w", err)
	}

	diff := strings.TrimSpace(string(out))
	if diff == "" {
		if all {
			return "", fmt.Errorf("no changes to review")
		}
		return "", fmt.Errorf("no staged changes (use -a to include unstaged changes)")
	}

	return diff, nil
}

func Review(diff string) (string, error) {
	prompt := `Review this git diff for potential issues. Look for:
- Bugs or logic errors
- Security vulnerabilities
- Performance issues
- Code style problems
- Missing error handling
- Edge cases not handled

Be concise. If the code looks good, say so briefly.
If there are issues, list them with file and context.

Diff:
` + diff

	cmd := exec.Command("claude", "-p", prompt)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("claude failed: %w\n%s", err, stderr.String())
	}

	return strings.TrimSpace(stdout.String()), nil
}
