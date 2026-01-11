package diff

import (
	"path/filepath"
	"strings"
	"unicode"

	"github.com/emiliopalmerini/grimorio/internal/lsp"
)

// Symbol kind weights for scoring.
const (
	WeightFunction    = 100.0
	WeightMethod      = 100.0
	WeightClass       = 90.0
	WeightStruct      = 90.0
	WeightInterface   = 90.0
	WeightConstructor = 85.0
	WeightEnum        = 80.0
	WeightProperty    = 50.0
	WeightField       = 50.0
	WeightVariable    = 40.0
	WeightConstant    = 40.0
	WeightModule      = 30.0
	WeightPackage     = 30.0
	WeightDefault     = 20.0

	// Bonus for exported/public symbols
	BonusExported = 50.0

	// Maximum bonus from hunk size
	MaxSizeBonus = 50.0
)

// symbolKindWeight returns the weight for a given LSP symbol kind.
func symbolKindWeight(kind string) float64 {
	switch kind {
	case "Function":
		return WeightFunction
	case "Method":
		return WeightMethod
	case "Class":
		return WeightClass
	case "Struct":
		return WeightStruct
	case "Interface":
		return WeightInterface
	case "Constructor":
		return WeightConstructor
	case "Enum":
		return WeightEnum
	case "Property":
		return WeightProperty
	case "Field":
		return WeightField
	case "Variable":
		return WeightVariable
	case "Constant":
		return WeightConstant
	case "Module":
		return WeightModule
	case "Package":
		return WeightPackage
	default:
		return WeightDefault
	}
}

// isExported checks if a symbol name is exported (starts with uppercase).
// This is primarily for Go, but the heuristic works for many languages.
func isExported(name string) bool {
	if name == "" {
		return false
	}
	r := []rune(name)
	return unicode.IsUpper(r[0])
}

// CategorizeFile determines the category of a file based on its path.
func CategorizeFile(path string) FileCategory {
	base := filepath.Base(path)
	ext := filepath.Ext(path)
	dir := filepath.Dir(path)

	// Check for generated files
	if strings.Contains(dir, "vendor/") ||
		strings.Contains(dir, "node_modules/") ||
		strings.HasSuffix(base, "_gen.go") ||
		strings.HasSuffix(base, ".gen.go") ||
		strings.HasSuffix(base, ".generated.go") ||
		strings.HasSuffix(base, ".pb.go") {
		return CategoryGenerated
	}

	// Check for test files
	if strings.HasSuffix(base, "_test.go") ||
		strings.HasSuffix(base, ".test.ts") ||
		strings.HasSuffix(base, ".test.js") ||
		strings.HasSuffix(base, ".spec.ts") ||
		strings.HasSuffix(base, ".spec.js") ||
		strings.HasPrefix(base, "test_") ||
		strings.Contains(path, "/test/") ||
		strings.Contains(path, "/tests/") ||
		strings.Contains(path, "/__tests__/") ||
		strings.HasPrefix(path, "test/") ||
		strings.HasPrefix(path, "tests/") ||
		strings.HasPrefix(path, "__tests__/") {
		return CategoryTest
	}

	// Check for documentation
	if ext == ".md" || ext == ".txt" || ext == ".rst" ||
		base == "README" || base == "CHANGELOG" || base == "LICENSE" {
		return CategoryDoc
	}

	// Check for config files
	if ext == ".json" || ext == ".yaml" || ext == ".yml" ||
		ext == ".toml" || ext == ".xml" || ext == ".ini" ||
		base == ".gitignore" || base == ".dockerignore" ||
		strings.HasPrefix(base, ".") {
		return CategoryConfig
	}

	// Check for source code
	sourceExts := map[string]bool{
		".go": true, ".py": true, ".rs": true, ".ts": true, ".tsx": true,
		".js": true, ".jsx": true, ".cs": true, ".java": true, ".rb": true,
		".php": true, ".swift": true, ".kt": true, ".scala": true,
		".c": true, ".cpp": true, ".h": true, ".hpp": true,
		".lua": true, ".nix": true, ".zig": true, ".odin": true,
	}
	if sourceExts[ext] {
		return CategorySource
	}

	return CategoryUnknown
}

// ScoreHunk calculates the priority score for a hunk.
func ScoreHunk(hunk *Hunk, symbols []lsp.DocumentSymbol) float64 {
	category := CategorizeFile(hunk.FilePath)
	multiplier := category.Multiplier()

	// Base score from hunk size
	lineCount := CountHunkLines(hunk)
	sizeBonus := float64(lineCount) * 0.1
	if sizeBonus > MaxSizeBonus {
		sizeBonus = MaxSizeBonus
	}

	// Find symbols affected by this hunk
	var maxSymbolScore float64
	var affectedSymbols []string

	for _, sym := range symbols {
		// Check if hunk overlaps with symbol range
		// Hunk affects lines NewStart to NewStart+NewCount-1
		hunkStart := hunk.NewStart
		hunkEnd := hunk.NewStart + hunk.NewCount - 1

		if overlaps(hunkStart, hunkEnd, sym.Line, sym.EndLine) {
			weight := symbolKindWeight(sym.Kind)
			if isExported(sym.Name) {
				weight += BonusExported
			}
			if weight > maxSymbolScore {
				maxSymbolScore = weight
			}
			affectedSymbols = append(affectedSymbols, sym.Name)
		}
	}

	// If no symbols found via LSP, use a base score
	if maxSymbolScore == 0 {
		maxSymbolScore = WeightDefault
	}

	hunk.Symbols = affectedSymbols
	hunk.Score = (maxSymbolScore + sizeBonus) * multiplier

	return hunk.Score
}

// ScoreFileDiff calculates scores for all hunks in a file diff.
func ScoreFileDiff(fd *FileDiff, symbols []lsp.DocumentSymbol) {
	for i := range fd.Hunks {
		ScoreHunk(&fd.Hunks[i], symbols)
	}
}

// overlaps checks if two ranges overlap.
func overlaps(start1, end1, start2, end2 int) bool {
	return start1 <= end2 && end1 >= start2
}
