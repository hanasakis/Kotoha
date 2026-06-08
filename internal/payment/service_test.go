package payment_test

import (
	"context"
	"os"
	"strings"
	"testing"

	"github.com/hanasakis/kotoha/internal/cart"
	"github.com/hanasakis/kotoha/internal/catalog"
	"github.com/hanasakis/kotoha/internal/order"
	"github.com/hanasakis/kotoha/internal/payment"
	"github.com/hanasakis/kotoha/internal/testutil"
	"github.com/hanasakis/kotoha/pkg/db"
	stripepkg "github.com/hanasakis/kotoha/pkg/stripe"
)

func hasStripeKey() bool {
	return os.Getenv("STRIPE_SECRET_KEY") != "" && os.Getenv("STRIPE_WEBHOOK_SECRET") != ""
}

func setupPaymentService(t *testing.T) (*payment.Service, *order.Repository, *catalog.Repository, func()) {
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
	orderRepo := order.NewRepository(database)
	orderSvc := order.NewService(orderRepo, cartRepo, catalogRepo, nil)

	stripeCli := stripepkg.New(
		os.Getenv("STRIPE_SECRET_KEY"),
		os.Getenv("STRIPE_WEBHOOK_SECRET"),
	)

	paymentSvc := payment.NewService(orderRepo, stripeCli)

	cleanup := func() {
		redisClient.RDB.FlushDB(context.Background())
		testutil.CleanTestDB(t, database)
		redisClient.Close()
	}

	_ = cartSvc
	_ = orderSvc

	return paymentSvc, orderRepo, catalogRepo, cleanup
}

func TestCreateCheckout(t *testing.T) {
	if !hasStripeKey() {
		t.Skip("STRIPE_SECRET_KEY and STRIPE_WEBHOOK_SECRET required")
	}

	svc, orderRepo, catalogRepo, cleanup := setupPaymentService(t)
	defer cleanup()
	ctx := context.Background()

	products, _, _ := catalogRepo.ListProducts(1, 1, "", 0)
	if len(products) == 0 || len(products[0].SKUs) == 0 {
		t.Fatal("expected at least one product with SKU")
	}
	sku := products[0].SKUs[0]

	// Create a pending order via the order service flow
	testOrder := &order.Order{
		UserID:      1,
		OrderNo:     "TEST-PAY-001",
		Status:      order.StatusPending,
		TotalAmount: sku.Price * 2,
		Currency:    "cny",
		Items: []order.OrderItem{
			{ProductID: sku.ProductID, SKUID: sku.ID, Name: products[0].Name, Price: sku.Price, Quantity: 2},
		},
	}
	if err := orderRepo.Create(testOrder); err != nil {
		t.Fatalf("failed to create test order: %v", err)
	}

	t.Run("create_checkout_success", func(t *testing.T) {
		sessionID, err := svc.CreateCheckout(testOrder.ID, 1, payment.CheckoutInput{
			SuccessURL: "https://example.com/success",
			CancelURL:  "https://example.com/cancel",
		})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if sessionID == "" {
			t.Error("expected non-empty session ID")
		}

		// Verify order was updated with Stripe session ID
		updated, err := orderRepo.GetByID(testOrder.ID, 1)
		if err != nil {
			t.Fatalf("failed to get updated order: %v", err)
		}
		if updated.StripeSID != sessionID {
			t.Errorf("order stripe_session_id %q, want %q", updated.StripeSID, sessionID)
		}
	})

	t.Run("order_not_found", func(t *testing.T) {
		_, err := svc.CreateCheckout(99999, 1, payment.CheckoutInput{
			SuccessURL: "https://example.com/success",
			CancelURL:  "https://example.com/cancel",
		})
		if err == nil {
			t.Fatal("expected error, got nil")
		}
		if !strings.Contains(err.Error(), "not_found") {
			t.Errorf("expected 'not_found' error, got %q", err.Error())
		}
	})

	t.Run("order_already_paid", func(t *testing.T) {
		paidOrder := &order.Order{
			UserID:      1,
			OrderNo:     "TEST-PAY-002",
			Status:      order.StatusPaid,
			TotalAmount: sku.Price,
			Currency:    "cny",
			StripeSID:   "cs_existing",
		}
		if err := orderRepo.Create(paidOrder); err != nil {
			t.Fatalf("failed to create paid order: %v", err)
		}

		_, err := svc.CreateCheckout(paidOrder.ID, 1, payment.CheckoutInput{
			SuccessURL: "https://example.com/success",
			CancelURL:  "https://example.com/cancel",
		})
		if err == nil {
			t.Fatal("expected error, got nil")
		}
		if !strings.Contains(err.Error(), "not_pending") {
			t.Errorf("expected 'not_pending' error, got %q", err.Error())
		}
	})

	// Clean up stripe session from test order
	_ = ctx
}

func TestHandleWebhook_InvalidSignature(t *testing.T) {
	if !hasStripeKey() {
		t.Skip("STRIPE_WEBHOOK_SECRET required")
	}

	svc, _, _, cleanup := setupPaymentService(t)
	defer cleanup()

	err := svc.HandleWebhook("invalid_signature", []byte(`{"type":"checkout.session.completed"}`))
	if err == nil {
		t.Fatal("expected error for invalid signature, got nil")
	}
	if !strings.Contains(err.Error(), "webhook_error") {
		t.Errorf("expected 'webhook_error', got %q", err.Error())
	}
}

func TestHandleWebhook_NonCheckoutEvent(t *testing.T) {
	t.Skip("requires real Stripe webhook signature generation")
}

func TestHandleWebhook_CheckoutCompleted(t *testing.T) {
	t.Skip("requires real Stripe webhook signature generation")
}

func TestPaymentService_New(t *testing.T) {
	svc := payment.NewService(nil, nil)
	if svc == nil {
		t.Fatal("expected non-nil service")
	}
}
