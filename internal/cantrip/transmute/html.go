package transmute

import (
	"bytes"
	"fmt"
	"regexp"
	"strings"
)

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

func escapeHTML(s string) string {
	s = strings.ReplaceAll(s, "&", "&amp;")
	s = strings.ReplaceAll(s, "<", "&lt;")
	s = strings.ReplaceAll(s, ">", "&gt;")
	s = strings.ReplaceAll(s, "\"", "&quot;")
	return s
}
