package augur

import (
	"fmt"
	"strings"

	"github.com/emiliopalmerini/grimorio/internal/clipboard"
	"github.com/emiliopalmerini/grimorio/internal/metrics"
	"github.com/emiliopalmerini/grimorio/internal/spell/augur"
	"github.com/spf13/cobra"
)

var Cmd = &cobra.Command{
	Use:   "augur [command]",
	Short: "[Spell] Run a command and analyze errors",
	Long: `Augur runs a command, captures its output, and analyzes any errors using Claude.

Examples:
  grimorio augur "go build"
  grimorio augur "npm test"
  grimorio augur "dotnet build"
  grimorio augur "cargo check"`,
	Args: cobra.MinimumNArgs(1),
	RunE: runAugur,
}

func runAugur(cmd *cobra.Command, args []string) error {
	return metrics.Track("augur", metrics.Spell, "", func() error {
		command := strings.Join(args, " ")

		fmt.Printf("Running: %s\n\n", command)
		result, err := augur.RunCommand(command)
		if err != nil {
			return err
		}

		if result.Stdout != "" {
			fmt.Println(result.Stdout)
		}
		if result.Stderr != "" {
			fmt.Println(result.Stderr)
		}

		if result.ExitCode == 0 && result.Stderr == "" {
			fmt.Println("Command succeeded.")
			return nil
		}

		fmt.Println("\nAuguring the output...")
		analysis, err := augur.Analyze(result)
		if err != nil {
			return err
		}

		fmt.Println(analysis)

		if err := clipboard.Copy(analysis); err == nil {
			fmt.Println("\n(Copied to clipboard)")
		}
		return nil
	})
}
