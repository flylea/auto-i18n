package translator

import (
	"testing"

	"github.com/flylea/auto-i18n/internal/config"
)

func TestCacheBasic(t *testing.T) {
	tmpDir := t.TempDir()
	cfg := &config.Config{OutputDir: tmpDir}

	cache, err := NewCache(cfg)
	if err != nil {
		t.Fatalf("NewCache() error = %v", err)
	}

	// Test Set and Get
	cache.Set("hello", "en-US", "deepseek-chat", "Hello")
	trans, ok := cache.Get("hello", "en-US", "deepseek-chat")
	if !ok {
		t.Fatal("Get() returned false, want true")
	}
	if trans != "Hello" {
		t.Errorf("Get() = %q, want %q", trans, "Hello")
	}

	// Test GetMulti
	cache.Set("world", "en-US", "deepseek-chat", "World")
	cache.Set("bye", "en-US", "deepseek-chat", "Goodbye")

	multi := cache.GetMulti([]string{"hello", "world", "nonexistent"}, "en-US", "deepseek-chat")
	if len(multi) != 2 {
		t.Errorf("GetMulti() returned %d keys, want 2", len(multi))
	}
}

func TestCacheMiss(t *testing.T) {
	tmpDir := t.TempDir()
	cfg := &config.Config{OutputDir: tmpDir}

	cache, _ := NewCache(cfg)

	_, ok := cache.Get("nonexistent", "en-US", "deepseek-chat")
	if ok {
		t.Error("Get() for nonexistent key returned true, want false")
	}
}

func TestCacheSaveLoad(t *testing.T) {
	tmpDir := t.TempDir()
	cfg := &config.Config{OutputDir: tmpDir}

	cache1, _ := NewCache(cfg)
	cache1.Set("key1", "en-US", "deepseek-chat", "Value1")
	cache1.Set("key2", "fr-FR", "deepseek-chat", "Valeur2")

	if err := cache1.Save(); err != nil {
		t.Fatalf("Save() error = %v", err)
	}

	// Create new cache instance - should load from disk
	cache2, _ := NewCache(cfg)

	trans, ok := cache2.Get("key1", "en-US", "deepseek-chat")
	if !ok || trans != "Value1" {
		t.Errorf("After Save/Load, Get() = %q, ok=%v, want %q, true", trans, ok, "Value1")
	}

	trans2, ok := cache2.Get("key2", "fr-FR", "deepseek-chat")
	if !ok || trans2 != "Valeur2" {
		t.Errorf("After Save/Load, Get() for key2 = %q, ok=%v, want %q, true", trans2, ok, "Valeur2")
	}
}

func TestCacheClear(t *testing.T) {
	tmpDir := t.TempDir()
	cfg := &config.Config{OutputDir: tmpDir}

	cache, _ := NewCache(cfg)
	cache.Set("hello", "en-US", "deepseek-chat", "Hello")

	cache.Clear()

	_, ok := cache.Get("hello", "en-US", "deepseek-chat")
	if ok {
		t.Error("After Clear(), Get() returned true, want false")
	}
}

func TestCacheInvalidate(t *testing.T) {
	tmpDir := t.TempDir()
	cfg := &config.Config{OutputDir: tmpDir}

	cache, _ := NewCache(cfg)
	cache.Set("hello", "en-US", "deepseek-chat", "Hello")
	cache.Set("world", "en-US", "deepseek-chat", "World")

	cache.Invalidate([]string{"hello"}, "en-US", "deepseek-chat")

	_, ok := cache.Get("hello", "en-US", "deepseek-chat")
	if ok {
		t.Error("After Invalidate(), Get() for hello returned true, want false")
	}

	_, ok = cache.Get("world", "en-US", "deepseek-chat")
	if !ok {
		t.Error("After Invalidate([hello]), Get() for world returned false, want true")
	}
}

func TestExtractPlaceholders(t *testing.T) {
	tests := []struct {
		input    string
		expected []string
	}{
		{"hello", nil},
		{"hello {name}", []string{"name"}},
		{"{count} items", []string{"count"}},
		{"{current} of {total}", []string{"current", "total"}},
		{"%s is invalid", []string{"%s"}},
		{"value: %d", []string{"%d"}},
		{"${variable}", []string{"variable"}},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := ExtractPlaceholders(tt.input)
			if len(result) != len(tt.expected) {
				t.Errorf("ExtractPlaceholders(%q) = %v, want %v", tt.input, result, tt.expected)
			}
		})
	}
}

func TestValidatePlaceholders(t *testing.T) {
	tests := []struct {
		name      string
		original  string
		translated string
		wantOK    bool
	}{
		{
			name:      "preserved",
			original:  "hello {name}",
			translated: "hola {name}",
			wantOK:    true,
		},
		{
			name:      "missing placeholder",
			original:  "hello {name}",
			translated: "hola nombre",
			wantOK:    false,
		},
		{
			name:      "no placeholders",
			original:  "hello",
			translated: "hola",
			wantOK:    true,
		},
		{
			name:      "different placeholder style",
			original:  "{count} items",
			translated: "{total} articulos",
			wantOK:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ok, _ := ValidatePlaceholders(tt.original, tt.translated)
			if ok != tt.wantOK {
				t.Errorf("ValidatePlaceholders() ok = %v, want %v", ok, tt.wantOK)
			}
		})
	}
}

func TestModelVersion(t *testing.T) {
	result := ModelVersion("deepseek", "deepseek-chat")
	if result != "deepseek-deepseek-chat" {
		t.Errorf("ModelVersion() = %q, want %q", result, "deepseek-deepseek-chat")
	}
}
