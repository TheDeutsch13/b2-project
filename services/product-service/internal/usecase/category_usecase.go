package usecase

import (
	"context"
	"errors"
	"strings"

	"github.com/TheDeutsch13/b2-project/services/product-service/internal/domain"
	"github.com/TheDeutsch13/b2-project/services/product-service/internal/repository/postgres"
)

type CategoryRepository interface {
	List(ctx context.Context) ([]domain.Category, error)
	Create(ctx context.Context, name string) (*domain.Category, error)
}

type CategoryUsecase struct {
	categoryRepo CategoryRepository
}

func NewCategoryUsecase(categoryRepo CategoryRepository) *CategoryUsecase {
	return &CategoryUsecase{categoryRepo: categoryRepo}
}

func (u *CategoryUsecase) List(ctx context.Context) ([]domain.Category, error) {
	return u.categoryRepo.List(ctx)
}

func (u *CategoryUsecase) Create(ctx context.Context, name string) (*domain.Category, error) {
	name = strings.TrimSpace(name)
	if name == "" {
		return nil, ErrInvalidInput
	}

	category, err := u.categoryRepo.Create(ctx, name)
	if err != nil {
		if errors.Is(err, postgres.ErrUniqueViolation) {
			return nil, ErrCategoryAlreadyExists
		}

		return nil, err
	}

	return category, nil
}
