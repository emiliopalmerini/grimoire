package augury

import (
	"fmt"
	"strings"

	"github.com/emiliopalmerini/grimorio/internal/clipboard"
	"github.com/emiliopalmerini/grimorio/internal/metrics"
	"github.com/emiliopalmerini/grimorio/internal/spell/augury"
	"github.com/spf13/cobra"
)

var Cmd = &cobra.Command{
	Use:   "augury [command]",
	Short: "[Spell] Run a command and analyze errors",
	Long: `Augury runs a command, captures its output, and analyzes any errors using Claude.

Examples:
  grimorio augury "go build"
  grimorio augury "npm test"
  grimorio augury "dotnet build"
  grimorio augury "cargo check"`,
	Args: cobra.MinimumNArgs(1),
	RunE: runAugury,
}

func runAugury(cmd *cobra.Command, args []string) error {
	return metrics.Track("augury", metrics.Spell, "", func() error {
		command := strings.Join(args, " ")

		fmt.Printf("Running: %s\n\n", command)
		result, err := augury.RunCommand(command)
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

		fmt.Println("\nReading the augury...")
		analysis, err := augury.Analyze(result)
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
