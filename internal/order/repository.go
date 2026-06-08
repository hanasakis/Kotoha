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

func (r *Repository) GetPaymentByOrderID(orderID uint) (*Payment, error) {
	var p Payment
	err := r.DB.Where("order_id = ?", orderID).First(&p).Error
	if err != nil {
		return nil, err
	}
	return &p, nil
}

func (r *Repository) SetRefunded(orderID uint, refundStripeID string, refundAmount int) error {
	return r.DB.Model(&Payment{}).Where("order_id = ?", orderID).
		Updates(map[string]interface{}{
			"refund_stripe_id": refundStripeID,
			"refund_amount":    refundAmount,
			"refunded_at":      gorm.Expr("NOW()"),
		}).Error
}

func (r *Repository) FindPaymentByStripePID(stripePID string) (*Payment, error) {
	var p Payment
	err := r.DB.Where("stripe_payment_intent_id = ?", stripePID).First(&p).Error
	if err != nil {
		return nil, err
	}
	return &p, nil
}

func (r *Repository) SetPaymentFailed(id uint) error {
	return r.DB.Model(&Order{}).Where("id = ?", id).
		Update("status", StatusPaymentFailed).Error
}

func (r *Repository) SetExpired(id uint) error {
	return r.DB.Model(&Order{}).Where("id = ?", id).
		Update("status", StatusExpired).Error
}

// ListAll returns all orders (admin only).
func (r *Repository) ListAll(page, pageSize int, status string) ([]Order, int64, error) {
	var orders []Order
	var total int64

	q := r.DB.Model(&Order{})
	if status != "" {
		q = q.Where("status = ?", status)
	}
	q.Count(&total)

	err := q.Preload("Items").Offset((page - 1) * pageSize).Limit(pageSize).
		Order("created_at DESC").Find(&orders).Error
	return orders, total, err
}

// GetByIDAdmin returns an order by ID without user scoping (admin only).
func (r *Repository) GetByIDAdmin(id uint) (*Order, error) {
	var o Order
	err := r.DB.Preload("Items").Where("id = ?", id).First(&o).Error
	if err != nil {
		return nil, err
	}
	return &o, nil
}
