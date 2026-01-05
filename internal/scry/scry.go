package scry

import (
	"fmt"
	"os/exec"
	"strings"

	"github.com/emiliopalmerini/grimorio/internal/claude"
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

	return claude.Run(claude.Opus, prompt)
}
