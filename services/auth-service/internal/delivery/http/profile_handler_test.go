package http

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/TheDeutsch13/b2-project/services/auth-service/internal/domain"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"golang.org/x/crypto/bcrypt"
)

func TestAuthHandler_UpdateProfile_Success(t *testing.T) {
	repo := new(mockUserRepository)
	refreshRepo := new(mockRefreshTokenRepository)
	router := setupTestRouter(t, repo, refreshRepo)

	repo.On("UpdateProfile", mock.Anything, int64(1), mock.AnythingOfType("domain.UserProfileInput")).
		Return(&domain.User{
			ID:        1,
			Email:     "user@example.com",
			Role:      domain.RoleUser,
			FirstName: "Ivan",
			LastName:  "Petrov",
		}, nil).Once()

	body := []byte(`{"first_name":"Ivan","last_name":"Petrov","gender":"male"}`)
	req := httptest.NewRequest(http.MethodPatch, "/api/auth/profile", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", authBearerToken(1, "user@example.com", domain.RoleUser))
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Contains(t, rec.Body.String(), "Ivan")
}

func TestAuthHandler_UpdateProfile_InvalidBirthDate(t *testing.T) {
	repo := new(mockUserRepository)
	refreshRepo := new(mockRefreshTokenRepository)
	router := setupTestRouter(t, repo, refreshRepo)

	body := []byte(`{"birth_date":"not-a-date"}`)
	req := httptest.NewRequest(http.MethodPatch, "/api/auth/profile", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", authBearerToken(1, "user@example.com", domain.RoleUser))
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusBadRequest, rec.Code)
}

func TestAuthHandler_ListUsers_Admin(t *testing.T) {
	repo := new(mockUserRepository)
	refreshRepo := new(mockRefreshTokenRepository)
	router := setupTestRouter(t, repo, refreshRepo)

	repo.On("List", mock.Anything).
		Return([]domain.User{
			{ID: 1, Email: "admin@example.com", Role: domain.RoleAdmin},
			{ID: 2, Email: "user@example.com", Role: domain.RoleUser},
		}, nil).Once()

	req := httptest.NewRequest(http.MethodGet, "/api/auth/users", nil)
	req.Header.Set("Authorization", authBearerToken(1, "admin@example.com", domain.RoleAdmin))
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)

	var users []map[string]interface{}
	assert.NoError(t, json.Unmarshal(rec.Body.Bytes(), &users))
	assert.Len(t, users, 2)
}

func TestAuthHandler_ListUsers_Forbidden(t *testing.T) {
	repo := new(mockUserRepository)
	refreshRepo := new(mockRefreshTokenRepository)
	router := setupTestRouter(t, repo, refreshRepo)

	req := httptest.NewRequest(http.MethodGet, "/api/auth/users", nil)
	req.Header.Set("Authorization", authBearerToken(2, "user@example.com", domain.RoleUser))
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusForbidden, rec.Code)
}

func TestAuthHandler_UpdateUserRole_Success(t *testing.T) {
	repo := new(mockUserRepository)
	refreshRepo := new(mockRefreshTokenRepository)
	router := setupTestRouter(t, repo, refreshRepo)

	repo.On("GetByID", mock.Anything, int64(1)).
		Return(&domain.User{ID: 1, Email: "admin@example.com", Role: domain.RoleAdmin}, nil).Once()
	repo.On("UpdateRole", mock.Anything, int64(2), domain.RoleModerator).
		Return(&domain.User{ID: 2, Email: "mod@example.com", Role: domain.RoleModerator}, nil).Once()

	body := []byte(`{"role":"moderator"}`)
	req := httptest.NewRequest(http.MethodPatch, "/api/auth/users/2/role", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", authBearerToken(1, "admin@example.com", domain.RoleAdmin))
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Contains(t, rec.Body.String(), "moderator")
}

func TestAuthHandler_UpdateUserRole_SelfChangeForbidden(t *testing.T) {
	repo := new(mockUserRepository)
	refreshRepo := new(mockRefreshTokenRepository)
	router := setupTestRouter(t, repo, refreshRepo)

	body := []byte(`{"role":"admin"}`)
	req := httptest.NewRequest(http.MethodPatch, "/api/auth/users/1/role", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", authBearerToken(1, "admin@example.com", domain.RoleAdmin))
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusBadRequest, rec.Code)
}

func TestAuthHandler_ListPublicUsers_Success(t *testing.T) {
	repo := new(mockUserRepository)
	refreshRepo := new(mockRefreshTokenRepository)
	router := setupTestRouter(t, repo, refreshRepo)

	repo.On("ListPublicByIDs", mock.Anything, []int64{1, 2}).
		Return([]domain.User{
			{ID: 1, Email: "a@example.com", Nickname: "Alpha"},
			{ID: 2, Email: "b@example.com", Nickname: "Beta"},
		}, nil).Once()

	req := httptest.NewRequest(http.MethodGet, "/api/auth/users/public?ids=1,2", nil)
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Contains(t, rec.Body.String(), "Alpha")
}

func TestAuthHandler_ListPublicUsers_EmptyIDs(t *testing.T) {
	repo := new(mockUserRepository)
	refreshRepo := new(mockRefreshTokenRepository)
	router := setupTestRouter(t, repo, refreshRepo)

	req := httptest.NewRequest(http.MethodGet, "/api/auth/users/public", nil)
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Equal(t, "[]", rec.Body.String())
}

func TestAuthHandler_ListPublicUsers_InvalidIDs(t *testing.T) {
	repo := new(mockUserRepository)
	refreshRepo := new(mockRefreshTokenRepository)
	router := setupTestRouter(t, repo, refreshRepo)

	req := httptest.NewRequest(http.MethodGet, "/api/auth/users/public?ids=abc", nil)
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusBadRequest, rec.Code)
}

func TestAuthHandler_Login_InvalidCredentials(t *testing.T) {
	repo := new(mockUserRepository)
	refreshRepo := new(mockRefreshTokenRepository)
	router := setupTestRouter(t, repo, refreshRepo)

	passwordHash, _ := bcrypt.GenerateFromPassword([]byte("secret"), bcrypt.DefaultCost)
	repo.On("GetByEmail", mock.Anything, "user@example.com").
		Return(&domain.User{
			ID:           1,
			Email:        "user@example.com",
			PasswordHash: string(passwordHash),
			Role:         domain.RoleUser,
		}, nil).Once()

	body := []byte(`{"email":"user@example.com","password":"wrong"}`)
	req := httptest.NewRequest(http.MethodPost, "/api/auth/login", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusUnauthorized, rec.Code)
}
