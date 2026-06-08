package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

func setupGin() (*gin.Context, *httptest.ResponseRecorder) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest("GET", "/test", nil)
	return c, w
}

func TestAuthRequired(t *testing.T) {
	secret := "test-secret"
	handler := AuthRequired(secret)

	t.Run("no_auth_header", func(t *testing.T) {
		c, w := setupGin()
		handler(c)
		if w.Code != http.StatusUnauthorized {
			t.Errorf("got status %d, want 401", w.Code)
		}
	})

	t.Run("malformed_header", func(t *testing.T) {
		c, w := setupGin()
		c.Request.Header.Set("Authorization", "InvalidFormat")
		handler(c)
		if w.Code != http.StatusUnauthorized {
			t.Errorf("got status %d, want 401", w.Code)
		}
	})

	t.Run("wrong_scheme", func(t *testing.T) {
		c, w := setupGin()
		c.Request.Header.Set("Authorization", "Basic xyz123")
		handler(c)
		if w.Code != http.StatusUnauthorized {
			t.Errorf("got status %d, want 401", w.Code)
		}
	})

	t.Run("invalid_token", func(t *testing.T) {
		c, w := setupGin()
		c.Request.Header.Set("Authorization", "Bearer invalid.token.here")
		handler(c)
		if w.Code != http.StatusUnauthorized {
			t.Errorf("got status %d, want 401", w.Code)
		}
	})

	t.Run("expired_token", func(t *testing.T) {
		claims := jwt.MapClaims{
			"sub":  "123",
			"role": "user",
			"exp":  time.Now().Add(-1 * time.Hour).Unix(),
		}
		token, _ := jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString([]byte(secret))

		c, w := setupGin()
		c.Request.Header.Set("Authorization", "Bearer "+token)
		handler(c)
		if w.Code != http.StatusUnauthorized {
			t.Errorf("got status %d, want 401 for expired token", w.Code)
		}
	})

	t.Run("valid_token", func(t *testing.T) {
		claims := jwt.MapClaims{
			"sub":  "42",
			"role": "user",
			"exp":  time.Now().Add(1 * time.Hour).Unix(),
		}
		token, _ := jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString([]byte(secret))

		c, w := setupGin()
		c.Request.Header.Set("Authorization", "Bearer "+token)
		handler(c)

		if w.Code != http.StatusOK {
			t.Errorf("got status %d, want 200", w.Code)
		}
		userID, _ := c.Get("user_id")
		if userID != "42" {
			t.Errorf("got user_id %v, want 42", userID)
		}
		role, _ := c.Get("role")
		if role != "user" {
			t.Errorf("got role %v, want user", role)
		}
	})

	t.Run("wrong_secret", func(t *testing.T) {
		claims := jwt.MapClaims{
			"sub":  "42",
			"role": "user",
			"exp":  time.Now().Add(1 * time.Hour).Unix(),
		}
		token, _ := jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString([]byte("different-secret"))

		c, w := setupGin()
		c.Request.Header.Set("Authorization", "Bearer "+token)
		handler(c)
		if w.Code != http.StatusUnauthorized {
			t.Errorf("got status %d, want 401 for wrong secret", w.Code)
		}
	})
}

func TestRoleRequired(t *testing.T) {
	t.Run("missing_role", func(t *testing.T) {
		c, w := setupGin()
		handler := RoleRequired("admin")
		handler(c)
		if w.Code != http.StatusForbidden {
			t.Errorf("got status %d, want 403", w.Code)
		}
	})

	t.Run("wrong_role", func(t *testing.T) {
		c, w := setupGin()
		c.Set("role", "user")
		handler := RoleRequired("admin")
		handler(c)
		if w.Code != http.StatusForbidden {
			t.Errorf("got status %d, want 403", w.Code)
		}
	})

	t.Run("correct_role", func(t *testing.T) {
		c, w := setupGin()
		c.Set("role", "admin")
		handler := RoleRequired("admin")
		handler(c)
		if w.Code != http.StatusOK {
			t.Errorf("got status %d, want 200", w.Code)
		}
	})

	t.Run("multiple_roles", func(t *testing.T) {
		c, w := setupGin()
		c.Set("role", "editor")
		handler := RoleRequired("admin", "editor", "moderator")
		handler(c)
		if w.Code != http.StatusOK {
			t.Errorf("got status %d, want 200", w.Code)
		}
	})

	t.Run("nil_role_value", func(t *testing.T) {
		c, w := setupGin()
		c.Set("role", nil)
		handler := RoleRequired("admin")
		handler(c)
		if w.Code != http.StatusForbidden {
			t.Errorf("got status %d, want 403 for nil role", w.Code)
		}
	})
}

func TestCORS(t *testing.T) {
	origins := []string{"http://localhost:3000", "https://app.example.com"}

	t.Run("allowed_origin", func(t *testing.T) {
		c, w := setupGin()
		c.Request.Header.Set("Origin", "http://localhost:3000")
		CORS(origins)(c)
		allowOrigin := w.Header().Get("Access-Control-Allow-Origin")
		if allowOrigin != "http://localhost:3000" {
			t.Errorf("got Allow-Origin %q, want http://localhost:3000", allowOrigin)
		}
	})

	t.Run("disallowed_origin", func(t *testing.T) {
		c, w := setupGin()
		c.Request.Header.Set("Origin", "https://evil.com")
		CORS(origins)(c)
		allowOrigin := w.Header().Get("Access-Control-Allow-Origin")
		if allowOrigin != "" {
			t.Errorf("expected empty Allow-Origin for disallowed origin, got %q", allowOrigin)
		}
	})

	t.Run("wildcard_origin", func(t *testing.T) {
		c, w := setupGin()
		c.Request.Header.Set("Origin", "https://any.app.com")
		CORS([]string{"*"})(c)
		allowOrigin := w.Header().Get("Access-Control-Allow-Origin")
		if allowOrigin != "https://any.app.com" {
			t.Errorf("got Allow-Origin %q, want https://any.app.com", allowOrigin)
		}
	})

	t.Run("cors_headers_set", func(t *testing.T) {
		c, w := setupGin()
		c.Request.Header.Set("Origin", "http://localhost:3000")
		CORS(origins)(c)
		if w.Header().Get("Access-Control-Allow-Methods") == "" {
			t.Error("expected Allow-Methods header")
		}
		if w.Header().Get("Access-Control-Allow-Headers") == "" {
			t.Error("expected Allow-Headers header")
		}
		if w.Header().Get("Access-Control-Allow-Credentials") != "true" {
			t.Error("expected Allow-Credentials: true")
		}
	})

	t.Run("preflight_options", func(t *testing.T) {
		c, w := setupGin()
		c.Request = httptest.NewRequest("OPTIONS", "/test", nil)
		c.Request.Header.Set("Origin", "http://localhost:3000")
		CORS(origins)(c)
		if w.Code != http.StatusNoContent {
			t.Errorf("got status %d, want 204 for preflight", w.Code)
		}
	})

	t.Run("no_origin_header", func(t *testing.T) {
		c, w := setupGin()
		CORS(origins)(c)
		// Should not set Allow-Origin when no Origin header
		if w.Header().Get("Access-Control-Allow-Origin") != "" {
			t.Error("expected empty Allow-Origin when no Origin header")
		}
		// But should still set other CORS headers
		if w.Header().Get("Access-Control-Allow-Methods") == "" {
			t.Error("expected Allow-Methods header even without Origin")
		}
	})
}

func TestRequestID(t *testing.T) {
	t.Run("generate_new_id", func(t *testing.T) {
		c, w := setupGin()
		RequestID()(c)

		rid, exists := c.Get("request_id")
		if !exists {
			t.Fatal("expected request_id in context")
		}
		if rid == "" {
			t.Error("expected non-empty request_id")
		}
		respID := w.Header().Get("X-Request-ID")
		if respID == "" {
			t.Error("expected X-Request-ID in response header")
		}
	})

	t.Run("propagate_existing_id", func(t *testing.T) {
		c, w := setupGin()
		c.Request.Header.Set("X-Request-ID", "my-custom-id-123")
		RequestID()(c)

		rid, _ := c.Get("request_id")
		if rid != "my-custom-id-123" {
			t.Errorf("got request_id %v, want my-custom-id-123", rid)
		}
		respID := w.Header().Get("X-Request-ID")
		if respID != "my-custom-id-123" {
			t.Errorf("got X-Request-ID %q, want my-custom-id-123", respID)
		}
	})
}

func TestI18n(t *testing.T) {
	t.Run("default_locale", func(t *testing.T) {
		c, _ := setupGin()
		I18n("zh", []string{"zh", "en"})(c)
		locale, _ := c.Get("locale")
		if locale != "zh" {
			t.Errorf("got locale %v, want zh", locale)
		}
	})

	t.Run("exact_match_header", func(t *testing.T) {
		c, _ := setupGin()
		c.Request.Header.Set("Accept-Language", "en")
		I18n("zh", []string{"zh", "en"})(c)
		locale, _ := c.Get("locale")
		if locale != "en" {
			t.Errorf("got locale %v, want en", locale)
		}
	})

	t.Run("short_code_match", func(t *testing.T) {
		c, _ := setupGin()
		c.Request.Header.Set("Accept-Language", "en-US")
		I18n("zh", []string{"zh", "en"})(c)
		locale, _ := c.Get("locale")
		if locale != "en" {
			t.Errorf("got locale %v, want en (short match)", locale)
		}
	})

	t.Run("quality_value_stripped", func(t *testing.T) {
		c, _ := setupGin()
		c.Request.Header.Set("Accept-Language", "en;q=0.9")
		I18n("zh", []string{"zh", "en"})(c)
		locale, _ := c.Get("locale")
		if locale != "en" {
			t.Errorf("got locale %v, want en (q= stripped)", locale)
		}
	})

	t.Run("first_preferred_wins", func(t *testing.T) {
		c, _ := setupGin()
		c.Request.Header.Set("Accept-Language", "fr, en;q=0.9")
		I18n("zh", []string{"zh", "en", "fr"})(c)
		locale, _ := c.Get("locale")
		if locale != "fr" {
			t.Errorf("got locale %v, want fr (first preferred)", locale)
		}
	})

	t.Run("unsupported_locale_falls_back", func(t *testing.T) {
		c, _ := setupGin()
		c.Request.Header.Set("Accept-Language", "ja")
		I18n("zh", []string{"zh", "en"})(c)
		locale, _ := c.Get("locale")
		if locale != "zh" {
			t.Errorf("got locale %v, want zh (fallback)", locale)
		}
	})

	t.Run("empty_header", func(t *testing.T) {
		c, _ := setupGin()
		c.Request.Header.Set("Accept-Language", "")
		I18n("zh", []string{"zh", "en"})(c)
		locale, _ := c.Get("locale")
		if locale != "zh" {
			t.Errorf("got locale %v, want zh (default for empty)", locale)
		}
	})

	t.Run("whitespace_trimmed", func(t *testing.T) {
		c, _ := setupGin()
		c.Request.Header.Set("Accept-Language", " en ")
		I18n("zh", []string{"zh", "en"})(c)
		locale, _ := c.Get("locale")
		if locale != "en" {
			t.Errorf("got locale %v, want en (whitespace trimmed)", locale)
		}
	})
}
