package polymorph

import (
	"fmt"

	"gopkg.in/yaml.v3"
)

func parseYAML(input []byte) (any, error) {
	var data any
	if err := yaml.Unmarshal(input, &data); err != nil {
		return nil, err
	}
	return normalizeYAMLMaps(data), nil
}

func renderYAML(data any) ([]byte, error) {
	return yaml.Marshal(data)
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
