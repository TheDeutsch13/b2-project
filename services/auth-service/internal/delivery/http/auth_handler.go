package http

import (
	"errors"
	stdhttp "net/http"

	"github.com/TheDeutsch13/b2-common/httperr"
	"github.com/TheDeutsch13/b2-project/services/auth-service/internal/usecase"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

type AuthHandler struct {
	authUsecase *usecase.AuthUsecase
	logger      *zap.Logger
}

func NewAuthHandler(authUsecase *usecase.AuthUsecase, logger *zap.Logger) *AuthHandler {
	return &AuthHandler{
		authUsecase: authUsecase,
		logger:      logger,
	}
}

type registerRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type loginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type refreshRequest struct {
	RefreshToken string `json:"refresh_token"`
}

type authResponse struct {
	AccessToken  string       `json:"access_token"`
	RefreshToken string       `json:"refresh_token"`
	User         userResponse `json:"user"`
}

type registerResponse struct {
	ID    int64  `json:"id"`
	Email string `json:"email"`
	Role  string `json:"role"`
}

// Register godoc
// @Summary Register user
// @Description Creates a new user account
// @Tags auth
// @Accept json
// @Produce json
// @Param request body registerRequest true "Register request"
// @Success 201 {object} registerResponse
// @Failure 400 {object} map[string]string
// @Failure 409 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/auth/register [post]
func (h *AuthHandler) Register(ctx *gin.Context) {
	var req registerRequest

	if err := ctx.ShouldBindJSON(&req); err != nil {
		httperr.BadRequest(ctx, "invalid request body")
		return
	}

	user, err := h.authUsecase.Register(ctx.Request.Context(), req.Email, req.Password)
	if err != nil {
		if errors.Is(err, usecase.ErrInvalidInput) {
			httperr.BadRequest(ctx, "email and password are required, password must be at least 6 characters")
			return
		}

		if errors.Is(err, usecase.ErrEmailAlreadyExists) {
			httperr.Conflict(ctx, "email already exists")
			return
		}

		h.logger.Error("failed to register user", zap.Error(err))
		httperr.Internal(ctx, "internal server error")
		return
	}

	ctx.JSON(stdhttp.StatusCreated, registerResponse{
		ID:    user.ID,
		Email: user.Email,
		Role:  user.Role,
	})
}

// Login godoc
// @Summary Login user
// @Description Authenticates user and returns JWT tokens
// @Tags auth
// @Accept json
// @Produce json
// @Param request body loginRequest true "Login request"
// @Success 200 {object} authResponse
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/auth/login [post]
func (h *AuthHandler) Login(ctx *gin.Context) {
	var req loginRequest

	if err := ctx.ShouldBindJSON(&req); err != nil {
		httperr.BadRequest(ctx, "invalid request body")
		return
	}

	tokens, err := h.authUsecase.Login(ctx.Request.Context(), req.Email, req.Password)
	if err != nil {
		if errors.Is(err, usecase.ErrInvalidInput) {
			httperr.BadRequest(ctx, "email and password are required")
			return
		}

		if errors.Is(err, usecase.ErrInvalidCredentials) {
			httperr.Unauthorized(ctx, "invalid email or password")
			return
		}

		if errors.Is(err, usecase.ErrAuthStorage) {
			h.logger.Error("login storage error", zap.Error(err))
			ctx.JSON(stdhttp.StatusServiceUnavailable, gin.H{
				"error": "Сервис авторизации не готов. Перезапустите auth-service (нужна миграция профиля).",
			})
			return
		}

		h.logger.Error("failed to login user", zap.Error(err))
		httperr.Internal(ctx, "internal server error")
		return
	}

	ctx.JSON(stdhttp.StatusOK, toAuthResponse(tokens))
}

// Refresh godoc
// @Summary Refresh tokens
// @Description Issues new access and refresh tokens
// @Tags auth
// @Accept json
// @Produce json
// @Param request body refreshRequest true "Refresh request"
// @Success 200 {object} authResponse
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/auth/refresh [post]
func (h *AuthHandler) Refresh(ctx *gin.Context) {
	var req refreshRequest

	if err := ctx.ShouldBindJSON(&req); err != nil {
		httperr.BadRequest(ctx, "invalid request body")
		return
	}

	tokens, err := h.authUsecase.Refresh(ctx.Request.Context(), req.RefreshToken)
	if err != nil {
		if errors.Is(err, usecase.ErrInvalidRefresh) {
			httperr.Unauthorized(ctx, "invalid refresh token")
			return
		}

		h.logger.Error("failed to refresh token", zap.Error(err))
		httperr.Internal(ctx, "internal server error")
		return
	}

	ctx.JSON(stdhttp.StatusOK, toAuthResponse(tokens))
}

func toAuthResponse(tokens *usecase.TokenPair) authResponse {
	return authResponse{
		AccessToken:  tokens.AccessToken,
		RefreshToken: tokens.RefreshToken,
		User:         toUserResponse(tokens.User),
	}
}
