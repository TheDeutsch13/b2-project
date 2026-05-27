package http

import (
	"errors"
	stdhttp "net/http"
	"strconv"
	"time"

	"github.com/TheDeutsch13/b2-common/httperr"
	"github.com/TheDeutsch13/b2-project/services/product-service/internal/domain"
	"github.com/TheDeutsch13/b2-project/services/product-service/internal/usecase"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

type ProductHandler struct {
	productUsecase *usecase.ProductUsecase
	logger         *zap.Logger
}

func NewProductHandler(productUsecase *usecase.ProductUsecase, logger *zap.Logger) *ProductHandler {
	return &ProductHandler{
		productUsecase: productUsecase,
		logger:         logger,
	}
}

type productSpecificationRequest struct {
	Label string `json:"label"`
	Value string `json:"value"`
}

type productReviewRequest struct {
	UserID    int64  `json:"user_id,omitempty"`
	Author    string `json:"author"`
	Rating    int    `json:"rating"`
	Text      string `json:"text"`
	CreatedAt string `json:"created_at,omitempty"`
}

type productPayload struct {
	Title          string                        `json:"title"`
	Description    string                        `json:"description"`
	Price          float64                       `json:"price"`
	CategoryID     *int64                        `json:"category_id"`
	Brand          string                        `json:"brand"`
	Stock          int                           `json:"stock"`
	Images         []string                      `json:"images"`
	Specifications []productSpecificationRequest `json:"specifications"`
	Variants       []string                      `json:"variants"`
	Reviews        []productReviewRequest        `json:"reviews"`
}

type productResponse struct {
	ID             int64                         `json:"id"`
	CategoryID     *int64                        `json:"category_id,omitempty"`
	CategoryName   string                        `json:"category_name,omitempty"`
	Title          string                        `json:"title"`
	Description    string                        `json:"description"`
	Price          float64                       `json:"price"`
	Brand          string                        `json:"brand"`
	Stock          int                           `json:"stock"`
	Images         []string                      `json:"images"`
	Specifications []productSpecificationRequest `json:"specifications"`
	Variants       []string                      `json:"variants"`
	Reviews        []productReviewRequest        `json:"reviews"`
	RatingAvg      float64                       `json:"rating_avg"`
	RatingCount    int                           `json:"rating_count"`
	CreatedAt      time.Time                     `json:"created_at"`
}

func toProductResponse(product domain.Product) productResponse {
	specifications := make([]productSpecificationRequest, 0, len(product.Specifications))
	for _, item := range product.Specifications {
		specifications = append(specifications, productSpecificationRequest{
			Label: item.Label,
			Value: item.Value,
		})
	}

	reviews := make([]productReviewRequest, 0, len(product.Reviews))
	for _, item := range product.Reviews {
		reviews = append(reviews, productReviewRequest{
			UserID:    item.UserID,
			Author:    item.Author,
			Rating:    item.Rating,
			Text:      item.Text,
			CreatedAt: item.CreatedAt,
		})
	}

	images := product.Images
	if images == nil {
		images = []string{}
	}

	variants := product.Variants
	if variants == nil {
		variants = []string{}
	}

	return productResponse{
		ID:             product.ID,
		CategoryID:     product.CategoryID,
		CategoryName:   product.CategoryName,
		Title:          product.Title,
		Description:    product.Description,
		Price:          product.Price,
		Brand:          product.Brand,
		Stock:          product.Stock,
		Images:         images,
		Specifications: specifications,
		Variants:       variants,
		Reviews:        reviews,
		RatingAvg:      product.RatingAvg,
		RatingCount:    product.RatingCount,
		CreatedAt:      product.CreatedAt,
	}
}

func (h *ProductHandler) payloadToInput(payload productPayload) usecase.ProductInput {
	specifications := make([]domain.ProductSpecification, 0, len(payload.Specifications))
	for _, item := range payload.Specifications {
		specifications = append(specifications, domain.ProductSpecification{
			Label: item.Label,
			Value: item.Value,
		})
	}

	reviews := make([]domain.ProductReview, 0, len(payload.Reviews))
	for _, item := range payload.Reviews {
		reviews = append(reviews, domain.ProductReview{
			Author: item.Author,
			Rating: item.Rating,
			Text:   item.Text,
		})
	}

	return usecase.ProductInput{
		Title:          payload.Title,
		Description:    payload.Description,
		Price:          payload.Price,
		CategoryID:     payload.CategoryID,
		Brand:          payload.Brand,
		Stock:          payload.Stock,
		Images:         payload.Images,
		Specifications: specifications,
		Variants:       payload.Variants,
		Reviews:        reviews,
	}
}

func (h *ProductHandler) handleUsecaseError(ctx *gin.Context, err error, action string) bool {
	if errors.Is(err, usecase.ErrInvalidInput) {
		httperr.BadRequest(ctx, "title is required and price/stock must be valid")
		return true
	}

	if errors.Is(err, domain.ErrProductNotFound) {
		ctx.JSON(stdhttp.StatusNotFound, gin.H{"error": "product not found"})
		return true
	}

	if errors.Is(err, usecase.ErrProductDuplicate) {
		ctx.JSON(stdhttp.StatusBadRequest, gin.H{
			"error": "Товар с таким названием, брендом, категорией и вариантом уже существует",
		})
		return true
	}

	if errors.Is(err, usecase.ErrProductInOrders) {
		ctx.JSON(stdhttp.StatusConflict, gin.H{
			"error": "Нельзя удалить товар: он есть в заказах",
		})
		return true
	}

		h.logger.Error("product handler error", zap.String("action", action), zap.Error(err))
		httperr.Internal(ctx, "internal server error")

		return true
	}

// Create godoc
// @Summary Create product
// @Tags products
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body productPayload true "Create product request"
// @Success 201 {object} productResponse
// @Router /api/products [post]
func (h *ProductHandler) Create(ctx *gin.Context) {
	var req productPayload

	if err := ctx.ShouldBindJSON(&req); err != nil {
		httperr.BadRequest(ctx, "invalid request body")
		return
	}

	product, err := h.productUsecase.Create(ctx.Request.Context(), h.payloadToInput(req))
	if err != nil {
		if h.handleUsecaseError(ctx, err, "create") {
			return
		}
	}

	ctx.JSON(stdhttp.StatusCreated, toProductResponse(*product))
}

// Update godoc
// @Summary Update product
// @Tags products
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "Product ID"
// @Param request body productPayload true "Update product request"
// @Success 200 {object} productResponse
// @Router /api/products/{id} [put]
func (h *ProductHandler) Update(ctx *gin.Context) {
	id, err := strconv.ParseInt(ctx.Param("id"), 10, 64)
	if err != nil {
		httperr.BadRequest(ctx, "invalid product id")
		return
	}

	var req productPayload

	if err := ctx.ShouldBindJSON(&req); err != nil {
		httperr.BadRequest(ctx, "invalid request body")
		return
	}

	product, err := h.productUsecase.Update(ctx.Request.Context(), id, h.payloadToInput(req))
	if err != nil {
		if h.handleUsecaseError(ctx, err, "update") {
			return
		}
	}

	ctx.JSON(stdhttp.StatusOK, toProductResponse(*product))
}

// Delete godoc
// @Summary Delete product
// @Tags products
// @Security BearerAuth
// @Param id path int true "Product ID"
// @Success 204
// @Router /api/products/{id} [delete]
func (h *ProductHandler) Delete(ctx *gin.Context) {
	id, err := strconv.ParseInt(ctx.Param("id"), 10, 64)
	if err != nil {
		httperr.BadRequest(ctx, "invalid product id")
		return
	}

	if err := h.productUsecase.Delete(ctx.Request.Context(), id); err != nil {
		if h.handleUsecaseError(ctx, err, "delete") {
			return
		}
	}

	ctx.Status(stdhttp.StatusNoContent)
}

// Get godoc
// @Summary Get product by ID
// @Tags products
// @Produce json
// @Param id path int true "Product ID"
// @Success 200 {object} productResponse
// @Router /api/products/{id} [get]
func (h *ProductHandler) Get(ctx *gin.Context) {
	id, err := strconv.ParseInt(ctx.Param("id"), 10, 64)
	if err != nil {
		httperr.BadRequest(ctx, "invalid product id")
		return
	}

	product, err := h.productUsecase.GetByID(ctx.Request.Context(), id)
	if err != nil {
		if h.handleUsecaseError(ctx, err, "get") {
			return
		}
	}

	ctx.JSON(stdhttp.StatusOK, toProductResponse(*product))
}

// List godoc
// @Summary List products
// @Tags products
// @Produce json
// @Param category_id query int false "Category ID"
// @Success 200 {array} productResponse
// @Router /api/products [get]
func (h *ProductHandler) List(ctx *gin.Context) {
	var categoryID *int64

	if rawCategoryID := ctx.Query("category_id"); rawCategoryID != "" {
		value, err := strconv.ParseInt(rawCategoryID, 10, 64)
		if err != nil {
			httperr.BadRequest(ctx, "invalid category_id")
			return
		}

		categoryID = &value
	}

	products, err := h.productUsecase.List(ctx.Request.Context(), categoryID)
	if err != nil {
		h.logger.Error("failed to list products", zap.Error(err))
		httperr.Internal(ctx, "internal server error")
		return
	}

	response := make([]productResponse, 0, len(products))
	for _, product := range products {
		response = append(response, toProductResponse(product))
	}

	ctx.JSON(stdhttp.StatusOK, response)
}
