package user

import (
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/hanasakis/kotoha/internal/middleware"
	resp "github.com/hanasakis/kotoha/pkg/response"
)

type Handler struct {
	svc *Service
}

func NewHandler(svc *Service) *Handler {
	return &Handler{svc: svc}
}

// @Summary      Get user profile
// @Tags         profile
// @Produce      json
// @Security     BearerAuth
// @Success      200 {object} Profile
// @Router       /profile [get]
func (h *Handler) GetProfile(c *gin.Context) {
	userID := middleware.UserIDFromContext(c)

	profile, err := h.svc.GetProfile(userID)
	if err != nil {
		resp.InternalError(c)
		return
	}

	resp.Success(c, profile)
}

// @Summary      Update user profile
// @Tags         profile
// @Accept       json
// @Produce      json
// @Param        body body Profile true "Profile data"
// @Security     BearerAuth
// @Success      200 {object} Profile
// @Router       /profile [put]
func (h *Handler) UpdateProfile(c *gin.Context) {
	userID := middleware.UserIDFromContext(c)

	var profile Profile
	if err := c.ShouldBindJSON(&profile); err != nil {
		resp.BadRequest(c, "common.invalid_request")
		return
	}
	profile.UserID = userID

	if err := h.svc.UpsertProfile(&profile); err != nil {
		resp.InternalError(c)
		return
	}

	resp.Success(c, profile)
}

// @Summary      List user addresses
// @Tags         profile
// @Produce      json
// @Security     BearerAuth
// @Success      200 {array} Address
// @Router       /addresses [get]
func (h *Handler) ListAddresses(c *gin.Context) {
	userID := middleware.UserIDFromContext(c)

	addresses, err := h.svc.GetAddresses(userID)
	if err != nil {
		resp.InternalError(c)
		return
	}

	resp.Success(c, addresses)
}

// @Summary      Create a new address
// @Tags         profile
// @Accept       json
// @Produce      json
// @Param        body body Address true "Address data"
// @Security     BearerAuth
// @Success      201 {object} Address
// @Router       /addresses [post]
func (h *Handler) CreateAddress(c *gin.Context) {
	userID := middleware.UserIDFromContext(c)

	var addr Address
	if err := c.ShouldBindJSON(&addr); err != nil {
		resp.BadRequest(c, "common.invalid_request")
		return
	}
	addr.UserID = userID
	addr.ID = 0

	if err := h.svc.CreateAddress(&addr); err != nil {
		resp.InternalError(c)
		return
	}

	resp.Created(c, addr)
}

// @Summary      Update an address
// @Tags         profile
// @Accept       json
// @Produce      json
// @Param        id   path int     true "Address ID"
// @Param        body body Address true "Updated address data"
// @Security     BearerAuth
// @Success      200 {object} Address
// @Router       /addresses/{id} [put]
func (h *Handler) UpdateAddress(c *gin.Context) {
	userID := middleware.UserIDFromContext(c)
	addrID, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		resp.BadRequest(c, "common.invalid_request")
		return
	}

	var addr Address
	if err := c.ShouldBindJSON(&addr); err != nil {
		resp.BadRequest(c, "common.invalid_request")
		return
	}
	addr.ID = uint(addrID)
	addr.UserID = userID

	if err := h.svc.UpdateAddress(&addr); err != nil {
		resp.InternalError(c)
		return
	}

	resp.Success(c, addr)
}

// @Summary      Delete an address
// @Tags         profile
// @Produce      json
// @Param        id path int true "Address ID"
// @Security     BearerAuth
// @Success      200 {object} map[string]interface{}
// @Router       /addresses/{id} [delete]
func (h *Handler) DeleteAddress(c *gin.Context) {
	userID := middleware.UserIDFromContext(c)
	addrID, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		resp.BadRequest(c, "common.invalid_request")
		return
	}

	if err := h.svc.DeleteAddress(userID, uint(addrID)); err != nil {
		resp.InternalError(c)
		return
	}

	resp.Success(c, gin.H{"message": "ok"})
}

// @Summary      Get user preferences
// @Tags         profile
// @Produce      json
// @Security     BearerAuth
// @Success      200 {object} Preference
// @Router       /preferences [get]
func (h *Handler) GetPreference(c *gin.Context) {
	userID := middleware.UserIDFromContext(c)

	pref, err := h.svc.GetPreference(userID)
	if err != nil {
		resp.InternalError(c)
		return
	}

	resp.Success(c, pref)
}

// @Summary      Update user preferences
// @Tags         profile
// @Accept       json
// @Produce      json
// @Param        body body Preference true "Preference data"
// @Security     BearerAuth
// @Success      200 {object} Preference
// @Router       /preferences [put]
func (h *Handler) UpdatePreference(c *gin.Context) {
	userID := middleware.UserIDFromContext(c)

	var pref Preference
	if err := c.ShouldBindJSON(&pref); err != nil {
		resp.BadRequest(c, "common.invalid_request")
		return
	}
	pref.UserID = userID

	if err := h.svc.UpsertPreference(&pref); err != nil {
		resp.InternalError(c)
		return
	}

	resp.Success(c, pref)
}
