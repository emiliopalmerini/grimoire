package mend

import (
	"context"
	"fmt"
	"os"

	"github.com/emiliopalmerini/grimorio/internal/cantrip/mend"
	"github.com/spf13/cobra"
)

var (
	checkOnly bool
	showDiff  bool
)

var Cmd = &cobra.Command{
	Use:   "mend [files...]",
	Short: "[Cantrip] Format files using LSP",
	Long: `Mend formats files using language server protocol (LSP) formatters.

Supports: Go, Python, Rust, C#, TypeScript, JavaScript, HTML, JSON, YAML, Nix, Lua

The appropriate LSP server must be installed and available in PATH.

Examples:
  grimorio mend file.go
  grimorio mend ./internal/...
  grimorio mend --check .
  grimorio mend --diff file.py`,
	Args: cobra.MinimumNArgs(1),
	RunE: runMend,
}

func init() {
	Cmd.Flags().BoolVarP(&checkOnly, "check", "c", false, "Check if files need formatting (exit 1 if changes needed)")
	Cmd.Flags().BoolVarP(&showDiff, "diff", "d", false, "Show diff of changes")
}

func runMend(cmd *cobra.Command, args []string) error {
	ctx := context.Background()
	files, err := mend.ExpandPaths(args)
	if err != nil {
		return err
	}

	if len(files) == 0 {
		return fmt.Errorf("no files found")
	}

	opts := mend.Options{
		Check: checkOnly,
		Diff:  showDiff,
	}

	var hasChanges bool
	var hasErrors bool

	for _, file := range files {
		result, err := mend.FormatFile(ctx, file, opts)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error formatting %s: %v\n", file, err)
			hasErrors = true
			continue
		}

		if result.Changed {
			hasChanges = true
			if checkOnly {
				fmt.Printf("Would format: %s\n", result.Path)
			} else {
				fmt.Printf("Formatted: %s\n", result.Path)
			}
			if showDiff && result.Diff != "" {
				fmt.Println(result.Diff)
			}
		}
	}

	if hasErrors {
		return fmt.Errorf("some files failed to format")
	}

	if checkOnly && hasChanges {
		return fmt.Errorf("files need formatting")
	}

	return nil
}
