package order

import (
	"context"
	"fmt"
	"time"

	"github.com/hanasakis/kotoha/internal/cart"
	"github.com/hanasakis/kotoha/internal/catalog"
)

type Service struct {
	repo        *Repository
	cartRepo    *cart.Repository
	catalogRepo *catalog.Repository
}

func NewService(repo *Repository, cartRepo *cart.Repository, catalogRepo *catalog.Repository) *Service {
	return &Service{repo: repo, cartRepo: cartRepo, catalogRepo: catalogRepo}
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
	return o, nil
}

func (s *Service) ListOrders(userID uint, page, pageSize int) ([]Order, int64, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 50 {
		pageSize = 20
	}
	return s.repo.ListByUser(userID, page, pageSize)
}

func (s *Service) CancelOrder(id, userID uint) error {
	o, err := s.repo.GetByID(id, userID)
	if err != nil {
		return fmt.Errorf("order.not_found")
	}
	if o.Status != StatusPending {
		return fmt.Errorf("order.cannot_cancel")
	}

	// Restore stock
	for _, item := range o.Items {
		s.catalogRepo.UpdateSKUStock(item.SKUID, item.Quantity)
	}

	return s.repo.UpdateStatus(id, StatusCancelled)
}

func generateOrderNo(userID uint) string {
	now := time.Now()
	return fmt.Sprintf("KO%d%s%04d", userID, now.Format("20060102150405"), now.Nanosecond()/1000000)
}
