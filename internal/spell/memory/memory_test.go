package memory

import "testing"

func TestStripCodeBlock(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{
			name:  "no code block",
			input: "just plain text",
			want:  "just plain text",
		},
		{
			name:  "with code block",
			input: "```\ncode here\n```",
			want:  "code here",
		},
		{
			name:  "with language specifier",
			input: "```go\npackage main\n```",
			want:  "package main",
		},
		{
			name:  "multiline code block",
			input: "```\nline 1\nline 2\nline 3\n```",
			want:  "line 1\nline 2\nline 3",
		},
		{
			name:  "incomplete code block - no closing",
			input: "```\ncode here",
			want:  "```\ncode here",
		},
		{
			name:  "single line - not a code block",
			input: "```",
			want:  "```",
		},
		{
			name:  "empty code block",
			input: "```\n```",
			want:  "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := stripCodeBlock(tt.input)
			if got != tt.want {
				t.Errorf("stripCodeBlock() = %q, want %q", got, tt.want)
			}
		})
	}
}
