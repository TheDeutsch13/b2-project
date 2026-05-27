package http

import (
	"errors"
	stdhttp "net/http"
	"time"

	"github.com/TheDeutsch13/b2-common/httperr"
	commonmiddleware "github.com/TheDeutsch13/b2-common/middleware"
	"github.com/TheDeutsch13/b2-project/services/auth-service/internal/domain"
	"github.com/TheDeutsch13/b2-project/services/auth-service/internal/usecase"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

type updateProfileRequest struct {
	FirstName string  `json:"first_name"`
	LastName  string  `json:"last_name"`
	Nickname  string  `json:"nickname"`
	BirthDate *string `json:"birth_date"`
	Gender    string  `json:"gender"`
	Phone     string  `json:"phone"`
	AvatarURL *string `json:"avatar_url"`
}

// UpdateProfile godoc
// @Summary Update current user profile
// @Tags auth
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body updateProfileRequest true "Profile"
// @Success 200 {object} userResponse
// @Router /api/auth/profile [patch]
func (h *AuthHandler) UpdateProfile(ctx *gin.Context) {
	userID, ok := commonmiddleware.GetUserID(ctx)
	if !ok {
		httperr.Unauthorized(ctx, "unauthorized")
		return
	}

	var req updateProfileRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		httperr.BadRequest(ctx, "invalid request body")
		return
	}

	var birthDate *time.Time
	if req.BirthDate != nil && *req.BirthDate != "" {
		parsed, err := usecase.ParseBirthDate(*req.BirthDate)
		if err != nil {
			httperr.BadRequest(ctx, "invalid birth_date")
			return
		}

		birthDate = parsed
	}

	user, err := h.authUsecase.UpdateProfile(ctx.Request.Context(), userID, domain.UserProfileInput{
		FirstName: req.FirstName,
		LastName:  req.LastName,
		Nickname:  req.Nickname,
		BirthDate: birthDate,
		Gender:    req.Gender,
		Phone:     req.Phone,
	})
	if err != nil {
		if errors.Is(err, usecase.ErrInvalidInput) {
			httperr.BadRequest(ctx, "invalid profile data")
			return
		}

		h.logger.Error("failed to update profile", zap.Error(err))
		httperr.Internal(ctx, "internal server error")
		return
	}

	if req.AvatarURL != nil {
		user, err = h.authUsecase.UpdateAvatar(ctx.Request.Context(), userID, *req.AvatarURL)
	}

	if err != nil {
		if errors.Is(err, usecase.ErrInvalidInput) {
			httperr.BadRequest(ctx, "invalid profile data")
			return
		}

		h.logger.Error("failed to update profile", zap.Error(err))
		httperr.Internal(ctx, "internal server error")
		return
	}

	ctx.JSON(stdhttp.StatusOK, toUserResponse(user))
}

// ListUsers godoc
// @Summary List users (admin)
// @Tags auth
// @Produce json
// @Security BearerAuth
// @Success 200 {array} userResponse
// @Router /api/auth/users [get]
func (h *AuthHandler) ListUsers(ctx *gin.Context) {
	users, err := h.authUsecase.ListUsers(ctx.Request.Context())
	if err != nil {
		h.logger.Error("failed to list users", zap.Error(err))
		httperr.Internal(ctx, "internal server error")
		return
	}

	ctx.JSON(stdhttp.StatusOK, toUserResponses(users))
}
