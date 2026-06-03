package i18n

import (
	"encoding/json"
	"os"
	"sync"
)

type Translator struct {
	mu            sync.RWMutex
	translations  map[string]map[string]string
	defaultLocale string
}

func New(defaultLocale string) *Translator {
	return &Translator{
		translations:  make(map[string]map[string]string),
		defaultLocale: defaultLocale,
	}
}

func (t *Translator) Load(locale, path string) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	var m map[string]string
	if err := json.Unmarshal(data, &m); err != nil {
		return err
	}
	t.mu.Lock()
	t.translations[locale] = m
	t.mu.Unlock()
	return nil
}

func (t *Translator) T(locale, key string) string {
	t.mu.RLock()
	defer t.mu.RUnlock()

	if m, ok := t.translations[locale]; ok {
		if v, ok := m[key]; ok {
			return v
		}
	}
	if m, ok := t.translations[t.defaultLocale]; ok {
		if v, ok := m[key]; ok {
			return v
		}
	}
	return key
}
