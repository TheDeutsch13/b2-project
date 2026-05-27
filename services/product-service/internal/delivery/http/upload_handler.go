package http

import (
	"fmt"
	stdhttp "net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/TheDeutsch13/b2-common/httperr"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

const maxUploadSize = 5 << 20

type UploadHandler struct {
	uploadDir string
	logger    *zap.Logger
}

func NewUploadHandler(uploadDir string, logger *zap.Logger) *UploadHandler {
	return &UploadHandler{
		uploadDir: uploadDir,
		logger:    logger,
	}
}

type uploadResponse struct {
	URL string `json:"url"`
}

// Upload godoc
// @Summary Upload product image
// @Tags uploads
// @Accept multipart/form-data
// @Produce json
// @Security BearerAuth
// @Param file formData file true "Image file"
// @Success 201 {object} uploadResponse
// @Router /api/upload [post]
func (h *UploadHandler) Upload(ctx *gin.Context) {
	ctx.Request.Body = stdhttp.MaxBytesReader(ctx.Writer, ctx.Request.Body, maxUploadSize)

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

	filename := fmt.Sprintf("%d-%s%s", time.Now().UnixNano(), uuid.NewString(), ext)
	targetPath := filepath.Join(h.uploadDir, filename)

	if err := ctx.SaveUploadedFile(fileHeader, targetPath); err != nil {
		h.logger.Error("failed to save uploaded file", zap.Error(err))
		httperr.Internal(ctx, "internal server error")
		return
	}

	ctx.JSON(stdhttp.StatusCreated, uploadResponse{
		URL: "/uploads/" + filename,
	})
}
