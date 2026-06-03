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
