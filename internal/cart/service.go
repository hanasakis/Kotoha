package cart

import (
	"context"
	"fmt"

	"github.com/hanasakis/kotoha/internal/catalog"
)

type Service struct {
	repo        *Repository
	catalogRepo *catalog.Repository
}

func NewService(repo *Repository, catalogRepo *catalog.Repository) *Service {
	return &Service{repo: repo, catalogRepo: catalogRepo}
}

func (s *Service) GetCart(ctx context.Context, userID uint) ([]CartItem, error) {
	raw, err := s.repo.GetItems(ctx, userID)
	if err != nil {
		return nil, err
	}
	if len(raw) == 0 {
		return []CartItem{}, nil
	}

	skuIDs := make([]uint, 0, len(raw))
	for id := range raw {
		skuIDs = append(skuIDs, id)
	}

	skus, err := s.catalogRepo.GetSKUsByIDs(skuIDs)
	if err != nil {
		return nil, err
	}

	productIDs := make([]uint, 0, len(skus))
	skuMap := make(map[uint]catalog.SKU, len(skus))
	for _, sku := range skus {
		skuMap[sku.ID] = sku
		productIDs = append(productIDs, sku.ProductID)
	}

	products, err := s.catalogRepo.GetProductsByIDs(productIDs)
	if err != nil {
		return nil, err
	}
	productMap := make(map[uint]catalog.Product, len(products))
	for _, p := range products {
		productMap[p.ID] = p
	}

	items := make([]CartItem, 0, len(skus))
	for _, sku := range skus {
		qty := raw[sku.ID]
		product := productMap[sku.ProductID]
		items = append(items, CartItem{
			SKUID:       sku.ID,
			ProductID:   sku.ProductID,
			ProductName: product.Name,
			SKUName:     sku.Name,
			Price:       sku.Price,
			Quantity:    qty,
			ImageURL:    sku.ImageURL,
			Stock:       sku.Stock,
		})
	}
	return items, nil
}

func (s *Service) AddItem(ctx context.Context, userID, skuID uint, qty int) error {
	sku, err := s.catalogRepo.GetSKU(skuID)
	if err != nil {
		return fmt.Errorf("cart.sku_not_found")
	}
	if sku.Stock < qty {
		return fmt.Errorf("cart.insufficient_stock")
	}
	raw, err := s.repo.GetItems(ctx, userID)
	if err != nil {
		return fmt.Errorf("cart.error: %w", err)
	}
	newQty := raw[skuID] + qty
	if newQty > sku.Stock {
		return fmt.Errorf("cart.insufficient_stock")
	}
	return s.repo.SetItem(ctx, userID, skuID, newQty)
}

func (s *Service) UpdateQty(ctx context.Context, userID, skuID uint, qty int) error {
	sku, err := s.catalogRepo.GetSKU(skuID)
	if err != nil {
		return fmt.Errorf("cart.sku_not_found")
	}
	if qty > sku.Stock {
		return fmt.Errorf("cart.insufficient_stock")
	}
	return s.repo.SetItem(ctx, userID, skuID, qty)
}

func (s *Service) RemoveItem(ctx context.Context, userID, skuID uint) error {
	return s.repo.RemoveItem(ctx, userID, skuID)
}

func (s *Service) ClearCart(ctx context.Context, userID uint) error {
	return s.repo.Clear(ctx, userID)
}
