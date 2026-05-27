package http

import (
	"errors"
	stdhttp "net/http"
	"time"

	"github.com/TheDeutsch13/b2-common/httperr"
	"github.com/TheDeutsch13/b2-project/services/product-service/internal/domain"
	"github.com/TheDeutsch13/b2-project/services/product-service/internal/usecase"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

type CategoryHandler struct {
	categoryUsecase *usecase.CategoryUsecase
	logger          *zap.Logger
}

func NewCategoryHandler(categoryUsecase *usecase.CategoryUsecase, logger *zap.Logger) *CategoryHandler {
	return &CategoryHandler{
		categoryUsecase: categoryUsecase,
		logger:          logger,
	}
}

type createCategoryRequest struct {
	Name string `json:"name"`
}

type categoryResponse struct {
	ID        int64     `json:"id"`
	Name      string    `json:"name"`
	CreatedAt time.Time `json:"created_at"`
}

func toCategoryResponse(category domain.Category) categoryResponse {
	return categoryResponse{
		ID:        category.ID,
		Name:      category.Name,
		CreatedAt: category.CreatedAt,
	}
}

// ListCategories godoc
// @Summary List categories
// @Tags categories
// @Produce json
// @Success 200 {array} categoryResponse
// @Router /api/categories [get]
func (h *CategoryHandler) List(ctx *gin.Context) {
	categories, err := h.categoryUsecase.List(ctx.Request.Context())
	if err != nil {
		h.logger.Error("failed to list categories", zap.Error(err))
		httperr.Internal(ctx, "internal server error")
		return
	}

	response := make([]categoryResponse, 0, len(categories))
	for _, category := range categories {
		response = append(response, toCategoryResponse(category))
	}

	ctx.JSON(stdhttp.StatusOK, response)
}

// CreateCategory godoc
// @Summary Create category
// @Tags categories
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body createCategoryRequest true "Create category request"
// @Success 201 {object} categoryResponse
// @Router /api/categories [post]
func (h *CategoryHandler) Create(ctx *gin.Context) {
	var req createCategoryRequest

	if err := ctx.ShouldBindJSON(&req); err != nil {
		httperr.BadRequest(ctx, "invalid request body")
		return
	}

	category, err := h.categoryUsecase.Create(ctx.Request.Context(), req.Name)
	if err != nil {
		if errors.Is(err, usecase.ErrInvalidInput) {
			httperr.BadRequest(ctx, "name is required")
			return
		}

		if errors.Is(err, usecase.ErrCategoryAlreadyExists) {
			httperr.Conflict(ctx, "category with this name already exists")
			return
		}

		h.logger.Error("failed to create category", zap.Error(err))
		httperr.Internal(ctx, "internal server error")
		return
	}

	ctx.JSON(stdhttp.StatusCreated, toCategoryResponse(*category))
}
