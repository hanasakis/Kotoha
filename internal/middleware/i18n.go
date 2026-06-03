package middleware

import (
	"strings"

	"github.com/gin-gonic/gin"
)

func I18n(defaultLocale string, supportedLocales []string) gin.HandlerFunc {
	supported := make(map[string]bool)
	for _, l := range supportedLocales {
		supported[strings.TrimSpace(l)] = true
	}

	return func(c *gin.Context) {
		locale := defaultLocale

		if lang := c.GetHeader("Accept-Language"); lang != "" {
			parts := strings.SplitN(lang, ",", 2)
			pref := strings.SplitN(parts[0], ";", 2)[0]
			pref = strings.TrimSpace(pref)
			short := strings.SplitN(pref, "-", 2)[0]
			if supported[pref] {
				locale = pref
			} else if supported[short] {
				locale = short
			}
		}

		c.Set("locale", locale)
		c.Next()
	}
}
