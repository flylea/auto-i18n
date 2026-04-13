package types

// TranslateResult: key -> lang -> translation
type TranslateResult map[string]map[string]string

// KeyFileMap: key -> set of files containing that key
type KeyFileMap map[string]map[string]bool
