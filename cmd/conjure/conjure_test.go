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

go 1.25
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

	expectedFiles := []string{
		"internal/user/models.go",
		"internal/user/service.go",
		"internal/user/repository.go",
		"internal/user/handler.go",
		"internal/user/routes.go",
	}

	for _, f := range expectedFiles {
		if _, err := os.Stat(filepath.Join(tmpDir, f)); os.IsNotExist(err) {
			t.Errorf("expected %s to exist", f)
		}
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

func TestConjureCommandWithHTMLViews(t *testing.T) {
	tmpDir := t.TempDir()

	if err := os.MkdirAll(filepath.Join(tmpDir, "internal"), 0755); err != nil {
		t.Fatalf("failed to create internal dir: %v", err)
	}

	goMod := `module github.com/test/project

go 1.25
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
	cmd.SetArgs([]string{"order", "--api=html"})

	err := cmd.Execute()
	if err != nil {
		t.Fatalf("conjure command failed: %v", err)
	}

	expectedPaths := []string{
		"internal/order/views/index.templ",
		"internal/order/views/form.templ",
	}

	for _, p := range expectedPaths {
		if _, err := os.Stat(filepath.Join(tmpDir, p)); os.IsNotExist(err) {
			t.Errorf("expected %s to exist", p)
		}
	}
}

func TestConjureCommandFlags(t *testing.T) {
	flags := []string{"transport", "api", "persistence"}
	for _, flag := range flags {
		if Cmd.Flags().Lookup(flag) == nil {
			t.Errorf("expected --%s flag to exist", flag)
		}
	}
}
