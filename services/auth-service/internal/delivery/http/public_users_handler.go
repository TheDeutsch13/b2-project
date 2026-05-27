package http

import (
	stdhttp "net/http"
	"strconv"
	"strings"

	"github.com/TheDeutsch13/b2-common/httperr"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// ListPublicUsers godoc
// @Summary List public user info by ids
// @Tags auth
// @Produce json
// @Param ids query string true "Comma-separated user ids"
// @Success 200 {array} publicUserResponse
// @Router /api/auth/users/public [get]
func (h *AuthHandler) ListPublicUsers(ctx *gin.Context) {
	raw := strings.TrimSpace(ctx.Query("ids"))
	if raw == "" {
		ctx.JSON(stdhttp.StatusOK, []publicUserResponse{})
		return
	}

	parts := strings.Split(raw, ",")
	seen := make(map[int64]struct{}, len(parts))
	ids := make([]int64, 0, len(parts))

	for _, part := range parts {
		value := strings.TrimSpace(part)
		if value == "" {
			continue
		}

		id, err := strconv.ParseInt(value, 10, 64)
		if err != nil || id <= 0 {
			httperr.BadRequest(ctx, "invalid ids")
			return
		}

		if _, ok := seen[id]; ok {
			continue
		}

		seen[id] = struct{}{}
		ids = append(ids, id)

		if len(ids) >= 50 {
			break
		}
	}

	users, err := h.authUsecase.ListPublicUsersByIDs(ctx.Request.Context(), ids)
	if err != nil {
		h.logger.Error("failed to list public users", zap.Error(err))
		httperr.Internal(ctx, "internal server error")
		return
	}

	ctx.JSON(stdhttp.StatusOK, toPublicUserResponses(users))
}

