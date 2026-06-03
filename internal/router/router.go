package router

import (
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"github.com/hanasakis/kotoha/internal/agent"
	"github.com/hanasakis/kotoha/internal/auth"
	"github.com/hanasakis/kotoha/internal/cart"
	"github.com/hanasakis/kotoha/internal/catalog"
	"github.com/hanasakis/kotoha/internal/config"
	"github.com/hanasakis/kotoha/internal/i18n"
	"github.com/hanasakis/kotoha/internal/middleware"
	"github.com/hanasakis/kotoha/internal/order"
	"github.com/hanasakis/kotoha/internal/payment"
	"github.com/hanasakis/kotoha/internal/search"
	"github.com/hanasakis/kotoha/internal/user"
	"github.com/hanasakis/kotoha/pkg/embedding"
	"github.com/hanasakis/kotoha/pkg/milvus"
	ollamapkg "github.com/hanasakis/kotoha/pkg/ollama"
	goredis "github.com/hanasakis/kotoha/pkg/redis"
	stripepkg "github.com/hanasakis/kotoha/pkg/stripe"
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

	catalogRepo := catalog.NewRepository(deps.DB)
	catalogSvc := catalog.NewService(catalogRepo)
	catalogH := catalog.NewHandler(catalogSvc)

	cartRepo := cart.NewRepository(deps.Redis)
	cartSvc := cart.NewService(cartRepo, catalogRepo)
	cartH := cart.NewHandler(cartSvc)

	orderRepo := order.NewRepository(deps.DB)
	orderSvc := order.NewService(orderRepo, cartRepo, catalogRepo)
	orderH := order.NewHandler(orderSvc)

	stripeCli := stripepkg.New(deps.Config.Stripe.SecretKey, deps.Config.Stripe.WebhookSecret)
	paymentSvc := payment.NewService(orderRepo, stripeCli)
	paymentH := payment.NewHandler(paymentSvc)

	milvusCli := milvus.New(deps.Config.MilvusAddr(), deps.Config.Milvus.DB)
	embeddingCli := embedding.New(deps.Config.Ollama.Host, deps.Config.Embedding.Model)
	searchSvc := search.NewService(milvusCli, embeddingCli, catalogRepo, deps.Config.Milvus.DenseWeight, deps.Config.Milvus.RecallTopK)
	searchH := search.NewHandler(searchSvc)

	ollamaCli := ollamapkg.New(deps.Config.Ollama.Host)
	agentExecutor := agent.NewToolExecutor(catalogRepo, searchSvc, cartSvc)
	agentSvc := agent.NewService(ollamaCli, agentExecutor, deps.Config.LLM.Model, deps.Config.LLM.Temperature, deps.Config.LLM.MaxTokens)
	agentH := agent.NewHandler(agentSvc)

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

		// Public catalog routes
		catalogGroup := v1.Group("/catalog")
		{
			catalogGroup.GET("/categories", catalogH.ListCategories)
			catalogGroup.GET("/products", catalogH.ListProducts)
			catalogGroup.GET("/products/:id", catalogH.GetProduct)
			catalogGroup.GET("/search", searchH.Search)
		}

		// Stripe webhook (public)
		v1.POST("/webhook", paymentH.HandleWebhook)

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

			// Cart
			protected.GET("/cart", cartH.GetCart)
			protected.POST("/cart/items", cartH.AddItem)
			protected.PUT("/cart/items/:skuID", cartH.UpdateQty)
			protected.DELETE("/cart/items/:skuID", cartH.RemoveItem)

			// Orders
			protected.POST("/orders", orderH.CreateOrder)
			protected.GET("/orders", orderH.ListOrders)
			protected.GET("/orders/:id", orderH.GetOrder)
			protected.POST("/orders/:id/cancel", orderH.CancelOrder)

			// Payment
			protected.POST("/orders/:id/checkout", paymentH.CreateCheckout)

			// Agent
			protected.POST("/agent/chat", agentH.Chat)

			// Admin routes
			admin := protected.Group("/admin")
			admin.Use(middleware.RoleRequired("admin"))
			{
				admin.POST("/products", catalogH.CreateProduct)
				admin.PUT("/products/:id", catalogH.UpdateProduct)
				admin.DELETE("/products/:id", catalogH.DeleteProduct)
				admin.POST("/search/index", searchH.IndexAll)
			}
		}
	}

	return r
}
