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
