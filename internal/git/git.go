package git

import (
	"fmt"
	"os/exec"
	"strings"
)

type DiffOptions struct {
	All    bool
	Staged bool
}

func GetDiff(opts DiffOptions) (string, error) {
	var cmd *exec.Cmd
	if opts.All {
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
		if opts.All {
			return "", ErrNoChanges
		}
		cmd = exec.Command("git", "diff")
		out, err = cmd.Output()
		if err != nil {
			return "", fmt.Errorf("failed to get diff: %w", err)
		}
		if strings.TrimSpace(string(out)) != "" {
			return "", ErrNoStagedChanges
		}
		return "", ErrNoChanges
	}

	return diff, nil
}

func GetRecentCommits(n int, format string) (string, error) {
	if format == "" {
		format = "%s"
	}
	cmd := exec.Command("git", "log", fmt.Sprintf("-%d", n), "--pretty=format:"+format)
	out, err := cmd.Output()
	if err != nil {
		return "", nil
	}
	return strings.TrimSpace(string(out)), nil
}

func StageAll() error {
	cmd := exec.Command("git", "add", "-A")
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("staging failed: %w", err)
	}
	return nil
}

func Commit(message string) error {
	cmd := exec.Command("git", "commit", "-m", message)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("commit failed: %w", err)
	}
	return nil
}
