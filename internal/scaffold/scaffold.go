package scaffold

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

type ModuleOptions struct {
	Name        string
	Transports  []string
	APIType     string
	WithCRUD    bool
	Persistence string
}

func CreateModule(opts ModuleOptions) error {
	internalDir, err := findInternalDir()
	if err != nil {
		return err
	}

	modulePath, err := findGoModulePath()
	if err != nil {
		return err
	}

	moduleDir := filepath.Join(internalDir, opts.Name)

	if err := createDirectories(moduleDir, opts); err != nil {
		return err
	}

	if err := createFiles(moduleDir, opts, modulePath); err != nil {
		return err
	}

	return nil
}

func findInternalDir() (string, error) {
	cwd, err := os.Getwd()
	if err != nil {
		return "", fmt.Errorf("failed to get current directory: %w", err)
	}

	internalDir := filepath.Join(cwd, "internal")
	if _, err := os.Stat(internalDir); os.IsNotExist(err) {
		return "", fmt.Errorf("internal/ directory not found in %s", cwd)
	}

	return internalDir, nil
}

func findGoModulePath() (string, error) {
	cwd, err := os.Getwd()
	if err != nil {
		return "", fmt.Errorf("failed to get current directory: %w", err)
	}

	goModPath := filepath.Join(cwd, "go.mod")
	file, err := os.Open(goModPath)
	if err != nil {
		return "", fmt.Errorf("go.mod not found in %s", cwd)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		if after, ok := strings.CutPrefix(line, "module "); ok {
			return after, nil
		}
	}

	return "", fmt.Errorf("module path not found in go.mod")
}

func createDirectories(moduleDir string, opts ModuleOptions) error {
	dirs := []string{
		moduleDir,
	}

	if opts.Persistence != "" {
		dirs = append(dirs, filepath.Join(moduleDir, "persistence"))
	}

	if opts.WithCRUD {
		dirs = append(dirs, filepath.Join(moduleDir, "commands"))
		dirs = append(dirs, filepath.Join(moduleDir, "queries"))
	}

	for _, transport := range opts.Transports {
		dirs = append(dirs, filepath.Join(moduleDir, "transport", transport))
	}

	if opts.APIType == "html" {
		dirs = append(dirs, filepath.Join(moduleDir, "views"))
	}

	for _, dir := range dirs {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return fmt.Errorf("failed to create directory %s: %w", dir, err)
		}
	}

	return nil
}

func createFiles(moduleDir string, opts ModuleOptions, goModulePath string) error {
	name := opts.Name
	namePascal := toPascalCase(name)
	moduleImportPath := fmt.Sprintf("%s/internal/%s", goModulePath, name)

	files := map[string]string{
		filepath.Join(moduleDir, "models.go"):     modelsTemplate(name, namePascal),
		filepath.Join(moduleDir, "service.go"):    serviceTemplate(name, namePascal),
		filepath.Join(moduleDir, "repository.go"): repositoryInterfaceTemplate(name, namePascal),
	}

	if opts.Persistence != "" {
		switch opts.Persistence {
		case "sqlite":
			files[filepath.Join(moduleDir, "persistence", "sqlite_repository.go")] = sqliteRepositoryTemplate(name, namePascal)
		case "postgres":
			files[filepath.Join(moduleDir, "persistence", "postgres_repository.go")] = postgresRepositoryTemplate(name, namePascal)
		case "mongodb":
			files[filepath.Join(moduleDir, "persistence", "mongo_repository.go")] = mongoRepositoryTemplate(name, namePascal)
		}
	}

	if opts.WithCRUD {
		files[filepath.Join(moduleDir, "commands", fmt.Sprintf("create_%s.go", name))] = createCommandTemplate(name, namePascal, moduleImportPath)
		files[filepath.Join(moduleDir, "commands", fmt.Sprintf("update_%s.go", name))] = updateCommandTemplate(name, namePascal, moduleImportPath)
		files[filepath.Join(moduleDir, "commands", fmt.Sprintf("delete_%s.go", name))] = deleteCommandTemplate(name, namePascal, moduleImportPath)
		files[filepath.Join(moduleDir, "queries", fmt.Sprintf("get_%s.go", name))] = getQueryTemplate(name, namePascal, moduleImportPath)
		files[filepath.Join(moduleDir, "queries", fmt.Sprintf("list_%s.go", name))] = listQueryTemplate(name, namePascal, moduleImportPath)
	}

	for _, transport := range opts.Transports {
		switch transport {
		case "http":
			files[filepath.Join(moduleDir, "transport", "http", "handler.go")] = httpHandlerTemplate(namePascal, moduleImportPath)
			files[filepath.Join(moduleDir, "transport", "http", "routes.go")] = httpRoutesTemplate(name)
		case "amqp":
			files[filepath.Join(moduleDir, "transport", "amqp", "consumer.go")] = amqpConsumerTemplate(name, namePascal, moduleImportPath)
		}
	}

	if opts.APIType == "html" {
		files[filepath.Join(moduleDir, "views", "index.html")] = indexViewTemplate(name, namePascal)
		files[filepath.Join(moduleDir, "views", "form.html")] = formViewTemplate(name, namePascal)
	}

	for path, content := range files {
		if err := os.WriteFile(path, []byte(content), 0644); err != nil {
			return fmt.Errorf("failed to write file %s: %w", path, err)
		}
	}

	return nil
}

func toPascalCase(s string) string {
	parts := strings.Split(s, "_")
	for i, part := range parts {
		if len(part) > 0 {
			parts[i] = strings.ToUpper(string(part[0])) + part[1:]
		}
	}
	return strings.Join(parts, "")
}
