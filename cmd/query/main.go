package main

import (
	"fmt"
	"os"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func main() {
	dsn := "host=localhost user=kotoha password=253121 dbname=kotoha port=5432 sslmode=disable"
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	var users []struct {
		ID        uint   `gorm:"column:id"`
		Email     string `gorm:"column:email"`
		Nickname  string `gorm:"column:nickname"`
		Role      string `gorm:"column:role"`
		CreatedAt string `gorm:"column:created_at"`
	}
	db.Table("users").Find(&users)
	for _, u := range users {
		fmt.Printf("ID=%-3d  %-30s  %-15s  %-12s  %s\n", u.ID, u.Email, u.Nickname, u.Role, u.CreatedAt)
	}
}
