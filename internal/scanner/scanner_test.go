package scanner

import (
	"testing"

	"github.com/flylea/auto-i18n/internal/config"
)

func TestExtractFromCode(t *testing.T) {
	s := &Scanner{}

	tests := []struct {
		name     string
		code     string
		expected []string
	}{
		{
			name:     "empty code",
			code:     "",
			expected: nil,
		},
		{
			name:     "t('key')",
			code:     "const msg = t('hello')",
			expected: []string{"hello"},
		},
		{
			name:     "$t('key')",
			code:     "const msg = $t('welcome')",
			expected: []string{"welcome"},
		},
		{
			name:     "i18n.t('key')",
			code:     "const msg = i18n.t('greeting')",
			expected: []string{"greeting"},
		},
		{
			name:     "useTranslation",
			code:     "const { t } = useTranslation(); t('message')",
			expected: []string{"message"},
		},
		{
			name:     "multiple keys",
			code:     "t('a'); t('b'); t('c')",
			expected: []string{"a", "b", "c"},
		},
		{
			name:     "key with dot",
			code:     "t('user.name')",
			expected: []string{"user.name"},
		},
		{
			name:     "double quoted",
			code:     `t("doublequoted")`,
			expected: []string{"doublequoted"},
		},
		{
			name:     "no false positive",
			code:     "const test = 'hello'; total += 1;",
			expected: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := s.extractFromCode(tt.code)

			// Convert to map for order-independent comparison
			resultMap := make(map[string]bool)
			for _, k := range result {
				resultMap[k] = true
			}

			for _, k := range tt.expected {
				if !resultMap[k] {
					t.Errorf("extractFromCode() missing key %q", k)
				}
			}
		})
	}
}

func TestExtractFromVue(t *testing.T) {
	s := &Scanner{}

	tests := []struct {
		name     string
		vue      string
		expected []string
	}{
		{
			name:     "empty",
			vue:      `<template><div>Hello</div></template>`,
			expected: nil,
		},
		{
			name: "script only",
			vue: `<script>
				export default {
					methods: {
						greet() {
							return t('hello')
						}
					}
				}
			</script>`,
			expected: []string{"hello"},
		},
		{
			name: "template only",
			vue: `<template>
				<div>{{ t('welcome') }}</div>
			</template>`,
			expected: []string{"welcome"},
		},
		{
			name: "both script and template",
			vue: `<template>
				<div>{{ t('greeting') }}</div>
			</template>
			<script>
				methods: {
					say() { return t('bye') }
				}
			</script>`,
			expected: []string{"greeting", "bye"},
		},
		{
			name: "script with lang attribute",
			vue: `<script lang="ts">
				t('typescript_key')
			</script>`,
			expected: []string{"typescript_key"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := s.extractFromVue(tt.vue)
			if err != nil {
				t.Fatalf("extractFromVue() error = %v", err)
			}

			resultMap := make(map[string]bool)
			for _, k := range result {
				resultMap[k] = true
			}

			for _, k := range tt.expected {
				if !resultMap[k] {
					t.Errorf("extractFromVue() missing key %q", k)
				}
			}
		})
	}
}

func TestMapToSlice(t *testing.T) {
	input := map[string]bool{"a": true, "b": true, "c": true}
	result := mapToSlice(input)

	if len(result) != 3 {
		t.Errorf("mapToSlice() got %d elements, want 3", len(result))
	}
}

func TestDeduplicate(t *testing.T) {
	tests := []struct {
		name     string
		input    []string
		expected []string
	}{
		{
			name:     "no duplicates",
			input:    []string{"a", "b", "c"},
			expected: []string{"a", "b", "c"},
		},
		{
			name:     "with duplicates",
			input:    []string{"a", "b", "a", "c", "b"},
			expected: []string{"a", "b", "c"},
		},
		{
			name:     "empty",
			input:    []string{},
			expected: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := deduplicate(tt.input)
			if len(result) != len(tt.expected) {
				t.Errorf("deduplicate() got %d, want %d", len(result), len(tt.expected))
				return
			}
			seen := make(map[string]bool)
			for _, s := range result {
				if seen[s] {
					t.Errorf("deduplicate() has duplicate %q", s)
				}
				seen[s] = true
			}
		})
	}
}

func TestNew(t *testing.T) {
	cfg := &config.Config{
		EntryDirs:  []string{"src"},
		Extensions: []string{"vue", "jsx"},
	}

	scanner := New(cfg, false)
	if scanner == nil {
		t.Fatal("New() returned nil")
	}
	if scanner.forceAll != false {
		t.Errorf("New() forceAll = %v, want false", scanner.forceAll)
	}

	scannerAll := New(cfg, true)
	if scannerAll.forceAll != true {
		t.Errorf("New(cfg, true) forceAll = %v, want true", scannerAll.forceAll)
	}
}
