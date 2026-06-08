package order

import (
	"context"
	"fmt"
	"time"

	"github.com/hanasakis/kotoha/internal/cart"
	"github.com/hanasakis/kotoha/internal/catalog"
	klog "github.com/hanasakis/kotoha/pkg/log"
	stripepkg "github.com/hanasakis/kotoha/pkg/stripe"
)

const paymentWindow = 10 * time.Minute

type Service struct {
	repo        *Repository
	cartRepo    *cart.Repository
	catalogRepo *catalog.Repository
	stripeCli   *stripepkg.Client
}

func NewService(repo *Repository, cartRepo *cart.Repository, catalogRepo *catalog.Repository, stripeCli *stripepkg.Client) *Service {
	return &Service{repo: repo, cartRepo: cartRepo, catalogRepo: catalogRepo, stripeCli: stripeCli}
}

type CreateOrderInput struct {
	AddressID uint   `json:"address_id" binding:"required"`
	Note      string `json:"note"`
}

func (s *Service) CreateOrder(ctx context.Context, userID uint, input CreateOrderInput) (*Order, error) {
	rawItems, err := s.cartRepo.GetItems(ctx, userID)
	if err != nil || len(rawItems) == 0 {
		return nil, fmt.Errorf("order.cart_empty")
	}

	var orderItems []OrderItem
	var totalAmount int

	for skuID, qty := range rawItems {
		sku, err := s.catalogRepo.GetSKU(skuID)
		if err != nil {
			return nil, fmt.Errorf("order.sku_not_found")
		}
		if sku.Stock < qty {
			return nil, fmt.Errorf("order.insufficient_stock")
		}

		product, err := s.catalogRepo.GetProduct(sku.ProductID)
		if err != nil {
			return nil, fmt.Errorf("order.product_not_found")
		}

		itemName := product.Name + " - " + sku.Name
		itemTotal := sku.Price * qty
		totalAmount += itemTotal

		orderItems = append(orderItems, OrderItem{
			ProductID: sku.ProductID,
			SKUID:     sku.ID,
			Name:      itemName,
			Price:     sku.Price,
			Quantity:  qty,
		})
	}

	// Deduct stock atomically
	for _, item := range orderItems {
		if err := s.catalogRepo.DecrStock(item.SKUID, item.Quantity); err != nil {
			return nil, fmt.Errorf("order.insufficient_stock")
		}
	}

	order := &Order{
		UserID:      userID,
		OrderNo:     generateOrderNo(userID),
		Status:      StatusPending,
		TotalAmount: totalAmount,
		Currency:    "cny",
		Items:       orderItems,
	}

	if err := s.repo.Create(order); err != nil {
		return nil, err
	}

	// Clear cart after order created
	s.cartRepo.Clear(ctx, userID)

	return order, nil
}

func (s *Service) GetOrder(id, userID uint) (*Order, error) {
	o, err := s.repo.GetByID(id, userID)
	if err != nil {
		return nil, fmt.Errorf("order.not_found")
	}
	s.expireIfStale(o)
	return o, nil
}

func (s *Service) ListOrders(userID uint, page, pageSize int) ([]Order, int64, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 50 {
		pageSize = 20
	}
	orders, total, err := s.repo.ListByUser(userID, page, pageSize)
	if err != nil {
		return nil, 0, err
	}
	for i := range orders {
		s.expireIfStale(&orders[i])
	}
	return orders, total, nil
}

func (s *Service) expireIfStale(o *Order) {
	if o.Status != StatusPending {
		return
	}
	if time.Since(o.CreatedAt) <= paymentWindow {
		return
	}
	klog.Infof("[order] expiring stale order %d (order_no=%s, created=%s)", o.ID, o.OrderNo, o.CreatedAt)
	// Restore stock
	for _, item := range o.Items {
		if err := s.catalogRepo.UpdateSKUStock(item.SKUID, item.Quantity); err != nil {
			klog.Warnf("[order] failed to restore stock for sku %d: %v", item.SKUID, err)
		}
	}
	if err := s.repo.UpdateStatus(o.ID, StatusExpired); err != nil {
		klog.Warnf("[order] failed to expire order %d: %v", o.ID, err)
		return
	}
	o.Status = StatusExpired
}

func (s *Service) CancelOrder(id, userID uint) error {
	o, err := s.repo.GetByID(id, userID)
	if err != nil {
		return fmt.Errorf("order.not_found")
	}

	switch o.Status {
	case StatusPending:
		// Restore stock, mark cancelled
		for _, item := range o.Items {
			s.catalogRepo.UpdateSKUStock(item.SKUID, item.Quantity)
		}
		return s.repo.UpdateStatus(id, StatusCancelled)

	case StatusPaid:
		// Get payment to find Stripe Payment Intent ID
		payment, err := s.repo.GetPaymentByOrderID(o.ID)
		if err != nil {
			return fmt.Errorf("payment.payment_not_found")
		}

		// Refund via Stripe
		refundID, refundAmount, err := s.stripeCli.CreateRefund(stripepkg.RefundParams{
			PaymentIntentID: payment.StripePID,
			Amount:          0, // full refund
		})
		if err != nil {
			return fmt.Errorf("payment.refund_error")
		}

		// Update payment record with refund info
		if err := s.repo.SetRefunded(o.ID, refundID, int(refundAmount)); err != nil {
			return err
		}

		// Restore stock
		for _, item := range o.Items {
			s.catalogRepo.UpdateSKUStock(item.SKUID, item.Quantity)
		}

		// Update order status
		return s.repo.UpdateStatus(id, StatusRefunded)

	default:
		return fmt.Errorf("order.cannot_cancel")
	}
}

func (s *Service) ListAllOrders(page, pageSize int, status string) ([]Order, int64, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}
	orders, total, err := s.repo.ListAll(page, pageSize, status)
	if err != nil {
		return nil, 0, err
	}
	for i := range orders {
		s.expireIfStale(&orders[i])
	}
	return orders, total, nil
}

func (s *Service) GetOrderAdmin(id uint) (*Order, error) {
	o, err := s.repo.GetByIDAdmin(id)
	if err != nil {
		return nil, fmt.Errorf("order.not_found")
	}
	s.expireIfStale(o)
	return o, nil
}

func generateOrderNo(userID uint) string {
	now := time.Now()
	return fmt.Sprintf("KO%d%s%04d", userID, now.Format("20060102150405"), now.Nanosecond()/1000000)
}
