# auto-i18n

## description

node cli tools for auto generate i18n files

## install

```bash
pnpm i i18n-a11y -D
```

## usage

```json
 "scripts": {
    "i18n": "i18n-a11y scan"
 }
```

## configuration

```json
   {
  entryDirs: ["src/views", "src/components"], // scan entry directory
  extensions: ["vue"], // scan file extensions
  output: "ts", // output file type
  outputDir: "src/i18n/locales", // output file directory
  withFileComment: false, // add file comment on file top
  withKeyFileComment: false, // add key file comment on key top
  i18nFns: ["t", "$t"], // i18n function name
  supportChineseKey: true,  // support chinese key
  languages: ["zh-CN", "en-US"],  // languages
};
```
## Environment Variables

Create a `.env.local` file in the project root and add:
``` bash
   DEEPSEEK_API_KEY=your-deepseek-key-here
   DEEPSEEK_API_URL=https://api.deepseek.com
```

## Output

After running the CLI, files will be generated in the outputDir, for example `src/i18n/locales`

```bash
src/i18n/locale/
├─ zh-CN.ts
├─ en-US.ts

```

## Features

Automatically scans project files to extract translation keys

Supports Vue, JS, and TS files

Generates multilingual locale files

Optional file-level and key-level comments

Fully configurable

Supports Chinese keys

Integrates Deepseek AI for automatic translations