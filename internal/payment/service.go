package payment

import (
	"encoding/json"
	"fmt"
	"time"

	stripepkg "github.com/hanasakis/kotoha/pkg/stripe"

	"github.com/hanasakis/kotoha/internal/order"
	"github.com/stripe/stripe-go/v84"
	"github.com/stripe/stripe-go/v84/webhook"
)

const paymentWindow = 10 * time.Minute

type Service struct {
	orderRepo *order.Repository
	stripeCli *stripepkg.Client
}

func NewService(orderRepo *order.Repository, stripeCli *stripepkg.Client) *Service {
	return &Service{orderRepo: orderRepo, stripeCli: stripeCli}
}

type CheckoutInput struct {
	SuccessURL     string `json:"success_url" binding:"required"`
	CancelURL      string `json:"cancel_url" binding:"required"`
	IdempotencyKey string `json:"idempotency_key"`
}

// SyncPayment checks the Stripe checkout session status and updates the order accordingly.
// This is a fallback for when webhooks are not delivered (e.g., local dev without Stripe CLI).
func (s *Service) SyncPayment(orderID, userID uint) (string, error) {
	o, err := s.orderRepo.GetByID(orderID, userID)
	if err != nil {
		return "", fmt.Errorf("order.not_found")
	}
	return s.syncOrderWithStripe(o)
}

// SyncPaymentByOrderNo is a public variant that only requires the order number (acts as a secret).
// This allows payment syncing after Stripe redirect even when the JWT has expired.
func (s *Service) SyncPaymentByOrderNo(orderNo string) (string, error) {
	o, err := s.orderRepo.FindByOrderNo(orderNo)
	if err != nil {
		return "", fmt.Errorf("order.not_found")
	}
	return s.syncOrderWithStripe(o)
}

func (s *Service) syncOrderWithStripe(o *order.Order) (string, error) {
	if o.Status != order.StatusPending {
		return o.Status, nil
	}
	if o.StripeSID == "" {
		return o.Status, nil
	}

	sessionStatus, paymentIntent, _, err := s.stripeCli.GetCheckoutSession(o.StripeSID)
	if err != nil {
		return o.Status, fmt.Errorf("payment.stripe_error: %w", err)
	}

	switch sessionStatus {
	case "complete":
		if err := s.orderRepo.SetPaid(o.ID); err != nil {
			return o.Status, err
		}
		if paymentIntent != "" {
			s.orderRepo.CreatePayment(&order.Payment{
				OrderID:        o.ID,
				StripePID:      paymentIntent,
				IdempotencyKey: paymentIntent,
				Amount:         o.TotalAmount,
				Currency:       o.Currency,
				Status:         "paid",
			})
		}
		return order.StatusPaid, nil
	case "expired":
		s.orderRepo.SetExpired(o.ID)
		return order.StatusExpired, nil
	default:
		return o.Status, nil
	}
}

func (s *Service) CreateCheckout(orderID, userID uint, input CheckoutInput) (string, error) {
	o, err := s.orderRepo.GetByID(orderID, userID)
	if err != nil {
		return "", fmt.Errorf("order.not_found")
	}

	// Prevent duplicate payments: if already paid, refunded, or cancelled
	switch o.Status {
	case order.StatusPaid, order.StatusRefunded, order.StatusShipped, order.StatusDelivered:
		return "", fmt.Errorf("payment.order_already_paid")
	case order.StatusCancelled, order.StatusExpired:
		return "", fmt.Errorf("payment.order_cancelled_or_expired")
	}

	if time.Since(o.CreatedAt) > paymentWindow {
		s.orderRepo.SetExpired(o.ID)
		return "", fmt.Errorf("payment.expired")
	}

	// If a Stripe session already exists, don't allow creating another
	if o.StripeSID != "" {
		status, _, _, _ := s.stripeCli.GetCheckoutSession(o.StripeSID)
		if status == "open" || status == "complete" {
			return "", fmt.Errorf("payment.session_exists")
		}
	}

	sessionID, checkoutURL, err := s.stripeCli.CreateCheckoutSession(stripepkg.CheckoutParams{
		OrderNo:        o.OrderNo,
		Amount:         int64(o.TotalAmount),
		Currency:       o.Currency,
		SuccessURL:     input.SuccessURL,
		CancelURL:      input.CancelURL,
		IdempotencyKey: input.IdempotencyKey,
	})
	if err != nil {
		return "", fmt.Errorf("payment.stripe_error")
	}

	if err := s.orderRepo.UpdateStripeSession(o.ID, sessionID); err != nil {
		return "", err
	}

	return checkoutURL, nil
}

func (s *Service) HandleWebhook(sigHeader string, payload []byte) error {
	event, err := webhook.ConstructEvent(payload, sigHeader, s.stripeCli.WebhookSecret())
	if err != nil {
		return fmt.Errorf("payment.webhook_error")
	}

	switch event.Type {
	case "checkout.session.completed":
		return s.handleCheckoutCompleted(event)
	case "checkout.session.async_payment_succeeded":
		return s.handleCheckoutCompleted(event) // same logic
	case "checkout.session.expired":
		return s.handleCheckoutExpired(event)
	case "checkout.session.async_payment_failed":
		return s.handleCheckoutFailed(event)
	case "payment_intent.payment_failed":
		return s.handlePaymentIntentFailed(event)
	case "charge.refunded":
		return s.handleChargeRefunded(event)
	}
	return nil
}

func (s *Service) handleCheckoutCompleted(event stripe.Event) error {
	var session struct {
		ID                string `json:"id"`
		ClientReferenceID string `json:"client_reference_id"`
		PaymentIntent     string `json:"payment_intent"`
	}

	data, err := json.Marshal(event.Data.Object)
	if err != nil {
		return fmt.Errorf("payment.webhook_error: marshal event object: %w", err)
	}
	if err := json.Unmarshal(data, &session); err != nil {
		return fmt.Errorf("payment.webhook_error: unmarshal session: %w", err)
	}

	o, err := s.orderRepo.FindByOrderNo(session.ClientReferenceID)
	if err != nil {
		return fmt.Errorf("payment.order_not_found: %w", err)
	}

	if err := s.orderRepo.SetPaid(o.ID); err != nil {
		return err
	}

	if err := s.orderRepo.CreatePayment(&order.Payment{
		OrderID:        o.ID,
		StripePID:      session.PaymentIntent,
		IdempotencyKey: session.PaymentIntent,
		Amount:         o.TotalAmount,
		Currency:       o.Currency,
		Status:         "paid",
	}); err != nil {
		return fmt.Errorf("payment.create_failed: %w", err)
	}
	return nil
}

func (s *Service) handleCheckoutExpired(event stripe.Event) error {
	var session struct {
		ClientReferenceID string `json:"client_reference_id"`
	}
	data, err := json.Marshal(event.Data.Object)
	if err != nil {
		return fmt.Errorf("payment.webhook_error: %w", err)
	}
	if err := json.Unmarshal(data, &session); err != nil {
		return fmt.Errorf("payment.webhook_error: %w", err)
	}
	o, err := s.orderRepo.FindByOrderNo(session.ClientReferenceID)
	if err != nil {
		return fmt.Errorf("payment.order_not_found: %w", err)
	}
	return s.orderRepo.SetExpired(o.ID)
}

func (s *Service) handleCheckoutFailed(event stripe.Event) error {
	var session struct {
		ClientReferenceID string `json:"client_reference_id"`
	}
	data, err := json.Marshal(event.Data.Object)
	if err != nil {
		return fmt.Errorf("payment.webhook_error: %w", err)
	}
	if err := json.Unmarshal(data, &session); err != nil {
		return fmt.Errorf("payment.webhook_error: %w", err)
	}
	o, err := s.orderRepo.FindByOrderNo(session.ClientReferenceID)
	if err != nil {
		return fmt.Errorf("payment.order_not_found: %w", err)
	}
	return s.orderRepo.SetPaymentFailed(o.ID)
}

func (s *Service) handlePaymentIntentFailed(event stripe.Event) error {
	var pi struct {
		ID     string `json:"id"`
		Amount int64  `json:"amount"`
	}
	data, err := json.Marshal(event.Data.Object)
	if err != nil {
		return fmt.Errorf("payment.webhook_error: %w", err)
	}
	if err := json.Unmarshal(data, &pi); err != nil {
		return fmt.Errorf("payment.webhook_error: %w", err)
	}
	// Record failed payment attempt
	payment := &order.Payment{
		StripePID: pi.ID,
		Amount:    int(pi.Amount),
		Currency:  "cny",
		Status:    "failed",
	}
	return s.orderRepo.CreatePayment(payment)
}

func (s *Service) handleChargeRefunded(event stripe.Event) error {
	var charge struct {
		ID             string `json:"id"`
		PaymentIntent  string `json:"payment_intent"`
		AmountRefunded int64  `json:"amount_refunded"`
		Refunds        struct {
			Data []struct {
				ID string `json:"id"`
			} `json:"data"`
		} `json:"refunds"`
	}

	data, err := json.Marshal(event.Data.Object)
	if err != nil {
		return fmt.Errorf("payment.webhook_error: %w", err)
	}
	if err := json.Unmarshal(data, &charge); err != nil {
		return fmt.Errorf("payment.webhook_error: %w", err)
	}

	payment, err := s.orderRepo.FindPaymentByStripePID(charge.PaymentIntent)
	if err != nil {
		return fmt.Errorf("payment.payment_not_found: %w", err)
	}

	refundID := ""
	if len(charge.Refunds.Data) > 0 {
		refundID = charge.Refunds.Data[0].ID
	}

	if err := s.orderRepo.SetRefunded(payment.OrderID, refundID, int(charge.AmountRefunded)); err != nil {
		return err
	}

	newStatus := order.StatusRefunded
	if int64(payment.Amount) > charge.AmountRefunded {
		newStatus = order.StatusPartialRefunded
	}
	return s.orderRepo.UpdateStatus(payment.OrderID, newStatus)
}
