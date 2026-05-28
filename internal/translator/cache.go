package translator

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sync"

	"github.com/flylea/auto-i18n/internal/config"
)

// Cache stores translation results to avoid redundant API calls
type Cache struct {
	mu      sync.RWMutex
	entries map[string]cacheEntry
	path   string
}

type cacheEntry struct {
	Translation string `json:"translation"`
	Lang       string `json:"lang"`
	Model      string `json:"model"`
}

func NewCache(cfg *config.Config) (*Cache, error) {
	cachePath := filepath.Join(cfg.OutputDir, ".i18n-cache.json")

	c := &Cache{
		entries: make(map[string]cacheEntry),
		path:    cachePath,
	}

	// Try to load existing cache
	if data, err := os.ReadFile(cachePath); err == nil {
		if err := json.Unmarshal(data, &c.entries); err != nil {
			// Invalid cache file, start fresh
			c.entries = make(map[string]cacheEntry)
		}
	}

	return c, nil
}

// cacheKey generates a unique key for (key, lang, model) pair
func cacheKey(key, lang, model string) string {
	return key + "|" + lang + "|" + model
}

// Get retrieves a cached translation
func (c *Cache) Get(key, lang, model string) (string, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	entry, ok := c.entries[cacheKey(key, lang, model)]
	if !ok {
		return "", false
	}
	if entry.Lang != lang {
		return "", false
	}
	return entry.Translation, true
}

// Set stores a translation in cache
func (c *Cache) Set(key, lang, model, translation string) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.entries[cacheKey(key, lang, model)] = cacheEntry{
		Translation: translation,
		Lang:       lang,
		Model:      model,
	}
}

// Save persists the cache to disk
func (c *Cache) Save() error {
	c.mu.RLock()
	defer c.mu.RUnlock()

	// Ensure directory exists
	if err := os.MkdirAll(filepath.Dir(c.path), 0755); err != nil {
		return err
	}

	data, err := json.MarshalIndent(c.entries, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(c.path, data, 0644)
}

// GetMulti retrieves multiple cached translations
func (c *Cache) GetMulti(keys []string, lang, model string) map[string]string {
	result := make(map[string]string)
	for _, key := range keys {
		if trans, ok := c.Get(key, lang, model); ok {
			result[key] = trans
		}
	}
	return result
}

// Invalidate removes entries for specific keys
func (c *Cache) Invalidate(keys []string, lang, model string) {
	c.mu.Lock()
	defer c.mu.Unlock()

	for _, key := range keys {
		delete(c.entries, cacheKey(key, lang, model))
	}
}

// Clear removes all cache entries
func (c *Cache) Clear() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.entries = make(map[string]cacheEntry)
}

// ExtractPlaceholders extracts placeholders like {name}, {{count}}, %s from a string
func ExtractPlaceholders(s string) []string {
	var placeholders []string
	seen := make(map[string]bool)

	// Match {name}, {count}, {0}, etc. - but NOT {{name}}
	// Use [^{}] to ensure we don't match nested braces
	re1 := regexp.MustCompile(`\{([^{}]+)\}`)
	for _, m := range re1.FindAllStringSubmatch(s, -1) {
		if len(m) > 1 && !seen[m[1]] {
			placeholders = append(placeholders, m[1])
			seen[m[1]] = true
		}
	}

	// Match %s, %d
	re3 := regexp.MustCompile(`%[sd]`)
	for _, m := range re3.FindAllStringSubmatch(s, -1) {
		if !seen[m[0]] {
			placeholders = append(placeholders, m[0])
			seen[m[0]] = true
		}
	}

	// Match ${variable}
	re4 := regexp.MustCompile(`\$\{([^}]+)\}`)
	for _, m := range re4.FindAllStringSubmatch(s, -1) {
		if len(m) > 1 && !seen[m[1]] {
			placeholders = append(placeholders, m[1])
			seen[m[1]] = true
		}
	}

	return placeholders
}

// ValidatePlaceholders checks if placeholders from original are preserved in translation
func ValidatePlaceholders(original, translated string) (bool, []string) {
	orig := ExtractPlaceholders(original)
	trans := ExtractPlaceholders(translated)

	transSet := make(map[string]bool)
	for _, p := range trans {
		transSet[p] = true
	}

	var missing []string
	for _, p := range orig {
		if !transSet[p] {
			missing = append(missing, p)
		}
	}

	return len(missing) == 0, missing
}

// ModelVersion returns a version string for the model (for cache invalidation)
func ModelVersion(provider, model string) string {
	return fmt.Sprintf("%s-%s", provider, model)
}
