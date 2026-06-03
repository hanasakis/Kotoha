package payment

import (
	"encoding/json"
	"fmt"

	stripepkg "github.com/hanasakis/kotoha/pkg/stripe"

	"github.com/hanasakis/kotoha/internal/order"
	"github.com/stripe/stripe-go/v84/webhook"
)

type Service struct {
	orderRepo *order.Repository
	stripeCli *stripepkg.Client
}

func NewService(orderRepo *order.Repository, stripeCli *stripepkg.Client) *Service {
	return &Service{orderRepo: orderRepo, stripeCli: stripeCli}
}

type CheckoutInput struct {
	SuccessURL string `json:"success_url" binding:"required"`
	CancelURL  string `json:"cancel_url" binding:"required"`
}

func (s *Service) CreateCheckout(orderID, userID uint, input CheckoutInput) (string, error) {
	o, err := s.orderRepo.GetByID(orderID, userID)
	if err != nil {
		return "", fmt.Errorf("order.not_found")
	}
	if o.Status != order.StatusPending {
		return "", fmt.Errorf("payment.order_not_pending")
	}

	sessionID, err := s.stripeCli.CreateCheckoutSession(stripepkg.CheckoutParams{
		OrderNo:    o.OrderNo,
		Amount:     int64(o.TotalAmount),
		Currency:   o.Currency,
		SuccessURL: input.SuccessURL,
		CancelURL:  input.CancelURL,
	})
	if err != nil {
		return "", fmt.Errorf("payment.stripe_error")
	}

	if err := s.orderRepo.UpdateStripeSession(o.ID, sessionID); err != nil {
		return "", err
	}

	return sessionID, nil
}

func (s *Service) HandleWebhook(sigHeader string, payload []byte) error {
	event, err := webhook.ConstructEvent(payload, sigHeader, s.stripeCli.WebhookSecret())
	if err != nil {
		return fmt.Errorf("payment.webhook_error")
	}

	if event.Type != "checkout.session.completed" {
		return nil
	}

	var session struct {
		ID                string `json:"id"`
		ClientReferenceID string `json:"client_reference_id"`
		PaymentIntent     string `json:"payment_intent"`
	}

	data, _ := json.Marshal(event.Data.Object)
	if err := json.Unmarshal(data, &session); err != nil {
		return nil
	}

	o, err := s.orderRepo.FindByOrderNo(session.ClientReferenceID)
	if err != nil {
		return nil
	}

	if err := s.orderRepo.SetPaid(o.ID); err != nil {
		return err
	}

	s.orderRepo.CreatePayment(&order.Payment{
		OrderID:   o.ID,
		StripePID: session.PaymentIntent,
		Amount:    o.TotalAmount,
		Currency:  o.Currency,
		Status:    "paid",
	})
	return nil
}
