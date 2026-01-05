package initializer

import (
	"fmt"
	"os"
	"path/filepath"
	"slices"
)

type ProjectOptions struct {
	Name       string
	ModulePath string
	GoVersion  string
	Type       string   // api, web, grpc
	Transports []string // http, grpc, amqp
}

func CreateProject(opts ProjectOptions) error {
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

	if hasTransport(opts.Transports, "grpc") {
		dirs = append(dirs, filepath.Join(projectDir, "api", "proto", "v1"))
		dirs = append(dirs, filepath.Join(projectDir, "gen", "proto", "v1"))
	}

	if hasTransport(opts.Transports, "amqp") {
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

	if hasTransport(opts.Transports, "http") {
		files[filepath.Join(projectDir, "internal", "server", "http.go")] = httpServerTemplate(opts.ModulePath, opts)
	}

	if hasTransport(opts.Transports, "grpc") {
		files[filepath.Join(projectDir, "internal", "server", "grpc.go")] = grpcServerTemplate(opts.ModulePath)
		files[filepath.Join(projectDir, "buf.yaml")] = bufYamlTemplate()
		files[filepath.Join(projectDir, "buf.gen.yaml")] = bufGenYamlTemplate()
	}

	if hasTransport(opts.Transports, "amqp") {
		files[filepath.Join(projectDir, "internal", "infra", "rabbitmq.go")] = rabbitmqTemplate()
	}

	// Middleware files
	files[filepath.Join(projectDir, "internal", "middleware", "logging.go")] = loggingMiddlewareTemplate()
	files[filepath.Join(projectDir, "internal", "middleware", "recovery.go")] = recoveryMiddlewareTemplate()
	files[filepath.Join(projectDir, "internal", "middleware", "requestid.go")] = requestIDMiddlewareTemplate()

	if opts.Type == "api" && hasTransport(opts.Transports, "http") {
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

func hasTransport(transports []string, t string) bool {
	return slices.Contains(transports, t)
}
