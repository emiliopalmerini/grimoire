package mend

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/emiliopalmerini/grimoire/internal/mend"
	"github.com/spf13/cobra"
)

var (
	checkOnly bool
	showDiff  bool
)

var Cmd = &cobra.Command{
	Use:   "mend [files...]",
	Short: "Format files using LSP",
	Long: `Mend formats files using language server protocol (LSP) formatters.

Supports: Go, Python, Rust, C#, TypeScript, JavaScript, HTML, JSON, YAML, Nix, Lua

The appropriate LSP server must be installed and available in PATH.

Examples:
  grimoire mend file.go
  grimoire mend ./internal/...
  grimoire mend --check .
  grimoire mend --diff file.py`,
	Args: cobra.MinimumNArgs(1),
	RunE: runMend,
}

func init() {
	Cmd.Flags().BoolVarP(&checkOnly, "check", "c", false, "Check if files need formatting (exit 1 if changes needed)")
	Cmd.Flags().BoolVarP(&showDiff, "diff", "d", false, "Show diff of changes")
}

func runMend(cmd *cobra.Command, args []string) error {
	ctx := context.Background()
	files, err := expandPaths(args)
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

func expandPaths(patterns []string) ([]string, error) {
	var files []string
	seen := make(map[string]bool)

	for _, pattern := range patterns {
		if strings.HasSuffix(pattern, "/...") {
			dir := strings.TrimSuffix(pattern, "/...")
			err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
				if err != nil {
					return err
				}
				if !info.IsDir() && isSupportedFile(path) {
					absPath, _ := filepath.Abs(path)
					if !seen[absPath] {
						seen[absPath] = true
						files = append(files, path)
					}
				}
				return nil
			})
			if err != nil {
				return nil, err
			}
		} else {
			info, err := os.Stat(pattern)
			if err != nil {
				return nil, err
			}
			if info.IsDir() {
				entries, err := os.ReadDir(pattern)
				if err != nil {
					return nil, err
				}
				for _, entry := range entries {
					if !entry.IsDir() {
						path := filepath.Join(pattern, entry.Name())
						if isSupportedFile(path) {
							absPath, _ := filepath.Abs(path)
							if !seen[absPath] {
								seen[absPath] = true
								files = append(files, path)
							}
						}
					}
				}
			} else {
				absPath, _ := filepath.Abs(pattern)
				if !seen[absPath] {
					seen[absPath] = true
					files = append(files, pattern)
				}
			}
		}
	}

	return files, nil
}

func isSupportedFile(path string) bool {
	return mend.DetectLanguage(path) != nil
}
