package translator

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/flylea/auto-i18n/internal/config"
	"github.com/flylea/auto-i18n/internal/types"
)

type Translator struct {
	cfg    *config.Config
	apiKey string
	apiURL string
	client *http.Client
}

func New(cfg *config.Config) *Translator {
	apiKey := os.Getenv("DEEPSEEK_API_KEY")
	apiURL := os.Getenv("DEEPSEEK_API_URL")
	if apiURL == "" {
		apiURL = "https://api.deepseek.com"
	}

	return &Translator{
		cfg:    cfg,
		apiKey: apiKey,
		apiURL: apiURL,
		client: &http.Client{Timeout: 120 * time.Second},
	}
}

func (t *Translator) Translate(keys []string) (types.TranslateResult, error) {
	results := make(types.TranslateResult)

	// 并发控制：同时翻译多个 key
	concurrency := 3
	sem := make(chan struct{}, concurrency)

	var wg sync.WaitGroup
	var mu sync.Mutex

	for _, key := range keys {
		wg.Add(1)
		go func(k string) {
			defer wg.Done()
			sem <- struct{}{}
			defer func() { <-sem }()

			translations := make(map[string]string)
			for _, lang := range t.cfg.Languages {
				trans, err := t.translateKey(k, lang)
				if err != nil {
					fmt.Printf("Warning: failed to translate %s to %s: %v\n", k, lang, err)
					trans = k // 回退到 key 本身
				} else {
					fmt.Printf("Translated %q to %s: %s\n", k, lang, trans)
				}
				translations[lang] = trans

				// API 限流保护：避免请求过于密集
				time.Sleep(100 * time.Millisecond)
			}

			mu.Lock()
			results[k] = translations
			mu.Unlock()
		}(key)
	}

	wg.Wait()
	return results, nil
}

// translateKey 单个 key 的翻译
func (t *Translator) translateKey(key, targetLang string) (string, error) {
	systemPrompt := `You are an i18n translation assistant.
- Output ONLY the translated string, no explanation, no markdown.
- If key contains ".", it indicates a nested path, translate only the last part.
Example: key "user.avatar" -> translate "avatar"
Rules:
1. Output must be a string.
2. Only return the target language content.`

	payload := map[string]interface{}{
		"model": "deepseek-chat",
		"messages": []map[string]string{
			{"role": "system", "content": systemPrompt},
			{"role": "user", "content": fmt.Sprintf(`Translate key: "%s", target language: %s`, key, targetLang)},
		},
		"temperature": 0,
	}

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	jsonPayload, err := json.Marshal(payload)
	if err != nil {
		return "", err
	}

	req, err := http.NewRequestWithContext(ctx, "POST", t.apiURL+"/chat/completions", bytes.NewBuffer(jsonPayload))
	if err != nil {
		return "", err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+t.apiKey)

	resp, err := t.client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return "", fmt.Errorf("API error: status %d", resp.StatusCode)
	}

	var result struct {
		Choices []struct {
			Message struct {
				Content string `json:"content"`
			} `json:"message"`
		} `json:"choices"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", err
	}

	if len(result.Choices) == 0 {
		return "", fmt.Errorf("empty response")
	}

	return strings.TrimSpace(result.Choices[0].Message.Content), nil
}
