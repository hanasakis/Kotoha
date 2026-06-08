package db

import (
	klog "github.com/hanasakis/kotoha/pkg/log"
	"gorm.io/gorm"

	"github.com/hanasakis/kotoha/internal/auth"
	"github.com/hanasakis/kotoha/internal/catalog"
	"github.com/hanasakis/kotoha/internal/order"
	"github.com/hanasakis/kotoha/internal/user"
)

func AutoMigrate(db *gorm.DB) error {
	klog.Info("Running auto-migration...", nil)
	if err := db.AutoMigrate(
		&auth.User{},
		&auth.Session{},
		&auth.PasswordResetToken{},
		&user.Profile{},
		&user.Address{},
		&user.Preference{},
		&catalog.Category{},
		&catalog.Product{},
		&catalog.SKU{},
		&order.Order{},
		&order.OrderItem{},
		&order.Payment{},
	); err != nil {
		return err
	}

	// Ensure columns that GORM may miss on repeated runs are present.
	for _, stmt := range []string{
		`ALTER TABLE orders ADD COLUMN IF NOT EXISTS stripe_session_id VARCHAR(512) DEFAULT ''`,
		`ALTER TABLE orders ALTER COLUMN stripe_session_id TYPE VARCHAR(512)`,
			`ALTER TABLE orders DROP COLUMN IF EXISTS stripe_sid`,
		`ALTER TABLE sessions ALTER COLUMN device_type TYPE VARCHAR(512)`,
	} {
		if err := db.Exec(stmt).Error; err != nil {
			klog.Warnf("Migration note: %v", err)
		}
	}
	return nil
}
