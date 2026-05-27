package http

import (
	"errors"
	stdhttp "net/http"
	"strconv"

	"github.com/TheDeutsch13/b2-common/httperr"
	commonmiddleware "github.com/TheDeutsch13/b2-common/middleware"
	"github.com/TheDeutsch13/b2-project/services/auth-service/internal/domain"
	"github.com/TheDeutsch13/b2-project/services/auth-service/internal/usecase"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

type updateUserRoleRequest struct {
	Role string `json:"role" binding:"required"`
}

// UpdateUserRole godoc
// @Summary Update user role (admin)
// @Tags auth
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "User ID"
// @Param request body updateUserRoleRequest true "Role"
// @Success 200 {object} userResponse
// @Router /api/auth/users/{id}/role [patch]
func (h *AuthHandler) UpdateUserRole(ctx *gin.Context) {
	actorID, ok := commonmiddleware.GetUserID(ctx)
	if !ok {
		httperr.Unauthorized(ctx, "unauthorized")
		return
	}

	targetUserID, err := strconv.ParseInt(ctx.Param("id"), 10, 64)
	if err != nil || targetUserID <= 0 {
		httperr.BadRequest(ctx, "invalid user id")
		return
	}

	var req updateUserRoleRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		httperr.BadRequest(ctx, "invalid request body")
		return
	}

	user, err := h.authUsecase.UpdateUserRole(
		ctx.Request.Context(),
		actorID,
		targetUserID,
		req.Role,
	)
	if err != nil {
		if errors.Is(err, usecase.ErrInvalidInput) {
			httperr.BadRequest(ctx, "invalid role or cannot change own role")
			return
		}

		if errors.Is(err, domain.ErrUserNotFound) {
			httperr.BadRequest(ctx, "user not found")
			return
		}

		h.logger.Error("failed to update user role", zap.Error(err))
		httperr.Internal(ctx, "internal server error")
		return
	}

	ctx.JSON(stdhttp.StatusOK, toUserResponse(user))
}
