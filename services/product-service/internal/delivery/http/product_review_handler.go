package http

import (
	"errors"
	stdhttp "net/http"
	"strconv"

	"github.com/TheDeutsch13/b2-common/httperr"
	commonmiddleware "github.com/TheDeutsch13/b2-common/middleware"
	"github.com/TheDeutsch13/b2-project/services/product-service/internal/domain"
	"github.com/TheDeutsch13/b2-project/services/product-service/internal/usecase"
	"github.com/gin-gonic/gin"
)

type upsertMyReviewRequest struct {
	Author string `json:"author"`
	Rating int    `json:"rating"`
	Text   string `json:"text"`
}

type myProductReviewResponse struct {
	ProductID    int64  `json:"product_id"`
	ProductTitle string `json:"product_title"`
	ProductImage string `json:"product_image,omitempty"`
	Author       string `json:"author"`
	Rating       int    `json:"rating"`
	Text         string `json:"text"`
	CreatedAt    string `json:"created_at,omitempty"`
}

func toMyProductReviewResponse(item domain.UserProductReview) myProductReviewResponse {
	return myProductReviewResponse{
		ProductID:    item.ProductID,
		ProductTitle: item.ProductTitle,
		ProductImage: item.ProductImage,
		Author:       item.Review.Author,
		Rating:       item.Review.Rating,
		Text:         item.Review.Text,
		CreatedAt:    item.Review.CreatedAt,
	}
}

func (h *ProductHandler) handleReviewUsecaseError(ctx *gin.Context, err error, action string) bool {
	if errors.Is(err, usecase.ErrInvalidInput) {
		httperr.BadRequest(ctx, "invalid review data")
		return true
	}

	if errors.Is(err, domain.ErrReviewNotAllowed) {
		ctx.JSON(stdhttp.StatusForbidden, gin.H{
			"error": "Отзыв можно оставить только после получения заказа",
		})
		return true
	}

	if errors.Is(err, domain.ErrReviewNotFound) {
		ctx.JSON(stdhttp.StatusNotFound, gin.H{"error": "review not found"})
		return true
	}

	if errors.Is(err, domain.ErrProductNotFound) {
		ctx.JSON(stdhttp.StatusNotFound, gin.H{"error": "product not found"})
		return true
	}

	return h.handleUsecaseError(ctx, err, action)
}

// UpsertMyReview godoc
// @Summary Upsert current user review for product
// @Tags products
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "Product ID"
// @Param request body upsertMyReviewRequest true "Review"
// @Success 200 {object} productResponse
// @Router /api/products/{id}/reviews [put]
func (h *ProductHandler) UpsertMyReview(ctx *gin.Context) {
	userID, ok := commonmiddleware.GetUserID(ctx)
	if !ok {
		httperr.Unauthorized(ctx, "unauthorized")
		return
	}

	productID, err := strconv.ParseInt(ctx.Param("id"), 10, 64)
	if err != nil {
		httperr.BadRequest(ctx, "invalid product id")
		return
	}

	var req upsertMyReviewRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		httperr.BadRequest(ctx, "invalid request body")
		return
	}

	product, err := h.productUsecase.UpsertUserReview(
		ctx.Request.Context(),
		userID,
		productID,
		req.Author,
		req.Rating,
		req.Text,
	)
	if err != nil {
		if h.handleReviewUsecaseError(ctx, err, "upsert review") {
			return
		}
	}

	ctx.JSON(stdhttp.StatusOK, toProductResponse(*product))
}

// DeleteMyReview godoc
// @Summary Delete current user review for product
// @Tags products
// @Produce json
// @Security BearerAuth
// @Param id path int true "Product ID"
// @Success 200 {object} productResponse
// @Router /api/products/{id}/reviews [delete]
func (h *ProductHandler) DeleteMyReview(ctx *gin.Context) {
	userID, ok := commonmiddleware.GetUserID(ctx)
	if !ok {
		httperr.Unauthorized(ctx, "unauthorized")
		return
	}

	productID, err := strconv.ParseInt(ctx.Param("id"), 10, 64)
	if err != nil {
		httperr.BadRequest(ctx, "invalid product id")
		return
	}

	product, err := h.productUsecase.DeleteUserReview(ctx.Request.Context(), userID, productID)
	if err != nil {
		if h.handleReviewUsecaseError(ctx, err, "delete review") {
			return
		}
	}

	ctx.JSON(stdhttp.StatusOK, toProductResponse(*product))
}

// ListMyReviews godoc
// @Summary List current user reviews
// @Tags reviews
// @Produce json
// @Security BearerAuth
// @Success 200 {array} myProductReviewResponse
// @Router /api/reviews/my [get]
func (h *ProductHandler) ListMyReviews(ctx *gin.Context) {
	userID, ok := commonmiddleware.GetUserID(ctx)
	if !ok {
		httperr.Unauthorized(ctx, "unauthorized")
		return
	}

	reviews, err := h.productUsecase.ListUserReviews(ctx.Request.Context(), userID)
	if err != nil {
		if h.handleReviewUsecaseError(ctx, err, "list reviews") {
			return
		}
	}

	response := make([]myProductReviewResponse, 0, len(reviews))
	for _, item := range reviews {
		response = append(response, toMyProductReviewResponse(item))
	}

	ctx.JSON(stdhttp.StatusOK, response)
}

type adminProductReviewResponse struct {
	ProductID    int64  `json:"product_id"`
	ProductTitle string `json:"product_title"`
	UserID       int64  `json:"user_id,omitempty"`
	Author       string `json:"author"`
	Rating       int    `json:"rating"`
	Text         string `json:"text"`
	CreatedAt    string `json:"created_at,omitempty"`
}

func toAdminProductReviewResponse(item domain.AdminProductReview) adminProductReviewResponse {
	return adminProductReviewResponse{
		ProductID:    item.ProductID,
		ProductTitle: item.ProductTitle,
		UserID:       item.UserID,
		Author:       item.Author,
		Rating:       item.Rating,
		Text:         item.Text,
		CreatedAt:    item.CreatedAt,
	}
}

// ListAdminReviews godoc
// @Summary List product reviews (admin)
// @Tags reviews
// @Produce json
// @Security BearerAuth
// @Param rating query int false "Filter by rating 1-5"
// @Param product_id query int false "Filter by product"
// @Param q query string false "Search in product, author, text"
// @Success 200 {array} adminProductReviewResponse
// @Router /api/reviews [get]
func (h *ProductHandler) ListAdminReviews(ctx *gin.Context) {
	filter := domain.ReviewListFilter{}

	if raw := ctx.Query("rating"); raw != "" {
		rating, err := strconv.Atoi(raw)
		if err != nil || rating < 1 || rating > 5 {
			httperr.BadRequest(ctx, "invalid rating filter")
			return
		}

		filter.Rating = rating
	}

	if raw := ctx.Query("product_id"); raw != "" {
		productID, err := strconv.ParseInt(raw, 10, 64)
		if err != nil || productID <= 0 {
			httperr.BadRequest(ctx, "invalid product_id filter")
			return
		}

		filter.ProductID = productID
	}

	filter.Query = ctx.Query("q")

	reviews, err := h.productUsecase.ListAllReviews(ctx.Request.Context(), filter)
	if err != nil {
		if h.handleReviewUsecaseError(ctx, err, "list admin reviews") {
			return
		}
	}

	response := make([]adminProductReviewResponse, 0, len(reviews))
	for _, item := range reviews {
		response = append(response, toAdminProductReviewResponse(item))
	}

	ctx.JSON(stdhttp.StatusOK, response)
}
