package conjure

import (
	"fmt"

	"github.com/emiliopalmerini/grimorio/internal/cantrip/scaffold"
	"github.com/spf13/cobra"
)

var (
	transports  []string
	apiType     string
	persistence string
)

var Cmd = &cobra.Command{
	Use:   "conjure [module-name]",
	Short: "[Cantrip] Conjure a new vertical slice module",
	Long: `Conjure summons a new module with the vertical slice architecture structure.

This incantation creates a module with service, repository, and transport layers.

Transports:
  http  - HTTP handlers with chi router (default)
  grpc  - gRPC service implementation
  amqp  - RabbitMQ consumer

Examples:
  grimorio conjure user
  grimorio conjure user --transport=http,grpc
  grimorio conjure user --api=html --persistence=postgres
  grimorio conjure order --transport=http,amqp`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		name := args[0]

		opts := scaffold.ModuleOptions{
			Name:        name,
			Transports:  transports,
			APIType:     apiType,
			Persistence: persistence,
		}

		if err := scaffold.CreateModule(opts); err != nil {
			return fmt.Errorf("conjuration failed: %w", err)
		}

		fmt.Printf("Module '%s' conjured successfully\n", name)
		return nil
	},
}

func init() {
	Cmd.Flags().StringSliceVarP(&transports, "transport", "T", []string{"http"}, "Transports: http, grpc, amqp")
	Cmd.Flags().StringVarP(&apiType, "api", "a", "json", "API type: json, html")
	Cmd.Flags().StringVarP(&persistence, "persistence", "p", "", "Persistence: sqlite, postgres, mongodb")
}
