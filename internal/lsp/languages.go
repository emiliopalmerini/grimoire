package lsp

import (
	"os/exec"
	"path/filepath"
	"strings"
)

type Language struct {
	Name       string
	Extensions []string
	Command    string
	Args       []string
}

var languages = []Language{
	{
		Name:       "go",
		Extensions: []string{".go"},
		Command:    "gopls",
		Args:       []string{},
	},
	{
		Name:       "python",
		Extensions: []string{".py"},
		Command:    "pyright-langserver",
		Args:       []string{"--stdio"},
	},
	{
		Name:       "rust",
		Extensions: []string{".rs"},
		Command:    "rust-analyzer",
		Args:       []string{},
	},
	{
		Name:       "csharp",
		Extensions: []string{".cs"},
		Command:    "OmniSharp",
		Args:       []string{"--languageserver"},
	},
	{
		Name:       "typescript",
		Extensions: []string{".ts", ".tsx", ".js", ".jsx"},
		Command:    "typescript-language-server",
		Args:       []string{"--stdio"},
	},
	{
		Name:       "html",
		Extensions: []string{".html", ".htm"},
		Command:    "vscode-html-language-server",
		Args:       []string{"--stdio"},
	},
	{
		Name:       "json",
		Extensions: []string{".json"},
		Command:    "vscode-json-language-server",
		Args:       []string{"--stdio"},
	},
	{
		Name:       "yaml",
		Extensions: []string{".yaml", ".yml"},
		Command:    "yaml-language-server",
		Args:       []string{"--stdio"},
	},
	{
		Name:       "nix",
		Extensions: []string{".nix"},
		Command:    "nil",
		Args:       []string{},
	},
	{
		Name:       "lua",
		Extensions: []string{".lua"},
		Command:    "lua-language-server",
		Args:       []string{},
	},
}

func DetectLanguage(filename string) *Language {
	ext := strings.ToLower(filepath.Ext(filename))
	for i := range languages {
		for _, e := range languages[i].Extensions {
			if e == ext {
				return &languages[i]
			}
		}
	}
	return nil
}

func (l *Language) Available() bool {
	_, err := exec.LookPath(l.Command)
	return err == nil
}
