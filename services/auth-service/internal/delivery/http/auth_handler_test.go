package http

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	commonjwt "github.com/TheDeutsch13/b2-common/jwt"
	"github.com/TheDeutsch13/b2-project/services/auth-service/internal/domain"
	"github.com/TheDeutsch13/b2-project/services/auth-service/internal/usecase"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"go.uber.org/zap"
	"golang.org/x/crypto/bcrypt"
)

type mockUserRepository struct {
	mock.Mock
}

func (m *mockUserRepository) ExistsByEmail(ctx context.Context, email string) (bool, error) {
	args := m.Called(ctx, email)
	return args.Bool(0), args.Error(1)
}

func (m *mockUserRepository) Create(ctx context.Context, user *domain.User) (*domain.User, error) {
	args := m.Called(ctx, user)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.User), args.Error(1)
}

func (m *mockUserRepository) GetByEmail(ctx context.Context, email string) (*domain.User, error) {
	args := m.Called(ctx, email)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.User), args.Error(1)
}

func (m *mockUserRepository) GetByID(ctx context.Context, id int64) (*domain.User, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.User), args.Error(1)
}

func (m *mockUserRepository) UpdateProfile(
	ctx context.Context,
	userID int64,
	input domain.UserProfileInput,
) (*domain.User, error) {
	args := m.Called(ctx, userID, input)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.User), args.Error(1)
}

func (m *mockUserRepository) UpdateAvatar(
	ctx context.Context,
	userID int64,
	avatarURL string,
) (*domain.User, error) {
	args := m.Called(ctx, userID, avatarURL)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.User), args.Error(1)
}

func (m *mockUserRepository) List(ctx context.Context) ([]domain.User, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]domain.User), args.Error(1)
}

func (m *mockUserRepository) ListPublicByIDs(ctx context.Context, ids []int64) ([]domain.User, error) {
	args := m.Called(ctx, ids)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]domain.User), args.Error(1)
}

type mockRefreshTokenRepository struct {
	mock.Mock
}

func (m *mockRefreshTokenRepository) Create(ctx context.Context, token *domain.RefreshToken) error {
	return m.Called(ctx, token).Error(0)
}

func (m *mockRefreshTokenRepository) GetByToken(ctx context.Context, token string) (*domain.RefreshToken, error) {
	args := m.Called(ctx, token)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.RefreshToken), args.Error(1)
}

func (m *mockRefreshTokenRepository) DeleteByToken(ctx context.Context, token string) error {
	return m.Called(ctx, token).Error(0)
}

func setupTestRouter(
	t *testing.T,
	repo *mockUserRepository,
	refreshRepo *mockRefreshTokenRepository,
) *gin.Engine {
	gin.SetMode(gin.TestMode)

	logger := zap.NewNop()
	jwtManager := commonjwt.NewManager("test-secret", time.Hour)
	authUsecase := usecase.NewAuthUsecase(repo, refreshRepo, jwtManager, 24*time.Hour)
	authHandler := NewAuthHandler(authUsecase, logger)
	uploadDir := t.TempDir()
	uploadHandler := NewUploadHandler(authUsecase, uploadDir, logger)

	return NewRouter(logger, authHandler, uploadHandler, jwtManager, uploadDir)
}

func TestAuthHandler_Register_Success(t *testing.T) {
	repo := new(mockUserRepository)
	refreshRepo := new(mockRefreshTokenRepository)
	router := setupTestRouter(t, repo, refreshRepo)

	email := "test@example.com"

	repo.On("ExistsByEmail", mock.Anything, email).Return(false, nil).Once()
	repo.On("Create", mock.Anything, mock.AnythingOfType("*domain.User")).
		Return(&domain.User{ID: 1, Email: email, Role: domain.RoleUser}, nil).Once()

	body := []byte(`{"email":"test@example.com","password":"123456"}`)
	req := httptest.NewRequest(http.MethodPost, "/api/auth/register", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusCreated, rec.Code)

	var response map[string]interface{}
	assert.NoError(t, json.Unmarshal(rec.Body.Bytes(), &response))
	assert.Equal(t, float64(1), response["id"])
}

func TestAuthHandler_Login_Success(t *testing.T) {
	repo := new(mockUserRepository)
	refreshRepo := new(mockRefreshTokenRepository)
	router := setupTestRouter(t, repo, refreshRepo)

	passwordHash, _ := bcrypt.GenerateFromPassword([]byte("123456"), bcrypt.DefaultCost)
	user := &domain.User{
		ID:           1,
		Email:        "test@example.com",
		PasswordHash: string(passwordHash),
		Role:         domain.RoleUser,
	}

	repo.On("GetByEmail", mock.Anything, user.Email).Return(user, nil).Once()
	refreshRepo.On("Create", mock.Anything, mock.AnythingOfType("*domain.RefreshToken")).Return(nil).Once()

	body := []byte(`{"email":"test@example.com","password":"123456"}`)
	req := httptest.NewRequest(http.MethodPost, "/api/auth/login", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Contains(t, rec.Body.String(), "access_token")
}

func TestAuthHandler_Register_InvalidInput(t *testing.T) {
	repo := new(mockUserRepository)
	refreshRepo := new(mockRefreshTokenRepository)
	router := setupTestRouter(t, repo, refreshRepo)

	body := []byte(`{"email":"","password":"123"}`)
	req := httptest.NewRequest(http.MethodPost, "/api/auth/register", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusBadRequest, rec.Code)
}

func TestAuthHandler_Register_EmailAlreadyExists(t *testing.T) {
	repo := new(mockUserRepository)
	refreshRepo := new(mockRefreshTokenRepository)
	router := setupTestRouter(t, repo, refreshRepo)

	repo.On("ExistsByEmail", mock.Anything, "test@example.com").Return(true, nil).Once()

	body := []byte(`{"email":"test@example.com","password":"123456"}`)
	req := httptest.NewRequest(http.MethodPost, "/api/auth/register", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusConflict, rec.Code)
}

func TestAuthHandler_Health(t *testing.T) {
	repo := new(mockUserRepository)
	refreshRepo := new(mockRefreshTokenRepository)
	router := setupTestRouter(t, repo, refreshRepo)

	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Contains(t, rec.Body.String(), "auth-service")
}
