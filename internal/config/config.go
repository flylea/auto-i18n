package config

import (
	"encoding/json"
	"os"
	"path/filepath"
)

type Config struct {
	EntryDirs          []string `json:"entryDirs"`
	Extensions         []string `json:"extensions"`
	Output             string   `json:"output"` // "ts", "js", or "json"
	OutputDir          string   `json:"outputDir"`
	WithFileComment    bool     `json:"withFileComment"`
	WithKeyFileComment bool     `json:"withKeyFileComment"`
	I18nFns            []string `json:"i18nFns"`
	SupportChineseKey  bool     `json:"supportChineseKey"`
	Languages          []string `json:"languages"`
	BatchSize          int      `json:"batchSize"` // 分组翻译每批数量，0 表示禁用分组
}

func Default() *Config {
	return &Config{
		EntryDirs:          []string{"src/views", "src/components"},
		Extensions:         []string{"vue", "jsx", "tsx"},
		Output:             "ts",
		OutputDir:          "src/i18n/locale",
		WithFileComment:    false,
		WithKeyFileComment: false,
		I18nFns:            []string{"t", "$t"},
		SupportChineseKey:  true,
		Languages:          []string{"zh-CN", "en-US"},
		BatchSize:          10, // 默认每批 10 个 key
	}
}

func Load(path string) (*Config, error) {
	if path == "" {
		path = "i18n.config.json"
	}

	data, err := os.ReadFile(path)
	if err != nil {
		// 尝试 TypeScript 配置作为备选
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
