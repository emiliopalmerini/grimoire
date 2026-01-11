package scrying

import (
	"github.com/emiliopalmerini/grimorio/internal/claude"
	"github.com/emiliopalmerini/grimorio/internal/diff"
	"github.com/emiliopalmerini/grimorio/internal/git"
)

// maxHighPriorityLines is the line budget for scrying.
// Lower than modify-memory since scrying uses Opus which is more expensive.
const maxHighPriorityLines = 300

func GetDiff(all bool) (string, error) {
	rawDiff, err := git.GetDiff(git.DiffOptions{All: all})
	if err != nil {
		return "", err
	}

	opts := diff.DefaultOptions()
	opts.MaxHighPriorityLines = maxHighPriorityLines
	prioritized, err := diff.Prioritize(rawDiff, opts)
	if err != nil {
		// Fall back to truncated diff on error
		return git.TruncateDiff(rawDiff, maxHighPriorityLines), nil
	}

	return diff.FormatForPrompt(prioritized), nil
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
