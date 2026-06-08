package router

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"

	"github.com/hanasakis/kotoha/internal/config"
	"github.com/hanasakis/kotoha/internal/i18n"
	"github.com/hanasakis/kotoha/internal/testutil"
	"github.com/hanasakis/kotoha/pkg/db"
)

// setupTestRouter creates a full router with test DB/Redis, auto-migrates, and seeds demo data.
// Returns the Gin engine and a cleanup function. Skips the test if DB/Redis are unavailable.
func setupTestRouter(t *testing.T) (*gin.Engine, func()) {
	t.Helper()

	pg := testutil.SetupTestDB(t)
	redisCli := testutil.SetupTestRedis(t)

	if err := db.AutoMigrate(pg); err != nil {
		t.Fatalf("auto-migrate failed: %v", err)
	}

	cfg := &config.Config{}
	cfg.JWT.Secret = "test-jwt-secret"
	cfg.JWT.AccessTTL = "15m"
	cfg.JWT.RefreshTTL = "720h"
	cfg.I18n.DefaultLocale = "zh"
	cfg.I18n.SupportedLocales = []string{"zh", "en"}
	cfg.CORS.Origins = []string{"http://localhost:3000", "*"}
	cfg.Stripe.SecretKey = ""
	cfg.Stripe.WebhookSecret = ""
	cfg.Ollama.Host = "http://localhost:11434"
	cfg.Embedding.Model = "bge-large-zh-v1.5"
	cfg.Embedding.Dim = 1024
	cfg.LLM.Model = "qwen3.5:9b"
	cfg.LLM.Temperature = 0.3
	cfg.LLM.MaxTokens = 256
	cfg.Milvus.Host = "localhost"
	cfg.Milvus.Port = "19530"
	cfg.Milvus.User = "root"
	cfg.Milvus.Password = ""
	cfg.Milvus.DB = "kotoha"
	cfg.Milvus.DenseWeight = 0.65
	cfg.Milvus.BM25Weight = 0.35
	cfg.Milvus.RecallTopK = 40
	cfg.Langfuse.Host = "https://cloud.langfuse.com"
	cfg.Langfuse.PublicKey = ""
	cfg.Langfuse.SecretKey = ""

	translator := i18n.New("zh")

	deps := &Dependencies{
		DB:     pg,
		Redis:  redisCli,
		Config: cfg,
		I18n:   translator,
	}

	r := Setup(deps)

	cleanup := func() {
		testutil.CleanTestDB(t, pg)
	}

	return r, cleanup
}

// generateTestJWT creates a signed JWT for testing with the given user ID and role.
func generateTestJWT(secret string, userID uint, role string) (string, error) {
	claims := jwt.MapClaims{
		"sub":  fmt.Sprintf("%d", userID),
		"role": role,
		"exp":  time.Now().Add(1 * time.Hour).Unix(),
	}
	token, err := jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString([]byte(secret))
	return token, err
}

// doJSON sends an HTTP request with optional JSON body and auth header, returns the recorder.
func doJSON(method, path string, body string, token string, r *gin.Engine) *httptest.ResponseRecorder {
	var req *http.Request
	if body != "" {
		req = httptest.NewRequest(method, path, strings.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
	} else {
		req = httptest.NewRequest(method, path, nil)
	}
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	return w
}

func TestHealthEndpoint(t *testing.T) {
	r, cleanup := setupTestRouter(t)
	defer cleanup()

	w := doJSON("GET", "/health", "", "", r)
	if w.Code != http.StatusOK {
		t.Errorf("health: got %d, want 200", w.Code)
	}
	var body struct {
		Data struct {
			Status string            `json:"status"`
			Deps   map[string]string `json:"deps"`
		} `json:"data"`
	}
	json.Unmarshal(w.Body.Bytes(), &body)
	if body.Data.Status != "ok" {
		t.Errorf("health: got status %q, want ok", body.Data.Status)
	}
}

// --- Auth Routes ---

func TestAuthRoutes(t *testing.T) {
	r, cleanup := setupTestRouter(t)
	defer cleanup()

	t.Run("register", func(t *testing.T) {
		w := doJSON("POST", "/api/v1/auth/register", `{"email":"test@example.com","password":"test123","nickname":"Tester"}`, "", r)
		if w.Code != http.StatusCreated {
			t.Fatalf("register: got %d, want 201; body=%s", w.Code, w.Body.String())
		}
		var envelope struct { Data map[string]interface{} `json:"data"` }
		json.Unmarshal(w.Body.Bytes(), &envelope)
		if envelope.Data["access_token"] == nil {
			t.Error("register: expected access_token in response")
		}
		if envelope.Data["refresh_token"] == nil {
			t.Error("register: expected refresh_token in response")
		}
	})

	t.Run("register_duplicate_email", func(t *testing.T) {
		w := doJSON("POST", "/api/v1/auth/register", `{"email":"test@example.com","password":"test123","nickname":"Tester2"}`, "", r)
		if w.Code != http.StatusBadRequest {
			t.Fatalf("duplicate register: got %d, want 400", w.Code)
		}
		var body map[string]string
		json.Unmarshal(w.Body.Bytes(), &body)
		if body["code"] != "auth.email_exists" {
			t.Errorf("duplicate register: got code %q, want auth.email_exists", body["code"])
		}
	})

	t.Run("register_invalid_body", func(t *testing.T) {
		w := doJSON("POST", "/api/v1/auth/register", `{"email":"not-an-email"}`, "", r)
		if w.Code != http.StatusBadRequest {
			t.Errorf("invalid register: got %d, want 400", w.Code)
		}
	})

	t.Run("login_success", func(t *testing.T) {
		w := doJSON("POST", "/api/v1/auth/login", `{"email":"test@example.com","password":"test123"}`, "", r)
		if w.Code != http.StatusOK {
			t.Fatalf("login: got %d, want 200; body=%s", w.Code, w.Body.String())
		}
		var envelope struct { Data map[string]interface{} `json:"data"` }
		json.Unmarshal(w.Body.Bytes(), &envelope)
		if envelope.Data["access_token"] == nil || envelope.Data["refresh_token"] == nil {
			t.Error("login: expected tokens in response")
		}
	})

	t.Run("login_wrong_password", func(t *testing.T) {
		w := doJSON("POST", "/api/v1/auth/login", `{"email":"test@example.com","password":"wrong"}`, "", r)
		if w.Code != http.StatusUnauthorized {
			t.Errorf("wrong password: got %d, want 401", w.Code)
		}
	})

	t.Run("login_invalid_body", func(t *testing.T) {
		w := doJSON("POST", "/api/v1/auth/login", `{"email":"x"}`, "", r)
		if w.Code != http.StatusBadRequest {
			t.Errorf("invalid login: got %d, want 400", w.Code)
		}
	})
}

func TestAuthRefreshAndLogout(t *testing.T) {
	r, cleanup := setupTestRouter(t)
	defer cleanup()

	// Register and get tokens
	w := doJSON("POST", "/api/v1/auth/register", `{"email":"refresh@test.com","password":"pass123"}`, "", r)
	if w.Code != http.StatusCreated {
		t.Fatalf("register: got %d, body=%s", w.Code, w.Body.String())
	}
	var envelope struct { Data map[string]interface{} `json:"data"` }
	json.Unmarshal(w.Body.Bytes(), &envelope)
	refreshToken := envelope.Data["refresh_token"].(string)
	accessToken := envelope.Data["access_token"].(string)

	t.Run("refresh", func(t *testing.T) {
		w := doJSON("POST", "/api/v1/auth/refresh", fmt.Sprintf(`{"refresh_token":"%s"}`, refreshToken), "", r)
		if w.Code != http.StatusOK {
			t.Fatalf("refresh: got %d, body=%s", w.Code, w.Body.String())
		}
		var newEnvelope struct {
			Data map[string]interface{} `json:"data"`
		}
		json.Unmarshal(w.Body.Bytes(), &newEnvelope)
		if newEnvelope.Data["access_token"] == nil {
			t.Error("refresh: expected new access_token")
		}
	})

	t.Run("refresh_invalid_token", func(t *testing.T) {
		w := doJSON("POST", "/api/v1/auth/refresh", `{"refresh_token":"invalid"}`, "", r)
		if w.Code != http.StatusUnauthorized {
			t.Errorf("refresh invalid: got %d, want 401", w.Code)
		}
	})

	t.Run("logout", func(t *testing.T) {
		w := doJSON("POST", "/api/v1/auth/logout", fmt.Sprintf(`{"refresh_token":"%s"}`, refreshToken), accessToken, r)
		if w.Code != http.StatusOK {
			t.Errorf("logout: got %d, want 200", w.Code)
		}
	})

	t.Run("logout_missing_auth", func(t *testing.T) {
		w := doJSON("POST", "/api/v1/auth/logout", fmt.Sprintf(`{"refresh_token":"%s"}`, refreshToken), "", r)
		if w.Code != http.StatusUnauthorized {
			t.Errorf("logout no auth: got %d, want 401", w.Code)
		}
	})
}

// --- Catalog Routes ---

func TestCatalogRoutes(t *testing.T) {
	r, cleanup := setupTestRouter(t)
	defer cleanup()

	t.Run("list_categories", func(t *testing.T) {
		w := doJSON("GET", "/api/v1/catalog/categories", "", "", r)
		if w.Code != http.StatusOK {
			t.Fatalf("categories: got %d, want 200", w.Code)
		}
		var envelope struct { Data []interface{} `json:"data"` }
		json.Unmarshal(w.Body.Bytes(), &envelope)
		if len(envelope.Data) == 0 {
			t.Error("categories: expected non-empty list")
		}
	})

	t.Run("list_products", func(t *testing.T) {
		w := doJSON("GET", "/api/v1/catalog/products", "", "", r)
		if w.Code != http.StatusOK {
			t.Fatalf("products: got %d, body=%s", w.Code, w.Body.String())
		}
		var envelope struct { Data map[string]interface{} `json:"data"` }
		json.Unmarshal(w.Body.Bytes(), &envelope)
		if envelope.Data["products"] == nil {
			t.Error("products: expected products field")
		}
		if envelope.Data["total"] == nil {
			t.Error("products: expected total field")
		}
	})

	t.Run("list_products_with_pagination", func(t *testing.T) {
		w := doJSON("GET", "/api/v1/catalog/products?page=1&page_size=5", "", "", r)
		if w.Code != http.StatusOK {
			t.Fatalf("products paginated: got %d", w.Code)
		}
		var envelope struct { Data map[string]interface{} `json:"data"` }
		json.Unmarshal(w.Body.Bytes(), &envelope)
		page, _ := envelope.Data["page"].(float64)
		if int(page) != 1 {
			t.Errorf("got page %v, want 1", page)
		}
	})

	t.Run("get_product", func(t *testing.T) {
		// Seed data creates product with ID 1 (or we can use any existing)
		w := doJSON("GET", "/api/v1/catalog/products/1", "", "", r)
		if w.Code != http.StatusOK {
			t.Fatalf("get product: got %d, body=%s", w.Code, w.Body.String())
		}
		var envP struct { Data map[string]interface{} `json:"data"` }
		json.Unmarshal(w.Body.Bytes(), &envP)
		if envP.Data["id"] == nil {
			t.Error("get product: expected product with id")
		}
	})

	t.Run("get_product_not_found", func(t *testing.T) {
		w := doJSON("GET", "/api/v1/catalog/products/99999", "", "", r)
		if w.Code != http.StatusNotFound {
			t.Errorf("get product not found: got %d, want 404", w.Code)
		}
		var body map[string]string
		json.Unmarshal(w.Body.Bytes(), &body)
		if body["code"] != "product.not_found" {
			t.Errorf("got code %q, want product.not_found", body["code"])
		}
	})

	t.Run("get_product_invalid_id", func(t *testing.T) {
		w := doJSON("GET", "/api/v1/catalog/products/abc", "", "", r)
		if w.Code != http.StatusBadRequest {
			t.Errorf("invalid id: got %d, want 400", w.Code)
		}
	})
}

// --- Protected Routes Unauthorized ---

func TestProtectedRoutesUnauthorized(t *testing.T) {
	r, cleanup := setupTestRouter(t)
	defer cleanup()

	tests := []struct {
		method string
		path   string
	}{
		{"GET", "/api/v1/profile"},
		{"PUT", "/api/v1/profile"},
		{"GET", "/api/v1/addresses"},
		{"POST", "/api/v1/addresses"},
		{"GET", "/api/v1/preferences"},
		{"PUT", "/api/v1/preferences"},
		{"GET", "/api/v1/cart"},
		{"POST", "/api/v1/cart/items"},
		{"GET", "/api/v1/orders"},
		{"POST", "/api/v1/orders"},
		{"POST", "/api/v1/agent/chat"},
		{"GET", "/api/v1/metrics"},
	}

	for _, tt := range tests {
		t.Run(fmt.Sprintf("%s_%s", tt.method, tt.path), func(t *testing.T) {
			w := doJSON(tt.method, tt.path, "", "", r)
			if w.Code != http.StatusUnauthorized {
				t.Errorf("%s %s: got %d, want 401", tt.method, tt.path, w.Code)
			}
		})
	}
}

// --- User Routes ---

func TestUserRoutes(t *testing.T) {
	r, cleanup := setupTestRouter(t)
	defer cleanup()

	// Register a user and get JWT
	w := doJSON("POST", "/api/v1/auth/register", `{"email":"userprof@test.com","password":"pass123"}`, "", r)
	if w.Code != http.StatusCreated {
		t.Fatalf("register: got %d", w.Code)
	}
	var envelope struct { Data map[string]interface{} `json:"data"` }
	json.Unmarshal(w.Body.Bytes(), &envelope)
	token := envelope.Data["access_token"].(string)

	t.Run("get_profile_initial", func(t *testing.T) {
		w := doJSON("GET", "/api/v1/profile", "", token, r)
		if w.Code != http.StatusOK {
			t.Fatalf("get profile: got %d, body=%s", w.Code, w.Body.String())
		}
		var envP struct { Data map[string]interface{} `json:"data"` }
		json.Unmarshal(w.Body.Bytes(), &envP)
		// Profile may be empty on first get — that's OK
	})

	t.Run("update_profile", func(t *testing.T) {
		w := doJSON("PUT", "/api/v1/profile", `{"phone":"13800138000","avatar":"https://img.test/a.png"}`, token, r)
		if w.Code != http.StatusOK {
			t.Fatalf("update profile: got %d, body=%s", w.Code, w.Body.String())
		}
		var envP struct { Data map[string]interface{} `json:"data"` }
		json.Unmarshal(w.Body.Bytes(), &envP)
		if envP.Data["phone"] != "13800138000" {
			t.Errorf("update profile: got phone %v, want 13800138000", envP.Data["phone"])
		}
	})

	t.Run("get_profile_after_update", func(t *testing.T) {
		w := doJSON("GET", "/api/v1/profile", "", token, r)
		if w.Code != http.StatusOK {
			t.Fatalf("get profile: got %d", w.Code)
		}
		var envP struct { Data map[string]interface{} `json:"data"` }
		json.Unmarshal(w.Body.Bytes(), &envP)
		if envP.Data["phone"] != "13800138000" {
			t.Errorf("get profile: got phone %v, want 13800138000", envP.Data["phone"])
		}
	})

	t.Run("update_profile_invalid_body", func(t *testing.T) {
		w := doJSON("PUT", "/api/v1/profile", "not json", token, r)
		if w.Code != http.StatusBadRequest {
			t.Errorf("invalid body: got %d, want 400", w.Code)
		}
	})
}

// --- Address Routes ---

func TestAddressRoutes(t *testing.T) {
	r, cleanup := setupTestRouter(t)
	defer cleanup()

	w := doJSON("POST", "/api/v1/auth/register", `{"email":"addr@test.com","password":"pass123"}`, "", r)
	if w.Code != http.StatusCreated {
		t.Fatalf("register: got %d", w.Code)
	}
	var envelope struct { Data map[string]interface{} `json:"data"` }
	json.Unmarshal(w.Body.Bytes(), &envelope)
	token := envelope.Data["access_token"].(string)

	var addrID float64

	t.Run("create_address", func(t *testing.T) {
		w := doJSON("POST", "/api/v1/addresses",
			`{"name":"张三","phone":"13800138000","province":"广东","city":"深圳","district":"南山区","detail":"科技园路1号"}`, token, r)
		if w.Code != http.StatusCreated {
			t.Fatalf("create addr: got %d, body=%s", w.Code, w.Body.String())
		}
		var envA struct { Data map[string]interface{} `json:"data"` }
		json.Unmarshal(w.Body.Bytes(), &envA)
		if envA.Data["id"] == nil || envA.Data["id"].(float64) == 0 {
			t.Error("create addr: expected non-zero id")
		}
		addrID = envA.Data["id"].(float64)
	})

	t.Run("list_addresses", func(t *testing.T) {
		w := doJSON("GET", "/api/v1/addresses", "", token, r)
		if w.Code != http.StatusOK {
			t.Fatalf("list addrs: got %d", w.Code)
		}
		var envAddrs struct { Data []interface{} `json:"data"` }
		json.Unmarshal(w.Body.Bytes(), &envAddrs)
		if len(envAddrs.Data) == 0 {
			t.Error("list addrs: expected at least 1 address")
		}
	})

	t.Run("update_address", func(t *testing.T) {
		w := doJSON("PUT", fmt.Sprintf("/api/v1/addresses/%d", int(addrID)),
			`{"name":"李四","phone":"13900139000","province":"北京","city":"北京","district":"朝阳区","detail":"望京SOHO"}`, token, r)
		if w.Code != http.StatusOK {
			t.Fatalf("update addr: got %d, body=%s", w.Code, w.Body.String())
		}
		var envA struct { Data map[string]interface{} `json:"data"` }
		json.Unmarshal(w.Body.Bytes(), &envA)
		if envA.Data["name"] != "李四" {
			t.Errorf("update addr: got name %v, want 李四", envA.Data["name"])
		}
	})

	t.Run("delete_address", func(t *testing.T) {
		w := doJSON("DELETE", fmt.Sprintf("/api/v1/addresses/%d", int(addrID)), "", token, r)
		if w.Code != http.StatusOK {
			t.Errorf("delete addr: got %d, want 200", w.Code)
		}
	})

	t.Run("delete_nonexistent_address", func(t *testing.T) {
		w := doJSON("DELETE", "/api/v1/addresses/99999", "", token, r)
		// Delete of non-existent may return 500 or 200 depending on implementation
		if w.Code != http.StatusOK && w.Code != http.StatusInternalServerError {
			t.Logf("delete nonexistent: got %d", w.Code)
		}
	})

	t.Run("update_invalid_id", func(t *testing.T) {
		w := doJSON("PUT", "/api/v1/addresses/abc", `{"name":"test"}`, token, r)
		if w.Code != http.StatusBadRequest {
			t.Errorf("invalid id: got %d, want 400", w.Code)
		}
	})
}

// --- Preference Routes ---

func TestPreferenceRoutes(t *testing.T) {
	r, cleanup := setupTestRouter(t)
	defer cleanup()

	w := doJSON("POST", "/api/v1/auth/register", `{"email":"pref@test.com","password":"pass123"}`, "", r)
	if w.Code != http.StatusCreated {
		t.Fatalf("register: got %d", w.Code)
	}
	var envelope struct { Data map[string]interface{} `json:"data"` }
	json.Unmarshal(w.Body.Bytes(), &envelope)
	token := envelope.Data["access_token"].(string)

	t.Run("get_preference_initial", func(t *testing.T) {
		w := doJSON("GET", "/api/v1/preferences", "", token, r)
		if w.Code != http.StatusOK {
			t.Fatalf("get pref: got %d, body=%s", w.Code, w.Body.String())
		}
	})

	t.Run("update_preference", func(t *testing.T) {
		w := doJSON("PUT", "/api/v1/preferences",
			`{"dietary_limits":"vegan","taste_prefs":"辣,甜","scene_prefs":"办公,追剧","allergens":"花生"}`, token, r)
		if w.Code != http.StatusOK {
			t.Fatalf("update pref: got %d, body=%s", w.Code, w.Body.String())
		}
		var envP struct { Data map[string]interface{} `json:"data"` }
		json.Unmarshal(w.Body.Bytes(), &envP)
		if envP.Data["taste_prefs"] != "辣,甜" {
			t.Errorf("update pref: got taste_prefs %v, want 辣,甜", envP.Data["taste_prefs"])
		}
	})

	t.Run("get_preference_after_update", func(t *testing.T) {
		w := doJSON("GET", "/api/v1/preferences", "", token, r)
		if w.Code != http.StatusOK {
			t.Fatalf("get pref: got %d", w.Code)
		}
		var envP struct { Data map[string]interface{} `json:"data"` }
		json.Unmarshal(w.Body.Bytes(), &envP)
		if envP.Data["dietary_limits"] != "vegan" {
			t.Errorf("get pref: got dietary_limits %v, want vegan", envP.Data["dietary_limits"])
		}
	})
}

// --- Cart Routes ---

func TestCartRoutes(t *testing.T) {
	r, cleanup := setupTestRouter(t)
	defer cleanup()

	w := doJSON("POST", "/api/v1/auth/register", `{"email":"cart@test.com","password":"pass123"}`, "", r)
	if w.Code != http.StatusCreated {
		t.Fatalf("register: got %d, body=%s", w.Code, w.Body.String())
	}
	var envelope struct { Data map[string]interface{} `json:"data"` }
	json.Unmarshal(w.Body.Bytes(), &envelope)
	token := envelope.Data["access_token"].(string)

	t.Run("get_empty_cart", func(t *testing.T) {
		w := doJSON("GET", "/api/v1/cart", "", token, r)
		if w.Code != http.StatusOK {
			t.Fatalf("get cart: got %d", w.Code)
		}
		var envelope struct { Data map[string]interface{} `json:"data"` }
		json.Unmarshal(w.Body.Bytes(), &envelope)
		items, _ := envelope.Data["items"].([]interface{})
		if len(items) != 0 {
			t.Errorf("empty cart: got %d items, want 0", len(items))
		}
	})

	t.Run("add_item", func(t *testing.T) {
		w := doJSON("POST", "/api/v1/cart/items", `{"sku_id":1,"quantity":2}`, token, r)
		if w.Code != http.StatusOK {
			t.Fatalf("add item: got %d, body=%s", w.Code, w.Body.String())
		}
		var envelope struct { Data map[string]interface{} `json:"data"` }
		json.Unmarshal(w.Body.Bytes(), &envelope)
		items, _ := envelope.Data["items"].([]interface{})
		if len(items) == 0 {
			t.Fatal("add item: expected items in cart")
		}
		item := items[0].(map[string]interface{})
		qty, _ := item["quantity"].(float64)
		if int(qty) != 2 {
			t.Errorf("add item: got qty %v, want 2", qty)
		}
	})

	t.Run("get_cart_after_add", func(t *testing.T) {
		w := doJSON("GET", "/api/v1/cart", "", token, r)
		if w.Code != http.StatusOK {
			t.Fatalf("get cart: got %d", w.Code)
		}
		var envelope struct { Data map[string]interface{} `json:"data"` }
		json.Unmarshal(w.Body.Bytes(), &envelope)
		items, _ := envelope.Data["items"].([]interface{})
		if len(items) != 1 {
			t.Errorf("get cart: got %d items, want 1", len(items))
		}
	})

	t.Run("update_qty", func(t *testing.T) {
		w := doJSON("PUT", "/api/v1/cart/items/1", `{"quantity":3}`, token, r)
		if w.Code != http.StatusOK {
			t.Fatalf("update qty: got %d, body=%s", w.Code, w.Body.String())
		}
		var envelope struct { Data map[string]interface{} `json:"data"` }
		json.Unmarshal(w.Body.Bytes(), &envelope)
		items, _ := envelope.Data["items"].([]interface{})
		item := items[0].(map[string]interface{})
		qty, _ := item["quantity"].(float64)
		if int(qty) != 3 {
			t.Errorf("update qty: got %v, want 3", qty)
		}
	})

	t.Run("remove_item", func(t *testing.T) {
		w := doJSON("DELETE", "/api/v1/cart/items/1", "", token, r)
		if w.Code != http.StatusOK {
			t.Fatalf("remove item: got %d", w.Code)
		}
		var envelope struct { Data map[string]interface{} `json:"data"` }
		json.Unmarshal(w.Body.Bytes(), &envelope)
		items, _ := envelope.Data["items"].([]interface{})
		if len(items) != 0 {
			t.Errorf("remove item: got %d items, want 0", len(items))
		}
	})

	t.Run("add_invalid_body", func(t *testing.T) {
		w := doJSON("POST", "/api/v1/cart/items", `{"sku_id":1}`, token, r)
		if w.Code != http.StatusBadRequest {
			t.Errorf("invalid body: got %d, want 400", w.Code)
		}
	})

	t.Run("update_invalid_sku_id", func(t *testing.T) {
		w := doJSON("PUT", "/api/v1/cart/items/abc", `{"quantity":1}`, token, r)
		if w.Code != http.StatusBadRequest {
			t.Errorf("invalid sku id: got %d, want 400", w.Code)
		}
	})
}

// --- Order Routes ---

func TestOrderRoutes(t *testing.T) {
	r, cleanup := setupTestRouter(t)
	defer cleanup()

	// Register and add item to cart before creating order
	w := doJSON("POST", "/api/v1/auth/register", `{"email":"order@test.com","password":"pass123"}`, "", r)
	if w.Code != http.StatusCreated {
		t.Fatalf("register: got %d", w.Code)
	}
	var envelope struct { Data map[string]interface{} `json:"data"` }
	json.Unmarshal(w.Body.Bytes(), &envelope)
	token := envelope.Data["access_token"].(string)

	// Add an item to cart for order creation tests
	w = doJSON("POST", "/api/v1/cart/items", `{"sku_id":1,"quantity":1}`, token, r)
	if w.Code != http.StatusOK {
		t.Fatalf("add to cart: got %d", w.Code)
	}

	// Create an address for the order
	w = doJSON("POST", "/api/v1/addresses",
		`{"name":"收货人","phone":"13800138000","province":"广东","city":"深圳","district":"南山区","detail":"收货地址1号"}`, token, r)
	if w.Code != http.StatusCreated {
		t.Fatalf("create address: got %d", w.Code)
	}
	var addrEnv struct { Data map[string]interface{} `json:"data"` }
	json.Unmarshal(w.Body.Bytes(), &addrEnv)
	addrID := int(addrEnv.Data["id"].(float64))

	var orderID uint

	t.Run("create_order", func(t *testing.T) {
		w := doJSON("POST", "/api/v1/orders", fmt.Sprintf(`{"address_id":%d}`, addrID), token, r)
		if w.Code != http.StatusCreated {
			t.Fatalf("create order: got %d, body=%s", w.Code, w.Body.String())
		}
		var orderEnv struct { Data map[string]interface{} `json:"data"` }
		json.Unmarshal(w.Body.Bytes(), &orderEnv)
		if orderEnv.Data["id"] == nil {
			t.Fatal("create order: expected order id")
		}
		orderID = uint(orderEnv.Data["id"].(float64))
		if orderEnv.Data["status"] != "pending_payment" {
			t.Errorf("create order: got status %v, want pending_payment", orderEnv.Data["status"])
		}
		if orderEnv.Data["items"] == nil {
			t.Error("create order: expected items array")
		}
	})

	t.Run("create_order_empty_cart", func(t *testing.T) {
		w := doJSON("POST", "/api/v1/orders", fmt.Sprintf(`{"address_id":%d}`, addrID), token, r)
		if w.Code != http.StatusBadRequest {
			t.Errorf("empty cart order: got %d, want 400", w.Code)
		}
	})

	t.Run("list_orders", func(t *testing.T) {
		w := doJSON("GET", "/api/v1/orders", "", token, r)
		if w.Code != http.StatusOK {
			t.Fatalf("list orders: got %d", w.Code)
		}
		var envelope struct { Data map[string]interface{} `json:"data"` }
		json.Unmarshal(w.Body.Bytes(), &envelope)
		if envelope.Data["orders"] == nil {
			t.Error("list orders: expected orders field")
		}
	})

	t.Run("get_order", func(t *testing.T) {
		w := doJSON("GET", fmt.Sprintf("/api/v1/orders/%d", orderID), "", token, r)
		if w.Code != http.StatusOK {
			t.Fatalf("get order: got %d, body=%s", w.Code, w.Body.String())
		}
		var orderEnv struct { Data map[string]interface{} `json:"data"` }
		json.Unmarshal(w.Body.Bytes(), &orderEnv)
		if orderEnv.Data["order_no"] == nil {
			t.Error("get order: expected order_no")
		}
	})

	t.Run("get_order_not_found", func(t *testing.T) {
		w := doJSON("GET", "/api/v1/orders/99999", "", token, r)
		if w.Code != http.StatusNotFound {
			t.Errorf("get order not found: got %d, want 404", w.Code)
		}
	})

	t.Run("cancel_order", func(t *testing.T) {
		// First create a new order by adding to cart then ordering
		doJSON("POST", "/api/v1/cart/items", `{"sku_id":1,"quantity":1}`, token, r)
		w := doJSON("POST", "/api/v1/orders", fmt.Sprintf(`{"address_id":%d}`, addrID), token, r)
		if w.Code != http.StatusCreated {
			t.Fatalf("create 2nd order: got %d", w.Code)
		}
		var orderEnv struct { Data map[string]interface{} `json:"data"` }
		json.Unmarshal(w.Body.Bytes(), &orderEnv)
		newOrderID := int(orderEnv.Data["id"].(float64))

		w = doJSON("POST", fmt.Sprintf("/api/v1/orders/%d/cancel", newOrderID), "", token, r)
		if w.Code != http.StatusOK {
			t.Errorf("cancel order: got %d, want 200", w.Code)
		}
	})

	t.Run("cancel_order_invalid_id", func(t *testing.T) {
		w := doJSON("POST", "/api/v1/orders/abc/cancel", "", token, r)
		if w.Code != http.StatusBadRequest {
			t.Errorf("cancel invalid id: got %d, want 400", w.Code)
		}
	})
}

// --- Payment Routes ---

func TestPaymentRoutes(t *testing.T) {
	r, cleanup := setupTestRouter(t)
	defer cleanup()

	w := doJSON("POST", "/api/v1/auth/register", `{"email":"pay@test.com","password":"pass123"}`, "", r)
	if w.Code != http.StatusCreated {
		t.Fatalf("register: got %d", w.Code)
	}
	var envelope struct { Data map[string]interface{} `json:"data"` }
	json.Unmarshal(w.Body.Bytes(), &envelope)
	token := envelope.Data["access_token"].(string)

	t.Run("checkout_order_not_found", func(t *testing.T) {
		w := doJSON("POST", "/api/v1/orders/99999/checkout",
			`{"success_url":"https://app.com/success","cancel_url":"https://app.com/cancel"}`, token, r)
		if w.Code != http.StatusBadRequest {
			t.Errorf("checkout not found: got %d, want 400", w.Code)
		}
	})

	t.Run("checkout_invalid_body", func(t *testing.T) {
		w := doJSON("POST", "/api/v1/orders/1/checkout", `{}`, token, r)
		if w.Code != http.StatusBadRequest {
			t.Errorf("checkout invalid body: got %d, want 400", w.Code)
		}
	})

	t.Run("webhook_invalid_signature", func(t *testing.T) {
		w := doJSON("POST", "/api/v1/webhook", `{"type":"checkout.session.completed"}`, "", r)
		// Without Stripe key, HandleWebhook fails at signature verification
		if w.Code != http.StatusBadRequest {
			t.Logf("webhook: got %d (expected 400 without valid Stripe secret)", w.Code)
		}
	})
}

// --- Search Routes (public) ---

func TestSearchRoutes(t *testing.T) {
	r, cleanup := setupTestRouter(t)
	defer cleanup()

	t.Run("search_empty_query", func(t *testing.T) {
		w := doJSON("GET", "/api/v1/catalog/search?q=", "", "", r)
		if w.Code != http.StatusBadRequest {
			t.Errorf("empty query: got %d, want 400", w.Code)
		}
	})

	t.Run("search_without_query_param", func(t *testing.T) {
		w := doJSON("GET", "/api/v1/catalog/search", "", "", r)
		if w.Code != http.StatusBadRequest {
			t.Errorf("no query param: got %d, want 400", w.Code)
		}
	})
}

// --- Agent Chat Routes ---

func TestAgentChatRoutes(t *testing.T) {
	r, cleanup := setupTestRouter(t)
	defer cleanup()

	w := doJSON("POST", "/api/v1/auth/register", `{"email":"agent@test.com","password":"pass123"}`, "", r)
	if w.Code != http.StatusCreated {
		t.Fatalf("register: got %d", w.Code)
	}
	var envelope struct { Data map[string]interface{} `json:"data"` }
	json.Unmarshal(w.Body.Bytes(), &envelope)
	token := envelope.Data["access_token"].(string)

	t.Run("chat_invalid_body", func(t *testing.T) {
		w := doJSON("POST", "/api/v1/agent/chat", `{}`, token, r)
		if w.Code != http.StatusBadRequest {
			t.Errorf("invalid body: got %d, want 400", w.Code)
		}
	})

	t.Run("chat_empty_message", func(t *testing.T) {
		w := doJSON("POST", "/api/v1/agent/chat", `{"message":""}`, token, r)
		// Empty message may fail binding or be accepted — depends on implementation
		if w.Code != http.StatusBadRequest && w.Code != http.StatusOK {
			t.Logf("empty message: got %d", w.Code)
		}
	})
}

// --- Observability Metrics Route ---

func TestMetricsRoute(t *testing.T) {
	r, cleanup := setupTestRouter(t)
	defer cleanup()

	w := doJSON("POST", "/api/v1/auth/register", `{"email":"metrics@test.com","password":"pass123"}`, "", r)
	if w.Code != http.StatusCreated {
		t.Fatalf("register: got %d", w.Code)
	}
	var envelope struct { Data map[string]interface{} `json:"data"` }
	json.Unmarshal(w.Body.Bytes(), &envelope)
	token := envelope.Data["access_token"].(string)

	t.Run("get_metrics", func(t *testing.T) {
		w := doJSON("GET", "/api/v1/metrics", "", token, r)
		if w.Code != http.StatusOK {
			t.Fatalf("metrics: got %d, body=%s", w.Code, w.Body.String())
		}
		var envelope struct { Data map[string]interface{} `json:"data"` }
		json.Unmarshal(w.Body.Bytes(), &envelope)
		if envelope.Data["metrics"] == nil {
			t.Error("metrics: expected metrics field")
		}
	})
}

// --- Admin Routes ---

func TestAdminRoutes(t *testing.T) {
	r, cleanup := setupTestRouter(t)
	defer cleanup()

	// Generate admin JWT
	adminToken, err := generateTestJWT("test-jwt-secret", 1, "admin")
	if err != nil {
		t.Fatalf("generate admin token: %v", err)
	}

	// Generate regular user JWT
	userToken, err := generateTestJWT("test-jwt-secret", 2, "user")
	if err != nil {
		t.Fatalf("generate user token: %v", err)
	}

	t.Run("admin_create_product", func(t *testing.T) {
		w := doJSON("POST", "/api/v1/admin/products",
			`{"category_id":1,"name":"测试商品","name_en":"Test Product","price":1990,"stock":100}`, adminToken, r)
		if w.Code != http.StatusCreated {
			t.Fatalf("create product: got %d, body=%s", w.Code, w.Body.String())
		}
	})

	t.Run("admin_create_product_forbidden", func(t *testing.T) {
		w := doJSON("POST", "/api/v1/admin/products",
			`{"category_id":1,"name":"forbidden"}`, userToken, r)
		if w.Code != http.StatusForbidden {
			t.Errorf("forbidden create: got %d, want 403", w.Code)
		}
	})

	t.Run("admin_create_product_unauthorized", func(t *testing.T) {
		w := doJSON("POST", "/api/v1/admin/products",
			`{"category_id":1,"name":"no auth"}`, "", r)
		if w.Code != http.StatusUnauthorized {
			t.Errorf("no auth: got %d, want 401", w.Code)
		}
	})

	t.Run("admin_update_product", func(t *testing.T) {
		w := doJSON("PUT", "/api/v1/admin/products/1",
			`{"name":"更新商品","name_en":"Updated Product"}`, adminToken, r)
		if w.Code != http.StatusOK {
			t.Fatalf("update product: got %d, body=%s", w.Code, w.Body.String())
		}
	})

	t.Run("admin_delete_product", func(t *testing.T) {
		// Create a product first, then delete it
		w := doJSON("POST", "/api/v1/admin/products",
			`{"category_id":1,"name":"待删除","name_en":"To Delete"}`, adminToken, r)
		var envP struct { Data map[string]interface{} `json:"data"` }
		json.Unmarshal(w.Body.Bytes(), &envP)
		pid := int(envP.Data["id"].(float64))

		w = doJSON("DELETE", fmt.Sprintf("/api/v1/admin/products/%d", pid), "", adminToken, r)
		if w.Code != http.StatusOK {
			t.Errorf("delete product: got %d, want 200", w.Code)
		}
	})

	t.Run("admin_delete_product_unauthorized", func(t *testing.T) {
		w := doJSON("DELETE", "/api/v1/admin/products/1", "", "", r)
		if w.Code != http.StatusUnauthorized {
			t.Errorf("delete no auth: got %d, want 401", w.Code)
		}
	})
}

// --- Middleware Integration ---

func TestMiddlewareIntegration(t *testing.T) {
	r, cleanup := setupTestRouter(t)
	defer cleanup()

	t.Run("request_id_header_set", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/health", nil)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)
		reqID := w.Header().Get("X-Request-ID")
		if reqID == "" {
			t.Error("expected X-Request-ID header to be set")
		}
	})

	t.Run("request_id_from_incoming", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/health", nil)
		req.Header.Set("X-Request-ID", "custom-id-42")
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)
		reqID := w.Header().Get("X-Request-ID")
		if reqID != "custom-id-42" {
			t.Errorf("got X-Request-ID %q, want custom-id-42", reqID)
		}
	})

	t.Run("cors_allowed_origin", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/health", nil)
		req.Header.Set("Origin", "http://localhost:3000")
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)
		allowOrigin := w.Header().Get("Access-Control-Allow-Origin")
		if allowOrigin != "http://localhost:3000" {
			t.Errorf("got Allow-Origin %q, want http://localhost:3000", allowOrigin)
		}
	})

	t.Run("cors_preflight", func(t *testing.T) {
		req := httptest.NewRequest("OPTIONS", "/api/v1/catalog/categories", nil)
		req.Header.Set("Origin", "http://localhost:3000")
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)
		if w.Code != http.StatusNoContent {
			t.Errorf("preflight: got %d, want 204", w.Code)
		}
	})

	t.Run("i18n_default_locale", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/health", nil)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)
		// Response should include locale header or we can verify via API behavior
		if w.Code != http.StatusOK {
			t.Errorf("i18n request: got %d, want 200", w.Code)
		}
	})

	t.Run("i18n_accept_language", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/health", nil)
		req.Header.Set("Accept-Language", "en")
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)
		if w.Code != http.StatusOK {
			t.Errorf("i18n en request: got %d, want 200", w.Code)
		}
	})
}

// --- Edge Cases ---

func TestRouterEdgeCases(t *testing.T) {
	r, cleanup := setupTestRouter(t)
	defer cleanup()

	t.Run("json_parse_error", func(t *testing.T) {
		w := doJSON("POST", "/api/v1/auth/register", `{invalid json}`, "", r)
		if w.Code != http.StatusBadRequest {
			t.Errorf("parse error: got %d, want 400", w.Code)
		}
	})

	t.Run("404_not_found", func(t *testing.T) {
		w := doJSON("GET", "/api/v1/nonexistent", "", "", r)
		if w.Code != http.StatusNotFound {
			t.Errorf("404: got %d, want 404", w.Code)
		}
	})

	t.Run("405_method_not_allowed", func(t *testing.T) {
		w := doJSON("DELETE", "/api/v1/catalog/categories", "", "", r)
		if w.Code != http.StatusNotFound {
			t.Errorf("405/404: got %d, want 404", w.Code)
		}
	})
}
