package order

import (
	"gorm.io/gorm"
)

type Repository struct {
	DB *gorm.DB
}

func NewRepository(db *gorm.DB) *Repository {
	return &Repository{DB: db}
}

func (r *Repository) Create(o *Order) error {
	return r.DB.Create(o).Error
}

func (r *Repository) GetByID(id, userID uint) (*Order, error) {
	var o Order
	err := r.DB.Preload("Items").Where("id = ? AND user_id = ?", id, userID).First(&o).Error
	if err != nil {
		return nil, err
	}
	return &o, nil
}

func (r *Repository) ListByUser(userID uint, page, pageSize int) ([]Order, int64, error) {
	var orders []Order
	var total int64

	query := r.DB.Model(&Order{}).Where("user_id = ?", userID)
	query.Count(&total)

	err := query.Preload("Items").Offset((page - 1) * pageSize).Limit(pageSize).
		Order("created_at DESC").Find(&orders).Error
	return orders, total, err
}

func (r *Repository) UpdateStatus(id uint, status string) error {
	return r.DB.Model(&Order{}).Where("id = ?", id).Update("status", status).Error
}

func (r *Repository) UpdateStripeSession(id uint, stripeSID string) error {
	return r.DB.Model(&Order{}).Where("id = ?", id).Update("stripe_session_id", stripeSID).Error
}

func (r *Repository) FindByOrderNo(orderNo string) (*Order, error) {
	var o Order
	err := r.DB.Preload("Items").Where("order_no = ?", orderNo).First(&o).Error
	if err != nil {
		return nil, err
	}
	return &o, nil
}

func (r *Repository) SetPaid(id uint) error {
	return r.DB.Model(&Order{}).Where("id = ?", id).
		Updates(map[string]interface{}{"status": StatusPaid, "paid_at": gorm.Expr("NOW()")}).Error
}

func (r *Repository) CreatePayment(p *Payment) error {
	return r.DB.Create(p).Error
}

func (r *Repository) UpdatePaymentStatus(orderID uint, status, stripePID string) error {
	return r.DB.Model(&Payment{}).Where("order_id = ?", orderID).
		Updates(map[string]interface{}{"status": status, "stripe_payment_intent_id": stripePID, "paid_at": gorm.Expr("NOW()")}).Error
}
