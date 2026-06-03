package router

import (
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"github.com/hanasakis/kotoha/internal/auth"
	"github.com/hanasakis/kotoha/internal/config"
	"github.com/hanasakis/kotoha/internal/i18n"
	"github.com/hanasakis/kotoha/internal/middleware"
	"github.com/hanasakis/kotoha/internal/user"
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

	// --- Services ---
	authRepo := auth.NewRepository(deps.DB)
	accessTTL, _ := time.ParseDuration(deps.Config.JWT.AccessTTL)
	refreshTTL, _ := time.ParseDuration(deps.Config.JWT.RefreshTTL)
	authSvc := auth.NewService(authRepo, deps.Config.JWT.Secret, accessTTL, refreshTTL)
	authH := auth.NewHandler(authSvc)

	userSvc := user.NewService(deps.DB)
	userH := user.NewHandler(userSvc)

	// --- Routes ---
	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})

	v1 := r.Group("/api/v1")
	{
		// Public auth routes
		authGroup := v1.Group("/auth")
		{
			authGroup.POST("/register", authH.Register)
			authGroup.POST("/login", authH.Login)
			authGroup.POST("/refresh", authH.Refresh)
		}

		// Protected routes
		protected := v1.Group("")
		protected.Use(middleware.AuthRequired(deps.Config.JWT.Secret))
		{
			// Auth
			protected.POST("/auth/logout", authH.Logout)

			// User profile
			protected.GET("/profile", userH.GetProfile)
			protected.PUT("/profile", userH.UpdateProfile)

			// Addresses
			protected.GET("/addresses", userH.ListAddresses)
			protected.POST("/addresses", userH.CreateAddress)
			protected.PUT("/addresses/:id", userH.UpdateAddress)
			protected.DELETE("/addresses/:id", userH.DeleteAddress)

			// Preferences
			protected.GET("/preferences", userH.GetPreference)
			protected.PUT("/preferences", userH.UpdatePreference)
		}
	}

	return r
}
