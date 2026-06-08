package response

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// Standard API response envelope.
// Success:  {"data": <payload>}
// Error:    {"code": "<error_key>", "message": "<translated>"}

func Success(c *gin.Context, data interface{}) {
	c.JSON(http.StatusOK, gin.H{"data": data})
}

func Created(c *gin.Context, data interface{}) {
	c.JSON(http.StatusCreated, gin.H{"data": data})
}

func NoContent(c *gin.Context) {
	c.Status(http.StatusNoContent)
}

func Error(c *gin.Context, status int, code string) {
	c.JSON(status, gin.H{"code": code})
}

func BadRequest(c *gin.Context, code string) {
	c.JSON(http.StatusBadRequest, gin.H{"code": code})
}

func Unauthorized(c *gin.Context, code string) {
	c.JSON(http.StatusUnauthorized, gin.H{"code": code})
}

func Forbidden(c *gin.Context, code string) {
	c.JSON(http.StatusForbidden, gin.H{"code": code})
}

func NotFound(c *gin.Context, code string) {
	c.JSON(http.StatusNotFound, gin.H{"code": code})
}

func InternalError(c *gin.Context) {
	c.JSON(http.StatusInternalServerError, gin.H{"code": "common.server_error"})
}
