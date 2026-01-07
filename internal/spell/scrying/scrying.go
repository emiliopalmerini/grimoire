package scrying

import (
	"github.com/emiliopalmerini/grimorio/internal/claude"
	"github.com/emiliopalmerini/grimorio/internal/git"
)

func GetDiff(all bool) (string, error) {
	return git.GetDiff(git.DiffOptions{All: all})
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

	return claude.Run(claude.Opus, "scrying", prompt)
}
