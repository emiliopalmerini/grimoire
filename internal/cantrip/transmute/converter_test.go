package transmute

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestDetectFormat(t *testing.T) {
	tests := []struct {
		filename string
		want     string
	}{
		{"data.json", "json"},
		{"config.yaml", "yaml"},
		{"config.yml", "yaml"},
		{"config.toml", "toml"},
		{"data.csv", "csv"},
		{"readme.md", "markdown"},
		{"index.html", "html"},
		{"data.xml", "xml"},
		{"unknown.xyz", ""},
		{"DATA.JSON", "json"},
	}

	for _, tt := range tests {
		t.Run(tt.filename, func(t *testing.T) {
			got := DetectFormat(tt.filename)
			if got != tt.want {
				t.Errorf("DetectFormat(%q) = %q, want %q", tt.filename, got, tt.want)
			}
		})
	}
}

func TestNormalizeFormat(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"json", "json"},
		{"JSON", "json"},
		{"md", "markdown"},
		{"yml", "yaml"},
		{"  yaml  ", "yaml"},
		{"unknown", "unknown"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got := normalizeFormat(tt.input)
			if got != tt.want {
				t.Errorf("normalizeFormat(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

func TestConvert_JSONToYAML(t *testing.T) {
	input := []byte(`{"name": "test", "value": 42}`)
	output, err := Convert(input, "json", "yaml")
	if err != nil {
		t.Fatalf("Convert failed: %v", err)
	}

	if !strings.Contains(string(output), "name: test") {
		t.Errorf("Expected YAML output to contain 'name: test', got: %s", output)
	}
	if !strings.Contains(string(output), "value: 42") {
		t.Errorf("Expected YAML output to contain 'value: 42', got: %s", output)
	}
}

func TestConvert_YAMLToJSON(t *testing.T) {
	input := []byte("name: test\nvalue: 42")
	output, err := Convert(input, "yaml", "json")
	if err != nil {
		t.Fatalf("Convert failed: %v", err)
	}

	var result map[string]any
	if err := json.Unmarshal(output, &result); err != nil {
		t.Fatalf("Failed to parse JSON output: %v", err)
	}

	if result["name"] != "test" {
		t.Errorf("Expected name=test, got %v", result["name"])
	}
	if result["value"].(float64) != 42 {
		t.Errorf("Expected value=42, got %v", result["value"])
	}
}

func TestConvert_JSONToCSV(t *testing.T) {
	input := []byte(`[{"name": "Alice", "age": 30}, {"name": "Bob", "age": 25}]`)
	output, err := Convert(input, "json", "csv")
	if err != nil {
		t.Fatalf("Convert failed: %v", err)
	}

	lines := strings.Split(strings.TrimSpace(string(output)), "\n")
	if len(lines) != 3 {
		t.Fatalf("Expected 3 lines (header + 2 data rows), got %d", len(lines))
	}

	if !strings.Contains(lines[0], "age") || !strings.Contains(lines[0], "name") {
		t.Errorf("Expected header with age and name, got: %s", lines[0])
	}
}

func TestConvert_CSVToJSON(t *testing.T) {
	input := []byte("name,age\nAlice,30\nBob,25")
	output, err := Convert(input, "csv", "json")
	if err != nil {
		t.Fatalf("Convert failed: %v", err)
	}

	var result []map[string]any
	if err := json.Unmarshal(output, &result); err != nil {
		t.Fatalf("Failed to parse JSON output: %v", err)
	}

	if len(result) != 2 {
		t.Fatalf("Expected 2 records, got %d", len(result))
	}

	if result[0]["name"] != "Alice" {
		t.Errorf("Expected first name=Alice, got %v", result[0]["name"])
	}
}

func TestConvert_JSONToMarkdown(t *testing.T) {
	input := []byte(`[{"name": "Alice", "age": 30}, {"name": "Bob", "age": 25}]`)
	output, err := Convert(input, "json", "markdown")
	if err != nil {
		t.Fatalf("Convert failed: %v", err)
	}

	if !strings.Contains(string(output), "|") {
		t.Error("Expected markdown table with | characters")
	}
	if !strings.Contains(string(output), "---") {
		t.Error("Expected markdown table separator ---")
	}
}

func TestConvert_JSONToHTML(t *testing.T) {
	input := []byte(`[{"name": "Alice"}, {"name": "Bob"}]`)
	output, err := Convert(input, "json", "html")
	if err != nil {
		t.Fatalf("Convert failed: %v", err)
	}

	if !strings.Contains(string(output), "<table>") {
		t.Error("Expected HTML table")
	}
	if !strings.Contains(string(output), "<th>") {
		t.Error("Expected HTML table headers")
	}
}

func TestConvert_JSONToXML(t *testing.T) {
	input := []byte(`{"user": {"name": "Alice"}}`)
	output, err := Convert(input, "json", "xml")
	if err != nil {
		t.Fatalf("Convert failed: %v", err)
	}

	if !strings.Contains(string(output), "<user>") {
		t.Error("Expected XML user element")
	}
	if !strings.Contains(string(output), "<name>Alice</name>") {
		t.Error("Expected XML name element with Alice")
	}
}

func TestConvert_UnsupportedFormat(t *testing.T) {
	input := []byte(`{"test": true}`)

	_, err := Convert(input, "unknown", "json")
	if err == nil {
		t.Error("Expected error for unsupported input format")
	}

	_, err = Convert(input, "json", "unknown")
	if err == nil {
		t.Error("Expected error for unsupported output format")
	}
}

func TestConvert_InvalidInput(t *testing.T) {
	_, err := Convert([]byte("not valid json"), "json", "yaml")
	if err == nil {
		t.Error("Expected error for invalid JSON input")
	}
}
