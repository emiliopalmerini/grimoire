package polymorph

import (
	"bytes"

	"github.com/BurntSushi/toml"
)

func parseTOML(input []byte) (any, error) {
	var data any
	if err := toml.Unmarshal(input, &data); err != nil {
		return nil, err
	}
	return data, nil
}

func renderTOML(data any) ([]byte, error) {
	var buf bytes.Buffer
	encoder := toml.NewEncoder(&buf)
	if err := encoder.Encode(data); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}
