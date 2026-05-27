package usecase

import (
	"context"
	"testing"

	"github.com/TheDeutsch13/b2-project/services/product-service/internal/domain"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type MockProductRepository struct {
	mock.Mock
}

func (m *MockProductRepository) Create(ctx context.Context, product *domain.Product) (*domain.Product, error) {
	args := m.Called(ctx, product)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Product), args.Error(1)
}

func (m *MockProductRepository) List(ctx context.Context, categoryID *int64) ([]domain.Product, error) {
	args := m.Called(ctx, categoryID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]domain.Product), args.Error(1)
}

func (m *MockProductRepository) GetByID(ctx context.Context, id int64) (*domain.Product, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Product), args.Error(1)
}

func (m *MockProductRepository) Update(ctx context.Context, product *domain.Product) (*domain.Product, error) {
	args := m.Called(ctx, product)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Product), args.Error(1)
}

func (m *MockProductRepository) Delete(ctx context.Context, id int64) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockProductRepository) ExistsDuplicate(
	ctx context.Context,
	title string,
	brand string,
	categoryID *int64,
	variantKey string,
	excludeID int64,
) (bool, error) {
	args := m.Called(ctx, title, brand, categoryID, variantKey, excludeID)
	return args.Bool(0), args.Error(1)
}

func TestProductUsecase_Create_Success(t *testing.T) {
	repo := new(MockProductRepository)
	productUsecase := NewProductUsecase(repo, nil, nil)

	repo.On(
		"ExistsDuplicate",
		mock.Anything,
		"Test Product",
		"",
		(*int64)(nil),
		"стандарт",
		int64(0),
	).Return(false, nil).Once()
	repo.On("Create", mock.Anything, mock.AnythingOfType("*domain.Product")).
		Return(&domain.Product{ID: 1, Title: "Test Product", Price: 99.99}, nil).Once()

	product, err := productUsecase.Create(context.Background(), ProductInput{
		Title:       "Test Product",
		Description: "Description",
		Price:       99.99,
	})

	assert.NoError(t, err)
	assert.Equal(t, int64(1), product.ID)
}

func TestProductUsecase_Create_InvalidInput(t *testing.T) {
	repo := new(MockProductRepository)
	productUsecase := NewProductUsecase(repo, nil, nil)

	product, err := productUsecase.Create(context.Background(), ProductInput{
		Title:       "",
		Description: "Description",
		Price:       99.99,
	})

	assert.Nil(t, product)
	assert.ErrorIs(t, err, ErrInvalidInput)
}

func TestProductUsecase_Create_InvalidPrice(t *testing.T) {
	repo := new(MockProductRepository)
	productUsecase := NewProductUsecase(repo, nil, nil)

	product, err := productUsecase.Create(context.Background(), ProductInput{
		Title:       "Test Product",
		Description: "Description",
		Price:       -10,
	})

	assert.Nil(t, product)
	assert.ErrorIs(t, err, ErrInvalidInput)
}

func TestProductUsecase_List_Success(t *testing.T) {
	repo := new(MockProductRepository)
	productUsecase := NewProductUsecase(repo, nil, nil)

	expectedProducts := []domain.Product{{ID: 1, Title: "Product 1"}, {ID: 2, Title: "Product 2"}}

	repo.On("List", mock.Anything, (*int64)(nil)).Return(expectedProducts, nil).Once()

	products, err := productUsecase.List(context.Background(), nil)

	assert.NoError(t, err)
	assert.Len(t, products, 2)
}

func TestProductUsecase_GetByID_Success(t *testing.T) {
	repo := new(MockProductRepository)
	productUsecase := NewProductUsecase(repo, nil, nil)

	repo.On("GetByID", mock.Anything, int64(1)).
		Return(&domain.Product{ID: 1, Title: "GPU"}, nil).Once()

	product, err := productUsecase.GetByID(context.Background(), 1)

	assert.NoError(t, err)
	assert.Equal(t, "GPU", product.Title)
}

func TestProductUsecase_GetByID_NotFound(t *testing.T) {
	repo := new(MockProductRepository)
	productUsecase := NewProductUsecase(repo, nil, nil)

	repo.On("GetByID", mock.Anything, int64(99)).
		Return(nil, domain.ErrProductNotFound).Once()

	product, err := productUsecase.GetByID(context.Background(), 99)

	assert.Nil(t, product)
	assert.ErrorIs(t, err, domain.ErrProductNotFound)
}

func TestProductUsecase_Delete_Success(t *testing.T) {
	repo := new(MockProductRepository)
	productUsecase := NewProductUsecase(repo, nil, nil)

	repo.On("Delete", mock.Anything, int64(1)).Return(nil).Once()

	err := productUsecase.Delete(context.Background(), 1)

	assert.NoError(t, err)
}

func TestProductUsecase_Update_Success(t *testing.T) {
	repo := new(MockProductRepository)
	productUsecase := NewProductUsecase(repo, nil, nil)

	repo.On("ExistsDuplicate", mock.Anything, "Updated", "", (*int64)(nil), "стандарт", int64(1)).
		Return(false, nil).Once()
	repo.On("Update", mock.Anything, mock.AnythingOfType("*domain.Product")).
		Return(&domain.Product{ID: 1, Title: "Updated", Price: 50}, nil).Once()

	product, err := productUsecase.Update(context.Background(), 1, ProductInput{
		Title:       "Updated",
		Description: "Desc",
		Price:       50,
	})

	assert.NoError(t, err)
	assert.Equal(t, "Updated", product.Title)
}

func TestProductUsecase_Create_Duplicate(t *testing.T) {
	repo := new(MockProductRepository)
	productUsecase := NewProductUsecase(repo, nil, nil)

	repo.On("ExistsDuplicate", mock.Anything, "Dup", "", (*int64)(nil), "стандарт", int64(0)).
		Return(true, nil).Once()

	product, err := productUsecase.Create(context.Background(), ProductInput{
		Title:       "Dup",
		Description: "Desc",
		Price:       10,
	})

	assert.Nil(t, product)
	assert.ErrorIs(t, err, ErrProductDuplicate)
}
