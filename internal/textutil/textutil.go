package textutil

import "strings"

// StripCodeBlock removes markdown code block fencing from a string.
// If the string is wrapped in ``` fences, returns the content without them.
func StripCodeBlock(s string) string {
	lines := strings.Split(s, "\n")
	if len(lines) >= 2 && strings.HasPrefix(lines[0], "```") && strings.HasSuffix(lines[len(lines)-1], "```") {
		return strings.TrimSpace(strings.Join(lines[1:len(lines)-1], "\n"))
	}
	return s
}
