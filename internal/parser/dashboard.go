package parser

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
)

func ParseDashboardFile(path string) (map[string]any, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	return ParseDashboardBytes(data)
}

func ParseDashboardBytes(data []byte) (map[string]any, error) {
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.UseNumber()

	var raw any
	if err := decoder.Decode(&raw); err != nil {
		return nil, err
	}

	var extra any
	if err := decoder.Decode(&extra); err != io.EOF {
		if err == nil {
			return nil, errors.New("multiple JSON values found")
		}
		return nil, err
	}

	topLevel, ok := raw.(map[string]any)
	if !ok {
		return nil, fmt.Errorf("dashboard JSON must be an object")
	}

	if wrapped, ok := topLevel["dashboard"].(map[string]any); ok {
		return wrapped, nil
	}

	return topLevel, nil
}
