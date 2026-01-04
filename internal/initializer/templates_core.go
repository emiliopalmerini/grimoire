package initializer

import "fmt"

func goModTemplate(modulePath, goVersion string, opts ProjectOptions) string {
	deps := `
require (
	github.com/kelseyhightower/envconfig v1.4.0
	github.com/google/uuid v1.6.0
`
	if hasTransport(opts.Transports, "http") {
		deps += `	github.com/go-chi/chi/v5 v5.1.0
`
	}

	if opts.Type == "web" {
		deps += `	github.com/alexedwards/scs/v2 v2.8.0
`
	}

	if hasTransport(opts.Transports, "grpc") {
		deps += `	google.golang.org/grpc v1.68.0
	google.golang.org/protobuf v1.35.2
`
	}

	if hasTransport(opts.Transports, "amqp") {
		deps += `	github.com/rabbitmq/amqp091-go v1.10.0
`
	}

	deps += ")"

	return fmt.Sprintf(`module %s

go %s
%s
`, modulePath, goVersion, deps)
}

func gitignoreTemplate() string {
	return `# Binaries
*.exe
*.exe~
*.dll
*.so
*.dylib

# Test binary
*.test

# Output
/bin/
/dist/

# Go workspace
go.work
go.work.sum

# IDE
.idea/
.vscode/
*.swp
*.swo

# OS
.DS_Store
Thumbs.db

# Environment
.env
.env.local

# Generated
/gen/
`
}

func flakeTemplate(goVersion string, opts ProjectOptions) string {
	nixGoVersion := "go"
	if len(goVersion) >= 4 {
		nixGoVersion = fmt.Sprintf("go_1_%s", goVersion[2:4])
	}

	packages := fmt.Sprintf("pkgs.%s", nixGoVersion)
	if hasTransport(opts.Transports, "grpc") {
		packages += " pkgs.buf pkgs.protobuf"
	}
	if opts.Type == "web" {
		packages += " pkgs.templ"
	}

	return fmt.Sprintf(`{
  inputs.nixpkgs.url = "github:NixOS/nixpkgs/nixos-unstable";

  outputs = { self, nixpkgs }:
    let
      systems = [ "x86_64-linux" "aarch64-linux" "x86_64-darwin" "aarch64-darwin" ];
      forAllSystems = nixpkgs.lib.genAttrs systems;
    in {
      devShells = forAllSystems (system:
        let pkgs = nixpkgs.legacyPackages.${system};
        in {
          default = pkgs.mkShell {
            packages = [ %s ];
          };
        });
    };
}
`, packages)
}
