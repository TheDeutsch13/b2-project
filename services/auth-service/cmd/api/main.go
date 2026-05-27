package main

import (
	"context"
	"log"
	"os"

	commonjwt "github.com/TheDeutsch13/b2-common/jwt"
	commonlogger "github.com/TheDeutsch13/b2-common/logger"
	commonpostgres "github.com/TheDeutsch13/b2-common/postgres"
	_ "github.com/TheDeutsch13/b2-project/services/auth-service/docs"
	"github.com/TheDeutsch13/b2-project/services/auth-service/internal/config"
	httpDelivery "github.com/TheDeutsch13/b2-project/services/auth-service/internal/delivery/http"
	postgresRepo "github.com/TheDeutsch13/b2-project/services/auth-service/internal/repository/postgres"
	"github.com/TheDeutsch13/b2-project/services/auth-service/internal/usecase"
	"go.uber.org/zap"
)

// @title Auth Service API
// @version 1.0
// @description API for user registration and authentication.
// @host localhost:8081
// @BasePath /
// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
func main() {
	logger, err := commonlogger.NewProduction()
	if err != nil {
		log.Fatal(err)
	}
	defer logger.Sync()

	cfg := config.New()

	ctx := context.Background()

	db, err := commonpostgres.Connect(ctx, cfg.DatabaseURL)
	if err != nil {
		logger.Fatal("failed to connect to database", zap.Error(err))
	}
	defer db.Close()

	if err := postgresRepo.EnsureUserProfileSchema(ctx, db, logger); err != nil {
		logger.Fatal("failed to ensure user profile schema", zap.Error(err))
	}

	jwtManager := commonjwt.NewManager(cfg.JWTSecret, cfg.AccessTokenTTL)

	userRepository := postgresRepo.NewUserRepository(db)
	refreshTokenRepository := postgresRepo.NewRefreshTokenRepository(db)
	authUsecase := usecase.NewAuthUsecase(
		userRepository,
		refreshTokenRepository,
		jwtManager,
		cfg.RefreshTokenTTL,
	)
	authHandler := httpDelivery.NewAuthHandler(authUsecase, logger)
	uploadHandler := httpDelivery.NewUploadHandler(authUsecase, cfg.UploadDir, logger)

	if err := os.MkdirAll(cfg.UploadDir, 0o755); err != nil {
		logger.Fatal("failed to create upload directory", zap.Error(err))
	}

	router := httpDelivery.NewRouter(
		logger,
		authHandler,
		uploadHandler,
		jwtManager,
		cfg.UploadDir,
	)

	logger.Info("auth-service started", zap.String("port", cfg.Port))

	if err := router.Run(":" + cfg.Port); err != nil {
		logger.Fatal("failed to start auth-service", zap.Error(err))
	}
}
