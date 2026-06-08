# Kotoha Go Backend — Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development or superpowers:executing-plans to implement this plan task-by-task.

**Goal:** Build a modular monolith Go backend for Kotoha snack e-commerce with AI Commerce Agent, Milvus hybrid search, Stripe payment, and Langfuse observability.

**Architecture:** Gin HTTP + GORM ORM + PostgreSQL + Redis + Milvus + MinIO. Internal packages: auth/user/catalog/search/cart/order/payment/agent/observability. REST API with JWT auth + i18n (zh/en).

**Tech Stack:** Go 1.22, Gin, GORM, golang-jwt, go-redis, milvus-sdk-go, stripe-go

---

## File Structure

```
kotoha/
├── cmd/server/main.go
├── internal/
│   ├── config/config.go
│   ├── middleware/{cors,auth,request_id,i18n}.go
│   ├── i18n/{i18n.go,locales/{zh,en}.json}
│   ├── auth/{handler,service,repository,models}.go
│   ├── user/{handler,service,models}.go
│   ├── catalog/{handler,service,repository,models,milvus}.go
│   ├── search/{handler,service,embedding,ranker}.go
│   ├── cart/{handler,service}.go
│   ├── order/{handler,service,repository,models}.go
│   ├── payment/{handler,service,webhook}.go
│   ├── agent/{handler,service,tools,prompts}.go
│   ├── observability/{langfuse,middleware,events}.go
│   └── router/router.go
├── pkg/
│   ├── db/{postgres,migration}.go
│   ├── redis/redis.go
│   ├── milvus/client.go
│   ├── stripe/client.go
│   ├── ollama/client.go
│   └── langfuse/client.go
├── migrations/
├── go.mod / go.sum / Makefile / Dockerfile
```

---

## Phase 1: Foundation (Tasks 1-7) — Complete project scaffold to running Gin server

### Task 1: Go module init + directory scaffold

**Create:** `go.mod`, `cmd/server/main.go`, `Makefile`, all empty dirs

- [ ] **Step 1: Init Go module**

```bash
go mod init github.com/hanasakis/kotoha
```

- [ ] **Step 2: Create directory structure**

```bash
mkdir -p cmd/server internal/{config,middleware,i18n/locales,auth,user,catalog,search,cart,order,payment,agent,observability,router} pkg/{db,redis,milvus,stripe,ollama,langfuse} migrations
```

- [ ] **Step 3: Write minimal main.go**

```go
package main

import "fmt"

func main() {
	fmt.Println("Kotoha server starting...")
}
```

- [ ] **Step 4: Write Makefile**

```makefile
.PHONY: run build test lint

run:
	go run ./cmd/server

build:
	go build -o bin/kotoha ./cmd/server

test:
	go test ./... -v
```

- [ ] **Step 5: Verify**

```bash
go run ./cmd/server
```
Expected: `Kotoha server starting...`

- [ ] **Step 6: Commit & push**

```bash
git add -A && git commit -m "chore: init Go module and project scaffold" && git push origin main
```

---

### Task 2: Config module (env loading with viper)

**Create:** `internal/config/config.go`

- [ ] **Step 1: Add viper**

```bash
go get github.com/spf13/viper
```

- [ ] **Step 2: Write config.go**

```go
package config

import (
	"strings"
	"github.com/spf13/viper"
)

type Config struct {
	Server    ServerConfig
	DB        DBConfig
	Redis     RedisConfig
	Milvus    MilvusConfig
	S3        S3Config
	Stripe    StripeConfig
	JWT       JWTConfig
	Langfuse  LangfuseConfig
	Ollama    OllamaConfig
	LLM       LLMConfig
	Embedding EmbeddingConfig
	I18n      I18nConfig
	CORS      CORSConfig
}

type ServerConfig   struct { Host, Port, Env, LogLevel string }
type DBConfig        struct { Host, Port, User, Password, DBName, SSLMode string; MaxOpenConns, MaxIdleConns int }
type RedisConfig     struct { Host, Port, Password string; DB int }
type MilvusConfig    struct { Host, Port, User, Password, DB string; HNSWM, HNSWEfConstruct, EfSearch, RecallTopK int; DenseWeight, BM25Weight float64 }
type S3Config        struct { Endpoint, AccessKey, SecretKey, Bucket string; UseSSL bool }
type StripeConfig    struct { SecretKey, PublishableKey, WebhookSecret, Currency string }
type JWTConfig       struct { Secret, AccessTTL, RefreshTTL string }
type LangfuseConfig  struct { PublicKey, SecretKey, Host string }
type OllamaConfig    struct { Host string }
type LLMConfig       struct { Model string; MaxTokens int; Temperature float64 }
type EmbeddingConfig struct { Model string; Dim int }
type I18nConfig      struct { DefaultLocale string; SupportedLocales []string }
type CORSConfig      struct { Origins []string }

func Load() (*Config, error) {
	viper.SetConfigFile(".env")
	viper.AutomaticEnv()
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	_ = viper.ReadInConfig()

	return &Config{
		Server:    ServerConfig{Host: getEnv("SERVER_HOST", "0.0.0.0"), Port: getEnv("SERVER_PORT", "8080"), Env: getEnv("ENV", "development"), LogLevel: getEnv("LOG_LEVEL", "debug")},
		DB:        DBConfig{Host: getEnv("DB_HOST", "localhost"), Port: getEnv("DB_PORT", "5432"), User: getEnv("DB_USER", "kotoha"), Password: getEnv("DB_PASSWORD", ""), DBName: getEnv("DB_NAME", "kotoha"), SSLMode: getEnv("DB_SSLMODE", "disable"), MaxOpenConns: getEnvInt("DB_MAX_OPEN_CONNS", 25), MaxIdleConns: getEnvInt("DB_MAX_IDLE_CONNS", 10)},
		Redis:     RedisConfig{Host: getEnv("REDIS_HOST", "localhost"), Port: getEnv("REDIS_PORT", "6379"), Password: getEnv("REDIS_PASSWORD", ""), DB: getEnvInt("REDIS_DB", 0)},
		Milvus:    MilvusConfig{Host: getEnv("MILVUS_HOST", "localhost"), Port: getEnv("MILVUS_PORT", "19530"), User: getEnv("MILVUS_USER", "root"), Password: getEnv("MILVUS_PASSWORD", ""), DB: getEnv("MILVUS_DB", "kotoha"), HNSWM: getEnvInt("MILVUS_HNSW_M", 30), HNSWEfConstruct: getEnvInt("MILVUS_HNSW_EF_CONSTRUCTION", 64), EfSearch: getEnvInt("MILVUS_EF_SEARCH", 80), DenseWeight: getEnvFloat("MILVUS_DENSE_WEIGHT", 0.65), BM25Weight: getEnvFloat("MILVUS_BM25_WEIGHT", 0.35), RecallTopK: getEnvInt("MILVUS_RECALL_TOP_K", 40)},
		S3:        S3Config{Endpoint: getEnv("S3_ENDPOINT", "localhost:9000"), AccessKey: getEnv("S3_ACCESS_KEY", ""), SecretKey: getEnv("S3_SECRET_KEY", ""), Bucket: getEnv("S3_BUCKET", "kotoha-images"), UseSSL: getEnvBool("S3_USE_SSL", false)},
		Stripe:    StripeConfig{SecretKey: getEnv("STRIPE_SECRET_KEY", ""), PublishableKey: getEnv("STRIPE_PUBLISHABLE_KEY", ""), WebhookSecret: getEnv("STRIPE_WEBHOOK_SECRET", ""), Currency: getEnv("STRIPE_CURRENCY", "cny")},
		JWT:       JWTConfig{Secret: getEnv("JWT_SECRET", ""), AccessTTL: getEnv("JWT_ACCESS_TTL", "15m"), RefreshTTL: getEnv("JWT_REFRESH_TTL", "720h")},
		Langfuse:  LangfuseConfig{PublicKey: getEnv("LANGFUSE_PUBLIC_KEY", ""), SecretKey: getEnv("LANGFUSE_SECRET_KEY", ""), Host: getEnv("LANGFUSE_HOST", "https://cloud.langfuse.com")},
		Ollama:    OllamaConfig{Host: getEnv("OLLAMA_HOST", "http://localhost:11434")},
		LLM:       LLMConfig{Model: getEnv("LLM_MODEL", "qwen3.5:9b"), MaxTokens: getEnvInt("LLM_MAX_TOKENS", 2048), Temperature: getEnvFloat("LLM_TEMPERATURE", 0.3)},
		Embedding: EmbeddingConfig{Model: getEnv("EMBEDDING_MODEL", "bge-large-zh-v1.5"), Dim: getEnvInt("EMBEDDING_DIM", 1024)},
		I18n:      I18nConfig{DefaultLocale: getEnv("DEFAULT_LOCALE", "zh"), SupportedLocales: strings.Split(getEnv("SUPPORTED_LOCALES", "zh,en"), ",")},
		CORS:      CORSConfig{Origins: strings.Split(getEnv("CORS_ORIGINS", "http://localhost:3000"), ",")},
	}, nil
}

func (c *Config) DSN() string {
	return "host=" + c.DB.Host + " user=" + c.DB.User + " password=" + c.DB.Password + " dbname=" + c.DB.DBName + " port=" + c.DB.Port + " sslmode=" + c.DB.SSLMode
}
func (c *Config) RedisAddr() string  { return c.Redis.Host + ":" + c.Redis.Port }
func (c *Config) MilvusAddr() string { return c.Milvus.Host + ":" + c.Milvus.Port }

func getEnv(key, def string) string {
	if v := viper.GetString(key); v != "" { return v }
	return def
}
func getEnvInt(key string, def int) int {
	if v := viper.GetInt(key); v != 0 || viper.IsSet(key) { return v }
	return def
}
func getEnvFloat(key string, def float64) float64 {
	if v := viper.GetFloat64(key); v != 0 || viper.IsSet(key) { return v }
	return def
}
func getEnvBool(key string, def bool) bool {
	if viper.IsSet(key) { return viper.GetBool(key) }
	return def
}
```

- [ ] **Step 3: Commit & push**

```bash
git add -A && git commit -m "feat: add config module with env loading" && git push origin main
```

---

### Task 3: Database connection + models + auto-migrate

**Create:** `pkg/db/postgres.go`, `pkg/db/migration.go`, `internal/auth/models.go`, `internal/user/models.go`, `internal/catalog/models.go`, `internal/order/models.go`

- [ ] **Step 1: Add GORM + pg driver**

```bash
go get gorm.io/gorm gorm.io/driver/postgres
```

- [ ] **Step 2: Write pkg/db/postgres.go**

```go
package db

import (
	"log"
	"time"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func NewPostgres(dsn string, maxOpen, maxIdle int) (*gorm.DB, error) {
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	})
	if err != nil {
		return nil, err
	}
	sqlDB, _ := db.DB()
	sqlDB.SetMaxOpenConns(maxOpen)
	sqlDB.SetMaxIdleConns(maxIdle)
	sqlDB.SetConnMaxLifetime(5 * time.Minute)
	log.Println("[DB] PostgreSQL connected")
	return db, nil
}
```

- [ ] **Step 3: Write models**

`internal/auth/models.go`:

```go
package auth

import (
	"time"
	"gorm.io/gorm"
)

type User struct {
	ID           uint           `gorm:"primaryKey" json:"id"`
	Email        string         `gorm:"uniqueIndex;size:255;not null" json:"email"`
	PasswordHash string         `gorm:"size:255;not null" json:"-"`
	Nickname     string         `gorm:"size:100" json:"nickname"`
	Role         string         `gorm:"size:20;default:user" json:"role"`
	CreatedAt    time.Time      `json:"created_at"`
	UpdatedAt    time.Time      `json:"updated_at"`
	DeletedAt    gorm.DeletedAt `gorm:"index" json:"-"`
}

type Session struct {
	ID           uint      `gorm:"primaryKey" json:"id"`
	UserID       uint      `gorm:"index;not null" json:"user_id"`
	RefreshToken string    `gorm:"size:512;not null" json:"-"`
	DeviceType   string    `gorm:"size:50" json:"device_type"`
	DeviceIP     string    `gorm:"size:50" json:"device_ip"`
	IsValid      bool      `gorm:"default:true" json:"is_valid"`
	ExpiresAt    time.Time `json:"expires_at"`
	CreatedAt    time.Time `json:"created_at"`
}
```

`internal/user/models.go`:

```go
package user

import "time"

type Profile struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	UserID    uint      `gorm:"uniqueIndex;not null" json:"user_id"`
	Avatar    string    `gorm:"size:500" json:"avatar"`
	Phone     string    `gorm:"size:20" json:"phone"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type Address struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	UserID    uint      `gorm:"index;not null" json:"user_id"`
	Name      string    `gorm:"size:50;not null" json:"name"`
	Phone     string    `gorm:"size:20;not null" json:"phone"`
	Province  string    `gorm:"size:50" json:"province"`
	City      string    `gorm:"size:50" json:"city"`
	District  string    `gorm:"size:50" json:"district"`
	Detail    string    `gorm:"size:500" json:"detail"`
	IsDefault bool      `gorm:"default:false" json:"is_default"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type Preference struct {
	ID            uint   `gorm:"primaryKey" json:"id"`
	UserID        uint   `gorm:"uniqueIndex;not null" json:"user_id"`
	DietaryLimits string `gorm:"size:500" json:"dietary_limits"`
	TastePrefs    string `gorm:"size:500" json:"taste_prefs"`
	ScenePrefs    string `gorm:"size:500" json:"scene_prefs"`
	Allergens     string `gorm:"size:500" json:"allergens"`
}
```

`internal/catalog/models.go`:

```go
package catalog

import (
	"time"
	"gorm.io/gorm"
)

type Category struct {
	ID        uint           `gorm:"primaryKey" json:"id"`
	Name      string         `gorm:"size:100;not null" json:"name"`
	NameEn    string         `gorm:"size:100" json:"name_en"`
	ParentID  *uint          `gorm:"index" json:"parent_id"`
	SortOrder int            `gorm:"default:0" json:"sort_order"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}

type Product struct {
	ID            uint           `gorm:"primaryKey" json:"id"`
	CategoryID    uint           `gorm:"index;not null" json:"category_id"`
	Name          string         `gorm:"size:200;not null" json:"name"`
	NameEn        string         `gorm:"size:200" json:"name_en"`
	Description   string         `gorm:"type:text" json:"description"`
	DescriptionEn string         `gorm:"type:text" json:"description_en"`
	ImageURL      string         `gorm:"size:500" json:"image_url"`
	Tags          string         `gorm:"size:500" json:"tags"`
	Scenes        string         `gorm:"size:500" json:"scenes"`
	Taste         string         `gorm:"size:200" json:"taste"`
	Allergens     string         `gorm:"size:500" json:"allergens"`
	Ingredients   string         `gorm:"type:text" json:"ingredients"`
	WeightGram    int            `json:"weight_gram"`
	ShelfDays     int            `json:"shelf_days"`
	IsActive      bool           `gorm:"default:true" json:"is_active"`
	ClickCount    int            `gorm:"default:0" json:"click_count"`
	BuyCount      int            `gorm:"default:0" json:"buy_count"`
	CreatedAt     time.Time      `json:"created_at"`
	UpdatedAt     time.Time      `json:"updated_at"`
	DeletedAt     gorm.DeletedAt `gorm:"index" json:"-"`
	SKUs          []SKU          `gorm:"foreignKey:ProductID" json:"skus,omitempty"`
}

type SKU struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	ProductID uint      `gorm:"index;not null" json:"product_id"`
	Name      string    `gorm:"size:200" json:"name"`
	Price     int       `gorm:"not null" json:"price"`
	Stock     int       `gorm:"not null;default:0" json:"stock"`
	ImageURL  string    `gorm:"size:500" json:"image_url"`
	IsDefault bool      `gorm:"default:false" json:"is_default"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}
```

`internal/order/models.go`:

```go
package order

import (
	"time"
	"gorm.io/gorm"
)

const (
	StatusPending   = "pending_payment"
	StatusPaid      = "paid"
	StatusShipped   = "shipped"
	StatusDelivered = "delivered"
	StatusCancelled = "cancelled"
)

type Order struct {
	ID          uint           `gorm:"primaryKey" json:"id"`
	UserID      uint           `gorm:"index;not null" json:"user_id"`
	OrderNo     string         `gorm:"uniqueIndex;size:32;not null" json:"order_no"`
	Status      string         `gorm:"size:30;default:pending_payment" json:"status"`
	TotalAmount int            `gorm:"not null" json:"total_amount"`
	Currency    string         `gorm:"size:10;default:cny" json:"currency"`
	StripeSID   string         `gorm:"size:100" json:"stripe_session_id"`
	PaidAt      *time.Time     `json:"paid_at"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `gorm:"index" json:"-"`
	Items       []OrderItem    `gorm:"foreignKey:OrderID" json:"items,omitempty"`
}

type OrderItem struct {
	ID        uint   `gorm:"primaryKey" json:"id"`
	OrderID   uint   `gorm:"index;not null" json:"order_id"`
	ProductID uint   `gorm:"not null" json:"product_id"`
	SKUID     uint   `gorm:"not null" json:"sku_id"`
	Name      string `gorm:"size:200" json:"name"`
	Price     int    `gorm:"not null" json:"price"`
	Quantity  int    `gorm:"not null" json:"quantity"`
}

type Payment struct {
	ID             uint       `gorm:"primaryKey" json:"id"`
	OrderID        uint       `gorm:"uniqueIndex;not null" json:"order_id"`
	StripePID      string     `gorm:"size:100" json:"stripe_payment_intent_id"`
	Amount         int        `gorm:"not null" json:"amount"`
	Currency       string     `gorm:"size:10" json:"currency"`
	Status         string     `gorm:"size:30" json:"status"`
	IdempotencyKey string     `gorm:"size:100" json:"idempotency_key"`
	PaidAt         *time.Time `json:"paid_at"`
	CreatedAt      time.Time  `json:"created_at"`
}
```

- [ ] **Step 4: Write pkg/db/migration.go**

```go
package db

import (
	"log"
	"gorm.io/gorm"
	"github.com/hanasakis/kotoha/internal/auth"
	"github.com/hanasakis/kotoha/internal/catalog"
	"github.com/hanasakis/kotoha/internal/order"
	"github.com/hanasakis/kotoha/internal/user"
)

func AutoMigrate(db *gorm.DB) error {
	log.Println("[DB] Running auto-migration...")
	return db.AutoMigrate(
		&auth.User{}, &auth.Session{},
		&user.Profile{}, &user.Address{}, &user.Preference{},
		&catalog.Category{}, &catalog.Product{}, &catalog.SKU{},
		&order.Order{}, &order.OrderItem{}, &order.Payment{},
	)
}
```

- [ ] **Step 5: Update cmd/server/main.go**

```go
package main

import (
	"log"
	"github.com/hanasakis/kotoha/internal/config"
	"github.com/hanasakis/kotoha/pkg/db"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	database, err := db.NewPostgres(cfg.DSN(), cfg.DB.MaxOpenConns, cfg.DB.MaxIdleConns)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	if err := db.AutoMigrate(database); err != nil {
		log.Fatalf("Failed to migrate database: %v", err)
	}

	log.Printf("Kotoha server starting on %s:%s", cfg.Server.Host, cfg.Server.Port)
}
```

- [ ] **Step 6: Verify build**

```bash
go mod tidy && go build ./...
```

- [ ] **Step 7: Commit & push**

```bash
git add -A && git commit -m "feat: add database connection, models, and auto-migration" && git push origin main
```

---

### Task 4: Redis connection

**Create:** `pkg/redis/redis.go`

- [ ] **Step 1: Add go-redis**

```bash
go get github.com/redis/go-redis/v9
```

- [ ] **Step 2: Write pkg/redis/redis.go**

```go
package redis

import (
	"context"
	"log"
	"time"
	goredis "github.com/redis/go-redis/v9"
)

type Client struct {
	RDB *goredis.Client
}

func New(addr, password string, db int) (*Client, error) {
	rdb := goredis.NewClient(&goredis.Options{
		Addr: addr, Password: password, DB: db,
		DialTimeout: 5 * time.Second, ReadTimeout: 3 * time.Second, WriteTimeout: 3 * time.Second,
	})
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := rdb.Ping(ctx).Err(); err != nil {
		return nil, err
	}
	log.Println("[Redis] connected")
	return &Client{RDB: rdb}, nil
}

func (c *Client) Close() error { return c.RDB.Close() }
```

- [ ] **Step 3: Update main.go, add after DB init:**

```go
redisClient, err := redis.New(cfg.RedisAddr(), cfg.Redis.Password, cfg.Redis.DB)
if err != nil {
    log.Fatalf("Failed to connect to Redis: %v", err)
}
defer redisClient.Close()
```

Also add import: `"github.com/hanasakis/kotoha/pkg/redis"`

- [ ] **Step 4: `go build ./...`** then commit & push

```bash
git add -A && git commit -m "feat: add Redis connection client" && git push origin main
```

---

### Task 5: i18n module

**Create:** `internal/i18n/i18n.go`, `internal/i18n/locales/zh.json`, `internal/i18n/locales/en.json`

zh.json keys: auth.login_success, auth.login_failed, auth.register_success, auth.email_exists, auth.invalid_token, auth.unauthorized, auth.forbidden, product.not_found, product.out_of_stock, cart.added, cart.removed, cart.empty, order.created, order.not_found, order.paid, order.cancelled, payment.pending, payment.failed, agent.recommendation, agent.no_result, common.server_error, common.invalid_request

en.json: same keys with English values.

```go
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
	if err != nil { return err }
	var m map[string]string
	if err := json.Unmarshal(data, &m); err != nil { return err }
	t.mu.Lock()
	t.translations[locale] = m
	t.mu.Unlock()
	return nil
}

func (t *Translator) T(locale, key string) string {
	t.mu.RLock()
	defer t.mu.RUnlock()
	if m, ok := t.translations[locale]; ok {
		if v, ok := m[key]; ok { return v }
	}
	if m, ok := t.translations[t.defaultLocale]; ok {
		if v, ok := m[key]; ok { return v }
	}
	return key
}
```

Commit: `feat: add i18n module with zh/en translations`

---

### Task 6: Middleware (CORS, RequestID, i18n)

**Create:** `internal/middleware/cors.go`, `internal/middleware/request_id.go`, `internal/middleware/i18n.go`

Add dependencies: `go get github.com/gin-gonic/gin github.com/google/uuid`

**cors.go:** Gin middleware, checks Origin against allowlist, sets CORS headers.
**request_id.go:** Reads `X-Request-ID` header or generates UUID, sets in context + response header.
**i18n.go:** Reads `Accept-Language` header, matches against supported locales, stores `locale` in Gin context.

Commit: `feat: add CORS, RequestID, and i18n detection middleware`

---

### Task 7: Router + server startup

**Create:** `internal/router/router.go`, update `cmd/server/main.go`

`router/router.go` defines `Dependencies` struct (DB, Redis, Config, I18n), `Setup()` function creating Gin engine with middleware chain, `/health` endpoint, `/api/v1` group.

`main.go` loads i18n JSON files, wires all deps into router, calls `r.Run(addr)`.

Commit: `feat: add router setup and Gin server entrypoint`

---

## Remaining Phases (summary)

**Phase 2 — Auth & User (Tasks 8-10):** auth handler/service/repository (register, login, refresh token, JWT generation, bcrypt), user profile/address CRUD

**Phase 3 — Catalog & Cart (Tasks 11-13):** product/category CRUD, Redis-backed cart (add/remove/item count), cart sync

**Phase 4 — Order & Payment (Tasks 14-16):** order creation (stock reservation, order number gen), Stripe Checkout Session, webhook handler (signature verification, idempotency, order->paid transition)

**Phase 5 — Search & Agent & Observability (Tasks 17-21):** Milvus client + collection setup, Ollama embedding + hybrid search service, Agent tools (search_products, get_product_detail, compare_products, recommend_bundle, add_to_cart, create_checkout_session), Langfuse trace middleware + business events, evaluation metrics

**Phase 6 — Seed Data & Polish:** 20-30 snack products seed data, 30-50 evaluation queries, integration tests
