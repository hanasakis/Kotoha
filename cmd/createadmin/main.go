package main

import (
	"fmt"
	"os"

	"github.com/hanasakis/kotoha/internal/config"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

type User struct {
	ID           uint   `gorm:"primaryKey"`
	Email        string `gorm:"uniqueIndex;size:255;not null"`
	PasswordHash string `gorm:"size:255;not null"`
	Nickname     string `gorm:"size:100"`
	Role         string `gorm:"size:20;default:user"`
}

func main() {
	if len(os.Args) < 3 {
		fmt.Println("Usage: createadmin <email> <password>")
		os.Exit(1)
	}
	email := os.Args[1]
	password := os.Args[2]

	cfg, err := config.Load()
	if err != nil {
		fmt.Println("Failed to load config:", err)
		os.Exit(1)
	}

	db, err := gorm.Open(postgres.Open(cfg.DSN()), &gorm.Config{})
	if err != nil {
		fmt.Println("Failed to connect to DB:", err)
		os.Exit(1)
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		fmt.Println("Failed to hash password:", err)
		os.Exit(1)
	}

	var existing User
	result := db.Where("email = ?", email).First(&existing)
	if result.Error == nil {
		db.Model(&existing).Updates(map[string]interface{}{
			"password_hash": string(hash),
			"role":          "admin",
		})
		fmt.Printf("Updated user %s to admin role\n", email)
	} else {
		user := User{
			Email:        email,
			PasswordHash: string(hash),
			Nickname:     "Admin",
			Role:         "admin",
		}
		db.Create(&user)
		fmt.Printf("Created admin user: %s (ID: %d)\n", email, user.ID)
	}
}
