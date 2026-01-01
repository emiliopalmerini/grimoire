package conjure

import (
	"fmt"

	"github.com/emiliopalmerini/grimoire/internal/scaffold"
	"github.com/spf13/cobra"
)

var (
	transports  []string
	apiType     string
	withCRUD    bool
	persistence string
)

var Cmd = &cobra.Command{
	Use:   "conjure [module-name]",
	Short: "Conjure a new vertical slice module",
	Long: `Conjure summons a new module with the vertical slice architecture structure.

This incantation creates the full directory structure with commands, queries,
transport layers, and persistence scaffolding.

Examples:
  grimoire conjure user --transport=http,amqp --api=html --crud
  grimoire conjure order --transport=http --api=json
  grimoire conjure product --persistence=postgres`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		name := args[0]

		opts := scaffold.ModuleOptions{
			Name:        name,
			Transports:  transports,
			APIType:     apiType,
			WithCRUD:    withCRUD,
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
	Cmd.Flags().StringSliceVarP(&transports, "transport", "t", []string{"http"}, "Transport layers (http, amqp)")
	Cmd.Flags().StringVarP(&apiType, "api", "a", "json", "API type (json, html)")
	Cmd.Flags().BoolVarP(&withCRUD, "crud", "c", true, "Generate CRUD commands and queries")
	Cmd.Flags().StringVarP(&persistence, "persistence", "p", "", "Persistence layer (sqlite, postgres, mongodb)")
}
