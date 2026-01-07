package editor

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
)

// Edit opens content in $EDITOR (defaults to vim) and returns the edited result.
// The tempPattern is used for the temporary file name (e.g., "commit-msg-*.txt").
func Edit(content, tempPattern string) (string, error) {
	editor := os.Getenv("EDITOR")
	if editor == "" {
		editor = "vim"
	}

	tmpfile, err := os.CreateTemp("", tempPattern)
	if err != nil {
		return "", err
	}
	defer os.Remove(tmpfile.Name())

	if _, err := tmpfile.WriteString(content); err != nil {
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
