package stripe

import (
	"github.com/stripe/stripe-go/v84"
	"github.com/stripe/stripe-go/v84/checkout/session"
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
	OrderNo    string
	Amount     int64
	Currency   string
	SuccessURL string
	CancelURL  string
}

func (c *Client) CreateCheckoutSession(p CheckoutParams) (string, error) {
	params := &stripe.CheckoutSessionParams{
		Mode:          stripe.String(string(stripe.CheckoutSessionModePayment)),
		SuccessURL:    stripe.String(p.SuccessURL),
		CancelURL:     stripe.String(p.CancelURL),
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

	s, err := session.New(params)
	if err != nil {
		return "", err
	}
	return s.ID, nil
}
