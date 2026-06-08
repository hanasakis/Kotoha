package testutil

import (
	"testing"

	"github.com/hanasakis/kotoha/internal/config"
	goredis "github.com/hanasakis/kotoha/pkg/redis"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// acquireLock uses a PostgreSQL advisory lock to serialize database setup
// across parallel test packages, preventing deadlocks during migration/seed.
func acquireLock(t *testing.T, db *gorm.DB) func() {
	t.Helper()
	// 42 is a magic number for this project's test lock
	if err := db.Exec("SELECT pg_advisory_lock(1542)").Error; err != nil {
		t.Logf("[testutil] failed to acquire advisory lock: %v", err)
		return func() {}
	}
	return func() {
		db.Exec("SELECT pg_advisory_unlock(1542)")
	}
}

func SetupTestDB(t *testing.T) *gorm.DB {
	t.Helper()

	cfg, err := config.Load()
	if err != nil {
		t.Skipf("skipping integration test: cannot load config: %v", err)
	}

	db, err := gorm.Open(postgres.Open(cfg.DSN()), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		t.Skipf("skipping integration test: cannot connect to database: %v", err)
	}

	unlock := acquireLock(t, db)
	t.Cleanup(unlock)

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
	db.Exec("TRUNCATE TABLE order_items, payments, orders, skus, products, categories, sessions, addresses, preferences, profiles, users RESTART IDENTITY CASCADE")
}
