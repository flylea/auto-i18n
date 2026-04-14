package writer

import (
	"testing"
)

func TestFlattenJSON(t *testing.T) {
	tests := []struct {
		name     string
		input    interface{}
		prefix   string
		expected map[string]string
	}{
		{
			name:   "simple key-value",
			input:  map[string]interface{}{"name": "John"},
			prefix: "",
			expected: map[string]string{
				"name": "John",
			},
		},
		{
			name: "nested object",
			input: map[string]interface{}{
				"user": map[string]interface{}{
					"name":  "John",
					"email": "john@example.com",
				},
			},
			prefix: "",
			expected: map[string]string{
				"user.name":  "John",
				"user.email": "john@example.com",
			},
		},
		{
			name: "deeply nested",
			input: map[string]interface{}{
				"a": map[string]interface{}{
					"b": map[string]interface{}{
						"c": "value",
					},
				},
			},
			prefix: "",
			expected: map[string]string{
				"a.b.c": "value",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := make(map[string]string)
			flattenJSON(tt.input, tt.prefix, result)

			if len(result) != len(tt.expected) {
				t.Errorf("flattenJSON() got %d keys, want %d", len(result), len(tt.expected))
				return
			}

			for key, val := range tt.expected {
				if result[key] != val {
					t.Errorf("flattenJSON()[%q] = %q, want %q", key, result[key], val)
				}
			}
		})
	}
}

func TestFlatToNested(t *testing.T) {
	tests := []struct {
		name     string
		input    map[string]string
		expected map[string]interface{}
	}{
		{
			name: "simple",
			input: map[string]string{
				"name": "John",
			},
			expected: map[string]interface{}{
				"name": "John",
			},
		},
		{
			name: "nested",
			input: map[string]string{
				"user.name":  "John",
				"user.email": "john@example.com",
			},
			expected: map[string]interface{}{
				"user": map[string]interface{}{
					"name":  "John",
					"email": "john@example.com",
				},
			},
		},
		{
			name: "mixed depth",
			input: map[string]string{
				"a":     "a value",
				"b.c":   "bc value",
				"b.d.e": "bde value",
			},
			expected: map[string]interface{}{
				"a": "a value",
				"b": map[string]interface{}{
					"c": "bc value",
					"d": map[string]interface{}{
						"e": "bde value",
					},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := flatToNested(tt.input)
			if !compareNested(result, tt.expected) {
				t.Errorf("flatToNested() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func compareNested(a, b map[string]interface{}) bool {
	if len(a) != len(b) {
		return false
	}
	for key, aVal := range a {
		bVal, ok := b[key]
		if !ok {
			return false
		}
		aMap, aIsMap := aVal.(map[string]interface{})
		bMap, bIsMap := bVal.(map[string]interface{})
		if aIsMap && bIsMap {
			if !compareNested(aMap, bMap) {
				return false
			}
		} else if aIsMap != bIsMap {
			return false
		} else if aVal != bVal {
			return false
		}
	}
	return true
}

func TestParseObject(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		prefix   string
		expected map[string]string
	}{
		{
			name:   "simple object",
			input:  `{ name: 'John', email: 'john@example.com' }`,
			prefix: "",
			expected: map[string]string{
				"name":  "John",
				"email": "john@example.com",
			},
		},
		{
			name:   "nested object",
			input:  `{ user: { name: 'John', email: 'john@example.com' } }`,
			prefix: "",
			expected: map[string]string{
				"user.name":  "John",
				"user.email": "john@example.com",
			},
		},
		{
			name:   "quoted keys",
			input:  `{ 'user.name': 'John', "user.email": 'john@example.com' }`,
			prefix: "",
			expected: map[string]string{
				"user.name":  "John",
				"user.email": "john@example.com",
			},
		},
		{
			name:   "empty object",
			input:  `{}`,
			prefix: "",
			expected: map[string]string{},
		},
		{
			name:   "number values",
			input:  `{ count: 42, ratio: 3.14 }`,
			prefix: "",
			expected: map[string]string{
				"count": "42",
				"ratio": "3.14",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := parseObject(tt.input, tt.prefix)
			if len(result) != len(tt.expected) {
				t.Errorf("parseObject() got %d keys, want %d", len(result), len(tt.expected))
				return
			}
			for key, val := range tt.expected {
				if result[key] != val {
					t.Errorf("parseObject()[%q] = %q, want %q", key, result[key], val)
				}
			}
		})
	}
}

func TestContainsChinese(t *testing.T) {
	tests := []struct {
		input    string
		expected bool
	}{
		{"hello", false},
		{"你好", true},
		{"hello世界", true},
		{"user_name", false},
		{"用户名", true},
		{"", false},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := containsChinese(tt.input)
			if result != tt.expected {
				t.Errorf("containsChinese(%q) = %v, want %v", tt.input, result, tt.expected)
			}
		})
	}
}
