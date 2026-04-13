# auto-i18n

A CLI tool that automatically scans your codebase for i18n translation keys and generates locale files using AI-powered translation.

## Features

- **AI Translation** - Leverages Deepseek API for accurate, contextual translations
- **Multi-framework Support** - Works with Vue (SFC), JavaScript, and TypeScript files
- **Configurable** - Customize entry directories, file extensions, output format, and more
- **Chinese Key Support** - Handles Chinese translation keys natively
- **Clean Output** - Optional file header and key-source file comments

## Requirements

- Node.js >= 22.0.0
- Deepseek API key

## Installation

```bash
# Install as dev dependency
pnpm add i18n-a11y -D

# Or use via npx
npx i18n-a11y scan
```

## Quick Start

1. Create a `.env.local` file in your project root:

```bash
DEEPSEEK_API_KEY=your-deepseek-api-key
DEEPSEEK_API_URL=https://api.deepseek.com
```

2. Configure `i18n.config.ts` in your project root:

```typescript
export default {
  // Directories to scan for translation keys
  entryDirs: ["src/views", "src/components"],

  // File extensions to scan
  extensions: ["vue"],

  // Output format: "ts" | "js" | "json"
  output: "ts",

  // Output directory for locale files
  outputDir: "src/i18n/locale",

  // Include file header comment
  withFileComment: false,

  // Include key → file path comments
  withKeyFileComment: false,

  // i18n function names to match
  i18nFns: ["t", "$t"],

  // Support Chinese keys
  supportChineseKey: true,

  // Target languages
  languages: ["zh-CN", "en-US"],
};
```

3. Run the scanner:

```bash
# Using npm scripts
pnpm i18n:local    # Run with TypeScript source (development)
pnpm i18n:build    # Run with compiled JavaScript (production)

# Or install globally and use directly
npm install -g i18n-a11y
auto-i18n scan
```

## CLI Commands

```bash
# Scan and generate i18n files
auto-i18n scan

# Interactive config generation (creates i18n.config.ts)
auto-i18n init
```

## Output

After running, locale files will be generated in your specified output directory:

```
src/i18n/locale/
├── zh-CN.ts
└── en-US.ts
```

Example output file:

```typescript
// zh-CN.ts
export default {
  user: {
    name: "用户名",
    email: "邮箱",
    avatar: "头像",
  },
  common: {
    save: "保存",
    cancel: "取消",
  },
};
```

## NPM Scripts

| Script | Description |
|--------|-------------|
| `pnpm build` | Compile TypeScript to JavaScript |
| `pnpm test` | Run tests |
| `pnpm test:watch` | Run tests in watch mode |
| `pnpm i18n:local` | Run with TypeScript source |
| `pnpm i18n:build` | Run with compiled JS |
| `pnpm release:patch` | Bump patch version |
| `pnpm release:minor` | Bump minor version |
| `pnpm release:major` | Bump major version |
| `pnpm publish:npm` | Publish to npm |

## Testing

### Integration Test (Playground)

The project includes a `playground/` directory with a sample Vue project for integration testing:

```bash
# Copy and configure environment
cd playground
cp .env.example .env.local
# Edit .env.local with your Deepseek API key

# Run the integration test
node test-cli.mjs
```

### Unit Tests

```bash
pnpm test        # Run tests once
pnpm test:watch  # Run tests in watch mode
```

## License

MIT
