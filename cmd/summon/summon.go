package summon

import (
	"fmt"

	"github.com/emiliopalmerini/grimoire/internal/initializer"
	"github.com/spf13/cobra"
)

var (
	modulePath  string
	goVersion   string
	projectType string
	transports  []string
)

var Cmd = &cobra.Command{
	Use:   "summon [project-name]",
	Short: "Summon a new Go project",
	Long: `Summon conjures a new Go project with a standard structure.

This incantation creates a project with the appropriate structure based on
the project type and transports specified.

Project types:
  api   - REST API with JSON responses (default)
  web   - Web application with sessions, CSRF, and templ
  grpc  - gRPC service

Examples:
  grimoire summon myapp
  grimoire summon myapi --type=api
  grimoire summon mysite --type=web
  grimoire summon myservice --type=grpc
  grimoire summon hybrid --type=api --transport=http,grpc`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		name := args[0]

		module := modulePath
		if module == "" {
			module = name
		}

		// Set default transports based on type if not specified
		if len(transports) == 0 {
			switch projectType {
			case "grpc":
				transports = []string{"grpc"}
			default:
				transports = []string{"http"}
			}
		}

		opts := initializer.ProjectOptions{
			Name:       name,
			ModulePath: module,
			GoVersion:  goVersion,
			Type:       projectType,
			Transports: transports,
		}

		if err := initializer.CreateProject(opts); err != nil {
			return fmt.Errorf("summoning failed: %w", err)
		}

		fmt.Printf("Project '%s' summoned successfully\n", name)
		return nil
	},
}

func init() {
	Cmd.Flags().StringVarP(&modulePath, "module", "m", "", "Go module path (default: project name)")
	Cmd.Flags().StringVarP(&goVersion, "go-version", "g", "1.25", "Go version for flake")
	Cmd.Flags().StringVarP(&projectType, "type", "t", "api", "Project type: api, web, grpc")
	Cmd.Flags().StringSliceVarP(&transports, "transport", "T", nil, "Transports: http, grpc, amqp (default based on type)")
}
