package initializer

import "fmt"

func goModTemplate(modulePath, goVersion string) string {
	return fmt.Sprintf(`module %s

go %s
`, modulePath, goVersion)
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
