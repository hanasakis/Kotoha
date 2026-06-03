package cart

import (
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

func (h *Handler) GetCart(c *gin.Context) {
	userID := userIDFromContext(c)
	items, err := h.svc.GetCart(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": "common.server_error"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"items": items})
}

func (h *Handler) AddItem(c *gin.Context) {
	userID := userIDFromContext(c)
	var req AddItemReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": "common.invalid_request"})
		return
	}

	if err := h.svc.AddItem(c.Request.Context(), userID, req.SKUID, req.Quantity); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": err.Error()})
		return
	}

	items, _ := h.svc.GetCart(c.Request.Context(), userID)
	c.JSON(http.StatusOK, gin.H{"items": items})
}

func (h *Handler) UpdateQty(c *gin.Context) {
	userID := userIDFromContext(c)
	skuID, err := strconv.ParseUint(c.Param("skuID"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": "common.invalid_request"})
		return
	}

	var req UpdateQtyReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": "common.invalid_request"})
		return
	}

	if err := h.svc.UpdateQty(c.Request.Context(), userID, uint(skuID), req.Quantity); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": err.Error()})
		return
	}

	items, _ := h.svc.GetCart(c.Request.Context(), userID)
	c.JSON(http.StatusOK, gin.H{"items": items})
}

func (h *Handler) RemoveItem(c *gin.Context) {
	userID := userIDFromContext(c)
	skuID, err := strconv.ParseUint(c.Param("skuID"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": "common.invalid_request"})
		return
	}

	if err := h.svc.RemoveItem(c.Request.Context(), userID, uint(skuID)); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": "common.server_error"})
		return
	}

	items, _ := h.svc.GetCart(c.Request.Context(), userID)
	c.JSON(http.StatusOK, gin.H{"items": items})
}
