package scaffold

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"slices"
	"strings"
)

var (
	validTransports   = []string{"http", "grpc", "amqp"}
	validAPITypes     = []string{"json", "html"}
	validPersistence  = []string{"sqlite", "postgres", "mongodb", ""}
	identifierPattern = regexp.MustCompile(`^[a-z][a-z0-9_]*$`)
)

type ModuleOptions struct {
	Name        string
	Transports  []string
	APIType     string
	Persistence string
}

func (o ModuleOptions) Validate() error {
	if o.Name == "" {
		return fmt.Errorf("module name is required")
	}
	if !identifierPattern.MatchString(o.Name) {
		return fmt.Errorf("invalid module name %q: must start with lowercase letter and contain only lowercase letters, numbers, and underscores", o.Name)
	}
	for _, t := range o.Transports {
		if !slices.Contains(validTransports, t) {
			return fmt.Errorf("invalid transport %q: must be one of %v", t, validTransports)
		}
	}
	if !slices.Contains(validAPITypes, o.APIType) {
		return fmt.Errorf("invalid API type %q: must be one of %v", o.APIType, validAPITypes)
	}
	if !slices.Contains(validPersistence, o.Persistence) {
		return fmt.Errorf("invalid persistence %q: must be one of %v", o.Persistence, validPersistence[:len(validPersistence)-1])
	}
	return nil
}

func CreateModule(opts ModuleOptions) error {
	if err := opts.Validate(); err != nil {
		return err
	}

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

	if opts.APIType == "html" {
		dirs = append(dirs, filepath.Join(moduleDir, "views"))
	}

	if hasTransport(opts.Transports, "grpc") {
		cwd, err := os.Getwd()
		if err != nil {
			return fmt.Errorf("failed to get current directory: %w", err)
		}
		dirs = append(dirs, filepath.Join(cwd, "api", "proto", "v1"))
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
			files[filepath.Join(moduleDir, "persistence", "sqlite.go")] = sqliteRepositoryTemplate(name, namePascal, moduleImportPath)
		case "postgres":
			files[filepath.Join(moduleDir, "persistence", "postgres.go")] = postgresRepositoryTemplate(name, namePascal, moduleImportPath)
		case "mongodb":
			files[filepath.Join(moduleDir, "persistence", "mongo.go")] = mongoRepositoryTemplate(name, namePascal, moduleImportPath)
		}
	}

	if hasTransport(opts.Transports, "http") {
		files[filepath.Join(moduleDir, "handler.go")] = httpHandlerTemplate(name, namePascal, moduleImportPath)
		files[filepath.Join(moduleDir, "routes.go")] = httpRoutesTemplate(name)
	}

	if hasTransport(opts.Transports, "grpc") {
		files[filepath.Join(moduleDir, "grpc_server.go")] = grpcServerTemplate(name, namePascal, goModulePath)
		cwd, err := os.Getwd()
		if err != nil {
			return fmt.Errorf("failed to get current directory: %w", err)
		}
		files[filepath.Join(cwd, "api", "proto", "v1", name+".proto")] = protoTemplate(name, namePascal, goModulePath)
	}

	if hasTransport(opts.Transports, "amqp") {
		files[filepath.Join(moduleDir, "consumer.go")] = amqpConsumerTemplate(name, namePascal, moduleImportPath)
	}

	if opts.APIType == "html" {
		files[filepath.Join(moduleDir, "views", "index.templ")] = indexTemplTemplate(name, namePascal)
		files[filepath.Join(moduleDir, "views", "form.templ")] = formTemplTemplate(name, namePascal)
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

func hasTransport(transports []string, t string) bool {
	return slices.Contains(transports, t)
}
