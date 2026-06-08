package order

import (
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/hanasakis/kotoha/internal/middleware"
	resp "github.com/hanasakis/kotoha/pkg/response"
)

type Handler struct {
	svc               *Service
	trackOrderCreated func()
}

func NewHandler(svc *Service, trackOrderCreated func()) *Handler {
	return &Handler{svc: svc, trackOrderCreated: trackOrderCreated}
}

// @Summary      Create an order from cart
// @Tags         orders
// @Accept       json
// @Produce      json
// @Param        body body CreateOrderInput true "Order details"
// @Security     BearerAuth
// @Success      201 {object} Order
// @Failure      400 {object} map[string]interface{}
// @Router       /orders [post]
func (h *Handler) CreateOrder(c *gin.Context) {
	userID := middleware.UserIDFromContext(c)
	var input CreateOrderInput
	if err := c.ShouldBindJSON(&input); err != nil {
		resp.BadRequest(c, "common.invalid_request")
		return
	}

	order, err := h.svc.CreateOrder(c.Request.Context(), userID, input)
	if err != nil {
		resp.BadRequest(c, err.Error())
		return
	}

	if h.trackOrderCreated != nil {
		h.trackOrderCreated()
	}

	resp.Created(c, order)
}

// @Summary      Get order by ID
// @Tags         orders
// @Produce      json
// @Param        id path int true "Order ID"
// @Security     BearerAuth
// @Success      200 {object} Order
// @Failure      404 {object} map[string]interface{}
// @Router       /orders/{id} [get]
func (h *Handler) GetOrder(c *gin.Context) {
	userID := middleware.UserIDFromContext(c)
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		resp.BadRequest(c, "common.invalid_request")
		return
	}

	order, err := h.svc.GetOrder(uint(id), userID)
	if err != nil {
		resp.NotFound(c, err.Error())
		return
	}

	resp.Success(c, order)
}

// @Summary      List user's orders
// @Tags         orders
// @Produce      json
// @Param        page      query int false "Page number" default(1)
// @Param        page_size query int false "Items per page" default(20)
// @Security     BearerAuth
// @Success      200 {object} map[string]interface{}
// @Router       /orders [get]
func (h *Handler) ListOrders(c *gin.Context) {
	userID := middleware.UserIDFromContext(c)
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))

	orders, total, err := h.svc.ListOrders(userID, page, pageSize)
	if err != nil {
		resp.InternalError(c)
		return
	}

	resp.Success(c, gin.H{
		"orders":    orders,
		"total":     total,
		"page":      page,
		"page_size": pageSize,
	})
}

// @Summary      Cancel an order
// @Tags         orders
// @Produce      json
// @Param        id path int true "Order ID"
// @Security     BearerAuth
// @Success      200 {object} map[string]interface{}
// @Failure      400 {object} map[string]interface{}
// @Router       /orders/{id}/cancel [post]
func (h *Handler) CancelOrder(c *gin.Context) {
	userID := middleware.UserIDFromContext(c)
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		resp.BadRequest(c, "common.invalid_request")
		return
	}

	if err := h.svc.CancelOrder(uint(id), userID); err != nil {
		resp.BadRequest(c, err.Error())
		return
	}

	resp.Success(c, gin.H{"message": "ok"})
}

// AdminListOrders returns all orders (admin only).
func (h *Handler) AdminListOrders(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))
	status := c.Query("status")

	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	orders, total, err := h.svc.ListAllOrders(page, pageSize, status)
	if err != nil {
		resp.InternalError(c)
		return
	}

	resp.Success(c, gin.H{
		"orders":    orders,
		"total":     total,
		"page":      page,
		"page_size": pageSize,
	})
}

// AdminGetOrder returns any order by ID (admin only).
func (h *Handler) AdminGetOrder(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		resp.BadRequest(c, "common.invalid_request")
		return
	}

	o, err := h.svc.GetOrderAdmin(uint(id))
	if err != nil {
		resp.NotFound(c, "order.not_found")
		return
	}

	resp.Success(c, o)
}
