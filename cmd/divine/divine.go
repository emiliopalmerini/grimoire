package divine

import (
	"fmt"

	"github.com/emiliopalmerini/grimorio/internal/divine"
	"github.com/spf13/cobra"
)

var symbol string

var Cmd = &cobra.Command{
	Use:   "divine [file]",
	Short: "[Spell] Explain code in plain language",
	Long: `Divine reads a file and explains its code using Claude.

Examples:
  grimorio divine main.go
  grimorio divine internal/auth/auth.go
  grimorio divine handler.go --symbol HandleLogin`,
	Args: cobra.ExactArgs(1),
	RunE: runDivine,
}

func init() {
	Cmd.Flags().StringVarP(&symbol, "symbol", "s", "", "Focus on a specific function/type")
}

func runDivine(cmd *cobra.Command, args []string) error {
	path := args[0]

	content, err := divine.ReadFile(path)
	if err != nil {
		return err
	}

	fmt.Println("Divining the code...")
	lspContext := divine.GetLSPContext(path, content)
	explanation, err := divine.Explain(content, symbol, lspContext)
	if err != nil {
		return err
	}

	fmt.Println(explanation)
	return nil
}
