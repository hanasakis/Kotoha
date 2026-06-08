package order_test

import (
	"context"
	"strings"
	"testing"

	"github.com/hanasakis/kotoha/internal/cart"
	"github.com/hanasakis/kotoha/internal/catalog"
	"github.com/hanasakis/kotoha/internal/order"
	"github.com/hanasakis/kotoha/internal/testutil"
	"github.com/hanasakis/kotoha/pkg/db"
)

func TestCreateOrder(t *testing.T) {
	database := testutil.SetupTestDB(t)
	defer testutil.CleanTestDB(t, database)

	if err := db.AutoMigrate(database); err != nil {
		t.Fatalf("failed to migrate: %v", err)
	}

	catalogRepo := catalog.NewRepository(database)
	catalogSvc := catalog.NewService(catalogRepo, "http://localhost:9000/kotoha-images/products/")
	if err := catalogSvc.SeedData(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	products, _, err := catalogRepo.ListProducts(1, 10, "", 0)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(products) == 0 {
		t.Fatal("expected at least one product")
	}

	redisClient := testutil.SetupTestRedis(t)
	defer redisClient.Close()

	cartRepo := cart.NewRepository(redisClient)
	cartSvc := cart.NewService(cartRepo, catalogRepo)
	ctx := context.Background()

	firstProduct := products[0]
	firstSKU := firstProduct.SKUs[0]
	err = cartSvc.AddItem(ctx, 1, firstSKU.ID, 2)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	t.Run("create_from_cart", func(t *testing.T) {
		orderRepo := order.NewRepository(database)
		orderSvc := order.NewService(orderRepo, cartRepo, catalogRepo, nil)

		created, err := orderSvc.CreateOrder(ctx, 1, order.CreateOrderInput{AddressID: 1})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if created.OrderNo == "" {
			t.Error("expected non-empty OrderNo")
		}
		if len(created.OrderNo) != 21 {
			t.Errorf("got OrderNo length %d, want 21", len(created.OrderNo))
		}
		if created.Status != order.StatusPending {
			t.Errorf("got status %q, want %q", created.Status, order.StatusPending)
		}
		if created.Currency != "cny" {
			t.Errorf("got currency %q, want cny", created.Currency)
		}
		if len(created.Items) != 1 {
			t.Errorf("got %d items, want 1", len(created.Items))
		}
		if created.Items[0].SKUID != firstSKU.ID {
			t.Errorf("got SKUID %d, want %d", created.Items[0].SKUID, firstSKU.ID)
		}
		if created.Items[0].Quantity != 2 {
			t.Errorf("got quantity %d, want 2", created.Items[0].Quantity)
		}
		if created.TotalAmount != firstSKU.Price*2 {
			t.Errorf("got TotalAmount %d, want %d", created.TotalAmount, firstSKU.Price*2)
		}

		items, _ := cartSvc.GetCart(ctx, 1)
		if len(items) != 0 {
			t.Errorf("expected empty cart after order, got %d items", len(items))
		}
	})

	t.Run("empty_cart_should_fail", func(t *testing.T) {
		orderRepo := order.NewRepository(database)
		orderSvc := order.NewService(orderRepo, cartRepo, catalogRepo, nil)

		_, err := orderSvc.CreateOrder(ctx, 2, order.CreateOrderInput{AddressID: 1})
		if err == nil {
			t.Fatal("expected error, got nil")
		}
		if !strings.Contains(err.Error(), "cart_empty") {
			t.Errorf("expected 'cart_empty' error, got %q", err.Error())
		}
	})
}

func TestListOrders(t *testing.T) {
	database := testutil.SetupTestDB(t)
	defer testutil.CleanTestDB(t, database)

	if err := db.AutoMigrate(database); err != nil {
		t.Fatalf("failed to migrate: %v", err)
	}

	orderRepo := order.NewRepository(database)

	t.Run("list_empty", func(t *testing.T) {
		orders, total, err := orderRepo.ListByUser(1, 1, 10)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if total != 0 {
			t.Errorf("got total %d, want 0", total)
		}
		if len(orders) != 0 {
			t.Errorf("expected empty orders, got %d", len(orders))
		}
	})
}

func TestCancelOrder(t *testing.T) {
	database := testutil.SetupTestDB(t)
	defer testutil.CleanTestDB(t, database)

	if err := db.AutoMigrate(database); err != nil {
		t.Fatalf("failed to migrate: %v", err)
	}

	catalogRepo := catalog.NewRepository(database)
	catalogSvc := catalog.NewService(catalogRepo, "http://localhost:9000/kotoha-images/products/")
	if err := catalogSvc.SeedData(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	products, _, _ := catalogRepo.ListProducts(1, 1, "", 0)
	if len(products) == 0 {
		t.Fatal("expected at least one product")
	}
	sku := products[0].SKUs[0]
	initialStock := sku.Stock

	redisClient := testutil.SetupTestRedis(t)
	defer redisClient.Close()

	cartRepo := cart.NewRepository(redisClient)
	orderRepo := order.NewRepository(database)
	orderSvc := order.NewService(orderRepo, cartRepo, catalogRepo, nil)

	testOrder := &order.Order{
		UserID:      1,
		OrderNo:     "TEST-ORDER-001",
		Status:      order.StatusPending,
		TotalAmount: sku.Price,
		Currency:    "cny",
		Items:       []order.OrderItem{{ProductID: sku.ProductID, SKUID: sku.ID, Name: "test", Price: sku.Price, Quantity: 1}},
	}
	if err := orderRepo.Create(testOrder); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	catalogRepo.DecrStock(sku.ID, 1)

	t.Run("cancel_pending", func(t *testing.T) {
		err := orderSvc.CancelOrder(testOrder.ID, 1)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		updated, err := orderRepo.GetByID(testOrder.ID, 1)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if updated.Status != order.StatusCancelled {
			t.Errorf("got status %q, want %q", updated.Status, order.StatusCancelled)
		}

		restoredSKU, _ := catalogRepo.GetSKU(sku.ID)
		if restoredSKU.Stock != initialStock {
			t.Errorf("got restored stock %d, want %d", restoredSKU.Stock, initialStock)
		}
	})

	t.Run("cancel_nonexistent", func(t *testing.T) {
		err := orderSvc.CancelOrder(99999, 1)
		if err == nil {
			t.Fatal("expected error, got nil")
		}
	})
}
