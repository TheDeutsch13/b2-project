package http

import (
	"fmt"
	stdhttp "net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/TheDeutsch13/b2-common/httperr"
	commonmiddleware "github.com/TheDeutsch13/b2-common/middleware"
	"github.com/TheDeutsch13/b2-project/services/auth-service/internal/usecase"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

const maxAvatarUploadSize = 2 << 20

type UploadHandler struct {
	authUsecase *usecase.AuthUsecase
	uploadDir   string
	logger      *zap.Logger
}

func NewUploadHandler(
	authUsecase *usecase.AuthUsecase,
	uploadDir string,
	logger *zap.Logger,
) *UploadHandler {
	return &UploadHandler{
		authUsecase: authUsecase,
		uploadDir:   uploadDir,
		logger:      logger,
	}
}

type uploadResponse struct {
	URL string `json:"url"`
}

// UploadAvatar godoc
// @Summary Upload user avatar
// @Tags auth
// @Accept multipart/form-data
// @Produce json
// @Security BearerAuth
// @Param file formData file true "Avatar image"
// @Success 201 {object} uploadResponse
// @Router /api/auth/upload/avatar [post]
func (h *UploadHandler) UploadAvatar(ctx *gin.Context) {
	userID, ok := commonmiddleware.GetUserID(ctx)
	if !ok {
		httperr.Unauthorized(ctx, "unauthorized")
		return
	}

	ctx.Request.Body = stdhttp.MaxBytesReader(ctx.Writer, ctx.Request.Body, maxAvatarUploadSize)

	fileHeader, err := ctx.FormFile("file")
	if err != nil {
		httperr.BadRequest(ctx, "file is required")
		return
	}

	ext := strings.ToLower(filepath.Ext(fileHeader.Filename))
	switch ext {
	case ".jpg", ".jpeg", ".png", ".webp":
	default:
		httperr.BadRequest(ctx, "only jpg, png and webp images are allowed")
		return
	}

	if err := os.MkdirAll(h.uploadDir, 0o755); err != nil {
		h.logger.Error("failed to create upload directory", zap.Error(err))
		httperr.Internal(ctx, "internal server error")
		return
	}

	filename := fmt.Sprintf("avatar-%d-%d%s", userID, time.Now().UnixNano(), ext)
	targetPath := filepath.Join(h.uploadDir, filename)

	if err := ctx.SaveUploadedFile(fileHeader, targetPath); err != nil {
		h.logger.Error("failed to save avatar", zap.Error(err))
		httperr.Internal(ctx, "internal server error")
		return
	}

	avatarURL := "/auth-uploads/" + filename

	user, err := h.authUsecase.UpdateAvatar(ctx.Request.Context(), userID, avatarURL)
	if err != nil {
		h.logger.Error("failed to update avatar url", zap.Error(err))
		httperr.Internal(ctx, "internal server error")
		return
	}

	ctx.JSON(stdhttp.StatusCreated, uploadResponse{URL: user.AvatarURL})
}
