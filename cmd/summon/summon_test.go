package summon

import (
	"testing"
)

func TestCmdRequiresProjectName(t *testing.T) {
	Cmd.SetArgs([]string{})
	err := Cmd.Execute()
	if err == nil {
		t.Error("expected error when no project name provided")
	}
}

func TestCmdFlagsExist(t *testing.T) {
	flags := []string{"module", "go-version", "type", "transport"}
	for _, flag := range flags {
		if Cmd.Flags().Lookup(flag) == nil {
			t.Errorf("expected --%s flag to exist", flag)
		}
	}
}

func TestCmdDefaultValues(t *testing.T) {
	tests := []struct {
		flag     string
		expected string
	}{
		{"go-version", "1.25"},
		{"type", "api"},
	}

	for _, tt := range tests {
		flag := Cmd.Flags().Lookup(tt.flag)
		if flag.DefValue != tt.expected {
			t.Errorf("expected default %s to be %s, got %s", tt.flag, tt.expected, flag.DefValue)
		}
	}
}
