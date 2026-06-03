package order_test

import (
	"context"
	"testing"

	"github.com/hanasakis/kotoha/internal/cart"
	"github.com/hanasakis/kotoha/internal/catalog"
	"github.com/hanasakis/kotoha/internal/order"
	"github.com/hanasakis/kotoha/internal/testutil"
	"github.com/hanasakis/kotoha/pkg/db"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCreateOrder(t *testing.T) {
	database := testutil.SetupTestDB(t)
	defer testutil.CleanTestDB(t, database)

	if err := db.AutoMigrate(database); err != nil {
		t.Fatalf("failed to migrate: %v", err)
	}

	catalogRepo := catalog.NewRepository(database)
	catalogSvc := catalog.NewService(catalogRepo)
	require.NoError(t, catalogSvc.SeedData())

	products, _, err := catalogRepo.ListProducts(1, 10)
	require.NoError(t, err)
	require.NotEmpty(t, products)

	redisClient := testutil.SetupTestRedis(t)
	defer redisClient.Close()

	cartRepo := cart.NewRepository(redisClient)
	cartSvc := cart.NewService(cartRepo, catalogRepo)
	ctx := context.Background()

	firstProduct := products[0]
	firstSKU := firstProduct.SKUs[0]
	err = cartSvc.AddItem(ctx, 1, firstSKU.ID, 2)
	require.NoError(t, err)

	t.Run("create_from_cart", func(t *testing.T) {
		orderRepo := order.NewRepository(database)
		orderSvc := order.NewService(orderRepo, cartRepo, catalogRepo)

		created, err := orderSvc.CreateOrder(ctx, 1, order.CreateOrderInput{AddressID: 1})
		require.NoError(t, err)
		assert.NotEmpty(t, created.OrderNo)
		assert.Len(t, created.OrderNo, 22) // KO + userID(1) + timestamp(14) + 4 random
		assert.Equal(t, order.StatusPending, created.Status)
		assert.Equal(t, "cny", created.Currency)
		assert.Len(t, created.Items, 1)
		assert.Equal(t, firstSKU.ID, created.Items[0].SKUID)
		assert.Equal(t, 2, created.Items[0].Quantity)
		assert.Equal(t, firstSKU.Price*2, created.TotalAmount)

		items, _ := cartSvc.GetCart(ctx, 1)
		assert.Empty(t, items)
	})

	t.Run("empty_cart_should_fail", func(t *testing.T) {
		orderRepo := order.NewRepository(database)
		orderSvc := order.NewService(orderRepo, cartRepo, catalogRepo)

		_, err := orderSvc.CreateOrder(ctx, 2, order.CreateOrderInput{AddressID: 1})
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "cart_empty")
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
		require.NoError(t, err)
		assert.Equal(t, int64(0), total)
		assert.Empty(t, orders)
	})
}

func TestCancelOrder(t *testing.T) {
	database := testutil.SetupTestDB(t)
	defer testutil.CleanTestDB(t, database)

	if err := db.AutoMigrate(database); err != nil {
		t.Fatalf("failed to migrate: %v", err)
	}

	catalogRepo := catalog.NewRepository(database)
	catalogSvc := catalog.NewService(catalogRepo)
	require.NoError(t, catalogSvc.SeedData())

	products, _, _ := catalogRepo.ListProducts(1, 1)
	require.NotEmpty(t, products)
	sku := products[0].SKUs[0]
	initialStock := sku.Stock

	redisClient := testutil.SetupTestRedis(t)
	defer redisClient.Close()

	cartRepo := cart.NewRepository(redisClient)
	orderRepo := order.NewRepository(database)
	orderSvc := order.NewService(orderRepo, cartRepo, catalogRepo)

	testOrder := &order.Order{
		UserID:      1,
		OrderNo:     "TEST-ORDER-001",
		Status:      order.StatusPending,
		TotalAmount: sku.Price,
		Currency:    "cny",
		Items:       []order.OrderItem{{ProductID: sku.ProductID, SKUID: sku.ID, Name: "test", Price: sku.Price, Quantity: 1}},
	}
	require.NoError(t, orderRepo.Create(testOrder))
	catalogRepo.DecrStock(sku.ID, 1)

	t.Run("cancel_pending", func(t *testing.T) {
		err := orderSvc.CancelOrder(testOrder.ID, 1)
		require.NoError(t, err)

		updated, err := orderRepo.GetByID(testOrder.ID, 1)
		require.NoError(t, err)
		assert.Equal(t, order.StatusCancelled, updated.Status)

		restoredSKU, _ := catalogRepo.GetSKU(sku.ID)
		assert.Equal(t, initialStock, restoredSKU.Stock)
	})

	t.Run("cancel_nonexistent", func(t *testing.T) {
		err := orderSvc.CancelOrder(99999, 1)
		assert.Error(t, err)
	})
}
