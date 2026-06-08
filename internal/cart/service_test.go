package cart_test

import (
	"context"
	"strings"
	"testing"

	"github.com/hanasakis/kotoha/internal/cart"
	"github.com/hanasakis/kotoha/internal/catalog"
	"github.com/hanasakis/kotoha/internal/testutil"
	"github.com/hanasakis/kotoha/pkg/db"
)

func setupCartService(t *testing.T) (*cart.Service, *catalog.Repository, func()) {
	t.Helper()

	database := testutil.SetupTestDB(t)
	if err := db.AutoMigrate(database); err != nil {
		t.Fatalf("failed to migrate: %v", err)
	}

	catalogRepo := catalog.NewRepository(database)
	catalogSvc := catalog.NewService(catalogRepo, "http://localhost:9000/kotoha-images/products/")
	if err := catalogSvc.SeedData(); err != nil {
		t.Fatalf("failed to seed: %v", err)
	}

	redisClient := testutil.SetupTestRedis(t)

	cartRepo := cart.NewRepository(redisClient)
	cartSvc := cart.NewService(cartRepo, catalogRepo)

	cleanup := func() {
		redisClient.RDB.FlushDB(context.Background())
		testutil.CleanTestDB(t, database)
		redisClient.Close()
	}

	return cartSvc, catalogRepo, cleanup
}

func TestAddItem(t *testing.T) {
	svc, catalogRepo, cleanup := setupCartService(t)
	defer cleanup()
	ctx := context.Background()

	products, _, err := catalogRepo.ListProducts(1, 1, "", 0)
	if err != nil {
		t.Fatalf("failed to list products: %v", err)
	}
	if len(products) == 0 {
		t.Fatal("expected at least one product")
	}
	if len(products[0].SKUs) == 0 {
		t.Fatal("expected at least one SKU")
	}
	sku := products[0].SKUs[0]

	t.Run("add_new_item", func(t *testing.T) {
		err := svc.AddItem(ctx, 1, sku.ID, 2)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		items, err := svc.GetCart(ctx, 1)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(items) != 1 {
			t.Fatalf("got %d items, want 1", len(items))
		}
		if items[0].SKUID != sku.ID {
			t.Errorf("got SKUID %d, want %d", items[0].SKUID, sku.ID)
		}
		if items[0].Quantity != 2 {
			t.Errorf("got quantity %d, want 2", items[0].Quantity)
		}
		if items[0].Price != sku.Price {
			t.Errorf("got price %d, want %d", items[0].Price, sku.Price)
		}
		if items[0].ProductName != products[0].Name {
			t.Errorf("got ProductName %q, want %q", items[0].ProductName, products[0].Name)
		}
	})

	t.Run("cumulative_add", func(t *testing.T) {
		err := svc.AddItem(ctx, 1, sku.ID, 3)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		items, _ := svc.GetCart(ctx, 1)
		if len(items) != 1 {
			t.Fatalf("got %d items, want 1", len(items))
		}
		if items[0].Quantity != 5 {
			t.Errorf("got quantity %d, want 5", items[0].Quantity)
		}
	})

	t.Run("insufficient_stock", func(t *testing.T) {
		err := svc.AddItem(ctx, 2, sku.ID, sku.Stock+1)
		if err == nil {
			t.Fatal("expected error, got nil")
		}
		if !strings.Contains(err.Error(), "insufficient_stock") {
			t.Errorf("expected error to contain 'insufficient_stock', got %q", err.Error())
		}
	})

	t.Run("sku_not_found", func(t *testing.T) {
		err := svc.AddItem(ctx, 3, 99999, 1)
		if err == nil {
			t.Fatal("expected error, got nil")
		}
		if !strings.Contains(err.Error(), "sku_not_found") {
			t.Errorf("expected error to contain 'sku_not_found', got %q", err.Error())
		}
	})

	t.Run("cumulative_exceeds_stock", func(t *testing.T) {
		err := svc.AddItem(ctx, 4, sku.ID, sku.Stock)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		err = svc.AddItem(ctx, 4, sku.ID, 1)
		if err == nil {
			t.Fatal("expected error, got nil")
		}
		if !strings.Contains(err.Error(), "insufficient_stock") {
			t.Errorf("expected error to contain 'insufficient_stock', got %q", err.Error())
		}
	})
}

func TestGetCart(t *testing.T) {
	svc, catalogRepo, cleanup := setupCartService(t)
	defer cleanup()
	ctx := context.Background()

	t.Run("empty_cart", func(t *testing.T) {
		items, err := svc.GetCart(ctx, 99)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(items) != 0 {
			t.Errorf("expected empty cart, got %d items", len(items))
		}
	})

	t.Run("single_item", func(t *testing.T) {
		products, _, _ := catalogRepo.ListProducts(1, 1, "", 0)
		if len(products) == 0 || len(products[0].SKUs) == 0 {
			t.Fatal("expected at least one product with SKU")
		}
		sku := products[0].SKUs[0]
		if err := svc.AddItem(ctx, 10, sku.ID, 1); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		items, err := svc.GetCart(ctx, 10)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(items) != 1 {
			t.Fatalf("got %d items, want 1", len(items))
		}
		if items[0].Price != sku.Price {
			t.Errorf("got price %d, want %d", items[0].Price, sku.Price)
		}
		if items[0].Stock <= 0 {
			t.Error("expected positive stock")
		}
		if items[0].ProductName == "" {
			t.Error("expected non-empty ProductName")
		}
	})

	t.Run("multiple_skus", func(t *testing.T) {
		products, _, _ := catalogRepo.ListProducts(1, 2, "", 0)
		if len(products) < 2 {
			t.Fatal("expected at least 2 products")
		}
		sku1 := products[0].SKUs[0]
		sku2 := products[1].SKUs[0]
		if err := svc.AddItem(ctx, 11, sku1.ID, 1); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if err := svc.AddItem(ctx, 11, sku2.ID, 2); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		items, err := svc.GetCart(ctx, 11)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(items) != 2 {
			t.Errorf("got %d items, want 2", len(items))
		}
	})
}

func TestUpdateQty(t *testing.T) {
	svc, catalogRepo, cleanup := setupCartService(t)
	defer cleanup()
	ctx := context.Background()

	products, _, _ := catalogRepo.ListProducts(1, 1, "", 0)
	if len(products) == 0 || len(products[0].SKUs) == 0 {
		t.Fatal("expected at least one product with SKU")
	}
	sku := products[0].SKUs[0]

	t.Run("update_success", func(t *testing.T) {
		if err := svc.AddItem(ctx, 20, sku.ID, 1); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if err := svc.UpdateQty(ctx, 20, sku.ID, 3); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		items, _ := svc.GetCart(ctx, 20)
		if len(items) != 1 {
			t.Fatalf("got %d items, want 1", len(items))
		}
		if items[0].Quantity != 3 {
			t.Errorf("got quantity %d, want 3", items[0].Quantity)
		}
	})

	t.Run("update_exceeds_stock", func(t *testing.T) {
		err := svc.UpdateQty(ctx, 20, sku.ID, sku.Stock+1)
		if err == nil {
			t.Fatal("expected error, got nil")
		}
		if !strings.Contains(err.Error(), "insufficient_stock") {
			t.Errorf("expected error to contain 'insufficient_stock', got %q", err.Error())
		}
	})

	t.Run("update_nonexistent_sku", func(t *testing.T) {
		err := svc.UpdateQty(ctx, 20, 99999, 1)
		if err == nil {
			t.Fatal("expected error, got nil")
		}
		if !strings.Contains(err.Error(), "sku_not_found") {
			t.Errorf("expected error to contain 'sku_not_found', got %q", err.Error())
		}
	})
}

func TestRemoveItem(t *testing.T) {
	svc, catalogRepo, cleanup := setupCartService(t)
	defer cleanup()
	ctx := context.Background()

	products, _, _ := catalogRepo.ListProducts(1, 1, "", 0)
	if len(products) == 0 || len(products[0].SKUs) == 0 {
		t.Fatal("expected at least one product with SKU")
	}
	sku := products[0].SKUs[0]

	t.Run("remove_existing", func(t *testing.T) {
		if err := svc.AddItem(ctx, 30, sku.ID, 1); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if err := svc.RemoveItem(ctx, 30, sku.ID); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		items, _ := svc.GetCart(ctx, 30)
		if len(items) != 0 {
			t.Errorf("expected empty cart, got %d items", len(items))
		}
	})

	t.Run("remove_nonexistent", func(t *testing.T) {
		err := svc.RemoveItem(ctx, 31, 99999)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})
}

func TestClearCart(t *testing.T) {
	svc, catalogRepo, cleanup := setupCartService(t)
	defer cleanup()
	ctx := context.Background()

	products, _, _ := catalogRepo.ListProducts(1, 2, "", 0)
	if len(products) < 2 {
		t.Fatal("expected at least 2 products")
	}
	sku1 := products[0].SKUs[0]
	sku2 := products[1].SKUs[0]

	t.Run("clear_populated_cart", func(t *testing.T) {
		if err := svc.AddItem(ctx, 40, sku1.ID, 1); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if err := svc.AddItem(ctx, 40, sku2.ID, 2); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if err := svc.ClearCart(ctx, 40); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		items, _ := svc.GetCart(ctx, 40)
		if len(items) != 0 {
			t.Errorf("expected empty cart, got %d items", len(items))
		}
	})

	t.Run("clear_empty_cart", func(t *testing.T) {
		err := svc.ClearCart(ctx, 41)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})
}
