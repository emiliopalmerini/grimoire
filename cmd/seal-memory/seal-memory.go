package sealmemory

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/emiliopalmerini/grimorio/internal/editor"
	"github.com/emiliopalmerini/grimorio/internal/metrics"
	"github.com/emiliopalmerini/grimorio/internal/spell/sealmemory"
	"github.com/spf13/cobra"
)

var (
	dryRun      bool
	description string
	baseBranch  string
)

var Cmd = &cobra.Command{
	Use:   "seal-memory",
	Short: "[Spell] Generate PR description from branch changes",
	Long: `Seal-memory analyzes your branch commits and generates a pull request description using Claude.

It compares the current branch against main/master and generates a title and description.

Examples:
  grimorio seal-memory
  grimorio seal-memory -m "Added user authentication"
  grimorio seal-memory -n
  grimorio seal-memory --base develop`,
	RunE: runSealMemory,
}

func init() {
	Cmd.Flags().BoolVarP(&dryRun, "dry-run", "n", false, "Output description without creating PR")
	Cmd.Flags().StringVarP(&description, "description", "m", "", "Additional context for the PR")
	Cmd.Flags().StringVarP(&baseBranch, "base", "b", "", "Base branch to compare against (default: auto-detect main/master)")
}

func runSealMemory(cmd *cobra.Command, args []string) error {
	flags, _ := json.Marshal(map[string]any{"dry-run": dryRun, "description": description, "base": baseBranch})
	return metrics.Track("seal-memory", metrics.Spell, string(flags), func() error {
		current, base, err := sealmemory.GetBranchInfo()
		if err != nil {
			return err
		}

		if baseBranch != "" {
			base = baseBranch
		}

		fmt.Printf("Comparing %s against %s...\n", current, base)

		diff, err := sealmemory.GetBranchDiff(base)
		if err != nil {
			return err
		}

		commits, _ := sealmemory.GetBranchCommits(base)

		fmt.Println("Generating PR description...")
		content, err := sealmemory.GeneratePRDescription(diff, commits, description)
		if err != nil {
			return err
		}

		title, body := sealmemory.ParseTitleBody(content)

		if dryRun {
			fmt.Printf("\n--- PR Title ---\n%s\n\n--- PR Body ---\n%s\n", title, body)
			return nil
		}

		fmt.Printf("\n--- PR Title ---\n%s\n\n--- PR Body ---\n%s\n", title, body)

		for {
			action, err := promptAction()
			if err != nil {
				return err
			}

			switch action {
			case "create":
				return createPR(base, title, body)
			case "edit":
				title, body, err = editContent(title, body)
				if err != nil {
					return err
				}
				fmt.Printf("\n--- PR Title ---\n%s\n\n--- PR Body ---\n%s\n", title, body)
				continue
			case "copy":
				return copyToClipboard(title, body)
			default:
				fmt.Println("Cancelled.")
				return nil
			}
		}
	})
}

func promptAction() (string, error) {
	fmt.Print("\nAction? [c]reate PR / [e]dit / c[o]py to clipboard / [q]uit: ")

	reader := bufio.NewReader(os.Stdin)
	input, err := reader.ReadString('\n')
	if err != nil {
		return "", err
	}

	switch strings.TrimSpace(strings.ToLower(input)) {
	case "c", "create":
		return "create", nil
	case "e", "edit":
		return "edit", nil
	case "o", "copy":
		return "copy", nil
	default:
		return "quit", nil
	}
}

func editContent(title, body string) (string, string, error) {
	content := title + "\n\n" + body
	edited, err := editor.Edit(content, "pr-description-*.md")
	if err != nil {
		return "", "", err
	}
	newTitle, newBody := sealmemory.ParseTitleBody(edited)
	return newTitle, newBody, nil
}

func createPR(base, title, body string) error {
	cmd := exec.Command("gh", "pr", "create", "--base", base, "--title", title, "--body", body)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to create PR: %w", err)
	}
	return nil
}

func copyToClipboard(title, body string) error {
	content := title + "\n\n" + body

	var cmd *exec.Cmd
	switch {
	case commandExists("pbcopy"):
		cmd = exec.Command("pbcopy")
	case commandExists("xclip"):
		cmd = exec.Command("xclip", "-selection", "clipboard")
	case commandExists("xsel"):
		cmd = exec.Command("xsel", "--clipboard", "--input")
	case commandExists("wl-copy"):
		cmd = exec.Command("wl-copy")
	default:
		return fmt.Errorf("no clipboard utility found (install xclip, xsel, or wl-copy)")
	}

	cmd.Stdin = strings.NewReader(content)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to copy to clipboard: %w", err)
	}

	fmt.Println("Copied to clipboard!")
	return nil
}

func commandExists(name string) bool {
	_, err := exec.LookPath(name)
	return err == nil
}
