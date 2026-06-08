package stripe

import (
	"fmt"

	"github.com/stripe/stripe-go/v84"
	"github.com/stripe/stripe-go/v84/checkout/session"
	"github.com/stripe/stripe-go/v84/refund"
)

type Client struct {
	webhookSecret string
}

func New(secretKey, webhookSecret string) *Client {
	stripe.Key = secretKey
	return &Client{webhookSecret: webhookSecret}
}

func (c *Client) WebhookSecret() string {
	return c.webhookSecret
}

type CheckoutParams struct {
	OrderNo        string
	Amount         int64
	Currency       string
	SuccessURL     string
	CancelURL      string
	IdempotencyKey string
}

func (c *Client) CreateCheckoutSession(p CheckoutParams) (sessionID string, checkoutURL string, err error) {
	params := &stripe.CheckoutSessionParams{
		Mode:              stripe.String(string(stripe.CheckoutSessionModePayment)),
		SuccessURL:        stripe.String(p.SuccessURL),
		CancelURL:         stripe.String(p.CancelURL),
		ClientReferenceID: stripe.String(p.OrderNo),
		LineItems: []*stripe.CheckoutSessionLineItemParams{
			{
				PriceData: &stripe.CheckoutSessionLineItemPriceDataParams{
					Currency:   stripe.String(p.Currency),
					UnitAmount: stripe.Int64(p.Amount),
					ProductData: &stripe.CheckoutSessionLineItemPriceDataProductDataParams{
						Name: stripe.String("Kotoha Order " + p.OrderNo),
					},
				},
				Quantity: stripe.Int64(1),
			},
		},
	}

	if p.IdempotencyKey != "" {
		params.Params.IdempotencyKey = stripe.String(p.IdempotencyKey)
	}

	s, err := session.New(params)
	if err != nil {
		return "", "", err
	}
	return s.ID, s.URL, nil
}

// GetCheckoutSession retrieves a Stripe checkout session and returns status, payment intent, and URL.
func (c *Client) GetCheckoutSession(sessionID string) (status, paymentIntent, url string, err error) {
	params := &stripe.CheckoutSessionParams{}
	s, err := session.Get(sessionID, params)
	if err != nil {
		return "", "", "", err
	}
	if s.PaymentIntent != nil {
		paymentIntent = s.PaymentIntent.ID
	}
	return string(s.Status), paymentIntent, s.URL, nil
}

type RefundParams struct {
	PaymentIntentID string
	Amount          int64 // 0 = full refund
}

func (c *Client) CreateRefund(p RefundParams) (string, int64, error) {
	params := &stripe.RefundParams{
		PaymentIntent: stripe.String(p.PaymentIntentID),
	}
	if p.Amount > 0 {
		params.Amount = stripe.Int64(p.Amount)
	}
	r, err := refund.New(params)
	if err != nil {
		return "", 0, fmt.Errorf("stripe.refund_error: %w", err)
	}
	return r.ID, r.Amount, nil
}
