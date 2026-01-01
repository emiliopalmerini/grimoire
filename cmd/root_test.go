package cmd

import (
	"bytes"
	"testing"
)

func TestRootCommand(t *testing.T) {
	cmd := rootCmd
	buf := new(bytes.Buffer)
	cmd.SetOut(buf)
	cmd.SetErr(buf)
	cmd.SetArgs([]string{})

	err := cmd.Execute()
	if err != nil {
		t.Fatalf("root command failed: %v", err)
	}
}

func TestRootCommandHelp(t *testing.T) {
	cmd := rootCmd
	buf := new(bytes.Buffer)
	cmd.SetOut(buf)
	cmd.SetErr(buf)
	cmd.SetArgs([]string{"--help"})

	err := cmd.Execute()
	if err != nil {
		t.Fatalf("root command help failed: %v", err)
	}

	output := buf.String()
	if len(output) == 0 {
		t.Error("expected help output")
	}
}

func TestRootCommandHasConjure(t *testing.T) {
	found := false
	for _, cmd := range rootCmd.Commands() {
		if cmd.Use == "conjure [module-name]" {
			found = true
			break
		}
	}

	if !found {
		t.Error("expected conjure command to be registered")
	}
}
