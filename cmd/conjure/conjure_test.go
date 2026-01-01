package conjure

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"
)

func TestConjureCommand(t *testing.T) {
	tmpDir := t.TempDir()

	if err := os.MkdirAll(filepath.Join(tmpDir, "internal"), 0755); err != nil {
		t.Fatalf("failed to create internal dir: %v", err)
	}

	goMod := `module github.com/test/project

go 1.21
`
	if err := os.WriteFile(filepath.Join(tmpDir, "go.mod"), []byte(goMod), 0644); err != nil {
		t.Fatalf("failed to create go.mod: %v", err)
	}

	originalWd, _ := os.Getwd()
	defer os.Chdir(originalWd)
	os.Chdir(tmpDir)

	cmd := Cmd
	buf := new(bytes.Buffer)
	cmd.SetOut(buf)
	cmd.SetErr(buf)
	cmd.SetArgs([]string{"user"})

	err := cmd.Execute()
	if err != nil {
		t.Fatalf("conjure command failed: %v", err)
	}

	if _, err := os.Stat(filepath.Join(tmpDir, "internal/user")); os.IsNotExist(err) {
		t.Error("expected user module to be created")
	}
}

func TestConjureCommandRequiresArg(t *testing.T) {
	cmd := Cmd
	buf := new(bytes.Buffer)
	cmd.SetOut(buf)
	cmd.SetErr(buf)
	cmd.SetArgs([]string{})

	err := cmd.Execute()
	if err == nil {
		t.Error("expected error when no module name provided")
	}
}

func TestConjureCommandFlags(t *testing.T) {
	tmpDir := t.TempDir()

	if err := os.MkdirAll(filepath.Join(tmpDir, "internal"), 0755); err != nil {
		t.Fatalf("failed to create internal dir: %v", err)
	}

	goMod := `module github.com/test/project

go 1.21
`
	if err := os.WriteFile(filepath.Join(tmpDir, "go.mod"), []byte(goMod), 0644); err != nil {
		t.Fatalf("failed to create go.mod: %v", err)
	}

	originalWd, _ := os.Getwd()
	defer os.Chdir(originalWd)
	os.Chdir(tmpDir)

	cmd := Cmd
	buf := new(bytes.Buffer)
	cmd.SetOut(buf)
	cmd.SetErr(buf)
	cmd.SetArgs([]string{"order", "--transport=http,amqp", "--api=html", "--crud=false"})

	err := cmd.Execute()
	if err != nil {
		t.Fatalf("conjure command failed: %v", err)
	}

	expectedPaths := []string{
		"internal/order/transport/http",
		"internal/order/transport/amqp",
		"internal/order/views",
	}

	for _, p := range expectedPaths {
		if _, err := os.Stat(filepath.Join(tmpDir, p)); os.IsNotExist(err) {
			t.Errorf("expected %s to exist", p)
		}
	}
}
