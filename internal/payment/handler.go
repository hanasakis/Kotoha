package payment

import (
	"io"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/hanasakis/kotoha/internal/middleware"
	resp "github.com/hanasakis/kotoha/pkg/response"
)

type Handler struct {
	svc *Service
}

func NewHandler(svc *Service) *Handler {
	return &Handler{svc: svc}
}

// @Summary      Create Stripe checkout session
// @Tags         orders
// @Accept       json
// @Produce      json
// @Param        id   path int           true "Order ID"
// @Param        body body CheckoutInput true "Checkout options"
// @Security     BearerAuth
// @Success      200 {object} map[string]interface{}
// @Failure      400 {object} map[string]interface{}
// @Router       /orders/{id}/checkout [post]
func (h *Handler) CreateCheckout(c *gin.Context) {
	userID := middleware.UserIDFromContext(c)
	orderID, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		resp.BadRequest(c, "common.invalid_request")
		return
	}

	var input CheckoutInput
	if err := c.ShouldBindJSON(&input); err != nil {
		resp.BadRequest(c, "common.invalid_request")
		return
	}

	sessionID, err := h.svc.CreateCheckout(uint(orderID), userID, input)
	if err != nil {
		resp.BadRequest(c, err.Error())
		return
	}

	resp.Success(c, gin.H{"checkout_url": sessionID})
}

// @Summary      Sync order payment status from Stripe
// @Tags         orders
// @Produce      json
// @Param        id path int true "Order ID"
// @Security     BearerAuth
// @Success      200 {object} map[string]interface{}
// @Router       /orders/{id}/sync-payment [post]
func (h *Handler) SyncPayment(c *gin.Context) {
	userID := middleware.UserIDFromContext(c)
	orderID, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		resp.BadRequest(c, "common.invalid_request")
		return
	}
	status, err := h.svc.SyncPayment(uint(orderID), userID)
	if err != nil {
		resp.BadRequest(c, err.Error())
		return
	}
	resp.Success(c, gin.H{"status": status})
}

// SyncPaymentPublicRequest is the public (no-auth) sync payload.
type SyncPaymentPublicRequest struct {
	OrderNo string `json:"order_no" binding:"required"`
}

// @Summary      Sync order payment from Stripe (public, no auth required)
// @Tags         payment
// @Accept       json
// @Produce      json
// @Param        body body SyncPaymentPublicRequest true "Order number"
// @Success      200 {object} map[string]interface{}
// @Router       /public/sync-payment [post]
func (h *Handler) SyncPaymentPublic(c *gin.Context) {
	var req SyncPaymentPublicRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		resp.BadRequest(c, "common.invalid_request")
		return
	}
	status, err := h.svc.SyncPaymentByOrderNo(req.OrderNo)
	if err != nil {
		resp.BadRequest(c, err.Error())
		return
	}
	resp.Success(c, gin.H{"status": status})
}

// @Summary      Handle Stripe webhook events
// @Tags         payment
// @Accept       json
// @Produce      json
// @Param        Stripe-Signature header string true "Stripe webhook signature"
// @Success      200 {object} map[string]interface{}
// @Failure      400 {object} map[string]interface{}
// @Router       /webhook [post]
func (h *Handler) HandleWebhook(c *gin.Context) {
	payload, err := io.ReadAll(c.Request.Body)
	if err != nil {
		resp.BadRequest(c, "common.invalid_request")
		return
	}

	sigHeader := c.GetHeader("Stripe-Signature")
	if err := h.svc.HandleWebhook(sigHeader, payload); err != nil {
		resp.BadRequest(c, err.Error())
		return
	}

	resp.Success(c, gin.H{"message": "ok"})
}
