package scrying

import (
	"encoding/json"
	"fmt"

	"github.com/emiliopalmerini/grimorio/internal/clipboard"
	"github.com/emiliopalmerini/grimorio/internal/metrics"
	"github.com/emiliopalmerini/grimorio/internal/spell/scrying"
	"github.com/spf13/cobra"
)

var allChanges bool

var Cmd = &cobra.Command{
	Use:   "scrying",
	Short: "[Spell] Review staged changes for bugs/issues",
	Long: `Scrying analyzes your staged changes and reviews them for potential issues using Claude.

Examples:
  grimorio scrying
  grimorio scrying -a`,
	RunE: runScrying,
}

func init() {
	Cmd.Flags().BoolVarP(&allChanges, "all", "a", false, "Include all changes, not just staged")
}

func runScrying(cmd *cobra.Command, args []string) error {
	flags, _ := json.Marshal(map[string]any{"all": allChanges})
	return metrics.Track("scrying", metrics.Spell, string(flags), func() error {
		diff, err := scrying.GetDiff(allChanges)
		if err != nil {
			return err
		}

		fmt.Println("Scrying the changes...")
		review, err := scrying.Review(diff)
		if err != nil {
			return err
		}

		fmt.Println(review)

		if err := clipboard.Copy(review); err == nil {
			fmt.Println("\n(Copied to clipboard)")
		}
		return nil
	})
}
