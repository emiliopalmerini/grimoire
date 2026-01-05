package scry

import (
	"encoding/json"
	"fmt"

	"github.com/emiliopalmerini/grimorio/internal/metrics"
	"github.com/emiliopalmerini/grimorio/internal/spell/scry"
	"github.com/spf13/cobra"
)

var allChanges bool

var Cmd = &cobra.Command{
	Use:   "scry",
	Short: "[Spell] Review staged changes for bugs/issues",
	Long: `Scry analyzes your staged changes and reviews them for potential issues using Claude.

Examples:
  grimorio scry
  grimorio scry -a`,
	RunE: runScry,
}

func init() {
	Cmd.Flags().BoolVarP(&allChanges, "all", "a", false, "Include all changes, not just staged")
}

func runScry(cmd *cobra.Command, args []string) error {
	flags, _ := json.Marshal(map[string]any{"all": allChanges})
	return metrics.Track("scry", metrics.Spell, string(flags), func() error {
		diff, err := scry.GetDiff(allChanges)
		if err != nil {
			return err
		}

		fmt.Println("Scrying the changes...")
		review, err := scry.Review(diff)
		if err != nil {
			return err
		}

		fmt.Println(review)
		return nil
	})
}
