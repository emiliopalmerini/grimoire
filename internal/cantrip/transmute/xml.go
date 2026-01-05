package transmute

import (
	"bytes"
	"encoding/xml"
	"fmt"
	"strings"
)

type xmlNode struct {
	name     string
	attrs    map[string]string
	children []xmlNode
	text     string
}

func parseXML(input []byte) (any, error) {
	decoder := xml.NewDecoder(bytes.NewReader(input))
	return parseXMLDocument(decoder)
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
