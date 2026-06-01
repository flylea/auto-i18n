package translator

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"
)

// DeepseekProvider implements Provider for Deepseek API
type DeepseekProvider struct {
	apiKey    string
	apiURL    string
	model     string
	client    *http.Client
	batchSize int
}

func NewDeepseekProvider(apiKey, model string, batchSize int) *DeepseekProvider {
	if model == "" {
		model = "deepseek-chat"
	}
	if batchSize <= 0 {
		batchSize = 10
	}

	return &DeepseekProvider{
		apiKey:    apiKey,
		apiURL:    "https://api.deepseek.com",
		model:     model,
		client:    &http.Client{Timeout: 120 * time.Second},
		batchSize: batchSize,
	}
}

func (d *DeepseekProvider) Name() string {
	return "deepseek"
}

// TranslateBatch translates multiple keys in a single API call
func (d *DeepseekProvider) TranslateBatch(keys []string, targetLang string) (map[string]string, error) {
	if len(keys) == 0 {
		return make(map[string]string), nil
	}

	// Build prompt with all keys
	keysList := strings.Join(keys, "\n")
	prompt := fmt.Sprintf(`Translate the following i18n keys to %s.
Each key is on its own line. Output must be in JSON format: {"key": "translation"}.
Keep placeholders like {name}, {count}, {{count}} unchanged in translations.

Keys:
%s

Output JSON only:`, targetLang, keysList)

	systemPrompt := `You are an i18n translation assistant.
- Output ONLY valid JSON: {"key1": "translation1", "key2": "translation2"}
- Preserve all placeholders exactly as they appear
- If a key contains ".", it indicates a nested path - translate only the meaningful part
- Do not add explanations or markdown`

	payload := map[string]interface{}{
		"model": d.model,
		"messages": []map[string]string{
			{"role": "system", "content": systemPrompt},
			{"role": "user", "content": prompt},
		},
		"temperature": 0,
	}

	ctx, cancel := context.WithTimeout(context.Background(), 120*time.Second)
	defer cancel()

	jsonPayload, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(ctx, "POST", d.apiURL+"/chat/completions", bytes.NewBuffer(jsonPayload))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+d.apiKey)

	resp, err := d.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("API error: status %d", resp.StatusCode)
	}

	var result struct {
		Choices []struct {
			Message struct {
				Content string `json:"content"`
			} `json:"message"`
		} `json:"choices"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	if len(result.Choices) == 0 {
		return nil, fmt.Errorf("empty response")
	}

	content := strings.TrimSpace(result.Choices[0].Message.Content)

	// Try to extract JSON from the response (in case there's extra text)
	jsonStart := strings.Index(content, "{")
	jsonEnd := strings.LastIndex(content, "}") + 1
	if jsonStart >= 0 && jsonEnd > jsonStart {
		content = content[jsonStart:jsonEnd]
	}

	var translations map[string]string
	if err := json.Unmarshal([]byte(content), &translations); err != nil {
		return nil, fmt.Errorf("failed to parse JSON response: %w", err)
	}

	return translations, nil
}

// TranslateKeys translates keys in batches, returns combined results
func (d *DeepseekProvider) TranslateKeys(allKeys []string, targetLang string) (map[string]string, error) {
	results := make(map[string]string)

	// Process in batches
	for i := 0; i < len(allKeys); i += d.batchSize {
		end := i + d.batchSize
		if end > len(allKeys) {
			end = len(allKeys)
		}
		batch := allKeys[i:end]

		batchResults, err := d.TranslateBatch(batch, targetLang)
		if err != nil {
			// Log error but continue with other batches
			fmt.Printf("Warning: batch translation failed for %d keys: %v\n", len(batch), err)
			// Fall back to individual translations for failed batch
			for _, key := range batch {
				trans, err := d.translateSingle(key, targetLang)
				if err != nil {
					fmt.Printf("Warning: failed to translate %s: %v\n", key, err)
					trans = key
				}
				results[key] = trans
			}
			continue
		}

		for key, trans := range batchResults {
			results[key] = trans
		}

		// Rate limiting between batches
		if end < len(allKeys) {
			time.Sleep(200 * time.Millisecond)
		}
	}

	return results, nil
}

// translateSingle translates a single key (fallback when batch fails)
func (d *DeepseekProvider) translateSingle(key, targetLang string) (string, error) {
	systemPrompt := `You are an i18n translation assistant.
- Output ONLY the translated string, no explanation, no markdown.
- If key contains ".", it indicates a nested path, translate only the last part.
Rules:
1. Output must be a string.
2. Only return the target language content.`

	payload := map[string]interface{}{
		"model": d.model,
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

	req, err := http.NewRequestWithContext(ctx, "POST", d.apiURL+"/chat/completions", bytes.NewBuffer(jsonPayload))
	if err != nil {
		return "", err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+d.apiKey)

	resp, err := d.client.Do(req)
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

// GetModel returns the model name
func (d *DeepseekProvider) GetModel() string {
	return d.model
}

// GetAPIKey returns the API key (for config)
func (d *DeepseekProvider) GetAPIKey() string {
	return d.apiKey
}

// OpenAIProvider implements Provider for OpenAI-compatible APIs
type OpenAIProvider struct {
	*DeepseekProvider // Reuse Deepseek implementation
}

func NewOpenAIProvider(apiKey, model string, batchSize int) *OpenAIProvider {
	if model == "" {
		model = "gpt-3.5-turbo"
	}
	p := NewDeepseekProvider(apiKey, model, batchSize)
	p.apiURL = "https://api.openai.com"
	return &OpenAIProvider{
		DeepseekProvider: p,
	}
}

func (o *OpenAIProvider) Name() string {
	return "openai"
}

// GetEnvConfig returns API config from environment variables
func GetEnvConfig() (apiKey, model string) {
	apiKey = os.Getenv("DEEPSEEK_API_KEY")
	model = os.Getenv("DEEPSEEK_MODEL")

	// Also check generic OpenAI env
	if apiKey == "" {
		apiKey = os.Getenv("OPENAI_API_KEY")
	}
	if model == "" {
		model = os.Getenv("OPENAI_MODEL")
	}

	return
}
