package agent

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/hanasakis/kotoha/internal/middleware"
	"github.com/hanasakis/kotoha/pkg/ollama"
)

type Handler struct {
	svc *Service
}

func NewHandler(svc *Service) *Handler {
	return &Handler{svc: svc}
}

type ChatRequest struct {
	Message string           `json:"message" binding:"required"`
	History []ollama.Message `json:"history"`
}

func (h *Handler) Chat(c *gin.Context) {
	userID := middleware.UserIDFromContext(c)

	var req ChatRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": "common.invalid_request"})
		return
	}

	output, err := h.svc.Chat(c.Request.Context(), userID, req.History, ChatInput{Message: req.Message})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": "agent.error"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"reply": output.Reply,
	})
}
