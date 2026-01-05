package transmute

import "encoding/json"

func parseJSON(input []byte) (any, error) {
	var data any
	if err := json.Unmarshal(input, &data); err != nil {
		return nil, err
	}
	return data, nil
}

func renderJSON(data any) ([]byte, error) {
	return json.MarshalIndent(data, "", "  ")
}
