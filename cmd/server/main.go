// @title           Kotoha API
// @version         1.0
// @description     AI-powered snack e-commerce platform with hybrid search, conversational agent, and Stripe payments.
// @host            localhost:8080
// @BasePath        /api/v1
// @schemes         http https
// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
// @description Type "Bearer" followed by a space and JWT token.

package main

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"runtime"
	"syscall"
	"time"

	"github.com/hanasakis/kotoha/internal/catalog"
	"github.com/hanasakis/kotoha/internal/config"
	"github.com/hanasakis/kotoha/internal/i18n"
	"github.com/hanasakis/kotoha/internal/router"
	"github.com/hanasakis/kotoha/pkg/db"
	klog "github.com/hanasakis/kotoha/pkg/log"
	goredis "github.com/hanasakis/kotoha/pkg/redis"

	_ "github.com/hanasakis/kotoha/docs"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		klog.Fatalf("Failed to load config: %v", err)
	}
	if err := cfg.Validate(); err != nil {
		klog.Fatalf("Invalid config: %v", err)
	}

	database, err := db.NewPostgres(cfg.DSN(), cfg.DB.MaxOpenConns, cfg.DB.MaxIdleConns)
	if err != nil {
		klog.Fatalf("Failed to connect to database: %v", err)
	}

	if err := db.AutoMigrate(database); err != nil {
		klog.Fatalf("Failed to migrate database: %v", err)
	}

	// Seed demo data on fresh install
	catalogRepo := catalog.NewRepository(database)
	catalogSvc := catalog.NewService(catalogRepo, cfg.S3PublicURL())
	if err := catalogSvc.SeedData(); err != nil {
		klog.Warnf("Failed to seed data: %v", err)
	}

	redisClient, err := goredis.New(cfg.RedisAddr(), cfg.Redis.Password, cfg.Redis.DB)
	if err != nil {
		klog.Fatalf("Failed to connect to Redis: %v", err)
	}
	defer redisClient.Close()

	// i18n: load translation files
	_, b, _, _ := runtime.Caller(0)
	basePath := filepath.Dir(filepath.Dir(b))
	translator := i18n.New(cfg.I18n.DefaultLocale)
	if err := translator.Load("zh", filepath.Join(basePath, "internal/i18n/locales/zh.json")); err != nil {
		klog.Warnf("Failed to load zh translations: %v", err)
	}
	if err := translator.Load("en", filepath.Join(basePath, "internal/i18n/locales/en.json")); err != nil {
		klog.Warnf("Failed to load en translations: %v", err)
	}

	deps := &router.Dependencies{
		DB:     database,
		Redis:  redisClient,
		Config: cfg,
		I18n:   translator,
	}

	r := router.Setup(deps)

	addr := cfg.Server.Host + ":" + cfg.Server.Port
	srv := &http.Server{
		Addr:    addr,
		Handler: r,
	}

	// Graceful shutdown on SIGINT / SIGTERM
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		klog.Infof("Kotoha server starting on %s", addr)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			klog.Fatalf("Failed to start server: %v", err)
		}
	}()

	<-quit
	klog.Info("Shutting down server...", nil)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		klog.Errorf("Server forced to shutdown: %v", err)
	}

	// Close database connection pool
	if sqlDB, err := database.DB(); err == nil {
		sqlDB.Close()
	}

	klog.Info("Server exited", nil)
}
