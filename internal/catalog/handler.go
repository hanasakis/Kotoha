package catalog

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

func (h *Handler) ListCategories(c *gin.Context) {
	cats, err := h.svc.ListCategories()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": "common.server_error"})
		return
	}
	c.JSON(http.StatusOK, cats)
}

func (h *Handler) ListProducts(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))

	products, total, err := h.svc.ListProducts(page, pageSize)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": "common.server_error"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data":      products,
		"total":     total,
		"page":      page,
		"page_size": pageSize,
	})
}

func (h *Handler) GetProduct(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": "common.invalid_request"})
		return
	}

	p, err := h.svc.GetProduct(uint(id))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"code": err.Error()})
		return
	}

	c.JSON(http.StatusOK, p)
}

func (h *Handler) CreateProduct(c *gin.Context) {
	var p Product
	if err := c.ShouldBindJSON(&p); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": "common.invalid_request"})
		return
	}

	if err := h.svc.CreateProduct(&p); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": "common.server_error"})
		return
	}

	c.JSON(http.StatusCreated, p)
}

func (h *Handler) UpdateProduct(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": "common.invalid_request"})
		return
	}

	var p Product
	if err := c.ShouldBindJSON(&p); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": "common.invalid_request"})
		return
	}
	p.ID = uint(id)

	if err := h.svc.UpdateProduct(&p); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": "common.server_error"})
		return
	}

	c.JSON(http.StatusOK, p)
}

func (h *Handler) DeleteProduct(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": "common.invalid_request"})
		return
	}

	if err := h.svc.DeleteProduct(uint(id)); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": "common.server_error"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "ok"})
}
