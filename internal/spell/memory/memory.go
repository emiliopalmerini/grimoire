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

const maxDiffLines = 500

func GetDiff(all bool) (string, error) {
	return git.GetDiff(git.DiffOptions{All: all, MaxLines: maxDiffLines})
}

func GetRecentCommits(n int) (string, error) {
	return git.GetRecentCommits(n, "%s")
}

func GenerateMessage(diff string, history string, description string) (string, error) {
	prompt := `Generate a conventional commit message (type(scope): description + body).
Types: feat/fix/docs/style/refactor/test/chore. Title under 50 chars. No emojis. Output only the message.
`
	if history != "" {
		prompt += "\nRecent commits:\n" + history + "\n"
	}
	if description != "" {
		prompt += "\nContext: " + description + "\n"
	}
	prompt += "\nDiff:\n" + diff

	claude.SetCommand("modify-memory")
	msg, err := claude.Run(claude.Haiku, prompt)
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
