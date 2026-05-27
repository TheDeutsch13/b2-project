package http

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	commonjwt "github.com/TheDeutsch13/b2-common/jwt"
	"github.com/TheDeutsch13/b2-project/services/auth-service/internal/domain"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestAuthHandler_GetMe_Success(t *testing.T) {
	repo := new(mockUserRepository)
	refreshRepo := new(mockRefreshTokenRepository)
	router := setupTestRouter(t, repo, refreshRepo)

	jwtManager := commonjwt.NewManager("test-secret", time.Hour)
	token, _ := jwtManager.Generate(1, "test@example.com", domain.RoleUser)

	repo.On("GetByID", mock.Anything, int64(1)).
		Return(&domain.User{
			ID:    1,
			Email: "test@example.com",
			Role:  domain.RoleUser,
		}, nil).Once()

	req := httptest.NewRequest(http.MethodGet, "/api/auth/me", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Contains(t, rec.Body.String(), "test@example.com")
}

func TestAuthHandler_GetMe_Unauthorized(t *testing.T) {
	repo := new(mockUserRepository)
	refreshRepo := new(mockRefreshTokenRepository)
	router := setupTestRouter(t, repo, refreshRepo)

	req := httptest.NewRequest(http.MethodGet, "/api/auth/me", nil)
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusUnauthorized, rec.Code)
}
