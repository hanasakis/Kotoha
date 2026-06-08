package search

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	resp "github.com/hanasakis/kotoha/pkg/response"
)

type Handler struct {
	svc         *Service
	trackSearch func(float64)
}

func NewHandler(svc *Service, trackSearch func(float64)) *Handler {
	return &Handler{svc: svc, trackSearch: trackSearch}
}

// @Summary      Search products (hybrid: dense + BM25)
// @Tags         catalog
// @Produce      json
// @Param        q     query string true  "Search query"
// @Param        limit query int    false "Max results" default(10)
// @Success      200 {object} map[string]interface{}
// @Router       /catalog/search [get]
func (h *Handler) Search(c *gin.Context) {
	query := c.Query("q")
	if query == "" {
		resp.BadRequest(c, "common.invalid_request")
		return
	}

	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))
	if limit < 1 || limit > 50 {
		limit = 10
	}

	results, err := h.svc.Search(query, limit)
	if err != nil {
		resp.Error(c, http.StatusInternalServerError, "search.error")
		return
	}

	if h.trackSearch != nil {
		h.trackSearch(0)
	}

	resp.Success(c, gin.H{"query": query, "results": results})
}

// @Summary      Re-index all products into Milvus (admin)
// @Tags         admin
// @Produce      json
// @Security     BearerAuth
// @Success      200 {object} map[string]interface{}
// @Router       /admin/search/index [post]
func (h *Handler) IndexAll(c *gin.Context) {
	if err := h.svc.IndexAll(); err != nil {
		resp.Error(c, http.StatusInternalServerError, err.Error())
		return
	}
	resp.Success(c, gin.H{"message": "ok"})
}
