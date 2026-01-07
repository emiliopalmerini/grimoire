package polymorph

import (
	"bytes"
	"encoding/csv"
	"fmt"
)

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
