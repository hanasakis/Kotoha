package payment

import (
	"io"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

type Handler struct {
	svc *Service
}

func NewHandler(svc *Service) *Handler {
	return &Handler{svc: svc}
}

func userIDFromContext(c *gin.Context) uint {
	v, exists := c.Get("user_id")
	if !exists {
		return 0
	}
	switch id := v.(type) {
	case float64:
		return uint(id)
	case uint:
		return id
	case string:
		n, _ := strconv.ParseUint(id, 10, 64)
		return uint(n)
	default:
		return 0
	}
}

func (h *Handler) CreateCheckout(c *gin.Context) {
	userID := userIDFromContext(c)
	orderID, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": "common.invalid_request"})
		return
	}

	var input CheckoutInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": "common.invalid_request"})
		return
	}

	sessionID, err := h.svc.CreateCheckout(uint(orderID), userID, input)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"session_id": sessionID})
}

func (h *Handler) HandleWebhook(c *gin.Context) {
	payload, err := io.ReadAll(c.Request.Body)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": "common.invalid_request"})
		return
	}

	sigHeader := c.GetHeader("Stripe-Signature")
	if err := h.svc.HandleWebhook(sigHeader, payload); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "ok"})
}
