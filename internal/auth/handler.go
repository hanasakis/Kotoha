package auth

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/hanasakis/kotoha/internal/middleware"
	"github.com/hanasakis/kotoha/pkg/mail"
	resp "github.com/hanasakis/kotoha/pkg/response"
)

type Handler struct {
	svc     *Service
	mailCli *mail.Client
}

func NewHandler(svc *Service, mailCli *mail.Client) *Handler {
	return &Handler{svc: svc, mailCli: mailCli}
}

// @Summary      Register a new user
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        body body RegisterInput true "Registration details"
// @Success      201 {object} map[string]interface{}
// @Failure      400 {object} map[string]interface{}
// @Router       /auth/register [post]
func (h *Handler) Register(c *gin.Context) {
	var input RegisterInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": "common.invalid_request", "error": err.Error()})
		return
	}

	result, err := h.svc.Register(input, c.GetHeader("User-Agent"), c.ClientIP())
	if err != nil {
		resp.BadRequest(c, err.Error())
		return
	}

	resp.Created(c, result)
}

// @Summary      Login
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        body body LoginInput true "Login credentials"
// @Success      200 {object} map[string]interface{}
// @Failure      401 {object} map[string]interface{}
// @Router       /auth/login [post]
func (h *Handler) Login(c *gin.Context) {
	var input LoginInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": "common.invalid_request", "error": err.Error()})
		return
	}

	result, err := h.svc.Login(input, c.GetHeader("User-Agent"), c.ClientIP())
	if err != nil {
		resp.Unauthorized(c, err.Error())
		return
	}

	resp.Success(c, result)
}

// @Summary      Refresh access token
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        body body object{refresh_token=string} true "Refresh token"
// @Success      200 {object} map[string]interface{}
// @Failure      401 {object} map[string]interface{}
// @Router       /auth/refresh [post]
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
		resp.Unauthorized(c, err.Error())
		return
	}

	resp.Success(c, result)
}

// @Summary      Logout
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        body body object{refresh_token=string} true "Refresh token"
// @Security     BearerAuth
// @Success      200 {object} map[string]interface{}
// @Failure      401 {object} map[string]interface{}
// @Router       /auth/logout [post]
func (h *Handler) Logout(c *gin.Context) {
	var input struct {
		RefreshToken string `json:"refresh_token" binding:"required"`
	}
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": "common.invalid_request", "error": err.Error()})
		return
	}

	if err := h.svc.Logout(input.RefreshToken); err != nil {
		resp.InternalError(c)
		return
	}

	resp.Success(c, gin.H{"message": "ok"})
}

type ForgotPasswordInput struct {
	Email string `json:"email" binding:"required,email"`
}

// @Summary      Request password reset email
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        body body ForgotPasswordInput true "Email address"
// @Success      200 {object} map[string]interface{}
// @Router       /auth/forgot-password [post]
func (h *Handler) ForgotPassword(c *gin.Context) {
	var input ForgotPasswordInput
	if err := c.ShouldBindJSON(&input); err != nil {
		resp.BadRequest(c, "common.invalid_request")
		return
	}
	token, resetURL, err := h.svc.ForgotPassword(input.Email, h.mailCli)
	if err != nil {
		resp.InternalError(c)
		return
	}
	// In dev mode (no SMTP), return the reset link directly
	result := gin.H{"message": "ok"}
	if token != "" && resetURL != "" {
		result["reset_url"] = resetURL
	}
	resp.Success(c, result)
}

type ResetPasswordInput struct {
	Token       string `json:"token" binding:"required"`
	NewPassword string `json:"new_password" binding:"required,min=6"`
}

// @Summary      Reset password with token
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        body body ResetPasswordInput true "Reset token and new password"
// @Success      200 {object} map[string]interface{}
// @Failure      400 {object} map[string]interface{}
// @Router       /auth/reset-password [post]
func (h *Handler) ResetPassword(c *gin.Context) {
	var input ResetPasswordInput
	if err := c.ShouldBindJSON(&input); err != nil {
		resp.BadRequest(c, "common.invalid_request")
		return
	}
	if err := h.svc.ResetPassword(input.Token, input.NewPassword); err != nil {
		resp.BadRequest(c, err.Error())
		return
	}
	resp.Success(c, gin.H{"message": "ok"})
}

type DeleteAccountInput struct {
	Password string `json:"password" binding:"required"`
}

// @Summary      Delete account (soft delete)
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        body body DeleteAccountInput true "Current password"
// @Security     BearerAuth
// @Success      200 {object} map[string]interface{}
// @Failure      400 {object} map[string]interface{}
// @Router       /auth/account [delete]
func (h *Handler) DeleteAccount(c *gin.Context) {
	userID := middleware.UserIDFromContext(c)

	var input DeleteAccountInput
	if err := c.ShouldBindJSON(&input); err != nil {
		resp.BadRequest(c, "common.invalid_request")
		return
	}
	if err := h.svc.DeleteAccount(userID, input.Password); err != nil {
		resp.BadRequest(c, err.Error())
		return
	}
	resp.Success(c, gin.H{"message": "ok"})
}
