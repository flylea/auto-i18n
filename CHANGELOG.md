# Changelog

## [2.0.0] - 2026-04-13

### Refactored
- Complete rewrite in Go for better performance
- Binary single-file distribution

### Added
- **Incremental translation** - Only translates new keys, preserves existing translations
- **Merge mode** - Merges with existing language files without overwriting
- **Dry run** - Preview changes without writing files
- **Key validation** - `auto-i18n check` detects missing/orphaned keys
- **React JSX/TSX support** - Extracts keys from React components
- **Concurrent translation** - Translates multiple keys in parallel

### Removed
- Node.js/TypeScript implementation deprecated

## [1.0.x] - Previous versions

See git history for Node.js version changes.
