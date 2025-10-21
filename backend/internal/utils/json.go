package utils

import (
	"encoding/json"
	"io"
	"strings"
)

// SplitMultipleJSONObjects attempts to split a string containing multiple JSON objects
// into individual JSON objects. Returns the original string in a slice if parsing fails.
//
// This is useful when dealing with concatenated JSON objects like "{}{}" or
// `{"a":1}{"b":2}{"c":3}` which can occur when multiple API calls are batched together.
//
// Examples:
//   - Input: "{}{}" -> Output: ["{}", "{}"]
//   - Input: `{"limit":1}{"limit":2}` -> Output: [`{"limit":1}`, `{"limit":2}`]
//   - Input: `{"single":true}` -> Output: [`{"single":true}`]
func SplitMultipleJSONObjects(input string) []string {
	if input == "" {
		return []string{""}
	}

	// Trim whitespace
	input = strings.TrimSpace(input)
	if input == "" {
		return []string{""}
	}

	var results []string
	reader := strings.NewReader(input)
	decoder := json.NewDecoder(reader)

	for {
		var obj json.RawMessage
		err := decoder.Decode(&obj)
		if err != nil {
			// If EOF and we already parsed some objects, return them
			if err == io.EOF && len(results) > 0 {
				return results
			}
			// Otherwise, return the original input
			return []string{input}
		}
		results = append(results, string(obj))

		// Check if there's more data to parse
		if !decoder.More() {
			break
		}
	}

	// If we successfully parsed multiple objects, return them
	if len(results) > 1 {
		return results
	}

	// If only one object was parsed, verify there's no extra data
	// by trying to unmarshal the original input
	var test json.RawMessage
	if err := json.Unmarshal([]byte(input), &test); err != nil {
		// Original input is not valid JSON, but we managed to parse something
		// This means there might be concatenated objects
		return results
	}

	// Original input is valid JSON as a single object
	return []string{input}
}
