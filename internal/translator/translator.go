package translator

import (
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/flylea/auto-i18n/internal/config"
	"github.com/flylea/auto-i18n/internal/types"
)

// Translator orchestrates translation using a provider and optional cache
type Translator struct {
	cfg      *config.Config
	provider Provider
	cache    *Cache
}

// New creates a new Translator with configured provider and cache
func New(cfg *config.Config) *Translator {
	apiKey, apiURL, model := GetEnvConfig()

	// Create provider based on API URL or default to Deepseek
	var provider Provider
	if strings.Contains(apiURL, "openai") || strings.Contains(apiURL, "api.openai") {
		provider = NewOpenAIProvider(apiKey, apiURL, model, cfg.BatchSize)
	} else {
		provider = NewDeepseekProvider(apiKey, apiURL, model, cfg.BatchSize)
	}

	// Create cache if output dir is configured
	var cache *Cache
	if cfg.OutputDir != "" {
		cache, _ = NewCache(cfg)
	}

	return &Translator{
		cfg:      cfg,
		provider: provider,
		cache:    cache,
	}
}

// Translate translates keys to all configured languages
// Returns results and cache statistics
func (t *Translator) Translate(keys []string) (types.TranslateResult, int, int, error) {
	results := make(types.TranslateResult)
	fromCache := 0
	translated := 0

	// Pre-load all cached translations
	cachedTranslations := make(map[string]map[string]string) // lang -> key -> translation
	for _, lang := range t.cfg.Languages {
		if t.cache != nil {
			cachedTranslations[lang] = t.cache.GetMulti(keys, lang, t.provider.Name())
			fromCache += len(cachedTranslations[lang])
		} else {
			cachedTranslations[lang] = make(map[string]string)
		}
	}

	// Determine which keys need translation
	var keysToTranslate []string
	for _, key := range keys {
		needsTranslation := false
		for _, lang := range t.cfg.Languages {
			if _, ok := cachedTranslations[lang][key]; !ok {
				needsTranslation = true
				break
			}
		}
		if needsTranslation {
			keysToTranslate = append(keysToTranslate, key)
		}
	}

	// Translate missing keys concurrently per language
	concurrency := 3
	sem := make(chan struct{}, concurrency)
	var wg sync.WaitGroup
	var mu sync.Mutex

	for _, lang := range t.cfg.Languages {
		// Filter keys that need this language
		var langKeys []string
		for _, key := range keysToTranslate {
			if _, ok := cachedTranslations[lang][key]; !ok {
				langKeys = append(langKeys, key)
			}
		}

		if len(langKeys) == 0 {
			// All keys cached for this language
			continue
		}

		wg.Add(1)
		go func(targetLang string, neededKeys []string) {
			defer wg.Done()
			sem <- struct{}{}
			defer func() { <-sem }()

			translations, err := t.provider.TranslateBatch(neededKeys, targetLang)
			if err != nil {
				fmt.Printf("Warning: failed to translate to %s: %v\n", targetLang, err)
				// Fallback: use key as translation
				translations = make(map[string]string)
				for _, k := range neededKeys {
					translations[k] = k
				}
			}

			// Validate placeholders and update cache
			for _, key := range neededKeys {
				if trans, ok := translations[key]; ok {
					// Validate placeholders
					if ok, missing := ValidatePlaceholders(key, trans); !ok && len(missing) > 0 {
						fmt.Printf("Warning: placeholder mismatch for key %q: missing %v\n", key, missing)
					}

					// Update cache
					if t.cache != nil {
						t.cache.Set(key, targetLang, t.provider.Name(), trans)
					}

					mu.Lock()
					translated++
					mu.Unlock()
				}
			}

			mu.Lock()
			for key, trans := range translations {
				if results[key] == nil {
					results[key] = make(map[string]string)
				}
				results[key][targetLang] = trans
			}
			mu.Unlock()

			// Rate limiting
			time.Sleep(100 * time.Millisecond)
		}(lang, langKeys)
	}

	wg.Wait()

	// Add cached translations to results
	for lang, keyMap := range cachedTranslations {
		for key, trans := range keyMap {
			if results[key] == nil {
				results[key] = make(map[string]string)
			}
			if results[key][lang] == "" {
				results[key][lang] = trans
			}
		}
	}

	// Fill in missing translations with key itself
	for _, key := range keys {
		if results[key] == nil {
			results[key] = make(map[string]string)
		}
		for _, lang := range t.cfg.Languages {
			if results[key][lang] == "" {
				results[key][lang] = key
			}
		}
	}

	// Save cache periodically
	if t.cache != nil && translated > 0 {
		t.cache.Save()
	}

	return results, fromCache, translated, nil
}

// TranslateWithCache is an alias for Translate that returns cache stats
func (t *Translator) TranslateWithCache(keys []string) (types.TranslateResult, error) {
	results, _, _, err := t.Translate(keys)
	return results, err
}

// InvalidateCache removes specific keys from cache
func (t *Translator) InvalidateCache(keys []string, lang string) {
	if t.cache != nil {
		t.cache.Invalidate(keys, lang, t.provider.Name())
	}
}

// ClearCache clears all cached translations
func (t *Translator) ClearCache() {
	if t.cache != nil {
		t.cache.Clear()
	}
}

// ProviderName returns the name of the current provider
func (t *Translator) ProviderName() string {
	return t.provider.Name()
}
