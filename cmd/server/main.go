package main

import (
	"log"
	"path/filepath"
	"runtime"

	"github.com/hanasakis/kotoha/internal/config"
	"github.com/hanasakis/kotoha/internal/i18n"
	"github.com/hanasakis/kotoha/internal/router"
	"github.com/hanasakis/kotoha/pkg/db"
	goredis "github.com/hanasakis/kotoha/pkg/redis"
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

	redisClient, err := goredis.New(cfg.RedisAddr(), cfg.Redis.Password, cfg.Redis.DB)
	if err != nil {
		log.Fatalf("Failed to connect to Redis: %v", err)
	}
	defer redisClient.Close()

	// i18n: load translation files
	_, b, _, _ := runtime.Caller(0)
	basePath := filepath.Dir(filepath.Dir(b))
	translator := i18n.New(cfg.I18n.DefaultLocale)
	if err := translator.Load("zh", filepath.Join(basePath, "internal/i18n/locales/zh.json")); err != nil {
		log.Printf("[WARN] Failed to load zh translations: %v", err)
	}
	if err := translator.Load("en", filepath.Join(basePath, "internal/i18n/locales/en.json")); err != nil {
		log.Printf("[WARN] Failed to load en translations: %v", err)
	}

	deps := &router.Dependencies{
		DB:     database,
		Redis:  redisClient,
		Config: cfg,
		I18n:   translator,
	}

	r := router.Setup(deps)

	addr := cfg.Server.Host + ":" + cfg.Server.Port
	log.Printf("Kotoha server starting on %s", addr)
	if err := r.Run(addr); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
