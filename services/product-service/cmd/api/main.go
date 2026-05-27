package main

import (
	"context"
	"log"
	"os"
	"time"

	commonjwt "github.com/TheDeutsch13/b2-common/jwt"
	commonlogger "github.com/TheDeutsch13/b2-common/logger"
	commonpostgres "github.com/TheDeutsch13/b2-common/postgres"
	_ "github.com/TheDeutsch13/b2-project/services/product-service/docs"
	"github.com/TheDeutsch13/b2-project/services/product-service/internal/config"
	httpDelivery "github.com/TheDeutsch13/b2-project/services/product-service/internal/delivery/http"
	"github.com/TheDeutsch13/b2-project/services/product-service/internal/delivery/ws"
	postgresRepo "github.com/TheDeutsch13/b2-project/services/product-service/internal/repository/postgres"
	"github.com/TheDeutsch13/b2-project/services/product-service/internal/usecase"
	"go.uber.org/zap"
)

// @title Product Service API
// @version 1.0
// @description API for products, categories and orders.
// @host localhost:8082
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

	if err := postgresRepo.EnsureProductSchema(ctx, db, logger); err != nil {
		logger.Fatal("failed to ensure product schema", zap.Error(err))
	}

	if err := postgresRepo.EnsureOrdersSchema(ctx, db, logger); err != nil {
		logger.Fatal("failed to ensure orders schema", zap.Error(err))
	}

	if err := postgresRepo.EnsureFixedCategories(ctx, db, logger); err != nil {
		logger.Fatal("failed to ensure fixed categories", zap.Error(err))
	}

	if err := postgresRepo.EnsureSupportSchema(ctx, db, logger); err != nil {
		logger.Fatal("failed to ensure support schema", zap.Error(err))
	}

	jwtManager := commonjwt.NewManager(cfg.JWTSecret, time.Hour)

	hub := ws.NewHub()
	go hub.Run()

	productRepository := postgresRepo.NewProductRepository(db)
	categoryRepository := postgresRepo.NewCategoryRepository(db)
	orderRepository := postgresRepo.NewOrderRepository(db)
	supportRepository := postgresRepo.NewSupportRepository(db)

	productUsecase := usecase.NewProductUsecase(
		productRepository,
		orderRepository,
		productRepository,
	)
	categoryUsecase := usecase.NewCategoryUsecase(categoryRepository)
	orderUsecase := usecase.NewOrderUsecase(orderRepository, productRepository)
	supportUsecase := usecase.NewSupportUsecase(supportRepository)

	if err := os.MkdirAll(cfg.UploadDir, 0o755); err != nil {
		logger.Fatal("failed to create upload directory", zap.Error(err))
	}

	productHandler := httpDelivery.NewProductHandler(productUsecase, logger)
	categoryHandler := httpDelivery.NewCategoryHandler(categoryUsecase, logger)
	orderHandler := httpDelivery.NewOrderHandler(orderUsecase, hub, logger)
	cdekHandler := httpDelivery.NewCdekHandler(cfg.CdekClientID, cfg.CdekClientSecret, logger)
	uploadHandler := httpDelivery.NewUploadHandler(cfg.UploadDir, logger)
	supportHandler := httpDelivery.NewSupportHandler(supportUsecase, hub, logger)
	wsHandler := ws.NewHandler(hub, jwtManager)

	router := httpDelivery.NewRouter(
		logger,
		productHandler,
		categoryHandler,
		orderHandler,
		cdekHandler,
		uploadHandler,
		supportHandler,
		wsHandler,
		jwtManager,
		cfg.UploadDir,
	)

	logger.Info("product-service started", zap.String("port", cfg.Port))

	if err := router.Run(":" + cfg.Port); err != nil {
		logger.Fatal("failed to start product-service", zap.Error(err))
	}
}
