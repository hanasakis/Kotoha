package router

import (
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	klog "github.com/hanasakis/kotoha/pkg/log"
	resp "github.com/hanasakis/kotoha/pkg/response"

	"github.com/hanasakis/kotoha/internal/agent"
	"github.com/hanasakis/kotoha/internal/auth"
	"github.com/hanasakis/kotoha/internal/cart"
	"github.com/hanasakis/kotoha/internal/catalog"
	"github.com/hanasakis/kotoha/internal/config"
	"github.com/hanasakis/kotoha/internal/i18n"
	"github.com/hanasakis/kotoha/internal/middleware"
	"github.com/hanasakis/kotoha/internal/observability"
	"github.com/hanasakis/kotoha/internal/order"
	"github.com/hanasakis/kotoha/internal/payment"
	"github.com/hanasakis/kotoha/internal/search"
	"github.com/hanasakis/kotoha/internal/user"
	"github.com/hanasakis/kotoha/docs"
	"github.com/hanasakis/kotoha/pkg/embedding"
	"github.com/hanasakis/kotoha/pkg/langfuse"
	"github.com/hanasakis/kotoha/pkg/milvus"
	ollamapkg "github.com/hanasakis/kotoha/pkg/ollama"
	goredis "github.com/hanasakis/kotoha/pkg/redis"
	stripepkg "github.com/hanasakis/kotoha/pkg/stripe"
	"github.com/hanasakis/kotoha/pkg/mail"

	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
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
	accessTTL, err := time.ParseDuration(deps.Config.JWT.AccessTTL)
	if err != nil {
		accessTTL = 15 * time.Minute
	}
	refreshTTL, err := time.ParseDuration(deps.Config.JWT.RefreshTTL)
	if err != nil {
		refreshTTL = 720 * time.Hour
	}
	authSvc := auth.NewService(authRepo, deps.Config.JWT.Secret, accessTTL, refreshTTL)
	mailCli := mail.New(
		deps.Config.SMTP.Host,
		deps.Config.SMTP.Port,
		deps.Config.SMTP.User,
		deps.Config.SMTP.Password,
		deps.Config.SMTP.From,
		deps.Config.SMTP.FromName,
	)
	authH := auth.NewHandler(authSvc, mailCli)

	userSvc := user.NewService(deps.DB)
	userH := user.NewHandler(userSvc)

	catalogRepo := catalog.NewRepository(deps.DB)
	catalogSvc := catalog.NewService(catalogRepo, deps.Config.S3PublicURL())
	if err := catalogSvc.SeedData(); err != nil {
		klog.Warnf(" Failed to seed data: %v", err)
	}
	catalogH := catalog.NewHandler(catalogSvc)

	cartRepo := cart.NewRepository(deps.Redis)
	cartSvc := cart.NewService(cartRepo, catalogRepo)

	orderRepo := order.NewRepository(deps.DB)
	stripeCli := stripepkg.New(deps.Config.Stripe.SecretKey, deps.Config.Stripe.WebhookSecret)
	orderSvc := order.NewService(orderRepo, cartRepo, catalogRepo, stripeCli)

	paymentSvc := payment.NewService(orderRepo, stripeCli)
	paymentH := payment.NewHandler(paymentSvc)

	milvusCli := milvus.New(deps.Config.MilvusAddr(), deps.Config.Milvus.DB, deps.Config.Milvus.User, deps.Config.Milvus.Password)
	embeddingCli := embedding.New(deps.Config.Ollama.Host, deps.Config.Embedding.Model)
	searchSvc := search.NewService(milvusCli, embeddingCli, catalogRepo, deps.Config.Milvus.DenseWeight, deps.Config.Milvus.BM25Weight, deps.Config.Milvus.RecallTopK, deps.Config.Embedding.Dim)
	if err := searchSvc.InitCollection(deps.Config.Embedding.Dim); err != nil {
		klog.Warnf(" Failed to init Milvus collection: %v (will retry on first search)", err)
	}

	ollamaCli := ollamapkg.New(deps.Config.Ollama.Host)
	agentExecutor := agent.NewToolExecutor(catalogRepo, searchSvc, cartSvc)
	agentSvc := agent.NewService(ollamaCli, agentExecutor, userSvc, deps.Config.LLM.Model, deps.Config.LLM.Temperature, deps.Config.LLM.MaxTokens)

	var langfuseCli *langfuse.Client
	if deps.Config.Langfuse.Host != "" && deps.Config.Langfuse.PublicKey != "" {
		klog.Infof("[bootstrap] Langfuse tracing enabled (host=%s)", deps.Config.Langfuse.Host)

		langfuseCli = langfuse.New(deps.Config.Langfuse.Host, deps.Config.Langfuse.PublicKey, deps.Config.Langfuse.SecretKey)
	}
	evalRepo := observability.NewEvalRepository(deps.DB)
	if err := evalRepo.Migrate(); err != nil {
		klog.Warnf(" Failed to migrate eval_runs: %v", err)
	}
	obsSvc := observability.NewService(agentSvc, langfuseCli, deps.Redis, evalRepo, deps.Config.LLM.Model, deps.Config.Server.Env)
	obsH := observability.NewHandler(obsSvc)

	cartH := cart.NewHandler(cartSvc, obsH.TrackCartAdd)
	orderH := order.NewHandler(orderSvc, obsH.TrackOrderCreated)
	searchH := search.NewHandler(searchSvc, obsH.TrackSearch)

	// --- Routes ---
	r.GET("/health", func(c *gin.Context) {
		status := "ok"
		healthDeps := map[string]string{}

		if sqlDB, err := deps.DB.DB(); err == nil {
			if err := sqlDB.Ping(); err != nil {
				healthDeps["postgres"] = "unhealthy: " + err.Error()
			} else {
				healthDeps["postgres"] = "ok"
			}
		} else {
			healthDeps["postgres"] = "unavailable"
		}

		if deps.Redis != nil && deps.Redis.RDB != nil {
			if err := deps.Redis.RDB.Ping(c.Request.Context()).Err(); err != nil {
				healthDeps["redis"] = "unhealthy: " + err.Error()
			} else {
				healthDeps["redis"] = "ok"
			}
		} else {
			healthDeps["redis"] = "unavailable"
		}

		for _, v := range healthDeps {
			if v != "ok" {
				status = "degraded"
				break
			}
		}

		resp.Success(c, gin.H{"status": status, "deps": healthDeps})
	})

	// Swagger UI
	docs.SwaggerInfo.Host = deps.Config.Server.Host + ":" + deps.Config.Server.Port
	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	v1 := r.Group("/api/v1")
	v1.Use(middleware.Tracing(langfuseCli))
	{
		// Public auth routes
		authGroup := v1.Group("/auth")
		{
			authGroup.POST("/register", authH.Register)
			authGroup.POST("/login", authH.Login)
			authGroup.POST("/refresh", authH.Refresh)
			authGroup.POST("/forgot-password", authH.ForgotPassword)
			authGroup.POST("/reset-password", authH.ResetPassword)
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
			// Public payment sync (no auth, uses order_no as secret)
			v1.POST("/public/sync-payment", paymentH.SyncPaymentPublic)

		// Protected routes
		protected := v1.Group("")
		protected.Use(middleware.AuthRequired(deps.Config.JWT.Secret))
		protected.Use(middleware.RateLimit(120, time.Minute, 20))
		{
			// Auth
			protected.POST("/auth/logout", authH.Logout)
			protected.DELETE("/auth/account", authH.DeleteAccount)

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
			protected.POST("/orders/:id/sync-payment", paymentH.SyncPayment)

			// Agent
			protected.POST("/agent/chat", obsH.Chat)

			// Observability
			protected.GET("/metrics", obsH.Metrics)
			protected.GET("/eval/runs", obsH.ListEvalRuns)
			protected.POST("/eval/runs", obsH.SaveEvalRun)

			// Admin routes
			admin := protected.Group("/admin")
			admin.Use(middleware.RoleRequired("admin"))
			{
				admin.POST("/products", catalogH.CreateProduct)
				admin.PUT("/products/:id", catalogH.UpdateProduct)
				admin.DELETE("/products/:id", catalogH.DeleteProduct)
				admin.POST("/search/index", searchH.IndexAll)
				admin.GET("/orders", orderH.AdminListOrders)
				admin.GET("/orders/:id", orderH.AdminGetOrder)
			}
		}
	}

	return r
}
