package cart

import (
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/hanasakis/kotoha/internal/middleware"
	resp "github.com/hanasakis/kotoha/pkg/response"
)

type Handler struct {
	svc          *Service
	trackCartAdd func()
}

func NewHandler(svc *Service, trackCartAdd func()) *Handler {
	return &Handler{svc: svc, trackCartAdd: trackCartAdd}
}

// @Summary      Get current user's cart
// @Tags         cart
// @Produce      json
// @Security     BearerAuth
// @Success      200 {object} map[string]interface{}
// @Router       /cart [get]
func (h *Handler) GetCart(c *gin.Context) {
	userID := middleware.UserIDFromContext(c)
	items, err := h.svc.GetCart(c.Request.Context(), userID)
	if err != nil {
		resp.InternalError(c)
		return
	}
	resp.Success(c, gin.H{"items": items})
}

// @Summary      Add item to cart
// @Tags         cart
// @Accept       json
// @Produce      json
// @Param        body body AddItemReq true "SKU ID and quantity"
// @Security     BearerAuth
// @Success      200 {object} map[string]interface{}
// @Failure      400 {object} map[string]interface{}
// @Router       /cart/items [post]
func (h *Handler) AddItem(c *gin.Context) {
	userID := middleware.UserIDFromContext(c)
	var req AddItemReq
	if err := c.ShouldBindJSON(&req); err != nil {
		resp.BadRequest(c, "common.invalid_request")
		return
	}

	if err := h.svc.AddItem(c.Request.Context(), userID, req.SKUID, req.Quantity); err != nil {
		resp.BadRequest(c, err.Error())
		return
	}

	if h.trackCartAdd != nil {
		h.trackCartAdd()
	}

	items, _ := h.svc.GetCart(c.Request.Context(), userID)
	resp.Success(c, gin.H{"items": items})
}

// @Summary      Update cart item quantity
// @Tags         cart
// @Accept       json
// @Produce      json
// @Param        skuID path int true "SKU ID"
// @Param        body body UpdateQtyReq true "New quantity"
// @Security     BearerAuth
// @Success      200 {object} map[string]interface{}
// @Failure      400 {object} map[string]interface{}
// @Router       /cart/items/{skuID} [put]
func (h *Handler) UpdateQty(c *gin.Context) {
	userID := middleware.UserIDFromContext(c)
	skuID, err := strconv.ParseUint(c.Param("skuID"), 10, 64)
	if err != nil {
		resp.BadRequest(c, "common.invalid_request")
		return
	}

	var req UpdateQtyReq
	if err := c.ShouldBindJSON(&req); err != nil {
		resp.BadRequest(c, "common.invalid_request")
		return
	}

	if err := h.svc.UpdateQty(c.Request.Context(), userID, uint(skuID), req.Quantity); err != nil {
		resp.BadRequest(c, err.Error())
		return
	}

	items, _ := h.svc.GetCart(c.Request.Context(), userID)
	resp.Success(c, gin.H{"items": items})
}

// @Summary      Remove item from cart
// @Tags         cart
// @Produce      json
// @Param        skuID path int true "SKU ID"
// @Security     BearerAuth
// @Success      200 {object} map[string]interface{}
// @Router       /cart/items/{skuID} [delete]
func (h *Handler) RemoveItem(c *gin.Context) {
	userID := middleware.UserIDFromContext(c)
	skuID, err := strconv.ParseUint(c.Param("skuID"), 10, 64)
	if err != nil {
		resp.BadRequest(c, "common.invalid_request")
		return
	}

	if err := h.svc.RemoveItem(c.Request.Context(), userID, uint(skuID)); err != nil {
		resp.InternalError(c)
		return
	}

	items, _ := h.svc.GetCart(c.Request.Context(), userID)
	resp.Success(c, gin.H{"items": items})
}
