# auto-i18n

AI-powered i18n CLI tool that automatically scans your codebase for translation keys and generates locale files.

## Features

- **AI Translation** - Leverages Deepseek API for accurate translations
- **Git-Aware Scanning** - Only scans files changed since last commit (uses `git diff`)
- **Multi-framework** - Supports Vue, React (JSX/TSX), and plain JS/TS
- **Incremental Translation** - Only translates new keys, preserves existing translations
- **Merge Mode** - Merges with existing language files without overwriting
- **Dry Run** - Preview changes before writing files
- **Key Validation** - Check for missing or orphaned keys in language files

## Requirements

- Go 1.21+
- Deepseek API key

## Installation

```bash
# Download pre-built binary
# Or build from source
go build -o auto-i18n ./cmd/auto-i18n
```

## Configuration

Create `i18n.config.json` in your project root:

```json
{
  "entryDirs": ["src/views", "src/components", "src/pages"],
  "extensions": ["vue", "jsx", "tsx"],
  "output": "ts",
  "outputDir": "src/i18n/locale",
  "withFileComment": false,
  "withKeyFileComment": false,
  "i18nFns": ["t", "$t"],
  "supportChineseKey": true,
  "languages": ["zh-CN", "en-US"]
}
```

## Environment Variables

```bash
DEEPSEEK_API_KEY=your-api-key
```

## Commands

```bash
# Scan and translate changed files (git-aware)
# Only scans files modified/added since last commit
auto-i18n

# Scan all files (ignore git changes)
# Useful for initial setup or full regeneration
auto-i18n --all

# Preview changes without writing files
auto-i18n --dry-run

# Incremental translation (only new keys, keep existing)
auto-i18n incremental

# Validate language files
auto-i18n check
```

## Git-Aware Scanning

By default, `auto-i18n` only scans files that have changed since the last commit:

```bash
# After making changes to code
git commit -m "update user profile"
auto-i18n

# CLI automatically detects:
# - Modified .vue/.tsx files
# - New .vue/.tsx files
# Only translates keys from changed files
```

Use `--all` flag to force full scan:

```bash
auto-i18n --all
```

## Supported Patterns

```javascript
// Vue
t('key')
$t('key')
i18n.t('key')

// React
useTranslation()['t']('key')
t('key')
```

## Output

Generated `src/i18n/locale/zh-CN.ts`:

```typescript
export default {
  user: {
    name: '用户名',
    email: '邮箱'
  },
  common: {
    save: '保存',
    cancel: '取消'
  }
};
```

## License

MIT
