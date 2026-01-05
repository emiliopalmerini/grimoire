package mend

import (
	"strings"
	"testing"

	"github.com/emiliopalmerini/grimorio/internal/lsp"
)

func TestApplyEdits_SingleEdit(t *testing.T) {
	content := "line 1\nline 2\nline 3"
	edits := []lsp.TextEdit{
		{
			Range: lsp.Range{
				Start: lsp.Position{Line: 1, Character: 0},
				End:   lsp.Position{Line: 1, Character: 6},
			},
			NewText: "replaced",
		},
	}

	result := ApplyEdits(content, edits)
	expected := "line 1\nreplaced\nline 3"

	if result != expected {
		t.Errorf("ApplyEdits() = %q, want %q", result, expected)
	}
}

func TestApplyEdits_MultipleEdits(t *testing.T) {
	content := "aaa\nbbb\nccc"
	edits := []lsp.TextEdit{
		{
			Range: lsp.Range{
				Start: lsp.Position{Line: 0, Character: 0},
				End:   lsp.Position{Line: 0, Character: 3},
			},
			NewText: "xxx",
		},
		{
			Range: lsp.Range{
				Start: lsp.Position{Line: 2, Character: 0},
				End:   lsp.Position{Line: 2, Character: 3},
			},
			NewText: "zzz",
		},
	}

	result := ApplyEdits(content, edits)
	expected := "xxx\nbbb\nzzz"

	if result != expected {
		t.Errorf("ApplyEdits() = %q, want %q", result, expected)
	}
}

func TestApplyEdits_InsertText(t *testing.T) {
	content := "hello world"
	edits := []lsp.TextEdit{
		{
			Range: lsp.Range{
				Start: lsp.Position{Line: 0, Character: 5},
				End:   lsp.Position{Line: 0, Character: 5},
			},
			NewText: " there",
		},
	}

	result := ApplyEdits(content, edits)
	expected := "hello there world"

	if result != expected {
		t.Errorf("ApplyEdits() = %q, want %q", result, expected)
	}
}

func TestApplyEdits_DeleteText(t *testing.T) {
	content := "hello world"
	edits := []lsp.TextEdit{
		{
			Range: lsp.Range{
				Start: lsp.Position{Line: 0, Character: 5},
				End:   lsp.Position{Line: 0, Character: 11},
			},
			NewText: "",
		},
	}

	result := ApplyEdits(content, edits)
	expected := "hello"

	if result != expected {
		t.Errorf("ApplyEdits() = %q, want %q", result, expected)
	}
}

func TestApplyEdits_MultilineEdit(t *testing.T) {
	content := "line 1\nline 2\nline 3"
	edits := []lsp.TextEdit{
		{
			Range: lsp.Range{
				Start: lsp.Position{Line: 0, Character: 5},
				End:   lsp.Position{Line: 2, Character: 5},
			},
			NewText: " replaced",
		},
	}

	result := ApplyEdits(content, edits)
	expected := "line  replaced3"

	if result != expected {
		t.Errorf("ApplyEdits() = %q, want %q", result, expected)
	}
}

func TestGenerateDiff_NoDifference(t *testing.T) {
	content := "same\ncontent"
	diff := generateDiff(content, content)

	if diff != "" {
		t.Errorf("generateDiff() for identical content = %q, want empty", diff)
	}
}

func TestGenerateDiff_SingleLineDifference(t *testing.T) {
	original := "line 1\nline 2\nline 3"
	modified := "line 1\nchanged\nline 3"

	diff := generateDiff(original, modified)

	if !strings.Contains(diff, "-2: line 2") {
		t.Error("Expected diff to show removed line 2")
	}
	if !strings.Contains(diff, "+2: changed") {
		t.Error("Expected diff to show added 'changed'")
	}
}

func TestGenerateDiff_AddedLine(t *testing.T) {
	original := "line 1\nline 2"
	modified := "line 1\nline 2\nline 3"

	diff := generateDiff(original, modified)

	if !strings.Contains(diff, "+3: line 3") {
		t.Error("Expected diff to show added line 3")
	}
}

func TestGenerateDiff_RemovedLine(t *testing.T) {
	original := "line 1\nline 2\nline 3"
	modified := "line 1\nline 2"

	diff := generateDiff(original, modified)

	if !strings.Contains(diff, "-3: line 3") {
		t.Error("Expected diff to show removed line 3")
	}
}
