package config

import (
	"encoding/json"
	"os"
	"path/filepath"
)

type Config struct {
	EntryDirs         []string `json:"entryDirs"`
	Extensions        []string `json:"extensions"`
	Output            string   `json:"output"` // "ts", "js", or "json"
	OutputDir         string   `json:"outputDir"`
	WithFileComment   bool     `json:"withFileComment"`
	WithKeyFileComment bool    `json:"withKeyFileComment"`
	I18nFns          []string `json:"i18nFns"`
	SupportChineseKey bool     `json:"supportChineseKey"`
	Languages        []string `json:"languages"`
}

func Default() *Config {
	return &Config{
		EntryDirs:         []string{"src/views", "src/components"},
		Extensions:        []string{"vue"},
		Output:            "ts",
		OutputDir:         "src/i18n/locale",
		WithFileComment:   false,
		WithKeyFileComment: false,
		I18nFns:          []string{"t", "$t"},
		SupportChineseKey: true,
		Languages:        []string{"zh-CN", "en-US"},
	}
}

func Load(path string) (*Config, error) {
	if path == "" {
		path = "i18n.config.json"
	}

	data, err := os.ReadFile(path)
	if err != nil {
		// Try .ts extension
		tsPath := filepath.Join(filepath.Dir(path), "i18n.config.ts")
		data, err = os.ReadFile(tsPath)
		if err != nil {
			return Default(), nil
		}
	}

	var cfg Config
	if err := json.Unmarshal(data, &cfg); err != nil {
		return nil, err
	}

	return &cfg, nil
}
