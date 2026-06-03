package auth

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

type Handler struct {
	svc *Service
}

func NewHandler(svc *Service) *Handler {
	return &Handler{svc: svc}
}

func (h *Handler) Register(c *gin.Context) {
	var input RegisterInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": "common.invalid_request", "error": err.Error()})
		return
	}

	result, err := h.svc.Register(input, c.GetHeader("User-Agent"), c.ClientIP())
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, result)
}

func (h *Handler) Login(c *gin.Context) {
	var input LoginInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": "common.invalid_request", "error": err.Error()})
		return
	}

	result, err := h.svc.Login(input, c.GetHeader("User-Agent"), c.ClientIP())
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"code": err.Error()})
		return
	}

	c.JSON(http.StatusOK, result)
}

func (h *Handler) Refresh(c *gin.Context) {
	var input struct {
		RefreshToken string `json:"refresh_token" binding:"required"`
	}
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": "common.invalid_request", "error": err.Error()})
		return
	}

	result, err := h.svc.Refresh(input.RefreshToken, c.GetHeader("User-Agent"), c.ClientIP())
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"code": err.Error()})
		return
	}

	c.JSON(http.StatusOK, result)
}

func (h *Handler) Logout(c *gin.Context) {
	var input struct {
		RefreshToken string `json:"refresh_token" binding:"required"`
	}
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": "common.invalid_request", "error": err.Error()})
		return
	}

	if err := h.svc.Logout(input.RefreshToken); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": "common.server_error"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "ok"})
}
