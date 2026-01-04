package modifymemory

import (
	"fmt"

	"github.com/emiliopalmerini/grimoire/internal/memory"
	"github.com/spf13/cobra"
)

var (
	allChanges bool
	dryRun     bool
)

var Cmd = &cobra.Command{
	Use:   "modify-memory",
	Short: "Generate commits from diffs using Claude",
	Long: `Modify-memory analyzes your git changes and generates conventional commit messages using Claude Code.

By default, it looks at staged changes. Use -a to include all changes.

Examples:
  grimoire modify-memory
  grimoire modify-memory -a
  grimoire modify-memory -n`,
	RunE: runModifyMemory,
}

func init() {
	Cmd.Flags().BoolVarP(&allChanges, "all", "a", false, "Include all changes, not just staged")
	Cmd.Flags().BoolVarP(&dryRun, "dry-run", "n", false, "Just output the message, don't prompt for commit")
}

func runModifyMemory(cmd *cobra.Command, args []string) error {
	diff, err := memory.GetDiff(allChanges)
	if err != nil {
		return err
	}

	history, _ := memory.GetRecentCommits(10)

	fmt.Println("Generating commit message...")
	message, err := memory.GenerateMessage(diff, history)
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
}
