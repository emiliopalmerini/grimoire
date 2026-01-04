package scaffold

import (
	"os"
	"path/filepath"
	"testing"
)

func TestToPascalCase(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"user", "User"},
		{"user_profile", "UserProfile"},
		{"order_item_detail", "OrderItemDetail"},
		{"a", "A"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := toPascalCase(tt.input)
			if result != tt.expected {
				t.Errorf("toPascalCase(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}

func TestCreateModule(t *testing.T) {
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

	opts := ModuleOptions{
		Name:       "user",
		Transports: []string{"http"},
		APIType:    "json",
	}

	if err := CreateModule(opts); err != nil {
		t.Fatalf("CreateModule failed: %v", err)
	}

	expectedFiles := []string{
		"internal/user/models.go",
		"internal/user/service.go",
		"internal/user/repository.go",
		"internal/user/handler.go",
		"internal/user/routes.go",
	}

	for _, file := range expectedFiles {
		path := filepath.Join(tmpDir, file)
		if _, err := os.Stat(path); os.IsNotExist(err) {
			t.Errorf("expected file %s to exist", file)
		}
	}
}

func TestCreateModuleWithHTML(t *testing.T) {
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

	opts := ModuleOptions{
		Name:       "product",
		Transports: []string{"http"},
		APIType:    "html",
	}

	if err := CreateModule(opts); err != nil {
		t.Fatalf("CreateModule failed: %v", err)
	}

	expectedFiles := []string{
		"internal/product/views/index.templ",
		"internal/product/views/form.templ",
	}

	for _, file := range expectedFiles {
		path := filepath.Join(tmpDir, file)
		if _, err := os.Stat(path); os.IsNotExist(err) {
			t.Errorf("expected file %s to exist", file)
		}
	}
}

func TestCreateModuleWithAMQP(t *testing.T) {
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

	opts := ModuleOptions{
		Name:       "order",
		Transports: []string{"http", "amqp"},
		APIType:    "json",
	}

	if err := CreateModule(opts); err != nil {
		t.Fatalf("CreateModule failed: %v", err)
	}

	expectedFiles := []string{
		"internal/order/handler.go",
		"internal/order/consumer.go",
	}

	for _, file := range expectedFiles {
		path := filepath.Join(tmpDir, file)
		if _, err := os.Stat(path); os.IsNotExist(err) {
			t.Errorf("expected file %s to exist", file)
		}
	}
}

func TestCreateModuleNoInternalDir(t *testing.T) {
	tmpDir := t.TempDir()

	originalWd, _ := os.Getwd()
	defer os.Chdir(originalWd)
	os.Chdir(tmpDir)

	opts := ModuleOptions{
		Name:       "user",
		Transports: []string{"http"},
		APIType:    "json",
	}

	err := CreateModule(opts)
	if err == nil {
		t.Error("expected error when internal/ directory doesn't exist")
	}
}

func TestCreateModuleNoGoMod(t *testing.T) {
	tmpDir := t.TempDir()

	if err := os.MkdirAll(filepath.Join(tmpDir, "internal"), 0755); err != nil {
		t.Fatalf("failed to create internal dir: %v", err)
	}

	originalWd, _ := os.Getwd()
	defer os.Chdir(originalWd)
	os.Chdir(tmpDir)

	opts := ModuleOptions{
		Name:       "user",
		Transports: []string{"http"},
		APIType:    "json",
	}

	err := CreateModule(opts)
	if err == nil {
		t.Error("expected error when go.mod doesn't exist")
	}
}

func TestCreateModuleWithSQLite(t *testing.T) {
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

	opts := ModuleOptions{
		Name:        "user",
		Transports:  []string{"http"},
		APIType:     "json",
		Persistence: "sqlite",
	}

	if err := CreateModule(opts); err != nil {
		t.Fatalf("CreateModule failed: %v", err)
	}

	expectedFiles := []string{
		"internal/user/repository.go",
		"internal/user/persistence/sqlite.go",
	}

	for _, file := range expectedFiles {
		path := filepath.Join(tmpDir, file)
		if _, err := os.Stat(path); os.IsNotExist(err) {
			t.Errorf("expected file %s to exist", file)
		}
	}
}

func TestCreateModuleWithPostgres(t *testing.T) {
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

	opts := ModuleOptions{
		Name:        "user",
		Transports:  []string{"http"},
		APIType:     "json",
		Persistence: "postgres",
	}

	if err := CreateModule(opts); err != nil {
		t.Fatalf("CreateModule failed: %v", err)
	}

	expectedFiles := []string{
		"internal/user/repository.go",
		"internal/user/persistence/postgres.go",
	}

	for _, file := range expectedFiles {
		path := filepath.Join(tmpDir, file)
		if _, err := os.Stat(path); os.IsNotExist(err) {
			t.Errorf("expected file %s to exist", file)
		}
	}
}

func TestCreateModuleWithMongoDB(t *testing.T) {
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

	opts := ModuleOptions{
		Name:        "user",
		Transports:  []string{"http"},
		APIType:     "json",
		Persistence: "mongodb",
	}

	if err := CreateModule(opts); err != nil {
		t.Fatalf("CreateModule failed: %v", err)
	}

	expectedFiles := []string{
		"internal/user/repository.go",
		"internal/user/persistence/mongo.go",
	}

	for _, file := range expectedFiles {
		path := filepath.Join(tmpDir, file)
		if _, err := os.Stat(path); os.IsNotExist(err) {
			t.Errorf("expected file %s to exist", file)
		}
	}
}

func TestCreateModuleWithGRPC(t *testing.T) {
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

	opts := ModuleOptions{
		Name:       "user",
		Transports: []string{"grpc"},
		APIType:    "json",
	}

	if err := CreateModule(opts); err != nil {
		t.Fatalf("CreateModule failed: %v", err)
	}

	expectedFiles := []string{
		"internal/user/grpc_server.go",
		"api/proto/v1/user.proto",
	}

	for _, file := range expectedFiles {
		path := filepath.Join(tmpDir, file)
		if _, err := os.Stat(path); os.IsNotExist(err) {
			t.Errorf("expected file %s to exist", file)
		}
	}
}
