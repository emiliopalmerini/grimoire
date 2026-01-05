package modifymemory

import (
	"encoding/json"
	"fmt"

	"github.com/emiliopalmerini/grimorio/internal/metrics"
	"github.com/emiliopalmerini/grimorio/internal/spell/memory"
	"github.com/spf13/cobra"
)

var (
	allChanges bool
	dryRun     bool
	motivation string
)

var Cmd = &cobra.Command{
	Use:   "modify-memory",
	Short: "[Spell] Generate commits from diffs using Claude",
	Long: `Modify-memory analyzes your git changes and generates conventional commit messages using Claude Code.

By default, it looks at staged changes. Use -a to include all changes.

Examples:
  grimorio modify-memory
  grimorio modify-memory -a
  grimorio modify-memory -m "refactoring auth flow"
  grimorio modify-memory -n`,
	RunE: runModifyMemory,
}

func init() {
	Cmd.Flags().BoolVarP(&allChanges, "all", "a", false, "Include all changes, not just staged")
	Cmd.Flags().BoolVarP(&dryRun, "dry-run", "n", false, "Just output the message, don't prompt for commit")
	Cmd.Flags().StringVarP(&motivation, "motivation", "m", "", "Motivation/context for the commit")
}

func runModifyMemory(cmd *cobra.Command, args []string) error {
	flags, _ := json.Marshal(map[string]any{"all": allChanges, "dry-run": dryRun, "motivation": motivation})
	return metrics.Track("modify-memory", metrics.Spell, string(flags), func() error {
		diff, err := memory.GetDiff(allChanges)
		if err != nil {
			return err
		}

		history, _ := memory.GetRecentCommits(5)

		fmt.Println("Generating commit message...")
		message, err := memory.GenerateMessage(diff, history, motivation)
		if err != nil {
			return err
		}

		if dryRun {
			fmt.Println(message)
			return nil
		}

		for {
			confirmed, edit, err := memory.Confirm(message)
			if err != nil {
				return err
			}

			if confirmed {
				return memory.Commit(message, allChanges)
			}

			if edit {
				message, err = memory.EditMessage(message)
				if err != nil {
					return err
				}
				continue
			}

			fmt.Println("Commit cancelled.")
			return nil
		}
	})
}
