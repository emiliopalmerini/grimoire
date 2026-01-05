package git

import "errors"

var (
	ErrNoChanges       = errors.New("no changes to commit")
	ErrNoStagedChanges = errors.New("no staged changes (use -a to include unstaged changes)")
)
