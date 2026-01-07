package git

import (
	"fmt"
	"os/exec"
	"strings"
)

// Repository defines the interface for git operations.
// This allows for mocking in tests.
type Repository interface {
	GetDiff(opts DiffOptions) (string, error)
	GetRecentCommits(n int, format string) (string, error)
	StageAll() error
	Commit(message string) error
}

// ExecRepository implements Repository using the actual git CLI.
type ExecRepository struct{}

func (r ExecRepository) GetDiff(opts DiffOptions) (string, error) {
	return GetDiff(opts)
}

func (r ExecRepository) GetRecentCommits(n int, format string) (string, error) {
	return GetRecentCommits(n, format)
}

func (r ExecRepository) StageAll() error {
	return StageAll()
}

func (r ExecRepository) Commit(message string) error {
	return Commit(message)
}

// DefaultRepository is the default Repository implementation.
var DefaultRepository Repository = ExecRepository{}

type DiffOptions struct {
	All      bool
	Staged   bool
	MaxLines int // 0 = unlimited
}

const DefaultMaxDiffLines = 500

func TruncateDiff(diff string, maxLines int) string {
	if maxLines <= 0 || diff == "" {
		return diff
	}

	lines := strings.Split(diff, "\n")
	if len(lines) <= maxLines {
		return diff
	}

	truncated := strings.Join(lines[:maxLines], "\n")
	omitted := len(lines) - maxLines
	return truncated + fmt.Sprintf("\n\n[... %d lines omitted, showing first %d of %d total lines ...]", omitted, maxLines, len(lines))
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

	if opts.MaxLines > 0 {
		diff = TruncateDiff(diff, opts.MaxLines)
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
