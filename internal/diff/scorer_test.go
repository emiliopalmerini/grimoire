package diff

import (
	"testing"

	"github.com/emiliopalmerini/grimorio/internal/lsp"
)

func TestCategorizeFile(t *testing.T) {
	tests := []struct {
		path     string
		expected FileCategory
	}{
		// Source files
		{"main.go", CategorySource},
		{"internal/handler.go", CategorySource},
		{"src/index.ts", CategorySource},
		{"lib/utils.py", CategorySource},
		{"app.rs", CategorySource},

		// Test files
		{"main_test.go", CategoryTest},
		{"handler.test.ts", CategoryTest},
		{"app.spec.js", CategoryTest},
		{"test_utils.py", CategoryTest},
		{"tests/unit/foo.go", CategoryTest},
		{"__tests__/app.tsx", CategoryTest},

		// Config files
		{"config.json", CategoryConfig},
		{"settings.yaml", CategoryConfig},
		{"docker-compose.yml", CategoryConfig},
		{"config.toml", CategoryConfig},
		{".gitignore", CategoryConfig},
		{".dockerignore", CategoryConfig},

		// Documentation
		{"README.md", CategoryDoc},
		{"CHANGELOG.md", CategoryDoc},
		{"docs/guide.txt", CategoryDoc},

		// Generated files
		{"vendor/github.com/foo/bar.go", CategoryGenerated},
		{"node_modules/lodash/index.js", CategoryGenerated},
		{"api_gen.go", CategoryGenerated},
		{"schema.pb.go", CategoryGenerated},
	}

	for _, tt := range tests {
		t.Run(tt.path, func(t *testing.T) {
			got := CategorizeFile(tt.path)
			if got != tt.expected {
				t.Errorf("CategorizeFile(%q) = %v, want %v", tt.path, got, tt.expected)
			}
		})
	}
}

func TestCategoryMultiplier(t *testing.T) {
	tests := []struct {
		category   FileCategory
		multiplier float64
	}{
		{CategorySource, 1.0},
		{CategoryTest, 0.6},
		{CategoryConfig, 0.4},
		{CategoryDoc, 0.2},
		{CategoryGenerated, 0.1},
	}

	for _, tt := range tests {
		t.Run(tt.category.String(), func(t *testing.T) {
			got := tt.category.Multiplier()
			if got != tt.multiplier {
				t.Errorf("Multiplier() = %v, want %v", got, tt.multiplier)
			}
		})
	}
}

func TestSymbolKindWeight(t *testing.T) {
	tests := []struct {
		kind   string
		weight float64
	}{
		{"Function", WeightFunction},
		{"Method", WeightMethod},
		{"Class", WeightClass},
		{"Struct", WeightStruct},
		{"Interface", WeightInterface},
		{"Variable", WeightVariable},
		{"Unknown", WeightDefault},
	}

	for _, tt := range tests {
		t.Run(tt.kind, func(t *testing.T) {
			got := symbolKindWeight(tt.kind)
			if got != tt.weight {
				t.Errorf("symbolKindWeight(%q) = %v, want %v", tt.kind, got, tt.weight)
			}
		})
	}
}

func TestIsExported(t *testing.T) {
	tests := []struct {
		name     string
		expected bool
	}{
		{"HandleRequest", true},
		{"handleRequest", false},
		{"MyStruct", true},
		{"myStruct", false},
		{"", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := isExported(tt.name)
			if got != tt.expected {
				t.Errorf("isExported(%q) = %v, want %v", tt.name, got, tt.expected)
			}
		})
	}
}

func TestScoreHunk(t *testing.T) {
	// Hunk affecting an exported function
	hunk := &Hunk{
		FilePath: "main.go",
		NewStart: 10,
		NewCount: 5,
		Content:  "+line1\n+line2\n+line3\n",
	}

	symbols := []lsp.DocumentSymbol{
		{Name: "HandleRequest", Kind: "Function", Line: 8, EndLine: 20},
	}

	score := ScoreHunk(hunk, symbols)

	// Should have function weight + exported bonus + size bonus, multiplied by source category
	expectedMin := (WeightFunction + BonusExported) * CategorySource.Multiplier()
	if score < expectedMin {
		t.Errorf("ScoreHunk() = %v, want >= %v", score, expectedMin)
	}

	// Check that symbols were recorded
	if len(hunk.Symbols) != 1 || hunk.Symbols[0] != "HandleRequest" {
		t.Errorf("Expected symbols [HandleRequest], got %v", hunk.Symbols)
	}
}

func TestScoreHunkNoSymbols(t *testing.T) {
	hunk := &Hunk{
		FilePath: "config.json",
		NewStart: 1,
		NewCount: 2,
		Content:  "+{}\n",
	}

	score := ScoreHunk(hunk, nil)

	// Should use default weight with config multiplier
	expected := WeightDefault * CategoryConfig.Multiplier()
	if score < expected {
		t.Errorf("ScoreHunk() = %v, want >= %v", score, expected)
	}
}

func TestOverlaps(t *testing.T) {
	tests := []struct {
		name                       string
		start1, end1, start2, end2 int
		expected                   bool
	}{
		{"no overlap before", 1, 5, 10, 15, false},
		{"no overlap after", 10, 15, 1, 5, false},
		{"partial overlap", 1, 10, 5, 15, true},
		{"contained", 1, 20, 5, 10, true},
		{"exact match", 5, 10, 5, 10, true},
		{"adjacent", 1, 5, 5, 10, true},
		{"single line overlap", 5, 5, 5, 5, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := overlaps(tt.start1, tt.end1, tt.start2, tt.end2)
			if got != tt.expected {
				t.Errorf("overlaps(%d,%d,%d,%d) = %v, want %v",
					tt.start1, tt.end1, tt.start2, tt.end2, got, tt.expected)
			}
		})
	}
}
