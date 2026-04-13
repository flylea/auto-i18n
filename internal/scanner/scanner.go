package scanner

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/flylea/auto-i18n/internal/config"
	"github.com/flylea/auto-i18n/internal/types"
)

type Scanner struct {
	cfg   *config.Config
	fnMap map[string]bool
}

func New(cfg *config.Config) *Scanner {
	fnMap := make(map[string]bool)
	for _, fn := range cfg.I18nFns {
		fnMap[fn] = true
	}
	return &Scanner{cfg: cfg, fnMap: fnMap}
}

func (s *Scanner) Scan() ([]string, types.KeyFileMap, error) {
	keyFileMap := make(types.KeyFileMap)
	allKeys := make(map[string]bool)

	extSet := make(map[string]bool)
	for _, ext := range s.cfg.Extensions {
		extSet["."+ext] = true
	}

	for _, dir := range s.cfg.EntryDirs {
		if _, err := os.Stat(dir); os.IsNotExist(err) {
			continue
		}

		err := filepath.WalkDir(dir, func(path string, info os.DirEntry, err error) error {
			if err != nil {
				return nil
			}
			if info.IsDir() {
				return nil
			}
			if !extSet[strings.ToLower(filepath.Ext(path))] {
				return nil
			}

			keys, err := s.extractKeys(path)
			if err != nil {
				fmt.Printf("Warning: failed to process %s: %v\n", path, err)
				return nil
			}

			for _, key := range keys {
				allKeys[key] = true
				if keyFileMap[key] == nil {
					keyFileMap[key] = make(map[string]bool)
				}
				keyFileMap[key][path] = true
			}
			return nil
		})
		if err != nil {
			continue
		}
	}

	keys := make([]string, 0, len(allKeys))
	for k := range allKeys {
		keys = append(keys, k)
	}

	return keys, keyFileMap, nil
}

func (s *Scanner) extractKeys(file string) ([]string, error) {
	content, err := os.ReadFile(file)
	if err != nil {
		return nil, err
	}

	ext := strings.ToLower(filepath.Ext(file))
	if ext == ".vue" {
		return s.extractFromVue(string(content))
	}
	return s.extractFromCode(string(content))
}

func (s *Scanner) extractFromVue(content string) ([]string, error) {
	keys := make(map[string]bool)

	// Extract script section
	scriptMatch := regexp.MustCompile(`<script[^>]*>([\s\S]*?)</script>`).FindStringSubmatch(content)
	if len(scriptMatch) > 1 {
		scriptKeys, _ := s.extractFromCode(scriptMatch[1])
		for _, k := range scriptKeys {
			keys[k] = true
		}
	}

	// Extract template section - look for {{ t('...') }} or v-bind:t
	templateMatch := regexp.MustCompile(`<template[^>]*>([\s\S]*?)</template>`).FindStringSubmatch(content)
	if len(templateMatch) > 1 {
		template := templateMatch[1]
		// Match t('key'), $t('key'), t("key"), etc.
		re := regexp.MustCompile(`(?:t|\$t)\s*\(\s*['"]([^'"]+)['"]\s*\)`)
		matches := re.FindAllStringSubmatch(template, -1)
		for _, m := range matches {
			if len(m) > 1 {
				keys[m[1]] = true
			}
		}
	}

	result := make([]string, 0, len(keys))
	for k := range keys {
		result = append(result, k)
	}
	return result, nil
}

func (s *Scanner) extractFromCode(content string) ([]string, error) {
	keys := make(map[string]bool)

	// Match patterns like t('key'), $t('key'), t("key")
	re := regexp.MustCompile(`(?:t|\$t)\s*\(\s*['"]([^'"]+)['"]\s*\)`)
	matches := re.FindAllStringSubmatch(content, -1)
	for _, m := range matches {
		if len(m) > 1 {
			keys[m[1]] = true
		}
	}

	result := make([]string, 0, len(keys))
	for k := range keys {
		result = append(result, k)
	}
	return result, nil
}
