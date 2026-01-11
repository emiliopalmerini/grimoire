package diff

import (
	"regexp"
	"strconv"
	"strings"
)

var (
	// Matches: diff --git a/path/to/file b/path/to/file
	diffHeaderRe = regexp.MustCompile(`^diff --git a/(.+) b/(.+)$`)
	// Matches: @@ -start,count +start,count @@ optional context
	hunkHeaderRe = regexp.MustCompile(`^@@ -(\d+)(?:,(\d+))? \+(\d+)(?:,(\d+))? @@`)
	// Matches: Binary files ... differ
	binaryRe = regexp.MustCompile(`^Binary files .+ differ$`)
	// Matches: new file mode
	newFileRe = regexp.MustCompile(`^new file mode`)
	// Matches: deleted file mode
	deletedFileRe = regexp.MustCompile(`^deleted file mode`)
	// Matches: rename from/to
	renameFromRe = regexp.MustCompile(`^rename from (.+)$`)
	renameToRe   = regexp.MustCompile(`^rename to (.+)$`)
)

// Parse parses a unified diff into structured FileDiff objects.
func Parse(diffText string) []FileDiff {
	if diffText == "" {
		return nil
	}

	lines := strings.Split(diffText, "\n")
	var files []FileDiff
	var currentFile *FileDiff
	var currentHunk *Hunk
	var hunkContent strings.Builder

	flushHunk := func() {
		if currentHunk != nil && currentFile != nil {
			currentHunk.Content = hunkContent.String()
			currentFile.Hunks = append(currentFile.Hunks, *currentHunk)
			currentHunk = nil
			hunkContent.Reset()
		}
	}

	flushFile := func() {
		flushHunk()
		if currentFile != nil {
			files = append(files, *currentFile)
			currentFile = nil
		}
	}

	for _, line := range lines {
		// New file diff starts
		if matches := diffHeaderRe.FindStringSubmatch(line); matches != nil {
			flushFile()
			currentFile = &FileDiff{
				OldPath: matches[1],
				NewPath: matches[2],
			}
			continue
		}

		if currentFile == nil {
			continue
		}

		// Check for binary file
		if binaryRe.MatchString(line) {
			currentFile.IsBinary = true
			continue
		}

		// Check for new file
		if newFileRe.MatchString(line) {
			currentFile.IsNew = true
			continue
		}

		// Check for deleted file
		if deletedFileRe.MatchString(line) {
			currentFile.IsDelete = true
			continue
		}

		// Check for rename
		if renameFromRe.MatchString(line) || renameToRe.MatchString(line) {
			currentFile.IsRename = true
			continue
		}

		// Hunk header
		if matches := hunkHeaderRe.FindStringSubmatch(line); matches != nil {
			flushHunk()

			// Regex guarantees digits, but handle errors defensively
			oldStart, err := strconv.Atoi(matches[1])
			if err != nil {
				oldStart = 1
			}
			oldCount := 1
			if matches[2] != "" {
				if c, err := strconv.Atoi(matches[2]); err == nil {
					oldCount = c
				}
			}
			newStart, err := strconv.Atoi(matches[3])
			if err != nil {
				newStart = 1
			}
			newCount := 1
			if matches[4] != "" {
				if c, err := strconv.Atoi(matches[4]); err == nil {
					newCount = c
				}
			}

			currentHunk = &Hunk{
				FilePath: currentFile.NewPath,
				OldStart: oldStart,
				OldCount: oldCount,
				NewStart: newStart,
				NewCount: newCount,
			}
			hunkContent.WriteString(line)
			hunkContent.WriteString("\n")
			continue
		}

		// Hunk content lines
		if currentHunk != nil {
			hunkContent.WriteString(line)
			hunkContent.WriteString("\n")
		}
	}

	flushFile()
	return files
}

// CountLines counts the total number of changed lines in a diff.
func CountLines(diffText string) int {
	count := 0
	for _, line := range strings.Split(diffText, "\n") {
		if strings.HasPrefix(line, "+") && !strings.HasPrefix(line, "+++") {
			count++
		}
		if strings.HasPrefix(line, "-") && !strings.HasPrefix(line, "---") {
			count++
		}
	}
	return count
}

// CountHunkLines counts the number of changed lines in a hunk.
func CountHunkLines(hunk *Hunk) int {
	if hunk == nil {
		return 0
	}
	count := 0
	for _, line := range strings.Split(hunk.Content, "\n") {
		if strings.HasPrefix(line, "+") && !strings.HasPrefix(line, "+++") {
			count++
		}
		if strings.HasPrefix(line, "-") && !strings.HasPrefix(line, "---") {
			count++
		}
	}
	return count
}
