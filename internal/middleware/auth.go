package middleware

import (
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	resp "github.com/hanasakis/kotoha/pkg/response"
)

// UserIDFromContext extracts the authenticated user ID from the gin context.
func UserIDFromContext(c *gin.Context) uint {
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

func AuthRequired(jwtSecret string) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			resp.Unauthorized(c, "auth.unauthorized")
			c.Abort()
			return
		}

		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || !strings.EqualFold(parts[0], "bearer") {
			resp.Unauthorized(c, "auth.unauthorized")
			c.Abort()
			return
		}

		token, err := jwt.Parse(parts[1], func(token *jwt.Token) (interface{}, error) {
			return []byte(jwtSecret), nil
		})
		if err != nil || !token.Valid {
			resp.Unauthorized(c, "auth.invalid_token")
			c.Abort()
			return
		}

		claims, ok := token.Claims.(jwt.MapClaims)
		if !ok {
			resp.Unauthorized(c, "auth.invalid_token")
			c.Abort()
			return
		}

		userID, _ := claims.GetSubject()
		c.Set("user_id", userID)
		c.Set("role", claims["role"])
		c.Next()
	}
}

func RoleRequired(roles ...string) gin.HandlerFunc {
	allowed := make(map[string]bool)
	for _, r := range roles {
		allowed[r] = true
	}

	return func(c *gin.Context) {
		role, _ := c.Get("role")
		roleStr, _ := role.(string)
		if !allowed[roleStr] {
			resp.Forbidden(c, "auth.forbidden")
			c.Abort()
			return
		}
		c.Next()
	}
}
