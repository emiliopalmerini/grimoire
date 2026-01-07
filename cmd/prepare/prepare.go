package prepare

import (
	"fmt"

	"github.com/emiliopalmerini/grimorio/internal/prepare"
	"github.com/spf13/cobra"
)

var (
	personal bool
	project  bool
)

var Cmd = &cobra.Command{
	Use:   "prepare",
	Short: "Install grimorio skills for Claude Code",
	Long: `Prepare installs grimorio skills to Claude Code.

Each spell and cantrip becomes a Claude Code skill that Claude can use
to help with development tasks.

Examples:
  grimorio prepare --personal    # Install to ~/.claude/skills/
  grimorio prepare --project     # Install to .claude/skills/`,
	RunE: runPrepare,
}

func init() {
	Cmd.Flags().BoolVarP(&personal, "personal", "p", false, "Install skills to personal directory (~/.claude/skills/)")
	Cmd.Flags().BoolVarP(&project, "project", "P", false, "Install skills to project directory (.claude/skills/)")
}

func runPrepare(cmd *cobra.Command, args []string) error {
	if !personal && !project {
		return fmt.Errorf("please specify --personal or --project")
	}

	commands := prepare.Commands()

	if personal {
		fmt.Println("Installing skills to personal directory...")
		if err := prepare.Install(prepare.Personal, commands); err != nil {
			return err
		}
	}

	if project {
		fmt.Println("Installing skills to project directory...")
		if err := prepare.Install(prepare.Project, commands); err != nil {
			return err
		}
	}

	fmt.Printf("\nInstalled %d skills successfully.\n", len(commands))
	return nil
}
