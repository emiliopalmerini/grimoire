package mending

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/emiliopalmerini/grimorio/internal/lsp"
)

type Options struct {
	Check bool
	Diff  bool
}

type Result struct {
	Path    string
	Changed bool
	Error   error
	Diff    string
}

func FormatFile(ctx context.Context, path string, opts Options) (*Result, error) {
	absPath, err := filepath.Abs(path)
	if err != nil {
		return nil, fmt.Errorf("failed to get absolute path: %w", err)
	}

	lang := lsp.DetectLanguage(path)
	if lang == nil {
		return nil, fmt.Errorf("unsupported file type: %s", filepath.Ext(path))
	}

	if !lang.Available() {
		return nil, fmt.Errorf("LSP server not found: %s (required for %s files)", lang.Command, lang.Name)
	}

	content, err := os.ReadFile(absPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %w", err)
	}

	original := string(content)
	uri := "file://" + absPath
	rootDir := filepath.Dir(absPath)

	client, err := lsp.NewClient(lang)
	if err != nil {
		return nil, fmt.Errorf("failed to start LSP client: %w", err)
	}
	defer client.Close()

	if err := client.Initialize(ctx, rootDir); err != nil {
		return nil, fmt.Errorf("failed to initialize LSP: %w", err)
	}

	if err := client.OpenDocument(uri, lang.Name, original); err != nil {
		return nil, fmt.Errorf("failed to open document: %w", err)
	}
	defer client.CloseDocument(uri)

	current := original

	importEdits, _ := client.OrganizeImports(uri, current)
	if len(importEdits) > 0 {
		current = ApplyEdits(current, importEdits)
		client.CloseDocument(uri)
		client.OpenDocument(uri, lang.Name, current)
	}

	formatEdits, err := client.Format(uri)
	if err != nil {
		return nil, fmt.Errorf("formatting failed: %w", err)
	}

	if len(formatEdits) > 0 {
		current = ApplyEdits(current, formatEdits)
	}

	changed := current != original

	result := &Result{
		Path:    path,
		Changed: changed,
	}

	if changed && opts.Diff {
		result.Diff = generateDiff(original, current)
	}

	if changed && !opts.Check {
		if err := os.WriteFile(absPath, []byte(current), 0644); err != nil {
			return nil, fmt.Errorf("failed to write file: %w", err)
		}
	}

	return result, nil
}

func ApplyEdits(content string, edits []lsp.TextEdit) string {
	lines := strings.Split(content, "\n")

	sort.Slice(edits, func(i, j int) bool {
		if edits[i].Range.Start.Line != edits[j].Range.Start.Line {
			return edits[i].Range.Start.Line > edits[j].Range.Start.Line
		}
		return edits[i].Range.Start.Character > edits[j].Range.Start.Character
	})

	for _, edit := range edits {
		lines = applyEdit(lines, edit)
	}

	return strings.Join(lines, "\n")
}

func applyEdit(lines []string, edit lsp.TextEdit) []string {
	startLine := edit.Range.Start.Line
	startChar := edit.Range.Start.Character
	endLine := edit.Range.End.Line
	endChar := edit.Range.End.Character

	if startLine >= len(lines) {
		for len(lines) <= startLine {
			lines = append(lines, "")
		}
	}
	if endLine >= len(lines) {
		endLine = len(lines) - 1
		endChar = len(lines[endLine])
	}

	if startChar > len(lines[startLine]) {
		startChar = len(lines[startLine])
	}
	if endChar > len(lines[endLine]) {
		endChar = len(lines[endLine])
	}

	before := lines[startLine][:startChar]
	after := lines[endLine][endChar:]

	newLines := strings.Split(edit.NewText, "\n")

	if len(newLines) == 1 {
		newLines[0] = before + newLines[0] + after
	} else {
		newLines[0] = before + newLines[0]
		newLines[len(newLines)-1] = newLines[len(newLines)-1] + after
	}

	result := make([]string, 0, len(lines)-endLine+startLine+len(newLines))
	result = append(result, lines[:startLine]...)
	result = append(result, newLines...)
	result = append(result, lines[endLine+1:]...)

	return result
}

func generateDiff(original, modified string) string {
	origLines := strings.Split(original, "\n")
	modLines := strings.Split(modified, "\n")

	var diff strings.Builder

	maxLen := len(origLines)
	if len(modLines) > maxLen {
		maxLen = len(modLines)
	}

	for i := 0; i < maxLen; i++ {
		var origLine, modLine string
		if i < len(origLines) {
			origLine = origLines[i]
		}
		if i < len(modLines) {
			modLine = modLines[i]
		}

		if origLine != modLine {
			if i < len(origLines) {
				diff.WriteString(fmt.Sprintf("-%d: %s\n", i+1, origLine))
			}
			if i < len(modLines) {
				diff.WriteString(fmt.Sprintf("+%d: %s\n", i+1, modLine))
			}
		}
	}

	return diff.String()
}
