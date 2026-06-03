package user

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

func userIDFromContext(c *gin.Context) (uint, error) {
	idStr, _ := c.Get("user_id")
	switch v := idStr.(type) {
	case float64:
		return uint(v), nil
	case string:
		id, _ := strconv.ParseUint(v, 10, 64)
		return uint(id), nil
	}
	return 0, nil
}

func (h *Handler) GetProfile(c *gin.Context) {
	userID, _ := userIDFromContext(c)

	profile, err := h.svc.GetProfile(userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": "common.server_error"})
		return
	}

	c.JSON(http.StatusOK, profile)
}

func (h *Handler) UpdateProfile(c *gin.Context) {
	userID, _ := userIDFromContext(c)

	var profile Profile
	if err := c.ShouldBindJSON(&profile); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": "common.invalid_request"})
		return
	}
	profile.UserID = userID

	if err := h.svc.UpsertProfile(&profile); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": "common.server_error"})
		return
	}

	c.JSON(http.StatusOK, profile)
}

func (h *Handler) ListAddresses(c *gin.Context) {
	userID, _ := userIDFromContext(c)

	addresses, err := h.svc.GetAddresses(userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": "common.server_error"})
		return
	}

	c.JSON(http.StatusOK, addresses)
}

func (h *Handler) CreateAddress(c *gin.Context) {
	userID, _ := userIDFromContext(c)

	var addr Address
	if err := c.ShouldBindJSON(&addr); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": "common.invalid_request"})
		return
	}
	addr.UserID = userID
	addr.ID = 0

	if err := h.svc.CreateAddress(&addr); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": "common.server_error"})
		return
	}

	c.JSON(http.StatusCreated, addr)
}

func (h *Handler) UpdateAddress(c *gin.Context) {
	userID, _ := userIDFromContext(c)
	addrID, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": "common.invalid_request"})
		return
	}

	var addr Address
	if err := c.ShouldBindJSON(&addr); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": "common.invalid_request"})
		return
	}
	addr.ID = uint(addrID)
	addr.UserID = userID

	if err := h.svc.UpdateAddress(&addr); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": "common.server_error"})
		return
	}

	c.JSON(http.StatusOK, addr)
}

func (h *Handler) DeleteAddress(c *gin.Context) {
	userID, _ := userIDFromContext(c)
	addrID, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": "common.invalid_request"})
		return
	}

	if err := h.svc.DeleteAddress(userID, uint(addrID)); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": "common.server_error"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "ok"})
}

func (h *Handler) GetPreference(c *gin.Context) {
	userID, _ := userIDFromContext(c)

	pref, err := h.svc.GetPreference(userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": "common.server_error"})
		return
	}

	c.JSON(http.StatusOK, pref)
}

func (h *Handler) UpdatePreference(c *gin.Context) {
	userID, _ := userIDFromContext(c)

	var pref Preference
	if err := c.ShouldBindJSON(&pref); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": "common.invalid_request"})
		return
	}
	pref.UserID = userID

	if err := h.svc.UpsertPreference(&pref); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": "common.server_error"})
		return
	}

	c.JSON(http.StatusOK, pref)
}
