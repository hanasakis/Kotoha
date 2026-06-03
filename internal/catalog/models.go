package catalog

import (
	"time"

	"gorm.io/gorm"
)

type Category struct {
	ID        uint           `gorm:"primaryKey" json:"id"`
	Name      string         `gorm:"size:100;not null" json:"name"`
	NameEn    string         `gorm:"size:100" json:"name_en"`
	ParentID  *uint          `gorm:"index" json:"parent_id"`
	SortOrder int            `gorm:"default:0" json:"sort_order"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}

type Product struct {
	ID            uint           `gorm:"primaryKey" json:"id"`
	CategoryID    uint           `gorm:"index;not null" json:"category_id"`
	Name          string         `gorm:"size:200;not null" json:"name"`
	NameEn        string         `gorm:"size:200" json:"name_en"`
	Description   string         `gorm:"type:text" json:"description"`
	DescriptionEn string         `gorm:"type:text" json:"description_en"`
	ImageURL      string         `gorm:"size:500" json:"image_url"`
	Tags          string         `gorm:"size:500" json:"tags"`
	Scenes        string         `gorm:"size:500" json:"scenes"`
	Taste         string         `gorm:"size:200" json:"taste"`
	Allergens     string         `gorm:"size:500" json:"allergens"`
	Ingredients   string         `gorm:"type:text" json:"ingredients"`
	WeightGram    int            `json:"weight_gram"`
	ShelfDays     int            `json:"shelf_days"`
	IsActive      bool           `gorm:"default:true" json:"is_active"`
	ClickCount    int            `gorm:"default:0" json:"click_count"`
	BuyCount      int            `gorm:"default:0" json:"buy_count"`
	CreatedAt     time.Time      `json:"created_at"`
	UpdatedAt     time.Time      `json:"updated_at"`
	DeletedAt     gorm.DeletedAt `gorm:"index" json:"-"`

	SKUs []SKU `gorm:"foreignKey:ProductID" json:"skus,omitempty"`
}

type SKU struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	ProductID uint      `gorm:"index;not null" json:"product_id"`
	Name      string    `gorm:"size:200" json:"name"`
	Price     int       `gorm:"not null" json:"price"`
	Stock     int       `gorm:"not null;default:0" json:"stock"`
	ImageURL  string    `gorm:"size:500" json:"image_url"`
	IsDefault bool      `gorm:"default:false" json:"is_default"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}
