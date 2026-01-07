package polymorph

import (
	"reflect"
	"testing"
)

func TestToSlice(t *testing.T) {
	tests := []struct {
		name   string
		input  any
		want   []any
		wantOk bool
	}{
		{
			name:   "[]any",
			input:  []any{"a", "b", "c"},
			want:   []any{"a", "b", "c"},
			wantOk: true,
		},
		{
			name:   "[]map[string]any",
			input:  []map[string]any{{"key": "value"}},
			want:   []any{map[string]any{"key": "value"}},
			wantOk: true,
		},
		{
			name:   "string - not a slice",
			input:  "not a slice",
			want:   nil,
			wantOk: false,
		},
		{
			name:   "map - not a slice",
			input:  map[string]any{"key": "value"},
			want:   nil,
			wantOk: false,
		},
		{
			name:   "nil",
			input:  nil,
			want:   nil,
			wantOk: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, ok := toSlice(tt.input)
			if ok != tt.wantOk {
				t.Errorf("toSlice() ok = %v, wantOk %v", ok, tt.wantOk)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("toSlice() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestToMap(t *testing.T) {
	tests := []struct {
		name   string
		input  any
		want   map[string]any
		wantOk bool
	}{
		{
			name:   "map[string]any",
			input:  map[string]any{"key": "value"},
			want:   map[string]any{"key": "value"},
			wantOk: true,
		},
		{
			name:   "map[any]any",
			input:  map[any]any{"key": "value", 123: "num"},
			want:   map[string]any{"key": "value", "123": "num"},
			wantOk: true,
		},
		{
			name:   "string - not a map",
			input:  "not a map",
			want:   nil,
			wantOk: false,
		},
		{
			name:   "slice - not a map",
			input:  []any{"a", "b"},
			want:   nil,
			wantOk: false,
		},
		{
			name:   "nil",
			input:  nil,
			want:   nil,
			wantOk: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, ok := toMap(tt.input)
			if ok != tt.wantOk {
				t.Errorf("toMap() ok = %v, wantOk %v", ok, tt.wantOk)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("toMap() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestExtractHeaders(t *testing.T) {
	rows := []any{
		map[string]any{"name": "Alice", "age": 30},
		map[string]any{"name": "Bob", "city": "NYC"},
	}

	headers := extractHeaders(rows)

	expected := []string{"age", "city", "name"}
	if !reflect.DeepEqual(headers, expected) {
		t.Errorf("extractHeaders() = %v, want %v", headers, expected)
	}
}

func TestExtractHeaders_EmptyRows(t *testing.T) {
	headers := extractHeaders([]any{})
	if len(headers) != 0 {
		t.Errorf("extractHeaders([]) = %v, want empty", headers)
	}
}

func TestExtractHeaders_NonMapRows(t *testing.T) {
	rows := []any{"string", 123, true}
	headers := extractHeaders(rows)
	if len(headers) != 0 {
		t.Errorf("extractHeaders(non-maps) = %v, want empty", headers)
	}
}

func TestSortedKeys(t *testing.T) {
	m := map[string]any{
		"zebra":    1,
		"apple":    2,
		"banana":   3,
		"cucumber": 4,
	}

	keys := sortedKeys(m)
	expected := []string{"apple", "banana", "cucumber", "zebra"}

	if !reflect.DeepEqual(keys, expected) {
		t.Errorf("sortedKeys() = %v, want %v", keys, expected)
	}
}

func TestSortedKeys_Empty(t *testing.T) {
	keys := sortedKeys(map[string]any{})
	if len(keys) != 0 {
		t.Errorf("sortedKeys({}) = %v, want empty", keys)
	}
}
