package identify

import (
	"encoding/json"
	"fmt"

	"github.com/emiliopalmerini/grimorio/internal/clipboard"
	"github.com/emiliopalmerini/grimorio/internal/metrics"
	"github.com/emiliopalmerini/grimorio/internal/spell/identify"
	"github.com/spf13/cobra"
)

var symbol string

var Cmd = &cobra.Command{
	Use:   "identify [file]",
	Short: "[Spell] Explain code in plain language",
	Long: `Identify reads a file and explains its code using Claude.

Examples:
  grimorio identify main.go
  grimorio identify internal/auth/auth.go
  grimorio identify handler.go --symbol HandleLogin`,
	Args: cobra.ExactArgs(1),
	RunE: runIdentify,
}

func init() {
	Cmd.Flags().StringVarP(&symbol, "symbol", "s", "", "Focus on a specific function/type")
}

func runIdentify(cmd *cobra.Command, args []string) error {
	flags, _ := json.Marshal(map[string]any{"symbol": symbol})
	return metrics.Track("identify", metrics.Spell, string(flags), func() error {
		path := args[0]

		content, err := identify.ReadFile(path)
		if err != nil {
			return err
		}

		fmt.Println("Identifying the code...")
		lspContext := identify.GetLSPContext(path, content)
		explanation, err := identify.Explain(content, symbol, lspContext)
		if err != nil {
			return err
		}

		fmt.Println(explanation)

		if err := clipboard.Copy(explanation); err == nil {
			fmt.Println("\n(Copied to clipboard)")
		}
		return nil
	})
}
