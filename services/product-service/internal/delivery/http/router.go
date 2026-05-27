package http

import (
	stdhttp "net/http"

	commonjwt "github.com/TheDeutsch13/b2-common/jwt"
	commonmiddleware "github.com/TheDeutsch13/b2-common/middleware"
	"github.com/TheDeutsch13/b2-project/services/product-service/internal/delivery/ws"
	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
	"go.uber.org/zap"
)

func NewRouter(
	logger *zap.Logger,
	productHandler *ProductHandler,
	categoryHandler *CategoryHandler,
	orderHandler *OrderHandler,
	cdekHandler *CdekHandler,
	uploadHandler *UploadHandler,
	supportHandler *SupportHandler,
	wsHandler *ws.Handler,
	jwtManager *commonjwt.Manager,
	uploadDir string,
) *gin.Engine {
	router := gin.New()

	router.Use(gin.Recovery())
	router.Use(commonmiddleware.CORS())

	router.GET("/health", func(ctx *gin.Context) {
		logger.Info("health check requested")

		ctx.JSON(stdhttp.StatusOK, gin.H{
			"service": "product-service",
			"status":  "ok",
		})
	})

	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
	router.GET("/ws/notifications", wsHandler.Notifications)
	router.Static("/uploads", uploadDir)

	api := router.Group("/api")
	{
		api.GET("/products", productHandler.List)
		api.GET("/products/:id", productHandler.Get)
		api.GET("/categories", categoryHandler.List)
		api.GET("/cdek/points", cdekHandler.ListPoints)

		protected := api.Group("")
		protected.Use(commonmiddleware.Auth(jwtManager))
		{
			protected.POST("/orders", orderHandler.Create)
			protected.GET("/orders/my", orderHandler.ListMy)
			protected.PUT("/products/:id/reviews", productHandler.UpsertMyReview)
			protected.DELETE("/products/:id/reviews", productHandler.DeleteMyReview)
			protected.GET("/reviews/my", productHandler.ListMyReviews)
			protected.GET("/support/my", supportHandler.GetMyThread)
			protected.POST("/support/my/messages", supportHandler.SendMyMessage)

			staff := protected.Group("")
			staff.Use(commonmiddleware.RequireAnyRole("admin", "moderator"))
			{
				staff.GET("/support/threads", supportHandler.ListThreads)
				staff.GET("/support/threads/:id", supportHandler.GetThread)
				staff.POST("/support/threads/:id/messages", supportHandler.SendThreadMessage)
				staff.PATCH("/support/threads/:id/close", supportHandler.CloseThread)
				staff.DELETE("/support/threads/:id", supportHandler.DeleteThread)
			}

			admin := protected.Group("")
			admin.Use(commonmiddleware.RequireRole("admin"))
			{
				admin.POST("/products", productHandler.Create)
				admin.PUT("/products/:id", productHandler.Update)
				admin.DELETE("/products/:id", productHandler.Delete)
				admin.POST("/upload", uploadHandler.Upload)
				admin.GET("/orders", orderHandler.ListAll)
				admin.PATCH("/orders/:id/status", orderHandler.UpdateStatus)
				admin.GET("/reviews", productHandler.ListAdminReviews)
			}

			delivery := protected.Group("")
			delivery.Use(commonmiddleware.RequireAnyRole("admin", "courier"))
			{
				delivery.GET("/courier/orders", orderHandler.ListAll)
				delivery.PATCH("/courier/orders/:id/status", orderHandler.UpdateStatus)
			}
		}
	}

	return router
}
