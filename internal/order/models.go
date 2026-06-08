package order

import (
	"time"

	"gorm.io/gorm"
)

const (
	StatusPending          = "pending_payment"
	StatusPaid             = "paid"
	StatusShipped          = "shipped"
	StatusDelivered        = "delivered"
	StatusCancelled        = "cancelled"
	StatusRefunded         = "refunded"
	StatusPartialRefunded  = "partially_refunded"
	StatusExpired          = "expired"
	StatusPaymentFailed    = "payment_failed"
)

type Order struct {
	ID          uint           `gorm:"primaryKey" json:"id"`
	UserID      uint           `gorm:"index;not null" json:"user_id"`
	OrderNo     string         `gorm:"uniqueIndex;size:32;not null" json:"order_no"`
	Status      string         `gorm:"size:30;default:pending_payment" json:"status"`
	TotalAmount int            `gorm:"not null" json:"total_amount"`
	Currency    string         `gorm:"size:10;default:cny" json:"currency"`
	StripeSID   string         `gorm:"column:stripe_session_id;size:512" json:"stripe_session_id"`
	PaidAt      *time.Time     `json:"paid_at"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `gorm:"index" json:"-"`

	Items []OrderItem `gorm:"foreignKey:OrderID" json:"items,omitempty"`
}

type OrderItem struct {
	ID        uint   `gorm:"primaryKey" json:"id"`
	OrderID   uint   `gorm:"index;not null" json:"order_id"`
	ProductID uint   `gorm:"not null" json:"product_id"`
	SKUID     uint   `gorm:"not null" json:"sku_id"`
	Name      string `gorm:"size:200" json:"name"`
	Price     int    `gorm:"not null" json:"price"`
	Quantity  int    `gorm:"not null" json:"quantity"`
}

type Payment struct {
	ID             uint       `gorm:"primaryKey" json:"id"`
	OrderID        uint       `gorm:"uniqueIndex;not null" json:"order_id"`
	StripePID      string     `gorm:"size:100" json:"stripe_payment_intent_id"`
	Amount         int        `gorm:"not null" json:"amount"`
	Currency       string     `gorm:"size:10" json:"currency"`
	Status         string     `gorm:"size:30" json:"status"`
	IdempotencyKey string     `gorm:"size:100" json:"idempotency_key"`
	PaidAt         *time.Time `json:"paid_at"`
	RefundStripeID string     `gorm:"size:100" json:"refund_stripe_id"`
	RefundAmount   int        `gorm:"default:0" json:"refund_amount"`
	RefundedAt     *time.Time `json:"refunded_at"`
	CreatedAt      time.Time  `json:"created_at"`
}
