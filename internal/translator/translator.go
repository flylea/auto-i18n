package translator

import (
	"fmt"
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
	apiKey, model := GetEnvConfig()

	// Default to Deepseek provider
	provider := NewDeepseekProvider(apiKey, model, cfg.BatchSize)

	// Create cache if output dir is configured
	var cache *Cache
	if cfg.OutputDir != "" {
		var err error
		cache, err = NewCache(cfg)
		if err != nil {
			fmt.Printf("Warning: failed to create cache: %v\n", err)
		}
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
	if len(keys) == 0 {
		return make(types.TranslateResult), 0, 0, nil
	}

	results := make(types.TranslateResult, len(keys))
	fromCache := 0
	translated := 0

	// Pre-load cached translations and determine per-language missing keys in one pass
	langKeysToTranslate := make(map[string][]string) // lang -> keys needing translation
	cachedTranslations := make(map[string]map[string]string)

	for _, lang := range t.cfg.Languages {
		if t.cache != nil {
			cachedTranslations[lang] = t.cache.GetMulti(keys, lang, t.provider.Name())
			fromCache += len(cachedTranslations[lang])
		} else {
			cachedTranslations[lang] = make(map[string]string)
		}

		// Build missing keys for this language while we have the cache map
		var missing []string
		for _, key := range keys {
			if _, ok := cachedTranslations[lang][key]; !ok {
				missing = append(missing, key)
			}
		}
		if len(missing) > 0 {
			langKeysToTranslate[lang] = missing
		}
	}

	// If all keys are cached, just build results from cache
	if len(langKeysToTranslate) == 0 {
		for _, key := range keys {
			results[key] = make(map[string]string)
			for _, lang := range t.cfg.Languages {
				results[key][lang] = cachedTranslations[lang][key]
			}
		}
		return results, fromCache, 0, nil
	}

	// Translate missing keys concurrently per language
	concurrency := 3
	sem := make(chan struct{}, concurrency)
	var wg sync.WaitGroup
	var mu sync.Mutex

	for lang, neededKeys := range langKeysToTranslate {
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
					if valid, missing := ValidatePlaceholders(key, trans); !valid && len(missing) > 0 {
						fmt.Printf("Warning: placeholder mismatch for key %q: missing %v\n", key, missing)
					}

					if t.cache != nil {
						t.cache.Set(key, targetLang, t.provider.Name(), trans)
					}

					mu.Lock()
					// Store translation directly in results
					if results[key] == nil {
						results[key] = make(map[string]string)
					}
					results[key][targetLang] = trans
					translated++
					mu.Unlock()
				}
			}

			// Rate limiting
			time.Sleep(100 * time.Millisecond)
		}(lang, neededKeys)
	}

	wg.Wait()

	// Add cached translations for keys not yet in results
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

	// Save cache if new translations were added
	if t.cache != nil && translated > 0 {
		_ = t.cache.Save()
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
