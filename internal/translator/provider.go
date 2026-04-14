package translator

// Provider defines the interface for translation services
type Provider interface {
	// TranslateBatch translates multiple keys to target language in one API call
	TranslateBatch(keys []string, targetLang string) (map[string]string, error)
	// Name returns the provider name
	Name() string
}
