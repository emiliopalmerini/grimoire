package memory

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/emiliopalmerini/grimorio/internal/claude"
	"github.com/emiliopalmerini/grimorio/internal/diff"
	"github.com/emiliopalmerini/grimorio/internal/editor"
	"github.com/emiliopalmerini/grimorio/internal/git"
	"github.com/emiliopalmerini/grimorio/internal/textutil"
)

func GetDiff(all bool) (string, error) {
	rawDiff, err := git.GetDiff(git.DiffOptions{All: all})
	if err != nil {
		return "", err
	}

	opts := diff.DefaultOptions()
	prioritized, err := diff.Prioritize(rawDiff, opts)
	if err != nil {
		// Fall back to raw diff on error
		return git.TruncateDiff(rawDiff, git.DefaultMaxDiffLines), nil
	}

	return diff.FormatForPrompt(prioritized), nil
}

func GetRecentCommits(n int) (string, error) {
	return git.GetRecentCommits(n, "%s")
}

func GenerateMessage(diff string, history string, description string) (string, error) {
	prompt := `Analyze this git diff and generate a conventional commit message with title and body.

Rules:
- Use conventional commits format for the title: type(scope): description
- Types: feat, fix, docs, style, refactor, test, chore
- Keep the title under 50 characters
- Add a blank line after the title
- Write a concise body explaining what changed and why
- Focus on the "why" not the "what"
- Do not use emojis
- Output ONLY the commit message (title + body), nothing else
`

	if history != "" {
		prompt += `
Recent commits (match this style):
` + history + `
`
	}

	if description != "" {
		prompt += `
User motivation:
` + description + `
`
	}

	prompt += `
Diff:
` + diff

	msg, err := claude.Run(claude.Haiku, "modify-memory", prompt)
	if err != nil {
		return "", err
	}
	if msg == "" {
		return "", fmt.Errorf("claude returned empty message")
	}

	msg = textutil.StripCodeBlock(msg)

	return msg, nil
}

func Confirm(message string) (bool, bool, error) {
	fmt.Println("\n" + message + "\n")
	fmt.Print("Commit with this message? [y/n/e(dit)] ")

	reader := bufio.NewReader(os.Stdin)
	input, err := reader.ReadString('\n')
	if err != nil {
		return false, false, err
	}

	input = strings.TrimSpace(strings.ToLower(input))
	switch input {
	case "y", "yes":
		return true, false, nil
	case "e", "edit":
		return false, true, nil
	default:
		return false, false, nil
	}
}

func EditMessage(message string) (string, error) {
	return editor.Edit(message, "commit-msg-*.txt")
}

func Commit(message string, stageAll bool) error {
	if stageAll {
		if err := git.StageAll(); err != nil {
			return err
		}
	}
	return git.Commit(message)
}
