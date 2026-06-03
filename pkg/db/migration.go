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
		&auth.User{},
		&auth.Session{},
		&user.Profile{},
		&user.Address{},
		&user.Preference{},
		&catalog.Category{},
		&catalog.Product{},
		&catalog.SKU{},
		&order.Order{},
		&order.OrderItem{},
		&order.Payment{},
	)
}
