package writer

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"time"

	"github.com/flylea/auto-i18n/internal/config"
	"github.com/flylea/auto-i18n/internal/types"
)

// MergeResults 合并已有翻译和新翻译
func MergeResults(cfg *config.Config, newResults types.TranslateResult) types.TranslateResult {
	for _, lang := range cfg.Languages {
		existingFile := filepath.Join(cfg.OutputDir, lang+"."+cfg.Output)
		if data, err := os.ReadFile(existingFile); err == nil {
			existing := parseExistingFile(string(data))
			for key, translations := range newResults {
				if _, exists := translations[lang]; !exists || translations[lang] == "" {
					if trans, hasTrans := existing[key]; hasTrans {
						newResults[key][lang] = trans
						fmt.Printf("  [merged] %s: %q -> %q (kept existing)\n", lang, key, trans)
					}
				}
			}
		}
	}
	return newResults
}

// parseExistingFile 解析现有语言文件，正确处理嵌套对象结构
func parseExistingFile(content string) map[string]string {
	result := make(map[string]string)

	// 移除 export default、注释
	re := regexp.MustCompile(`//[^\n]*`)
	content = re.ReplaceAllString(content, "")
	re = regexp.MustCompile(`/\*[\s\S]*?\*/`)
	content = re.ReplaceAllString(content, "")
	content = strings.Replace(content, "export default", "", -1)
	content = strings.Replace(content, "export", "", -1)
	content = strings.TrimSpace(content)
	content = strings.Trim(content, "{}")

	if strings.HasPrefix(strings.TrimSpace(content), "{") || strings.Contains(content, ":") {
		// 尝试 JSON 解析
		var data interface{}
		if err := json.Unmarshal([]byte(content), &data); err == nil {
			flattenJSON(data, "", result)
			return result
		}

		// TS/JS 对象解析：从内向外递归解析
		result = parseObject(content, "")
	}

	return result
}

// parseObject 递归解析对象字面量，返回 key.path -> value 的映射
func parseObject(content string, prefix string) map[string]string {
	result := make(map[string]string)
	content = strings.TrimSpace(content)

	// 移除首尾大括号
	content = strings.Trim(content, "{}")

	i := 0
	for i < len(content) {
		// 跳过空白
		for i < len(content) && (content[i] == ' ' || content[i] == '\n' || content[i] == '\t' || content[i] == ',') {
			i++
		}
		if i >= len(content) {
			break
		}

		// 解析 key（可能是 'key' 或 "key" 或 key）
		var key string
		if content[i] == '\'' || content[i] == '"' {
			quote := content[i]
			i++
			start := i
			for i < len(content) && content[i] != quote {
				i++
			}
			key = content[start:i]
			i++ // 跳过结束引号
		} else {
			// 普通标识符 key (可能包含 . 如 user.name)
			start := i
			for i < len(content) && content[i] != ':' && content[i] != ' ' && content[i] != '\n' && content[i] != '\t' {
				i++
			}
			key = content[start:i]
		}

		// 跳过空白和冒号
		for i < len(content) && (content[i] == ' ' || content[i] == '\n' || content[i] == '\t' || content[i] == ':') {
			i++
		}

		if i >= len(content) {
			break
		}

		// 构建完整 key 路径
		fullKey := key
		if prefix != "" {
			fullKey = prefix + "." + key
		}

		// 解析 value
		if content[i] == '{' {
			// 嵌套对象：递归解析
			depth := 1
			start := i + 1
			for i+1 < len(content) && depth > 0 {
				i++
				if content[i] == '{' {
					depth++
				} else if content[i] == '}' {
					depth--
				}
			}
			nestedContent := content[start:i]
			nested := parseObject(nestedContent, fullKey)
			for k, v := range nested {
				result[k] = v
			}
			i++ // 跳过结束 }
		} else if content[i] == '\'' || content[i] == '"' {
			// 字符串值
			quote := content[i]
			i++
			start := i
			for i < len(content) && content[i] != quote {
				if content[i] == '\\' && i+1 < len(content) {
					i++ // 跳过转义符
				}
				i++
			}
			value := content[start:i]
			result[fullKey] = value
			i++ // 跳过结束引号
		} else {
			// 其他值（数字、布尔等），读取到逗号或结束
			start := i
			for i < len(content) && content[i] != ',' && content[i] != '\n' {
				i++
			}
			value := strings.TrimSpace(content[start:i])
			value = strings.Trim(value, ",")
			if value != "" && value != "{" && value != "}" {
				result[fullKey] = value
			}
		}
	}

	return result
}

// flattenJSON 将嵌套 JSON 展平为 key.path 格式
func flattenJSON(obj interface{}, prefix string, result map[string]string) {
	if m, ok := obj.(map[string]interface{}); ok {
		for k, v := range m {
			newKey := k
			if prefix != "" {
				newKey = prefix + "." + k
			}
			flattenJSON(v, newKey, result)
		}
	} else if s, ok := obj.(string); ok {
		result[prefix] = s
	}
}

// Preview 预览将要生成的翻译内容（Dry run 模式）
func Preview(cfg *config.Config, results types.TranslateResult) {
	for _, lang := range cfg.Languages {
		fmt.Printf("\n=== %s (dry run) ===\n", lang)
		fmt.Println("export default {")
		flat := flattenResults(results, lang)
		nested := flatToNested(flat)
		writeObjectPreview(nested, 1)
		fmt.Println("};")
	}
}

func flattenResults(results types.TranslateResult, lang string) map[string]string {
	output := make(map[string]string)
	for key, translations := range results {
		if trans, ok := translations[lang]; ok {
			output[key] = trans
		}
	}
	return output
}

func writeObjectPreview(obj map[string]interface{}, level int) {
	indent := strings.Repeat("  ", level)

	keys := make([]string, 0, len(obj))
	for k := range obj {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	for i, k := range keys {
		v := obj[k]
		keyStr := k

		if m, ok := v.(map[string]interface{}); ok {
			fmt.Printf("%s%s: {\n", indent, keyStr)
			writeObjectPreview(m, level+1)
			closingIndent := strings.Repeat("  ", level)
			if i < len(keys)-1 {
				fmt.Printf("%s},\n", closingIndent)
			} else {
				fmt.Printf("%s}\n", closingIndent)
			}
		} else {
			val := fmt.Sprintf("%v", v)
			if i < len(keys)-1 {
				fmt.Printf("%s%s: '%s',\n", indent, keyStr, val)
			} else {
				fmt.Printf("%s%s: '%s'\n", indent, keyStr, val)
			}
		}
	}
}

// Check 校验语言包，检测多余或缺失的 key
func Check(cfg *config.Config, lang string, sourceKeys []string) {
	fmt.Printf("\n=== Checking %s ===\n", lang)

	filename := filepath.Join(cfg.OutputDir, lang+"."+cfg.Output)
	data, err := os.ReadFile(filename)
	if err != nil {
		fmt.Printf("  ⚠️  File not found: %s\n", filename)
		fmt.Printf("  Missing %d keys from source\n", len(sourceKeys))
		return
	}

	existing := parseExistingFile(string(data))
	fmt.Printf("  Parsed %d keys from file\n", len(existing))

	sourceSet := make(map[string]bool)
	for _, k := range sourceKeys {
		sourceSet[k] = true
	}

	var missing []string
	for _, key := range sourceKeys {
		if _, exists := existing[key]; !exists || existing[key] == "" {
			missing = append(missing, key)
		}
	}

	var orphaned []string
	for key := range existing {
		if !sourceSet[key] {
			orphaned = append(orphaned, key)
		}
	}

	if len(missing) > 0 {
		fmt.Printf("  ❌ Missing keys (%d):\n", len(missing))
		for _, k := range missing {
			fmt.Printf("     - %s\n", k)
		}
	} else {
		fmt.Printf("  ✅ No missing keys\n")
	}

	if len(orphaned) > 0 {
		fmt.Printf("  ⚠️  Orphaned keys (unused in source, consider removing):\n")
		for _, k := range orphaned {
			fmt.Printf("     - %s\n", k)
		}
	} else {
		fmt.Printf("  ✅ No orphaned keys\n")
	}
}

// Write 生成语言文件
func Write(cfg *config.Config, results types.TranslateResult, keyFileMap types.KeyFileMap) error {
	outputDir := cfg.OutputDir
	if !filepath.IsAbs(outputDir) {
		cwd, _ := os.Getwd()
		outputDir = filepath.Join(cwd, outputDir)
	}

	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return fmt.Errorf("failed to create output dir: %w", err)
	}

	isJSON := cfg.Output == "json"

	for _, lang := range cfg.Languages {
		output := flattenResults(results, lang)

		var content string
		if isJSON {
			content = toJSON(output)
		} else {
			content = toTypeScript(output, lang, cfg, keyFileMap)
		}

		filename := filepath.Join(outputDir, lang+"."+cfg.Output)
		if err := os.WriteFile(filename, []byte(content), 0644); err != nil {
			return fmt.Errorf("failed to write %s: %w", filename, err)
		}
		fmt.Printf("Generated %s\n", filepath.Base(filename))
	}

	return nil
}

func toJSON(flat map[string]string) string {
	nested := flatToNested(flat)
	data, _ := json.MarshalIndent(nested, "", "  ")
	return string(data)
}

func flatToNested(flat map[string]string) map[string]interface{} {
	nested := make(map[string]interface{})

	keys := make([]string, 0, len(flat))
	for k := range flat {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	for _, fullKey := range keys {
		value := flat[fullKey]
		segments := strings.Split(fullKey, ".")
		cur := nested

		for i, seg := range segments {
			if i == len(segments)-1 {
				cur[seg] = value
			} else {
				if _, exists := cur[seg]; !exists {
					cur[seg] = make(map[string]interface{})
				}
				cur = cur[seg].(map[string]interface{})
			}
		}
	}

	return nested
}

func toTypeScript(flat map[string]string, lang string, cfg *config.Config, keyFileMap types.KeyFileMap) string {
	var b strings.Builder

	if cfg.WithFileComment {
		b.WriteString("/**\n")
		b.WriteString(fmt.Sprintf(" * I18n locale file: %s\n", lang))
		b.WriteString(" * Auto-generated by auto-i18n\n")
		b.WriteString(fmt.Sprintf(" * Generated at: %s\n", time.Now().Format(time.RFC1123)))
		b.WriteString(" */\n")
	}

	b.WriteString("export default {\n")

	nested := flatToNested(flat)
	topLevelFileMap := buildTopLevelFileMap(keyFileMap)
	writeObject(&b, nested, 1, cfg, topLevelFileMap)

	b.WriteString("};\n")
	return b.String()
}

func buildTopLevelFileMap(keyFileMap types.KeyFileMap) map[string]map[string]bool {
	result := make(map[string]map[string]bool)
	for fullKey, files := range keyFileMap {
		topKey := strings.Split(fullKey, ".")[0]
		if result[topKey] == nil {
			result[topKey] = make(map[string]bool)
		}
		for f := range files {
			result[topKey][f] = true
		}
	}
	return result
}

func writeObject(b *strings.Builder, obj map[string]interface{}, level int, cfg *config.Config, topLevelFileMap map[string]map[string]bool) {
	indent := strings.Repeat("  ", level)
	nextIndent := strings.Repeat("  ", level+1)

	keys := make([]string, 0, len(obj))
	for k := range obj {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	for i, k := range keys {
		v := obj[k]

		// 顶层 key 写入来源文件注释
		if level == 1 && cfg.WithKeyFileComment {
			if files, ok := topLevelFileMap[k]; ok && len(files) > 0 {
				b.WriteString("\n")
				for f := range files {
					b.WriteString(fmt.Sprintf("%s// %s\n", nextIndent, f))
				}
			}
		}

		keyStr := k
		// 中文 key 需要用引号包裹
		if cfg.SupportChineseKey && containsChinese(k) {
			keyStr = "'" + k + "'"
		}

		if m, ok := v.(map[string]interface{}); ok {
			b.WriteString(fmt.Sprintf("%s%s: {\n", indent, keyStr))
			writeObject(b, m, level+1, cfg, topLevelFileMap)
			closingIndent := strings.Repeat("  ", level)
			if i < len(keys)-1 {
				b.WriteString(fmt.Sprintf("%s},\n", closingIndent))
			} else {
				b.WriteString(fmt.Sprintf("%s}\n", closingIndent))
			}
		} else {
			val := fmt.Sprintf("%v", v)
			val = strings.ReplaceAll(val, "'", "\\'")
			if i < len(keys)-1 {
				b.WriteString(fmt.Sprintf("%s%s: '%s',\n", indent, keyStr, val))
			} else {
				b.WriteString(fmt.Sprintf("%s%s: '%s'\n", indent, keyStr, val))
			}
		}
	}
}

func containsChinese(s string) bool {
	for _, r := range s {
		if r >= 0x4e00 && r <= 0x9fa5 {
			return true
		}
	}
	return false
}
