package memory

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/emiliopalmerini/grimorio/internal/claude"
	"github.com/emiliopalmerini/grimorio/internal/git"
)

func GetDiff(all bool) (string, error) {
	return git.GetDiff(git.DiffOptions{All: all})
}

func GetRecentCommits(n int) (string, error) {
	return git.GetRecentCommits(n, "%s")
}

func AskDescription() (string, error) {
	fmt.Print("What's the motivation for this change? (optional): ")
	reader := bufio.NewReader(os.Stdin)
	input, err := reader.ReadString('\n')
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(input), nil
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

	claude.SetCommand("modify-memory")
	msg, err := claude.Run(claude.Sonnet, prompt)
	if err != nil {
		return "", err
	}
	if msg == "" {
		return "", fmt.Errorf("claude returned empty message")
	}

	msg = stripCodeBlock(msg)

	return msg, nil
}

func stripCodeBlock(s string) string {
	lines := strings.Split(s, "\n")
	if len(lines) >= 2 && strings.HasPrefix(lines[0], "```") && strings.HasSuffix(lines[len(lines)-1], "```") {
		return strings.TrimSpace(strings.Join(lines[1:len(lines)-1], "\n"))
	}
	return s
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
	editor := os.Getenv("EDITOR")
	if editor == "" {
		editor = "vim"
	}

	tmpfile, err := os.CreateTemp("", "commit-msg-*.txt")
	if err != nil {
		return "", err
	}
	defer os.Remove(tmpfile.Name())

	if _, err := tmpfile.WriteString(message); err != nil {
		return "", err
	}
	tmpfile.Close()

	cmd := exec.Command(editor, tmpfile.Name())
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("editor failed: %w", err)
	}

	edited, err := os.ReadFile(tmpfile.Name())
	if err != nil {
		return "", err
	}

	return strings.TrimSpace(string(edited)), nil
}

func Commit(message string, stageAll bool) error {
	if stageAll {
		if err := git.StageAll(); err != nil {
			return err
		}
	}
	return git.Commit(message)
}
