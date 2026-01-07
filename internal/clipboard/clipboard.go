package clipboard

import (
	"fmt"
	"os/exec"
	"strings"
)

// Copy copies content to the system clipboard.
// Supports macOS (pbcopy), Linux X11 (xclip, xsel), and Wayland (wl-copy).
func Copy(content string) error {
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

	return nil
}

func commandExists(name string) bool {
	_, err := exec.LookPath(name)
	return err == nil
}
