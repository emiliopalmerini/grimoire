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

go 1.21
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
		WithCRUD:   true,
	}

	if err := CreateModule(opts); err != nil {
		t.Fatalf("CreateModule failed: %v", err)
	}

	expectedDirs := []string{
		"internal/user",
		"internal/user/commands",
		"internal/user/queries",
		"internal/user/transport/http",
	}

	for _, dir := range expectedDirs {
		path := filepath.Join(tmpDir, dir)
		if _, err := os.Stat(path); os.IsNotExist(err) {
			t.Errorf("expected directory %s to exist", dir)
		}
	}

	expectedFiles := []string{
		"internal/user/models.go",
		"internal/user/service.go",
		"internal/user/repository.go",
		"internal/user/commands/create_user.go",
		"internal/user/commands/update_user.go",
		"internal/user/commands/delete_user.go",
		"internal/user/queries/get_user.go",
		"internal/user/queries/list_user.go",
		"internal/user/transport/http/handler.go",
		"internal/user/transport/http/routes.go",
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

go 1.21
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
		WithCRUD:   true,
	}

	if err := CreateModule(opts); err != nil {
		t.Fatalf("CreateModule failed: %v", err)
	}

	expectedFiles := []string{
		"internal/product/views/index.html",
		"internal/product/views/form.html",
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

go 1.21
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
		WithCRUD:   false,
	}

	if err := CreateModule(opts); err != nil {
		t.Fatalf("CreateModule failed: %v", err)
	}

	expectedFiles := []string{
		"internal/order/transport/http/handler.go",
		"internal/order/transport/amqp/consumer.go",
	}

	for _, file := range expectedFiles {
		path := filepath.Join(tmpDir, file)
		if _, err := os.Stat(path); os.IsNotExist(err) {
			t.Errorf("expected file %s to exist", file)
		}
	}

	unexpectedDirs := []string{
		"internal/order/commands",
		"internal/order/queries",
	}

	for _, dir := range unexpectedDirs {
		path := filepath.Join(tmpDir, dir)
		if _, err := os.Stat(path); !os.IsNotExist(err) {
			t.Errorf("expected directory %s to not exist when WithCRUD=false", dir)
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
		WithCRUD:   true,
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
		WithCRUD:   true,
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

go 1.21
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
		WithCRUD:    true,
		Persistence: "sqlite",
	}

	if err := CreateModule(opts); err != nil {
		t.Fatalf("CreateModule failed: %v", err)
	}

	expectedFiles := []string{
		"internal/user/repository.go",
		"internal/user/persistence/sqlite_repository.go",
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

go 1.21
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
		WithCRUD:    true,
		Persistence: "postgres",
	}

	if err := CreateModule(opts); err != nil {
		t.Fatalf("CreateModule failed: %v", err)
	}

	expectedFiles := []string{
		"internal/user/repository.go",
		"internal/user/persistence/postgres_repository.go",
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

go 1.21
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
		WithCRUD:    true,
		Persistence: "mongodb",
	}

	if err := CreateModule(opts); err != nil {
		t.Fatalf("CreateModule failed: %v", err)
	}

	expectedFiles := []string{
		"internal/user/repository.go",
		"internal/user/persistence/mongo_repository.go",
	}

	for _, file := range expectedFiles {
		path := filepath.Join(tmpDir, file)
		if _, err := os.Stat(path); os.IsNotExist(err) {
			t.Errorf("expected file %s to exist", file)
		}
	}
}
