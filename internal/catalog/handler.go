package catalog

import (
	"strconv"

	"github.com/gin-gonic/gin"
	resp "github.com/hanasakis/kotoha/pkg/response"
)

type Handler struct {
	svc *Service
}

func NewHandler(svc *Service) *Handler {
	return &Handler{svc: svc}
}

// @Summary      List all categories
// @Tags         catalog
// @Produce      json
// @Success      200 {array} Category
// @Router       /catalog/categories [get]
func (h *Handler) ListCategories(c *gin.Context) {
	cats, err := h.svc.ListCategories()
	if err != nil {
		resp.InternalError(c)
		return
	}
	resp.Success(c, cats)
}

// @Summary      List products with pagination
// @Tags         catalog
// @Produce      json
// @Param        page      query int    false "Page number" default(1)
// @Param        page_size query int    false "Items per page" default(20)
// @Param        keyword   query string false "Keyword search (name, tags, description)"
// @Success      200 {object} map[string]interface{}
// @Router       /catalog/products [get]
func (h *Handler) ListProducts(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))
	keyword := c.Query("keyword")
	categoryID, _ := strconv.ParseUint(c.Query("category_id"), 10, 64)

	products, total, err := h.svc.ListProducts(page, pageSize, keyword, uint(categoryID))
	if err != nil {
		resp.InternalError(c)
		return
	}

	resp.Success(c, gin.H{
		"products":  products,
		"total":     total,
		"page":      page,
		"page_size": pageSize,
	})
}

// @Summary      Get product by ID
// @Tags         catalog
// @Produce      json
// @Param        id path int true "Product ID"
// @Success      200 {object} Product
// @Failure      404 {object} map[string]interface{}
// @Router       /catalog/products/{id} [get]
func (h *Handler) GetProduct(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		resp.BadRequest(c, "common.invalid_request")
		return
	}

	p, err := h.svc.GetProduct(uint(id))
	if err != nil {
		resp.NotFound(c, err.Error())
		return
	}

	resp.Success(c, p)
}

// @Summary      Create a product (admin)
// @Tags         admin
// @Accept       json
// @Produce      json
// @Param        body body Product true "Product data"
// @Security     BearerAuth
// @Success      201 {object} Product
// @Failure      400 {object} map[string]interface{}
// @Router       /admin/products [post]
func (h *Handler) CreateProduct(c *gin.Context) {
	var p Product
	if err := c.ShouldBindJSON(&p); err != nil {
		resp.BadRequest(c, "common.invalid_request")
		return
	}

	if err := h.svc.CreateProduct(&p); err != nil {
		resp.InternalError(c)
		return
	}

	resp.Created(c, p)
}

// @Summary      Update a product (admin)
// @Tags         admin
// @Accept       json
// @Produce      json
// @Param        id   path int     true "Product ID"
// @Param        body body Product true "Updated product data"
// @Security     BearerAuth
// @Success      200 {object} Product
// @Failure      400 {object} map[string]interface{}
// @Router       /admin/products/{id} [put]
func (h *Handler) UpdateProduct(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		resp.BadRequest(c, "common.invalid_request")
		return
	}

	var p Product
	if err := c.ShouldBindJSON(&p); err != nil {
		resp.BadRequest(c, "common.invalid_request")
		return
	}
	p.ID = uint(id)

	if err := h.svc.UpdateProduct(&p); err != nil {
		resp.InternalError(c)
		return
	}

	resp.Success(c, p)
}

// @Summary      Delete a product (admin)
// @Tags         admin
// @Produce      json
// @Param        id path int true "Product ID"
// @Security     BearerAuth
// @Success      200 {object} map[string]interface{}
// @Router       /admin/products/{id} [delete]
func (h *Handler) DeleteProduct(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		resp.BadRequest(c, "common.invalid_request")
		return
	}

	if err := h.svc.DeleteProduct(uint(id)); err != nil {
		resp.InternalError(c)
		return
	}

	resp.Success(c, gin.H{"message": "ok"})
}
