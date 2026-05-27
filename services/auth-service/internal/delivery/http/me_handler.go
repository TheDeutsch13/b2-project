package http

import (
	"errors"
	stdhttp "net/http"

	"github.com/TheDeutsch13/b2-common/httperr"
	commonmiddleware "github.com/TheDeutsch13/b2-common/middleware"
	"github.com/TheDeutsch13/b2-project/services/auth-service/internal/usecase"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// Me godoc
// @Summary Get current user
// @Description Returns authenticated user profile from database
// @Tags auth
// @Produce json
// @Security BearerAuth
// @Success 200 {object} userResponse
// @Failure 401 {object} map[string]string
// @Router /api/auth/me [get]
func (h *AuthHandler) Me(ctx *gin.Context) {
	userID, ok := commonmiddleware.GetUserID(ctx)
	if !ok {
		httperr.Unauthorized(ctx, "unauthorized")
		return
	}

	user, err := h.authUsecase.GetProfile(ctx.Request.Context(), userID)
	if err != nil {
		if errors.Is(err, usecase.ErrInvalidInput) {
			httperr.BadRequest(ctx, "invalid user id")
			return
		}

		h.logger.Error("failed to get profile", zap.Error(err))
		httperr.Internal(ctx, "internal server error")
		return
	}

	ctx.JSON(stdhttp.StatusOK, toUserResponse(user))
}
