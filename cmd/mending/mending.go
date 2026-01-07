package mending

import (
	"context"
	"encoding/json"
	"fmt"
	"os"

	"github.com/emiliopalmerini/grimorio/internal/cantrip/mending"
	"github.com/emiliopalmerini/grimorio/internal/metrics"
	"github.com/spf13/cobra"
)

var (
	checkOnly bool
	showDiff  bool
)

var Cmd = &cobra.Command{
	Use:   "mending [files...]",
	Short: "[Cantrip] Format files using LSP",
	Long: `Mending formats files using language server protocol (LSP) formatters.

Supports: Go, Python, Rust, C#, TypeScript, JavaScript, HTML, JSON, YAML, Nix, Lua

The appropriate LSP server must be installed and available in PATH.

Examples:
  grimorio mending file.go
  grimorio mending ./internal/...
  grimorio mending --check .
  grimorio mending --diff file.py`,
	Args: cobra.MinimumNArgs(1),
	RunE: runMending,
}

func init() {
	Cmd.Flags().BoolVarP(&checkOnly, "check", "c", false, "Check if files need formatting (exit 1 if changes needed)")
	Cmd.Flags().BoolVarP(&showDiff, "diff", "d", false, "Show diff of changes")
}

func runMending(cmd *cobra.Command, args []string) error {
	flags, _ := json.Marshal(map[string]any{"check": checkOnly, "diff": showDiff})
	return metrics.Track("mending", metrics.Cantrip, string(flags), func() error {
		ctx := context.Background()
		files, err := mending.ExpandPaths(args)
		if err != nil {
			return err
		}

		if len(files) == 0 {
			return fmt.Errorf("no files found")
		}

		opts := mending.Options{
			Check: checkOnly,
			Diff:  showDiff,
		}

		var hasChanges bool
		var hasErrors bool

		for _, file := range files {
			result, err := mending.FormatFile(ctx, file, opts)
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
	})
}
