package memory

import (
	"bufio"
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"strings"
)

type Options struct {
	All    bool
	DryRun bool
}

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
			return "", fmt.Errorf("no changes to commit")
		}
		cmd = exec.Command("git", "diff")
		out, err = cmd.Output()
		if err != nil {
			return "", fmt.Errorf("failed to get diff: %w", err)
		}
		diff = strings.TrimSpace(string(out))
		if diff == "" {
			return "", fmt.Errorf("no changes to commit")
		}
		return "", fmt.Errorf("no staged changes (use -a to include unstaged changes)")
	}

	return diff, nil
}

func GetRecentCommits(n int) (string, error) {
	cmd := exec.Command("git", "log", fmt.Sprintf("-%d", n), "--pretty=format:%s")
	out, err := cmd.Output()
	if err != nil {
		return "", nil // not fatal, just skip history
	}
	return strings.TrimSpace(string(out)), nil
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

	cmd := exec.Command("claude", "-p", "--no-session-persistence", "--model", "sonnet", prompt)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("claude failed: %w\n%s", err, stderr.String())
	}

	msg := strings.TrimSpace(stdout.String())
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
		add := exec.Command("git", "add", "-A")
		add.Stdout = os.Stdout
		add.Stderr = os.Stderr
		if err := add.Run(); err != nil {
			return fmt.Errorf("staging failed: %w", err)
		}
	}

	cmd := exec.Command("git", "commit", "-m", message)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("commit failed: %w", err)
	}
	return nil
}
