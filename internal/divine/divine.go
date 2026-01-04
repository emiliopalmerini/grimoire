package divine

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/emiliopalmerini/grimoire/internal/mend"
)

func ReadFile(path string) (string, error) {
	content, err := os.ReadFile(path)
	if err != nil {
		return "", fmt.Errorf("failed to read file: %w", err)
	}
	return string(content), nil
}

func GetLSPContext(path string, content string) string {
	lang := mend.DetectLanguage(path)
	if lang == nil || !lang.Available() {
		return ""
	}

	client, err := mend.NewLSPClient(lang)
	if err != nil {
		return ""
	}
	defer client.Close()

	ctx := context.Background()
	rootPath := filepath.Dir(path)
	if err := client.Initialize(ctx, rootPath); err != nil {
		return ""
	}

	absPath, _ := filepath.Abs(path)
	uri := "file://" + absPath

	if err := client.OpenDocument(uri, lang.Name, content); err != nil {
		return ""
	}
	defer client.CloseDocument(uri)

	symbols, err := client.DocumentSymbols(uri)
	if err != nil || len(symbols) == 0 {
		return ""
	}

	var sb strings.Builder
	sb.WriteString("Document symbols:\n")
	for _, sym := range symbols {
		sb.WriteString(fmt.Sprintf("- %s (%s) at line %d\n", sym.Name, sym.Kind, sym.Line+1))
	}

	return sb.String()
}

func Explain(content string, symbol string, lspContext string) (string, error) {
	prompt := `Explain this code in plain language. Be concise but thorough.
Focus on:
- What the code does
- Key functions/types and their purpose
- Important patterns or techniques used
`

	if symbol != "" {
		prompt += fmt.Sprintf("\nFocus specifically on: %s\n", symbol)
	}

	if lspContext != "" {
		prompt += "\n" + lspContext + "\n"
	}

	prompt += "\nCode:\n" + content

	cmd := exec.Command("claude", "-p", prompt)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("claude failed: %w\n%s", err, stderr.String())
	}

	return strings.TrimSpace(stdout.String()), nil
}
