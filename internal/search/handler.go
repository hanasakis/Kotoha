package search

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

func (h *Handler) Search(c *gin.Context) {
	query := c.Query("q")
	if query == "" {
		c.JSON(http.StatusBadRequest, gin.H{"code": "common.invalid_request"})
		return
	}

	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))
	if limit < 1 || limit > 50 {
		limit = 10
	}

	results, err := h.svc.Search(query, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": "search.error"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"query": query, "results": results})
}

func (h *Handler) IndexAll(c *gin.Context) {
	if err := h.svc.IndexAll(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "ok"})
}
