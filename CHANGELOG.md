# Changelog

All notable changes will be documented in this file.

## [1.0.8] - 2026-04-13

### Changed
- Refactored to use Node.js 22+ native TypeScript execution (`--experimental-strip-types`)
- Removed `tsx` dependency
- Moved `dotenv` to devDependencies
- Updated `i18n:local` script to use native TS execution
- Changed `prepublishOnly` to `prepare` lifecycle script
- Removed unused `p-limit` dependency

### Added
- `engines` field to specify Node.js >=22 requirement
- `exports` field for proper module exports
- `vitest` test framework with translation test cases
- `test` and `test:watch` npm scripts

## [1.0.7] - Previous Release

### Features
- AI-powered translation via Deepseek API
- Support for Vue, JS, and TS files
- Configurable i18n function names (`t`, `$t`, `i18n.t`)
- Optional file-level and key-level comments
- Chinese key support
- Interactive config generator
