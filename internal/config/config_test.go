package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestDefault(t *testing.T) {
	cfg := Default()

	if len(cfg.EntryDirs) != 2 {
		t.Errorf("Default() EntryDirs = %v, want 2 elements", len(cfg.EntryDirs))
	}
	if cfg.Output != "ts" {
		t.Errorf("Default() Output = %q, want %q", cfg.Output, "ts")
	}
	if cfg.OutputDir != "src/i18n/locale" {
		t.Errorf("Default() OutputDir = %q, want %q", cfg.OutputDir, "src/i18n/locale")
	}
	if len(cfg.Languages) != 2 {
		t.Errorf("Default() Languages = %v, want 2 elements", len(cfg.Languages))
	}
	if cfg.BatchSize != 10 {
		t.Errorf("Default() BatchSize = %d, want 10", cfg.BatchSize)
	}
}

func TestLoad(t *testing.T) {
	// Create temp config file
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "i18n.config.json")

	configContent := `{
		"entryDirs": ["src/views"],
		"extensions": ["vue"],
		"output": "json",
		"outputDir": "locale",
		"languages": ["ja-JP"],
		"batchSize": 5
	}`

	if err := os.WriteFile(configPath, []byte(configContent), 0644); err != nil {
		t.Fatalf("Failed to write temp config: %v", err)
	}

	cfg, err := Load(configPath)
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	if len(cfg.EntryDirs) != 1 || cfg.EntryDirs[0] != "src/views" {
		t.Errorf("Load() EntryDirs = %v, want [src/views]", cfg.EntryDirs)
	}
	if cfg.Output != "json" {
		t.Errorf("Load() Output = %q, want %q", cfg.Output, "json")
	}
	if len(cfg.Languages) != 1 || cfg.Languages[0] != "ja-JP" {
		t.Errorf("Load() Languages = %v, want [ja-JP]", cfg.Languages)
	}
	if cfg.BatchSize != 5 {
		t.Errorf("Load() BatchSize = %d, want 5", cfg.BatchSize)
	}
}

func TestLoadNonExistent(t *testing.T) {
	cfg, err := Load("nonexistent.json")
	if err != nil {
		t.Fatalf("Load() error = %v, want nil for nonexistent file (should use default)", err)
	}

	// Should return default config
	if cfg.Output != "ts" {
		t.Errorf("Load() for nonexistent file should return Default(), got Output = %q", cfg.Output)
	}
}

func TestLoadInvalidJSON(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "invalid.json")

	if err := os.WriteFile(configPath, []byte("{invalid json}"), 0644); err != nil {
		t.Fatalf("Failed to write temp config: %v", err)
	}

	_, err := Load(configPath)
	if err == nil {
		t.Error("Load() should return error for invalid JSON")
	}
}
