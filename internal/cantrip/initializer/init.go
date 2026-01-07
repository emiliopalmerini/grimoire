package initializer

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"slices"
)

var (
	validProjectTypes = []string{"api", "web", "grpc"}
	validTransports   = []string{"http", "grpc", "amqp"}
	projectNameRegex  = regexp.MustCompile(`^[a-z][a-z0-9_-]*$`)
)

type ProjectOptions struct {
	Name       string
	ModulePath string
	GoVersion  string
	Type       string   // api, web, grpc
	Transports []string // http, grpc, amqp
}

func (o ProjectOptions) Validate() error {
	if o.Name == "" {
		return fmt.Errorf("project name is required")
	}
	if !projectNameRegex.MatchString(o.Name) {
		return fmt.Errorf("invalid project name %q: must start with lowercase letter and contain only lowercase letters, numbers, hyphens, and underscores", o.Name)
	}
	if !slices.Contains(validProjectTypes, o.Type) {
		return fmt.Errorf("invalid project type %q: must be one of %v", o.Type, validProjectTypes)
	}
	for _, t := range o.Transports {
		if !slices.Contains(validTransports, t) {
			return fmt.Errorf("invalid transport %q: must be one of %v", t, validTransports)
		}
	}
	return nil
}

func CreateProject(opts ProjectOptions) error {
	if err := opts.Validate(); err != nil {
		return err
	}

	projectDir := opts.Name

	if err := createDirectories(projectDir, opts); err != nil {
		return err
	}

	if err := createFiles(projectDir, opts); err != nil {
		return err
	}

	return nil
}

func createDirectories(projectDir string, opts ProjectOptions) error {
	dirs := []string{
		projectDir,
		filepath.Join(projectDir, "cmd"),
		filepath.Join(projectDir, "internal", "app"),
		filepath.Join(projectDir, "internal", "server"),
		filepath.Join(projectDir, "internal", "middleware"),
		filepath.Join(projectDir, ".github", "workflows"),
	}

	if slices.Contains(opts.Transports, "grpc") {
		dirs = append(dirs, filepath.Join(projectDir, "api", "proto", "v1"))
		dirs = append(dirs, filepath.Join(projectDir, "gen", "proto", "v1"))
	}

	if slices.Contains(opts.Transports, "amqp") {
		dirs = append(dirs, filepath.Join(projectDir, "internal", "infra"))
	}

	if opts.Type == "web" {
		dirs = append(dirs, filepath.Join(projectDir, "static"))
	}

	for _, dir := range dirs {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return fmt.Errorf("failed to create directory %s: %w", dir, err)
		}
	}

	return nil
}

func createFiles(projectDir string, opts ProjectOptions) error {
	files := map[string]string{
		filepath.Join(projectDir, "go.mod"):                          goModTemplate(opts.ModulePath, opts.GoVersion, opts),
		filepath.Join(projectDir, "cmd", "main.go"):                  mainTemplate(opts.ModulePath, opts),
		filepath.Join(projectDir, "internal", "app", "app.go"):       appTemplate(opts),
		filepath.Join(projectDir, "internal", "server", "health.go"): healthTemplate(),
		filepath.Join(projectDir, "flake.nix"):                       flakeTemplate(opts.GoVersion, opts),
		filepath.Join(projectDir, ".gitignore"):                      gitignoreTemplate(),
		filepath.Join(projectDir, "Makefile"):                        makefileTemplate(opts.Name),
		filepath.Join(projectDir, ".github", "workflows", "ci.yml"):  ciWorkflowTemplate(opts.Name, opts.GoVersion, opts),
	}

	if slices.Contains(opts.Transports, "http") {
		files[filepath.Join(projectDir, "internal", "server", "http.go")] = httpServerTemplate(opts.ModulePath, opts)
	}

	if slices.Contains(opts.Transports, "grpc") {
		files[filepath.Join(projectDir, "internal", "server", "grpc.go")] = grpcServerTemplate(opts.ModulePath)
		files[filepath.Join(projectDir, "buf.yaml")] = bufYamlTemplate()
		files[filepath.Join(projectDir, "buf.gen.yaml")] = bufGenYamlTemplate()
	}

	if slices.Contains(opts.Transports, "amqp") {
		files[filepath.Join(projectDir, "internal", "infra", "rabbitmq.go")] = rabbitmqTemplate()
	}

	// Middleware files
	files[filepath.Join(projectDir, "internal", "middleware", "logging.go")] = loggingMiddlewareTemplate()
	files[filepath.Join(projectDir, "internal", "middleware", "recovery.go")] = recoveryMiddlewareTemplate()
	files[filepath.Join(projectDir, "internal", "middleware", "requestid.go")] = requestIDMiddlewareTemplate()

	if opts.Type == "api" && slices.Contains(opts.Transports, "http") {
		files[filepath.Join(projectDir, "internal", "middleware", "cors.go")] = corsMiddlewareTemplate()
	}

	if opts.Type == "web" {
		files[filepath.Join(projectDir, "internal", "middleware", "session.go")] = sessionMiddlewareTemplate()
		files[filepath.Join(projectDir, "internal", "middleware", "csrf.go")] = csrfMiddlewareTemplate()
		files[filepath.Join(projectDir, "static", ".gitkeep")] = ""
	}

	for path, content := range files {
		if err := os.WriteFile(path, []byte(content), 0644); err != nil {
			return fmt.Errorf("failed to write file %s: %w", path, err)
		}
	}

	return nil
}
