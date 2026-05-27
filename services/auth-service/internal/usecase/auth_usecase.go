package usecase

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"strings"
	"time"

	commonjwt "github.com/TheDeutsch13/b2-common/jwt"
	"github.com/TheDeutsch13/b2-project/services/auth-service/internal/domain"
	"golang.org/x/crypto/bcrypt"
)

func isUserNotFound(err error) bool {
	return errors.Is(err, domain.ErrUserNotFound)
}

var (
	ErrInvalidInput       = errors.New("invalid input")
	ErrEmailAlreadyExists = errors.New("email already exists")
	ErrInvalidCredentials = errors.New("invalid credentials")
	ErrInvalidRefresh     = errors.New("invalid refresh token")
	ErrAuthStorage        = errors.New("auth storage error")
)

type UserRepository interface {
	ExistsByEmail(ctx context.Context, email string) (bool, error)
	Create(ctx context.Context, user *domain.User) (*domain.User, error)
	GetByEmail(ctx context.Context, email string) (*domain.User, error)
	GetByID(ctx context.Context, id int64) (*domain.User, error)
	UpdateProfile(ctx context.Context, userID int64, input domain.UserProfileInput) (*domain.User, error)
	UpdateAvatar(ctx context.Context, userID int64, avatarURL string) (*domain.User, error)
	List(ctx context.Context) ([]domain.User, error)
	ListPublicByIDs(ctx context.Context, ids []int64) ([]domain.User, error)
	UpdateRole(ctx context.Context, userID int64, role string) (*domain.User, error)
}

type RefreshTokenRepository interface {
	Create(ctx context.Context, token *domain.RefreshToken) error
	GetByToken(ctx context.Context, token string) (*domain.RefreshToken, error)
	DeleteByToken(ctx context.Context, token string) error
}

type TokenPair struct {
	AccessToken  string
	RefreshToken string
	User         *domain.User
}

type AuthUsecase struct {
	userRepo         UserRepository
	refreshTokenRepo RefreshTokenRepository
	jwtManager       *commonjwt.Manager
	refreshTTL       time.Duration
}

func NewAuthUsecase(
	userRepo UserRepository,
	refreshTokenRepo RefreshTokenRepository,
	jwtManager *commonjwt.Manager,
	refreshTTL time.Duration,
) *AuthUsecase {
	return &AuthUsecase{
		userRepo:         userRepo,
		refreshTokenRepo: refreshTokenRepo,
		jwtManager:       jwtManager,
		refreshTTL:       refreshTTL,
	}
}

func (u *AuthUsecase) Register(ctx context.Context, email string, password string) (*domain.User, error) {
	email = strings.TrimSpace(email)

	if email == "" || password == "" {
		return nil, ErrInvalidInput
	}

	if len(password) < 6 {
		return nil, ErrInvalidInput
	}

	exists, err := u.userRepo.ExistsByEmail(ctx, email)
	if err != nil {
		return nil, err
	}

	if exists {
		return nil, ErrEmailAlreadyExists
	}

	passwordHash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	user := &domain.User{
		Email:        email,
		PasswordHash: string(passwordHash),
		Role:         domain.RoleUser,
	}

	return u.userRepo.Create(ctx, user)
}

func (u *AuthUsecase) Login(ctx context.Context, email string, password string) (*TokenPair, error) {
	email = strings.TrimSpace(email)

	if email == "" || password == "" {
		return nil, ErrInvalidInput
	}

	user, err := u.userRepo.GetByEmail(ctx, email)
	if err != nil {
		if isUserNotFound(err) {
			return nil, ErrInvalidCredentials
		}

		return nil, ErrAuthStorage
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)); err != nil {
		return nil, ErrInvalidCredentials
	}

	return u.issueTokens(ctx, user)
}

func (u *AuthUsecase) Refresh(ctx context.Context, refreshToken string) (*TokenPair, error) {
	refreshToken = strings.TrimSpace(refreshToken)
	if refreshToken == "" {
		return nil, ErrInvalidRefresh
	}

	storedToken, err := u.refreshTokenRepo.GetByToken(ctx, refreshToken)
	if err != nil {
		return nil, ErrInvalidRefresh
	}

	if time.Now().After(storedToken.ExpiresAt) {
		_ = u.refreshTokenRepo.DeleteByToken(ctx, refreshToken)
		return nil, ErrInvalidRefresh
	}

	user, err := u.userRepo.GetByID(ctx, storedToken.UserID)
	if err != nil {
		return nil, ErrInvalidRefresh
	}

	_ = u.refreshTokenRepo.DeleteByToken(ctx, refreshToken)

	return u.issueTokens(ctx, user)
}

func (u *AuthUsecase) issueTokens(ctx context.Context, user *domain.User) (*TokenPair, error) {
	accessToken, err := u.jwtManager.Generate(user.ID, user.Email, user.Role)
	if err != nil {
		return nil, err
	}

	refreshToken, err := generateRefreshToken()
	if err != nil {
		return nil, err
	}

	err = u.refreshTokenRepo.Create(ctx, &domain.RefreshToken{
		UserID:    user.ID,
		Token:     refreshToken,
		ExpiresAt: time.Now().Add(u.refreshTTL),
	})
	if err != nil {
		return nil, err
	}

	return &TokenPair{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		User:         user,
	}, nil
}

func generateRefreshToken() (string, error) {
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}

	return hex.EncodeToString(bytes), nil
}
