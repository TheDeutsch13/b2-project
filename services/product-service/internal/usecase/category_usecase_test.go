package usecase

import (
	"context"
	"testing"

	"github.com/TheDeutsch13/b2-project/services/product-service/internal/domain"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type MockCategoryRepository struct {
	mock.Mock
}

func (m *MockCategoryRepository) List(ctx context.Context) ([]domain.Category, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]domain.Category), args.Error(1)
}

func (m *MockCategoryRepository) Create(ctx context.Context, name string) (*domain.Category, error) {
	args := m.Called(ctx, name)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Category), args.Error(1)
}

func TestCategoryUsecase_List_Success(t *testing.T) {
	repo := new(MockCategoryRepository)
	usecase := NewCategoryUsecase(repo)

	repo.On("List", mock.Anything).Return([]domain.Category{{ID: 1, Name: "GPU"}}, nil).Once()

	categories, err := usecase.List(context.Background())

	assert.NoError(t, err)
	assert.Len(t, categories, 1)
}

func TestCategoryUsecase_Create_InvalidInput(t *testing.T) {
	usecase := NewCategoryUsecase(new(MockCategoryRepository))

	category, err := usecase.Create(context.Background(), "  ")

	assert.Nil(t, category)
	assert.ErrorIs(t, err, ErrInvalidInput)
}
