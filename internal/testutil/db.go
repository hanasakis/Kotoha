package testutil

import (
	"testing"

	"github.com/hanasakis/kotoha/internal/config"
	goredis "github.com/hanasakis/kotoha/pkg/redis"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func SetupTestDB(t *testing.T) *gorm.DB {
	t.Helper()

	cfg, err := config.Load()
	if err != nil {
		t.Skipf("skipping integration test: cannot load config: %v", err)
	}

	db, err := gorm.Open(postgres.Open(cfg.DSN()), &gorm.Config{})
	if err != nil {
		t.Skipf("skipping integration test: cannot connect to database: %v", err)
	}

	return db
}

func SetupTestRedis(t *testing.T) *goredis.Client {
	t.Helper()

	cfg, err := config.Load()
	if err != nil {
		t.Skipf("skipping integration test: cannot connect to config: %v", err)
	}

	rdb, err := goredis.New(cfg.RedisAddr(), cfg.Redis.Password, cfg.Redis.DB)
	if err != nil {
		t.Skipf("skipping integration test: cannot connect to redis: %v", err)
	}

	return rdb
}

func CleanTestDB(t *testing.T, db *gorm.DB) {
	t.Helper()
	db.Exec("TRUNCATE TABLE order_items, payments, orders, skus, products, categories, sessions, addresses, preferences, profiles, users CASCADE")
}
