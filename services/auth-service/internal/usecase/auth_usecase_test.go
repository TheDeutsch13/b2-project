package usecase

import (
	"context"
	"testing"
	"time"

	commonjwt "github.com/TheDeutsch13/b2-common/jwt"
	"github.com/TheDeutsch13/b2-project/services/auth-service/internal/domain"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"golang.org/x/crypto/bcrypt"
)

type MockUserRepository struct {
	mock.Mock
}

func (m *MockUserRepository) ExistsByEmail(ctx context.Context, email string) (bool, error) {
	args := m.Called(ctx, email)
	return args.Bool(0), args.Error(1)
}

func (m *MockUserRepository) Create(ctx context.Context, user *domain.User) (*domain.User, error) {
	args := m.Called(ctx, user)

	if args.Get(0) == nil {
		return nil, args.Error(1)
	}

	return args.Get(0).(*domain.User), args.Error(1)
}

func (m *MockUserRepository) GetByEmail(ctx context.Context, email string) (*domain.User, error) {
	args := m.Called(ctx, email)

	if args.Get(0) == nil {
		return nil, args.Error(1)
	}

	return args.Get(0).(*domain.User), args.Error(1)
}

func (m *MockUserRepository) GetByID(ctx context.Context, id int64) (*domain.User, error) {
	args := m.Called(ctx, id)

	if args.Get(0) == nil {
		return nil, args.Error(1)
	}

	return args.Get(0).(*domain.User), args.Error(1)
}

func (m *MockUserRepository) UpdateProfile(
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

func (m *MockUserRepository) UpdateAvatar(
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

func (m *MockUserRepository) List(ctx context.Context) ([]domain.User, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]domain.User), args.Error(1)
}

func (m *MockUserRepository) ListPublicByIDs(ctx context.Context, ids []int64) ([]domain.User, error) {
	args := m.Called(ctx, ids)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]domain.User), args.Error(1)
}

type MockRefreshTokenRepository struct {
	mock.Mock
}

func (m *MockRefreshTokenRepository) Create(ctx context.Context, token *domain.RefreshToken) error {
	args := m.Called(ctx, token)
	return args.Error(0)
}

func (m *MockRefreshTokenRepository) GetByToken(ctx context.Context, token string) (*domain.RefreshToken, error) {
	args := m.Called(ctx, token)

	if args.Get(0) == nil {
		return nil, args.Error(1)
	}

	return args.Get(0).(*domain.RefreshToken), args.Error(1)
}

func (m *MockRefreshTokenRepository) DeleteByToken(ctx context.Context, token string) error {
	args := m.Called(ctx, token)
	return args.Error(0)
}

func newTestAuthUsecase(userRepo *MockUserRepository, refreshRepo *MockRefreshTokenRepository) *AuthUsecase {
	jwtManager := commonjwt.NewManager("test-secret", time.Hour)
	return NewAuthUsecase(userRepo, refreshRepo, jwtManager, 24*time.Hour)
}

func TestAuthUsecase_Register_Success(t *testing.T) {
	repo := new(MockUserRepository)
	refreshRepo := new(MockRefreshTokenRepository)
	authUsecase := newTestAuthUsecase(repo, refreshRepo)

	email := "test@example.com"
	password := "123456"

	repo.On("ExistsByEmail", mock.Anything, email).Return(false, nil).Once()
	repo.On("Create", mock.Anything, mock.MatchedBy(func(user *domain.User) bool {
		if user.Email != email || user.Role != domain.RoleUser {
			return false
		}

		return bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)) == nil
	})).Return(&domain.User{ID: 1, Email: email, Role: domain.RoleUser}, nil).Once()

	user, err := authUsecase.Register(context.Background(), email, password)

	assert.NoError(t, err)
	assert.Equal(t, int64(1), user.ID)
	repo.AssertExpectations(t)
}

func TestAuthUsecase_Register_InvalidInput(t *testing.T) {
	repo := new(MockUserRepository)
	refreshRepo := new(MockRefreshTokenRepository)
	authUsecase := newTestAuthUsecase(repo, refreshRepo)

	user, err := authUsecase.Register(context.Background(), "", "123")

	assert.Nil(t, user)
	assert.ErrorIs(t, err, ErrInvalidInput)
}

func TestAuthUsecase_Register_EmailAlreadyExists(t *testing.T) {
	repo := new(MockUserRepository)
	refreshRepo := new(MockRefreshTokenRepository)
	authUsecase := newTestAuthUsecase(repo, refreshRepo)

	email := "test@example.com"

	repo.On("ExistsByEmail", mock.Anything, email).Return(true, nil).Once()

	user, err := authUsecase.Register(context.Background(), email, "123456")

	assert.Nil(t, user)
	assert.ErrorIs(t, err, ErrEmailAlreadyExists)
}

func TestAuthUsecase_Login_Success(t *testing.T) {
	repo := new(MockUserRepository)
	refreshRepo := new(MockRefreshTokenRepository)
	authUsecase := newTestAuthUsecase(repo, refreshRepo)

	passwordHash, _ := bcrypt.GenerateFromPassword([]byte("123456"), bcrypt.DefaultCost)
	user := &domain.User{
		ID:           1,
		Email:        "test@example.com",
		PasswordHash: string(passwordHash),
		Role:         domain.RoleUser,
	}

	repo.On("GetByEmail", mock.Anything, user.Email).Return(user, nil).Once()
	refreshRepo.On("Create", mock.Anything, mock.AnythingOfType("*domain.RefreshToken")).Return(nil).Once()

	tokens, err := authUsecase.Login(context.Background(), user.Email, "123456")

	assert.NoError(t, err)
	assert.NotEmpty(t, tokens.AccessToken)
	assert.NotEmpty(t, tokens.RefreshToken)
}

func TestAuthUsecase_Login_InvalidCredentials(t *testing.T) {
	repo := new(MockUserRepository)
	refreshRepo := new(MockRefreshTokenRepository)
	authUsecase := newTestAuthUsecase(repo, refreshRepo)

	repo.On("GetByEmail", mock.Anything, "missing@example.com").Return(nil, domain.ErrUserNotFound).Once()

	tokens, err := authUsecase.Login(context.Background(), "missing@example.com", "123456")

	assert.Nil(t, tokens)
	assert.ErrorIs(t, err, ErrInvalidCredentials)
}

func TestAuthUsecase_Refresh_Success(t *testing.T) {
	repo := new(MockUserRepository)
	refreshRepo := new(MockRefreshTokenRepository)
	authUsecase := newTestAuthUsecase(repo, refreshRepo)

	user := &domain.User{ID: 1, Email: "test@example.com", Role: domain.RoleUser}
	refreshToken := &domain.RefreshToken{
		UserID:    1,
		Token:     "refresh-token",
		ExpiresAt: time.Now().Add(time.Hour),
	}

	refreshRepo.On("GetByToken", mock.Anything, "refresh-token").Return(refreshToken, nil).Once()
	repo.On("GetByID", mock.Anything, int64(1)).Return(user, nil).Once()
	refreshRepo.On("DeleteByToken", mock.Anything, "refresh-token").Return(nil).Once()
	refreshRepo.On("Create", mock.Anything, mock.AnythingOfType("*domain.RefreshToken")).Return(nil).Once()

	tokens, err := authUsecase.Refresh(context.Background(), "refresh-token")

	assert.NoError(t, err)
	assert.NotEmpty(t, tokens.AccessToken)
}
