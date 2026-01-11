package diff

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/emiliopalmerini/grimorio/internal/lsp"
)

// Options configures the prioritization behavior.
type Options struct {
	MaxHighPriorityLines int           // Maximum lines in high-priority section (default: 400)
	IncludeSummary       bool          // Include summary of low-priority changes (default: true)
	LSPTimeout           time.Duration // Timeout for LSP operations (default: 5s)
	WorkDir              string        // Working directory for file paths
}

// DefaultOptions returns sensible default options.
func DefaultOptions() Options {
	return Options{
		MaxHighPriorityLines: 400,
		IncludeSummary:       true,
		LSPTimeout:           5 * time.Second,
	}
}

// scoredHunk pairs a hunk with its file diff for sorting.
type scoredHunk struct {
	Hunk     Hunk
	FileDiff *FileDiff
}

// Prioritize analyzes a raw diff and returns a prioritized version.
func Prioritize(rawDiff string, opts Options) (*PrioritizedDiff, error) {
	if opts.MaxHighPriorityLines == 0 {
		opts.MaxHighPriorityLines = 400
	}

	files := Parse(rawDiff)
	if len(files) == 0 {
		return &PrioritizedDiff{
			HighPriority: rawDiff,
		}, nil
	}

	// Score all hunks
	ctx, cancel := context.WithTimeout(context.Background(), opts.LSPTimeout)
	defer cancel()

	for i := range files {
		symbols := getSymbolsForFile(ctx, &files[i], opts.WorkDir)
		ScoreFileDiff(&files[i], symbols)
	}

	// Collect all hunks and sort by score
	var allHunks []scoredHunk
	for i := range files {
		for _, h := range files[i].Hunks {
			allHunks = append(allHunks, scoredHunk{Hunk: h, FileDiff: &files[i]})
		}
	}

	sort.Slice(allHunks, func(i, j int) bool {
		return allHunks[i].Hunk.Score > allHunks[j].Hunk.Score
	})

	// Partition into high-priority and low-priority
	var highPriority []scoredHunk
	var lowPriority []scoredHunk
	currentLines := 0

	for _, sh := range allHunks {
		hunkLines := CountHunkLines(&sh.Hunk)
		if currentLines+hunkLines <= opts.MaxHighPriorityLines {
			highPriority = append(highPriority, sh)
			currentLines += hunkLines
		} else {
			lowPriority = append(lowPriority, sh)
		}
	}

	// Build the result
	result := &PrioritizedDiff{
		Stats: computeStats(files),
	}

	// Generate high-priority diff output
	result.HighPriority = buildHighPriorityDiff(highPriority, files)

	// Generate summary for low-priority changes
	if opts.IncludeSummary && len(lowPriority) > 0 {
		result.Summary = buildSummary(lowPriority)
	}

	return result, nil
}

// getSymbolsForFile attempts to get LSP symbols for a file.
func getSymbolsForFile(ctx context.Context, fd *FileDiff, workDir string) []lsp.DocumentSymbol {
	if fd.IsBinary || fd.IsDelete {
		return nil
	}

	path := fd.NewPath
	if workDir != "" {
		path = filepath.Join(workDir, path)
	}

	// Check if file exists
	absPath, err := filepath.Abs(path)
	if err != nil {
		return nil
	}

	content, err := os.ReadFile(absPath)
	if err != nil {
		return nil
	}

	lang := lsp.DetectLanguage(path)
	if lang == nil || !lang.Available() {
		return nil
	}

	client, err := lsp.NewClient(lang)
	if err != nil {
		return nil
	}
	defer client.Close()

	if err := client.Initialize(ctx, filepath.Dir(absPath)); err != nil {
		return nil
	}

	uri := "file://" + absPath
	if err := client.OpenDocument(uri, lang.Name, string(content)); err != nil {
		return nil
	}
	defer client.CloseDocument(uri)

	symbols, err := client.DocumentSymbols(uri)
	if err != nil {
		return nil
	}

	return symbols
}

// buildHighPriorityDiff reconstructs the diff for high-priority hunks.
func buildHighPriorityDiff(hunks []scoredHunk, allFiles []FileDiff) string {
	if len(hunks) == 0 {
		return ""
	}

	// Group hunks by file
	fileHunks := make(map[string][]Hunk)
	for _, sh := range hunks {
		path := sh.FileDiff.NewPath
		fileHunks[path] = append(fileHunks[path], sh.Hunk)
	}

	var result strings.Builder

	for _, fd := range allFiles {
		hunksForFile, ok := fileHunks[fd.NewPath]
		if !ok {
			continue
		}

		// Sort hunks by line number for consistent output
		sort.Slice(hunksForFile, func(i, j int) bool {
			return hunksForFile[i].NewStart < hunksForFile[j].NewStart
		})

		// Write file header
		result.WriteString(fmt.Sprintf("diff --git a/%s b/%s\n", fd.OldPath, fd.NewPath))
		if fd.IsNew {
			result.WriteString("new file mode 100644\n")
		}
		if fd.IsDelete {
			result.WriteString("deleted file mode 100644\n")
		}

		// Handle /dev/null for new and deleted files
		oldPath := "a/" + fd.OldPath
		newPath := "b/" + fd.NewPath
		if fd.IsNew {
			oldPath = "/dev/null"
		}
		if fd.IsDelete {
			newPath = "/dev/null"
		}
		result.WriteString(fmt.Sprintf("--- %s\n", oldPath))
		result.WriteString(fmt.Sprintf("+++ %s\n", newPath))

		// Write hunks
		for _, h := range hunksForFile {
			result.WriteString(h.Content)
		}
	}

	return result.String()
}

// buildSummary generates a summary of low-priority changes.
func buildSummary(lowPriority []scoredHunk) string {
	if len(lowPriority) == 0 {
		return ""
	}

	// Group by category
	categoryFiles := make(map[FileCategory]map[string]int) // category -> file -> lines
	for _, sh := range lowPriority {
		cat := CategorizeFile(sh.Hunk.FilePath)
		if categoryFiles[cat] == nil {
			categoryFiles[cat] = make(map[string]int)
		}
		categoryFiles[cat][sh.Hunk.FilePath] += CountHunkLines(&sh.Hunk)
	}

	var parts []string
	order := []FileCategory{CategoryTest, CategoryConfig, CategoryDoc, CategoryGenerated, CategorySource, CategoryUnknown}

	for _, cat := range order {
		files := categoryFiles[cat]
		if len(files) == 0 {
			continue
		}

		totalLines := 0
		var fileNames []string
		for f, lines := range files {
			totalLines += lines
			fileNames = append(fileNames, filepath.Base(f))
		}

		// Sort for deterministic output
		sort.Strings(fileNames)

		if len(fileNames) > 3 {
			fileNames = append(fileNames[:3], "...")
		}

		parts = append(parts, fmt.Sprintf("%d %s file(s) (%d lines): %s",
			len(files), cat.String(), totalLines, strings.Join(fileNames, ", ")))
	}

	if len(parts) == 0 {
		return ""
	}

	return "\n[Also modified]\n- " + strings.Join(parts, "\n- ")
}

// computeStats calculates aggregate statistics for a diff.
func computeStats(files []FileDiff) DiffStats {
	var stats DiffStats

	for _, fd := range files {
		cat := CategorizeFile(fd.NewPath)
		fileLines := 0

		for _, h := range fd.Hunks {
			fileLines += CountHunkLines(&h)
		}

		stats.TotalFiles++
		stats.TotalLines += fileLines

		switch cat {
		case CategorySource:
			stats.SourceFiles++
			stats.SourceLines += fileLines
		case CategoryTest:
			stats.TestFiles++
			stats.TestLines += fileLines
		case CategoryConfig:
			stats.ConfigFiles++
			stats.ConfigLines += fileLines
		case CategoryDoc:
			stats.DocFiles++
			stats.DocLines += fileLines
		case CategoryGenerated:
			stats.GeneratedFiles++
			stats.GeneratedLines += fileLines
		}
	}

	return stats
}

// FormatForPrompt formats the prioritized diff for inclusion in an AI prompt.
func FormatForPrompt(pd *PrioritizedDiff) string {
	var result strings.Builder

	if pd.HighPriority != "" {
		result.WriteString(pd.HighPriority)
	}

	if pd.Summary != "" {
		result.WriteString(pd.Summary)
	}

	return result.String()
}
