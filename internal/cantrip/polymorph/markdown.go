package polymorph

import (
	"bytes"
	"fmt"
	"regexp"
	"strings"
)

func parseMarkdown(input []byte) (any, error) {
	content := string(input)

	if table := parseMarkdownTable(content); table != nil {
		return table, nil
	}

	if list := parseMarkdownList(content); list != nil {
		return list, nil
	}

	if kv := parseMarkdownKeyValue(content); kv != nil {
		return kv, nil
	}

	text := strings.TrimSpace(content)
	if text != "" {
		return map[string]any{"content": text}, nil
	}

	return map[string]any{}, nil
}

func parseMarkdownTable(content string) []any {
	lines := strings.Split(content, "\n")
	var tableLines []string
	inTable := false

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "|") && strings.HasSuffix(trimmed, "|") {
			inTable = true
			tableLines = append(tableLines, trimmed)
		} else if inTable {
			break
		}
	}

	if len(tableLines) < 2 {
		return nil
	}

	headerLine := tableLines[0]
	headers := parseMarkdownTableRow(headerLine)
	if len(headers) == 0 {
		return nil
	}

	startIdx := 1
	if len(tableLines) > 1 {
		sepLine := tableLines[1]
		if strings.Contains(sepLine, "---") || strings.Contains(sepLine, ":-") {
			startIdx = 2
		}
	}

	var result []any
	for i := startIdx; i < len(tableLines); i++ {
		cells := parseMarkdownTableRow(tableLines[i])
		obj := make(map[string]any)
		for j, cell := range cells {
			if j < len(headers) {
				obj[headers[j]] = cell
			}
		}
		if len(obj) > 0 {
			result = append(result, obj)
		}
	}

	return result
}

func parseMarkdownTableRow(line string) []string {
	line = strings.Trim(line, "|")
	parts := strings.Split(line, "|")
	var result []string
	for _, p := range parts {
		result = append(result, strings.TrimSpace(p))
	}
	return result
}

func parseMarkdownList(content string) []any {
	lines := strings.Split(content, "\n")
	var result []any
	listRe := regexp.MustCompile(`^\s*[-*+]\s+(.+)$`)
	numListRe := regexp.MustCompile(`^\s*\d+\.\s+(.+)$`)

	for _, line := range lines {
		if m := listRe.FindStringSubmatch(line); m != nil {
			result = append(result, strings.TrimSpace(m[1]))
		} else if m := numListRe.FindStringSubmatch(line); m != nil {
			result = append(result, strings.TrimSpace(m[1]))
		}
	}

	if len(result) == 0 {
		return nil
	}
	return result
}

func parseMarkdownKeyValue(content string) map[string]any {
	lines := strings.Split(content, "\n")
	result := make(map[string]any)
	kvRe := regexp.MustCompile(`^\*\*(.+?)\*\*:\s*(.+)$`)
	colonRe := regexp.MustCompile(`^([^:]+):\s+(.+)$`)

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if m := kvRe.FindStringSubmatch(line); m != nil {
			result[strings.TrimSpace(m[1])] = strings.TrimSpace(m[2])
		} else if m := colonRe.FindStringSubmatch(line); m != nil {
			key := strings.TrimSpace(m[1])
			if !strings.HasPrefix(key, "#") && !strings.HasPrefix(key, "-") {
				result[key] = strings.TrimSpace(m[2])
			}
		}
	}

	if len(result) == 0 {
		return nil
	}
	return result
}

func renderMarkdown(data any) ([]byte, error) {
	var buf bytes.Buffer

	if rows, ok := toSlice(data); ok && len(rows) > 0 {
		if _, isMap := toMap(rows[0]); isMap {
			renderMarkdownTable(&buf, rows)
		} else {
			renderMarkdownList(&buf, rows)
		}
	} else if obj, ok := toMap(data); ok {
		renderMarkdownKeyValue(&buf, obj)
	} else {
		buf.WriteString(fmt.Sprintf("%v\n", data))
	}

	return buf.Bytes(), nil
}

func renderMarkdownTable(buf *bytes.Buffer, rows []any) {
	headers := extractHeaders(rows)
	if len(headers) == 0 {
		return
	}

	buf.WriteString("| " + strings.Join(headers, " | ") + " |\n")
	buf.WriteString("|" + strings.Repeat(" --- |", len(headers)) + "\n")

	for _, row := range rows {
		obj, _ := toMap(row)
		vals := make([]string, len(headers))
		for i, h := range headers {
			vals[i] = escapeMarkdown(fmt.Sprintf("%v", obj[h]))
		}
		buf.WriteString("| " + strings.Join(vals, " | ") + " |\n")
	}
}

func renderMarkdownList(buf *bytes.Buffer, items []any) {
	for _, item := range items {
		buf.WriteString(fmt.Sprintf("- %v\n", item))
	}
}

func renderMarkdownKeyValue(buf *bytes.Buffer, obj map[string]any) {
	keys := sortedKeys(obj)
	for _, k := range keys {
		v := obj[k]
		if nested, ok := toMap(v); ok {
			buf.WriteString(fmt.Sprintf("## %s\n\n", k))
			renderMarkdownKeyValue(buf, nested)
			buf.WriteString("\n")
		} else if arr, ok := toSlice(v); ok {
			buf.WriteString(fmt.Sprintf("## %s\n\n", k))
			if len(arr) > 0 {
				if _, isMap := toMap(arr[0]); isMap {
					renderMarkdownTable(buf, arr)
				} else {
					renderMarkdownList(buf, arr)
				}
			}
			buf.WriteString("\n")
		} else {
			buf.WriteString(fmt.Sprintf("**%s**: %v\n\n", k, v))
		}
	}
}

func escapeMarkdown(s string) string {
	s = strings.ReplaceAll(s, "|", "\\|")
	s = strings.ReplaceAll(s, "\n", " ")
	return s
}
