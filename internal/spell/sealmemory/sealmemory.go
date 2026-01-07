package sealmemory

import (
	"fmt"
	"strings"

	"github.com/emiliopalmerini/grimorio/internal/claude"
	"github.com/emiliopalmerini/grimorio/internal/git"
	"github.com/emiliopalmerini/grimorio/internal/textutil"
)

func GetBranchInfo() (current, base string, err error) {
	current, err = git.GetCurrentBranch()
	if err != nil {
		return "", "", err
	}

	base, err = git.GetBaseBranch()
	if err != nil {
		return "", "", err
	}

	if current == base {
		return "", "", fmt.Errorf("already on %s branch, nothing to compare", base)
	}

	return current, base, nil
}

func GetBranchDiff(base string) (string, error) {
	return git.GetBranchDiff(base, git.DefaultMaxDiffLines)
}

func GetBranchCommits(base string) (string, error) {
	return git.GetBranchCommits(base)
}

func GeneratePRDescription(diff, commits, description string) (string, error) {
	prompt := `Analyze this git branch and generate a pull request title and description.

Rules:
- First line is the PR title: concise, under 72 characters
- Add a blank line after the title
- Write a "## Summary" section with 2-4 bullet points explaining the key changes
- Write a "## Changes" section briefly describing what was modified
- If applicable, add a "## Testing" section with suggested test steps
- Do not use emojis
- Output ONLY the PR title and description, nothing else
`

	if description != "" {
		prompt += `
User context:
` + description + `
`
	}

	if commits != "" {
		prompt += `
Branch commits:
` + commits + `
`
	}

	prompt += `
Diff:
` + diff

	msg, err := claude.Run(claude.Haiku, "seal-memory", prompt)
	if err != nil {
		return "", err
	}
	if msg == "" {
		return "", fmt.Errorf("claude returned empty message")
	}

	return textutil.StripCodeBlock(msg), nil
}

func ParseTitleBody(content string) (title, body string) {
	lines := strings.SplitN(content, "\n", 2)
	title = strings.TrimSpace(lines[0])
	if len(lines) > 1 {
		body = strings.TrimSpace(lines[1])
	}
	return title, body
}
