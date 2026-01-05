package git

import (
	"errors"
	"testing"
)

func TestErrors(t *testing.T) {
	if ErrNoChanges == nil {
		t.Error("ErrNoChanges should not be nil")
	}
	if ErrNoStagedChanges == nil {
		t.Error("ErrNoStagedChanges should not be nil")
	}

	if !errors.Is(ErrNoChanges, ErrNoChanges) {
		t.Error("ErrNoChanges should match itself with errors.Is")
	}
	if !errors.Is(ErrNoStagedChanges, ErrNoStagedChanges) {
		t.Error("ErrNoStagedChanges should match itself with errors.Is")
	}

	if errors.Is(ErrNoChanges, ErrNoStagedChanges) {
		t.Error("ErrNoChanges should not match ErrNoStagedChanges")
	}
}

func TestErrorMessages(t *testing.T) {
	if ErrNoChanges.Error() == "" {
		t.Error("ErrNoChanges should have a message")
	}
	if ErrNoStagedChanges.Error() == "" {
		t.Error("ErrNoStagedChanges should have a message")
	}
}
