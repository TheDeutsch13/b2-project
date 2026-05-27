package usecase

import (
	"context"
	"strings"
	"time"

	"github.com/TheDeutsch13/b2-project/services/auth-service/internal/domain"
)

func (u *AuthUsecase) GetProfile(ctx context.Context, userID int64) (*domain.User, error) {
	if userID == 0 {
		return nil, ErrInvalidInput
	}

	return u.userRepo.GetByID(ctx, userID)
}

func (u *AuthUsecase) UpdateProfile(
	ctx context.Context,
	userID int64,
	input domain.UserProfileInput,
) (*domain.User, error) {
	if userID == 0 {
		return nil, ErrInvalidInput
	}

	input.FirstName = strings.TrimSpace(input.FirstName)
	input.LastName = strings.TrimSpace(input.LastName)
	input.Nickname = strings.TrimSpace(input.Nickname)
	input.Phone = strings.TrimSpace(input.Phone)
	input.Gender = strings.TrimSpace(input.Gender)

	if input.Gender != "" &&
		input.Gender != domain.GenderMale &&
		input.Gender != domain.GenderFemale {
		return nil, ErrInvalidInput
	}

	return u.userRepo.UpdateProfile(ctx, userID, input)
}

func (u *AuthUsecase) UpdateAvatar(
	ctx context.Context,
	userID int64,
	avatarURL string,
) (*domain.User, error) {
	if userID == 0 {
		return nil, ErrInvalidInput
	}

	avatarURL = strings.TrimSpace(avatarURL)

	return u.userRepo.UpdateAvatar(ctx, userID, avatarURL)
}

func (u *AuthUsecase) ListUsers(ctx context.Context) ([]domain.User, error) {
	return u.userRepo.List(ctx)
}

func (u *AuthUsecase) ListPublicUsersByIDs(ctx context.Context, ids []int64) ([]domain.User, error) {
	return u.userRepo.ListPublicByIDs(ctx, ids)
}

func (u *AuthUsecase) UpdateUserRole(
	ctx context.Context,
	actorID int64,
	targetUserID int64,
	role string,
) (*domain.User, error) {
	if actorID == 0 || targetUserID == 0 {
		return nil, ErrInvalidInput
	}

	if actorID == targetUserID {
		return nil, ErrInvalidInput
	}

	role = strings.TrimSpace(role)
	if !domain.IsValidRole(role) {
		return nil, ErrInvalidInput
	}

	actor, err := u.userRepo.GetByID(ctx, actorID)
	if err != nil {
		return nil, err
	}

	if actor.Role != domain.RoleAdmin {
		return nil, ErrInvalidInput
	}

	return u.userRepo.UpdateRole(ctx, targetUserID, role)
}

func ParseBirthDate(value string) (*time.Time, error) {
	value = strings.TrimSpace(value)
	if value == "" {
		return nil, nil
	}

	parsed, err := time.Parse("2006-01-02", value)
	if err != nil {
		return nil, ErrInvalidInput
	}

	return &parsed, nil
}
