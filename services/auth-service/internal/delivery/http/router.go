package http

import (
	stdhttp "net/http"

	commonjwt "github.com/TheDeutsch13/b2-common/jwt"
	commonmiddleware "github.com/TheDeutsch13/b2-common/middleware"
	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
	"go.uber.org/zap"
)

func NewRouter(
	logger *zap.Logger,
	authHandler *AuthHandler,
	uploadHandler *UploadHandler,
	jwtManager *commonjwt.Manager,
	uploadDir string,
) *gin.Engine {
	router := gin.New()

	router.Use(gin.Recovery())
	router.Use(commonmiddleware.CORS())

	router.GET("/health", func(ctx *gin.Context) {
		logger.Info("health check requested")

		ctx.JSON(stdhttp.StatusOK, gin.H{
			"service": "auth-service",
			"status":  "ok",
		})
	})

	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
	router.Static("/auth-uploads", uploadDir)

	api := router.Group("/api")
	{
		auth := api.Group("/auth")
		{
			auth.POST("/register", authHandler.Register)
			auth.POST("/login", authHandler.Login)
			auth.POST("/refresh", authHandler.Refresh)
			auth.GET("/users/public", authHandler.ListPublicUsers)

			protected := auth.Group("")
			protected.Use(commonmiddleware.Auth(jwtManager))
			{
				protected.GET("/me", authHandler.Me)
				protected.PATCH("/profile", authHandler.UpdateProfile)
				protected.POST("/upload/avatar", uploadHandler.UploadAvatar)

				admin := protected.Group("")
				admin.Use(commonmiddleware.RequireRole("admin"))
				{
					admin.GET("/users", authHandler.ListUsers)
					admin.PATCH("/users/:id/role", authHandler.UpdateUserRole)
				}
			}
		}
	}

	return router
}
