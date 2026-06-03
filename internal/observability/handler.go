package observability

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/hanasakis/kotoha/internal/agent"
	"github.com/hanasakis/kotoha/pkg/ollama"
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

type ChatRequest struct {
	Message string           `json:"message" binding:"required"`
	History []ollama.Message `json:"history"`
}

func (h *Handler) Chat(c *gin.Context) {
	userID := userIDFromContext(c)

	var req ChatRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": "common.invalid_request"})
		return
	}

	output, err := h.svc.Chat(c.Request.Context(), userID, req.History, agent.ChatInput{Message: req.Message})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": "agent.error"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"reply": output.Reply})
}

func (h *Handler) Metrics(c *gin.Context) {
	metrics := h.svc.GetMetrics()
	c.JSON(http.StatusOK, gin.H{"metrics": metrics})
}
