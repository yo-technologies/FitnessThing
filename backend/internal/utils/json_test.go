package utils

import (
	"reflect"
	"testing"
)

func TestSplitMultipleJSONObjects(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected []string
	}{
		{
			name:     "empty string",
			input:    "",
			expected: []string{""},
		},
		{
			name:     "single object",
			input:    `{"limit":5}`,
			expected: []string{`{"limit":5}`},
		},
		{
			name:     "two empty objects",
			input:    `{}{}`,
			expected: []string{`{}`, `{}`},
		},
		{
			name:     "three empty objects",
			input:    `{}{}{}`,
			expected: []string{`{}`, `{}`, `{}`},
		},
		{
			name:     "two objects with data",
			input:    `{"limit":1,"muscle_group_ids":[],"exclude_exercise_ids":[]}{"limit":2,"muscle_group_ids":[],"exclude_exercise_ids":[]}`,
			expected: []string{`{"limit":1,"muscle_group_ids":[],"exclude_exercise_ids":[]}`, `{"limit":2,"muscle_group_ids":[],"exclude_exercise_ids":[]}`},
		},
		{
			name:     "multiple objects with whitespace",
			input:    `{"a":1}  {"b":2}  {"c":3}`,
			expected: []string{`{"a":1}`, `{"b":2}`, `{"c":3}`},
		},
		{
			name:     "invalid json",
			input:    `{invalid}`,
			expected: []string{`{invalid}`},
		},
		{
			name:     "complex nested object",
			input:    `{"nested":{"key":"value"}}`,
			expected: []string{`{"nested":{"key":"value"}}`},
		},
		{
			name:     "array values",
			input:    `{"arr":[1,2,3]}{"arr":[4,5,6]}`,
			expected: []string{`{"arr":[1,2,3]}`, `{"arr":[4,5,6]}`},
		},
		{
			name:     "whitespace only",
			input:    "   \n\t  ",
			expected: []string{""},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := SplitMultipleJSONObjects(tt.input)
			if !reflect.DeepEqual(result, tt.expected) {
				t.Errorf("SplitMultipleJSONObjects() = %v, want %v", result, tt.expected)
			}
		})
	}
}
