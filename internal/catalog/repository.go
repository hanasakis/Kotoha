package catalog

import (
	"gorm.io/gorm"
)

type Repository struct {
	DB *gorm.DB
}

func NewRepository(db *gorm.DB) *Repository {
	return &Repository{DB: db}
}

// --- Categories ---

func (r *Repository) ListCategories() ([]Category, error) {
	var cats []Category
	err := r.DB.Order("sort_order ASC").Find(&cats).Error
	return cats, err
}

func (r *Repository) GetCategory(id uint) (*Category, error) {
	var cat Category
	err := r.DB.First(&cat, id).Error
	if err != nil {
		return nil, err
	}
	return &cat, nil
}

func (r *Repository) CreateCategory(cat *Category) error {
	return r.DB.Create(cat).Error
}

// --- Products ---

func (r *Repository) ListProducts(page, pageSize int) ([]Product, int64, error) {
	var products []Product
	var total int64

	query := r.DB.Model(&Product{}).Where("is_active = ?", true)
	query.Count(&total)

	err := query.Preload("SKUs").Offset((page - 1) * pageSize).Limit(pageSize).
		Order("created_at DESC").Find(&products).Error
	return products, total, err
}

func (r *Repository) GetProduct(id uint) (*Product, error) {
	var p Product
	err := r.DB.Preload("SKUs").First(&p, id).Error
	if err != nil {
		return nil, err
	}
	return &p, nil
}

func (r *Repository) CreateProduct(p *Product) error {
	return r.DB.Create(p).Error
}

func (r *Repository) UpdateProduct(p *Product) error {
	return r.DB.Save(p).Error
}

func (r *Repository) DeleteProduct(id uint) error {
	return r.DB.Delete(&Product{}, id).Error
}

func (r *Repository) GetProductsByIDs(ids []uint) ([]Product, error) {
	var products []Product
	err := r.DB.Preload("SKUs").Where("id IN ? AND is_active = ?", ids, true).Find(&products).Error
	return products, err
}

func (r *Repository) IncrementClickCount(id uint) {
	r.DB.Model(&Product{}).Where("id = ?", id).UpdateColumn("click_count", gorm.Expr("click_count + 1"))
}

func (r *Repository) IncrementBuyCount(id uint) {
	r.DB.Model(&Product{}).Where("id = ?", id).UpdateColumn("buy_count", gorm.Expr("buy_count + 1"))
}

// --- SKUs ---

func (r *Repository) GetSKU(id uint) (*SKU, error) {
	var sku SKU
	err := r.DB.First(&sku, id).Error
	if err != nil {
		return nil, err
	}
	return &sku, nil
}

func (r *Repository) UpdateSKUStock(id uint, delta int) error {
	return r.DB.Model(&SKU{}).Where("id = ?", id).
		UpdateColumn("stock", gorm.Expr("stock + ?", delta)).Error
}

func (r *Repository) DecrStock(skuID uint, qty int) error {
	return r.DB.Model(&SKU{}).Where("id = ? AND stock >= ?", skuID, qty).
		UpdateColumn("stock", gorm.Expr("stock - ?", qty)).Error
}
