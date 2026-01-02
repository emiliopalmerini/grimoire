package transmute

import (
	"bytes"
	"encoding/csv"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"path/filepath"
	"regexp"
	"sort"
	"strings"

	"github.com/BurntSushi/toml"
	"gopkg.in/yaml.v3"
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

func parseJSON(input []byte) (any, error) {
	var data any
	if err := json.Unmarshal(input, &data); err != nil {
		return nil, err
	}
	return data, nil
}

func parseYAML(input []byte) (any, error) {
	var data any
	if err := yaml.Unmarshal(input, &data); err != nil {
		return nil, err
	}
	return normalizeYAMLMaps(data), nil
}

func parseTOML(input []byte) (any, error) {
	var data any
	if err := toml.Unmarshal(input, &data); err != nil {
		return nil, err
	}
	return data, nil
}

func parseCSV(input []byte) (any, error) {
	reader := csv.NewReader(bytes.NewReader(input))
	records, err := reader.ReadAll()
	if err != nil {
		return nil, err
	}
	if len(records) < 1 {
		return []any{}, nil
	}

	headers := records[0]
	var result []any
	for _, row := range records[1:] {
		obj := make(map[string]any)
		for i, val := range row {
			if i < len(headers) {
				obj[headers[i]] = val
			}
		}
		result = append(result, obj)
	}
	return result, nil
}

func parseXML(input []byte) (any, error) {
	decoder := xml.NewDecoder(bytes.NewReader(input))
	return parseXMLDocument(decoder)
}

type xmlNode struct {
	name     string
	attrs    map[string]string
	children []xmlNode
	text     string
}

func parseXMLDocument(decoder *xml.Decoder) (any, error) {
	var stack []*xmlNode
	var root *xmlNode

	for {
		token, err := decoder.Token()
		if err != nil {
			break
		}

		switch t := token.(type) {
		case xml.StartElement:
			node := &xmlNode{
				name:  t.Name.Local,
				attrs: make(map[string]string),
			}
			for _, attr := range t.Attr {
				node.attrs[attr.Name.Local] = attr.Value
			}
			if len(stack) > 0 {
				parent := stack[len(stack)-1]
				parent.children = append(parent.children, *node)
				stack = append(stack, &parent.children[len(parent.children)-1])
			} else {
				root = node
				stack = append(stack, node)
			}

		case xml.EndElement:
			if len(stack) > 0 {
				stack = stack[:len(stack)-1]
			}

		case xml.CharData:
			text := strings.TrimSpace(string(t))
			if text != "" && len(stack) > 0 {
				current := stack[len(stack)-1]
				current.text = text
			}
		}
	}

	if root == nil {
		return map[string]any{}, nil
	}
	return map[string]any{root.name: nodeToMap(root)}, nil
}

func nodeToMap(node *xmlNode) any {
	if len(node.children) == 0 && len(node.attrs) == 0 {
		return node.text
	}

	result := make(map[string]any)

	for k, v := range node.attrs {
		result["@"+k] = v
	}

	if node.text != "" && len(node.children) == 0 {
		if len(node.attrs) > 0 {
			result["#text"] = node.text
		} else {
			return node.text
		}
	}

	childMap := make(map[string][]any)
	for i := range node.children {
		child := &node.children[i]
		childMap[child.name] = append(childMap[child.name], nodeToMap(child))
	}

	for name, values := range childMap {
		if len(values) == 1 {
			result[name] = values[0]
		} else {
			result[name] = values
		}
	}

	return result
}

func parseHTML(input []byte) (any, error) {
	content := string(input)

	if table := parseHTMLTable(content); table != nil {
		return table, nil
	}

	if list := parseHTMLList(content); list != nil {
		return list, nil
	}

	if dl := parseHTMLDefinitionList(content); dl != nil {
		return dl, nil
	}

	text := stripHTMLTags(content)
	text = strings.TrimSpace(text)
	if text != "" {
		return map[string]any{"content": text}, nil
	}

	return map[string]any{}, nil
}

func parseHTMLTable(content string) []any {
	tableRe := regexp.MustCompile(`(?is)<table[^>]*>(.*?)</table>`)
	tableMatch := tableRe.FindStringSubmatch(content)
	if tableMatch == nil {
		return nil
	}

	tableContent := tableMatch[1]

	var headers []string
	headerRe := regexp.MustCompile(`(?is)<th[^>]*>(.*?)</th>`)
	headerMatches := headerRe.FindAllStringSubmatch(tableContent, -1)
	for _, m := range headerMatches {
		headers = append(headers, strings.TrimSpace(stripHTMLTags(m[1])))
	}

	if len(headers) == 0 {
		trRe := regexp.MustCompile(`(?is)<tr[^>]*>(.*?)</tr>`)
		trMatches := trRe.FindAllStringSubmatch(tableContent, -1)
		if len(trMatches) > 0 {
			tdRe := regexp.MustCompile(`(?is)<td[^>]*>(.*?)</td>`)
			firstRowCells := tdRe.FindAllStringSubmatch(trMatches[0][1], -1)
			for _, cell := range firstRowCells {
				headers = append(headers, strings.TrimSpace(stripHTMLTags(cell[1])))
			}
		}
	}

	if len(headers) == 0 {
		return nil
	}

	var result []any
	rowRe := regexp.MustCompile(`(?is)<tr[^>]*>(.*?)</tr>`)
	cellRe := regexp.MustCompile(`(?is)<td[^>]*>(.*?)</td>`)

	rows := rowRe.FindAllStringSubmatch(tableContent, -1)
	startIdx := 0
	if len(headerMatches) == 0 && len(rows) > 0 {
		startIdx = 1
	}

	for i := startIdx; i < len(rows); i++ {
		cells := cellRe.FindAllStringSubmatch(rows[i][1], -1)
		if len(cells) == 0 {
			continue
		}
		obj := make(map[string]any)
		for j, cell := range cells {
			if j < len(headers) {
				obj[headers[j]] = strings.TrimSpace(stripHTMLTags(cell[1]))
			}
		}
		if len(obj) > 0 {
			result = append(result, obj)
		}
	}

	return result
}

func parseHTMLList(content string) []any {
	listRe := regexp.MustCompile(`(?is)<[uo]l[^>]*>(.*?)</[uo]l>`)
	listMatch := listRe.FindStringSubmatch(content)
	if listMatch == nil {
		return nil
	}

	var result []any
	itemRe := regexp.MustCompile(`(?is)<li[^>]*>(.*?)</li>`)
	items := itemRe.FindAllStringSubmatch(listMatch[1], -1)
	for _, item := range items {
		text := strings.TrimSpace(stripHTMLTags(item[1]))
		if text != "" {
			result = append(result, text)
		}
	}

	return result
}

func parseHTMLDefinitionList(content string) map[string]any {
	dlRe := regexp.MustCompile(`(?is)<dl[^>]*>(.*?)</dl>`)
	dlMatch := dlRe.FindStringSubmatch(content)
	if dlMatch == nil {
		return nil
	}

	result := make(map[string]any)
	dtRe := regexp.MustCompile(`(?is)<dt[^>]*>(.*?)</dt>\s*<dd[^>]*>(.*?)</dd>`)
	pairs := dtRe.FindAllStringSubmatch(dlMatch[1], -1)
	for _, pair := range pairs {
		key := strings.TrimSpace(stripHTMLTags(pair[1]))
		value := strings.TrimSpace(stripHTMLTags(pair[2]))
		if key != "" {
			result[key] = value
		}
	}

	return result
}

func stripHTMLTags(s string) string {
	tagRe := regexp.MustCompile(`<[^>]*>`)
	result := tagRe.ReplaceAllString(s, "")
	result = strings.ReplaceAll(result, "&amp;", "&")
	result = strings.ReplaceAll(result, "&lt;", "<")
	result = strings.ReplaceAll(result, "&gt;", ">")
	result = strings.ReplaceAll(result, "&quot;", "\"")
	result = strings.ReplaceAll(result, "&#39;", "'")
	result = strings.ReplaceAll(result, "&apos;", "'")
	return result
}

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

func renderJSON(data any) ([]byte, error) {
	return json.MarshalIndent(data, "", "  ")
}

func renderYAML(data any) ([]byte, error) {
	return yaml.Marshal(data)
}

func renderTOML(data any) ([]byte, error) {
	var buf bytes.Buffer
	encoder := toml.NewEncoder(&buf)
	if err := encoder.Encode(data); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func renderCSV(data any) ([]byte, error) {
	rows, ok := toSlice(data)
	if !ok || len(rows) == 0 {
		return nil, fmt.Errorf("CSV output requires an array of objects")
	}

	headers := extractHeaders(rows)
	if len(headers) == 0 {
		return nil, fmt.Errorf("CSV output requires objects with keys")
	}

	var buf bytes.Buffer
	writer := csv.NewWriter(&buf)

	writer.Write(headers)
	for _, row := range rows {
		obj, ok := toMap(row)
		if !ok {
			continue
		}
		record := make([]string, len(headers))
		for i, h := range headers {
			record[i] = fmt.Sprintf("%v", obj[h])
		}
		writer.Write(record)
	}
	writer.Flush()
	return buf.Bytes(), nil
}

func renderXML(data any) ([]byte, error) {
	var buf bytes.Buffer
	buf.WriteString(xml.Header)

	if obj, ok := toMap(data); ok {
		if len(obj) == 1 {
			for k, v := range obj {
				writeXMLElement(&buf, k, v, 0)
			}
		} else {
			buf.WriteString("<root>\n")
			writeXMLContent(&buf, obj, 1)
			buf.WriteString("</root>\n")
		}
	} else if arr, ok := toSlice(data); ok {
		buf.WriteString("<root>\n")
		for _, item := range arr {
			writeXMLElement(&buf, "item", item, 1)
		}
		buf.WriteString("</root>\n")
	} else {
		buf.WriteString(fmt.Sprintf("<value>%s</value>\n", escapeXML(fmt.Sprintf("%v", data))))
	}

	return buf.Bytes(), nil
}

func writeXMLElement(buf *bytes.Buffer, name string, value any, indent int) {
	indentStr := strings.Repeat("  ", indent)

	if obj, ok := toMap(value); ok {
		attrs, children := separateXMLAttrs(obj)
		buf.WriteString(indentStr + "<" + xmlSafeName(name))
		for k, v := range attrs {
			buf.WriteString(fmt.Sprintf(" %s=\"%s\"", k[1:], escapeXML(fmt.Sprintf("%v", v))))
		}
		if len(children) == 0 {
			buf.WriteString("/>\n")
		} else if text, hasText := children["#text"]; hasText && len(children) == 1 {
			buf.WriteString(">" + escapeXML(fmt.Sprintf("%v", text)) + "</" + xmlSafeName(name) + ">\n")
		} else {
			buf.WriteString(">\n")
			writeXMLContent(buf, children, indent+1)
			buf.WriteString(indentStr + "</" + xmlSafeName(name) + ">\n")
		}
	} else if arr, ok := toSlice(value); ok {
		for _, item := range arr {
			writeXMLElement(buf, name, item, indent)
		}
	} else {
		buf.WriteString(indentStr + "<" + xmlSafeName(name) + ">" + escapeXML(fmt.Sprintf("%v", value)) + "</" + xmlSafeName(name) + ">\n")
	}
}

func writeXMLContent(buf *bytes.Buffer, obj map[string]any, indent int) {
	keys := sortedKeys(obj)
	for _, k := range keys {
		if k == "#text" {
			continue
		}
		writeXMLElement(buf, k, obj[k], indent)
	}
}

func separateXMLAttrs(obj map[string]any) (attrs map[string]any, children map[string]any) {
	attrs = make(map[string]any)
	children = make(map[string]any)
	for k, v := range obj {
		if strings.HasPrefix(k, "@") {
			attrs[k] = v
		} else {
			children[k] = v
		}
	}
	return
}

func xmlSafeName(name string) string {
	safe := strings.Map(func(r rune) rune {
		if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') || r == '_' || r == '-' || r == '.' {
			return r
		}
		return '_'
	}, name)
	if len(safe) > 0 && (safe[0] >= '0' && safe[0] <= '9') {
		safe = "_" + safe
	}
	if safe == "" {
		safe = "element"
	}
	return safe
}

func escapeXML(s string) string {
	s = strings.ReplaceAll(s, "&", "&amp;")
	s = strings.ReplaceAll(s, "<", "&lt;")
	s = strings.ReplaceAll(s, ">", "&gt;")
	s = strings.ReplaceAll(s, "\"", "&quot;")
	s = strings.ReplaceAll(s, "'", "&apos;")
	return s
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

func renderHTML(data any) ([]byte, error) {
	var buf bytes.Buffer

	if rows, ok := toSlice(data); ok && len(rows) > 0 {
		if _, isMap := toMap(rows[0]); isMap {
			renderHTMLTable(&buf, rows)
		} else {
			renderHTMLList(&buf, rows)
		}
	} else if obj, ok := toMap(data); ok {
		renderHTMLKeyValue(&buf, obj)
	} else {
		buf.WriteString(fmt.Sprintf("<p>%v</p>\n", escapeHTML(fmt.Sprintf("%v", data))))
	}

	return buf.Bytes(), nil
}

func renderHTMLTable(buf *bytes.Buffer, rows []any) {
	headers := extractHeaders(rows)
	if len(headers) == 0 {
		return
	}

	buf.WriteString("<table>\n<thead>\n<tr>\n")
	for _, h := range headers {
		buf.WriteString(fmt.Sprintf("  <th>%s</th>\n", escapeHTML(h)))
	}
	buf.WriteString("</tr>\n</thead>\n<tbody>\n")

	for _, row := range rows {
		obj, _ := toMap(row)
		buf.WriteString("<tr>\n")
		for _, h := range headers {
			buf.WriteString(fmt.Sprintf("  <td>%s</td>\n", escapeHTML(fmt.Sprintf("%v", obj[h]))))
		}
		buf.WriteString("</tr>\n")
	}
	buf.WriteString("</tbody>\n</table>\n")
}

func renderHTMLList(buf *bytes.Buffer, items []any) {
	buf.WriteString("<ul>\n")
	for _, item := range items {
		buf.WriteString(fmt.Sprintf("  <li>%s</li>\n", escapeHTML(fmt.Sprintf("%v", item))))
	}
	buf.WriteString("</ul>\n")
}

func renderHTMLKeyValue(buf *bytes.Buffer, obj map[string]any) {
	buf.WriteString("<dl>\n")
	keys := sortedKeys(obj)
	for _, k := range keys {
		v := obj[k]
		buf.WriteString(fmt.Sprintf("  <dt>%s</dt>\n", escapeHTML(k)))
		if nested, ok := toMap(v); ok {
			buf.WriteString("  <dd>\n")
			renderHTMLKeyValue(buf, nested)
			buf.WriteString("  </dd>\n")
		} else if arr, ok := toSlice(v); ok {
			buf.WriteString("  <dd>\n")
			if len(arr) > 0 {
				if _, isMap := toMap(arr[0]); isMap {
					renderHTMLTable(buf, arr)
				} else {
					renderHTMLList(buf, arr)
				}
			}
			buf.WriteString("  </dd>\n")
		} else {
			buf.WriteString(fmt.Sprintf("  <dd>%s</dd>\n", escapeHTML(fmt.Sprintf("%v", v))))
		}
	}
	buf.WriteString("</dl>\n")
}

func toSlice(v any) ([]any, bool) {
	switch s := v.(type) {
	case []any:
		return s, true
	case []map[string]any:
		result := make([]any, len(s))
		for i, m := range s {
			result[i] = m
		}
		return result, true
	}
	return nil, false
}

func toMap(v any) (map[string]any, bool) {
	switch m := v.(type) {
	case map[string]any:
		return m, true
	case map[any]any:
		result := make(map[string]any)
		for k, val := range m {
			result[fmt.Sprintf("%v", k)] = val
		}
		return result, true
	}
	return nil, false
}

func extractHeaders(rows []any) []string {
	seen := make(map[string]bool)
	var headers []string

	for _, row := range rows {
		obj, ok := toMap(row)
		if !ok {
			continue
		}
		for k := range obj {
			if !seen[k] {
				seen[k] = true
				headers = append(headers, k)
			}
		}
	}
	sort.Strings(headers)
	return headers
}

func sortedKeys(m map[string]any) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}

func escapeMarkdown(s string) string {
	s = strings.ReplaceAll(s, "|", "\\|")
	s = strings.ReplaceAll(s, "\n", " ")
	return s
}

func escapeHTML(s string) string {
	s = strings.ReplaceAll(s, "&", "&amp;")
	s = strings.ReplaceAll(s, "<", "&lt;")
	s = strings.ReplaceAll(s, ">", "&gt;")
	s = strings.ReplaceAll(s, "\"", "&quot;")
	return s
}

func normalizeYAMLMaps(v any) any {
	switch x := v.(type) {
	case map[any]any:
		m := make(map[string]any)
		for k, val := range x {
			m[fmt.Sprintf("%v", k)] = normalizeYAMLMaps(val)
		}
		return m
	case map[string]any:
		m := make(map[string]any)
		for k, val := range x {
			m[k] = normalizeYAMLMaps(val)
		}
		return m
	case []any:
		for i, val := range x {
			x[i] = normalizeYAMLMaps(val)
		}
		return x
	default:
		return v
	}
}
