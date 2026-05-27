package http

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	commonjwt "github.com/TheDeutsch13/b2-common/jwt"

	"github.com/TheDeutsch13/b2-project/services/auth-service/internal/domain"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestAuthHandler_Refresh_Success(t *testing.T) {
	repo := new(mockUserRepository)
	refreshRepo := new(mockRefreshTokenRepository)
	router := setupTestRouter(t, repo, refreshRepo)

	refreshRepo.On("GetByToken", mock.Anything, "refresh-token").
		Return(&domain.RefreshToken{
			UserID:    1,
			Token:     "refresh-token",
			ExpiresAt: time.Now().Add(time.Hour),
		}, nil).Once()
	repo.On("GetByID", mock.Anything, int64(1)).
		Return(&domain.User{ID: 1, Email: "test@example.com", Role: domain.RoleUser}, nil).Once()
	refreshRepo.On("DeleteByToken", mock.Anything, "refresh-token").Return(nil).Once()
	refreshRepo.On("Create", mock.Anything, mock.AnythingOfType("*domain.RefreshToken")).Return(nil).Once()

	body := []byte(`{"refresh_token":"refresh-token"}`)
	req := httptest.NewRequest(http.MethodPost, "/api/auth/refresh", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
}

func TestAuthHandler_Me_Success(t *testing.T) {
	repo := new(mockUserRepository)
	refreshRepo := new(mockRefreshTokenRepository)
	router := setupTestRouter(t, repo, refreshRepo)

	jwtManager := commonjwt.NewManager("test-secret", time.Hour)
	token, _ := jwtManager.Generate(1, "user@example.com", "user")

	repo.On("GetByID", mock.Anything, int64(1)).
		Return(&domain.User{ID: 1, Email: "user@example.com", Role: domain.RoleUser}, nil).Once()

	req := httptest.NewRequest(http.MethodGet, "/api/auth/me", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
}

func TestAuthHandler_Me_Unauthorized(t *testing.T) {
	repo := new(mockUserRepository)
	refreshRepo := new(mockRefreshTokenRepository)
	router := setupTestRouter(t, repo, refreshRepo)

	req := httptest.NewRequest(http.MethodGet, "/api/auth/me", nil)
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusUnauthorized, rec.Code)
}
