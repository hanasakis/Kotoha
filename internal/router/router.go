package router

import (
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"github.com/hanasakis/kotoha/internal/config"
	"github.com/hanasakis/kotoha/internal/i18n"
	"github.com/hanasakis/kotoha/internal/middleware"
	goredis "github.com/hanasakis/kotoha/pkg/redis"
)

type Dependencies struct {
	DB     *gorm.DB
	Redis  *goredis.Client
	Config *config.Config
	I18n   *i18n.Translator
}

func Setup(deps *Dependencies) *gin.Engine {
	r := gin.New()

	r.Use(
		gin.Recovery(),
		middleware.RequestID(),
		middleware.CORS(deps.Config.CORS.Origins),
		middleware.I18n(deps.Config.I18n.DefaultLocale, deps.Config.I18n.SupportedLocales),
	)

	// Health check
	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})

	// API v1
	v1 := r.Group("/api/v1")
	{
		// Public routes — no auth required
		public := v1.Group("")
		{
			public.GET("/ping", func(c *gin.Context) {
				locale, _ := c.Get("locale")
				c.JSON(200, gin.H{
					"message": "pong",
					"locale":  locale,
				})
			})
		}

		// Protected routes — JWT required
		protected := v1.Group("")
		protected.Use(middleware.AuthRequired(deps.Config.JWT.Secret))
		{
			protected.GET("/me", func(c *gin.Context) {
				userID, _ := c.Get("user_id")
				role, _ := c.Get("role")
				c.JSON(200, gin.H{
					"user_id": userID,
					"role":    role,
				})
			})
		}
	}

	return r
}
