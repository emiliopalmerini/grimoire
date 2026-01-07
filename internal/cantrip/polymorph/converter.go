package polymorph

import (
	"fmt"
	"path/filepath"
	"strings"
)

var formatExtensions = map[string]string{
	".json": "json",
	".yaml": "yaml",
	".yml":  "yaml",
	".toml": "toml",
	".csv":  "csv",
	".md":   "markdown",
	".html": "html",
	".xml":  "xml",
}

var formatAliases = map[string]string{
	"md":       "markdown",
	"yml":      "yaml",
	"markdown": "markdown",
	"json":     "json",
	"yaml":     "yaml",
	"toml":     "toml",
	"csv":      "csv",
	"html":     "html",
	"xml":      "xml",
}

func DetectFormat(filename string) string {
	ext := strings.ToLower(filepath.Ext(filename))
	return formatExtensions[ext]
}

func normalizeFormat(format string) string {
	f := strings.ToLower(strings.TrimSpace(format))
	if alias, ok := formatAliases[f]; ok {
		return alias
	}
	return f
}

func Convert(input []byte, from, to string) ([]byte, error) {
	from = normalizeFormat(from)
	to = normalizeFormat(to)

	data, err := parse(input, from)
	if err != nil {
		return nil, fmt.Errorf("parse error: %w", err)
	}

	return render(data, to)
}

func parse(input []byte, format string) (any, error) {
	switch format {
	case "json":
		return parseJSON(input)
	case "yaml":
		return parseYAML(input)
	case "toml":
		return parseTOML(input)
	case "csv":
		return parseCSV(input)
	case "xml":
		return parseXML(input)
	case "html":
		return parseHTML(input)
	case "markdown":
		return parseMarkdown(input)
	default:
		return nil, fmt.Errorf("unsupported input format: %s", format)
	}
}

func render(data any, format string) ([]byte, error) {
	switch format {
	case "json":
		return renderJSON(data)
	case "yaml":
		return renderYAML(data)
	case "toml":
		return renderTOML(data)
	case "csv":
		return renderCSV(data)
	case "markdown":
		return renderMarkdown(data)
	case "html":
		return renderHTML(data)
	case "xml":
		return renderXML(data)
	default:
		return nil, fmt.Errorf("unsupported output format: %s", format)
	}
}
