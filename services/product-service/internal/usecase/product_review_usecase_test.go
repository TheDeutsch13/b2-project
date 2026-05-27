package usecase

import (
	"context"
	"testing"

	"github.com/TheDeutsch13/b2-project/services/product-service/internal/domain"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type MockOrderReviewRepository struct {
	mock.Mock
}

func (m *MockOrderReviewRepository) UserHasReceivedProduct(
	ctx context.Context,
	userID int64,
	productID int64,
) (bool, error) {
	args := m.Called(ctx, userID, productID)
	return args.Bool(0), args.Error(1)
}

type MockProductReviewRepository struct {
	mock.Mock
}

func (m *MockProductReviewRepository) ListReviewsByUserID(
	ctx context.Context,
	userID int64,
) ([]domain.UserProductReview, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]domain.UserProductReview), args.Error(1)
}

func (m *MockProductReviewRepository) ListAllReviews(
	ctx context.Context,
	filter domain.ReviewListFilter,
) ([]domain.AdminProductReview, error) {
	args := m.Called(ctx, filter)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]domain.AdminProductReview), args.Error(1)
}

func TestProductUsecase_UpsertUserReview_Success(t *testing.T) {
	productRepo := new(MockProductRepository)
	orderRepo := new(MockOrderReviewRepository)
	usecase := NewProductUsecase(productRepo, orderRepo, nil)

	orderRepo.On("UserHasReceivedProduct", mock.Anything, int64(1), int64(10)).
		Return(true, nil).Once()
	productRepo.On("GetByID", mock.Anything, int64(10)).
		Return(&domain.Product{ID: 10, Title: "Mouse", Reviews: []domain.ProductReview{}}, nil).Once()
	productRepo.On("Update", mock.Anything, mock.AnythingOfType("*domain.Product")).
		Return(&domain.Product{ID: 10, Title: "Mouse", RatingCount: 1}, nil).Once()

	product, err := usecase.UpsertUserReview(
		context.Background(),
		1,
		10,
		"Ivan",
		5,
		"Great mouse",
	)

	assert.NoError(t, err)
	assert.Equal(t, int64(10), product.ID)
}

func TestProductUsecase_UpsertUserReview_NotAllowed(t *testing.T) {
	productRepo := new(MockProductRepository)
	orderRepo := new(MockOrderReviewRepository)
	usecase := NewProductUsecase(productRepo, orderRepo, nil)

	orderRepo.On("UserHasReceivedProduct", mock.Anything, int64(1), int64(10)).
		Return(false, nil).Once()

	product, err := usecase.UpsertUserReview(
		context.Background(),
		1,
		10,
		"Ivan",
		5,
		"Great mouse",
	)

	assert.Nil(t, product)
	assert.ErrorIs(t, err, domain.ErrReviewNotAllowed)
}

func TestProductUsecase_ListUserReviews_EmptyRepo(t *testing.T) {
	usecase := NewProductUsecase(new(MockProductRepository), nil, nil)

	reviews, err := usecase.ListUserReviews(context.Background(), 1)

	assert.NoError(t, err)
	assert.Empty(t, reviews)
}

func TestProductUsecase_ListAllReviews_WithRepo(t *testing.T) {
	productRepo := new(MockProductRepository)
	reviewRepo := new(MockProductReviewRepository)
	usecase := NewProductUsecase(productRepo, nil, reviewRepo)

	expected := []domain.AdminProductReview{{ProductID: 1, Rating: 5, Text: "Nice"}}
	reviewRepo.On("ListAllReviews", mock.Anything, domain.ReviewListFilter{Rating: 5}).
		Return(expected, nil).Once()

	reviews, err := usecase.ListAllReviews(context.Background(), domain.ReviewListFilter{Rating: 5})

	assert.NoError(t, err)
	assert.Len(t, reviews, 1)
}

func TestProductUsecase_DeleteUserReview_Success(t *testing.T) {
	productRepo := new(MockProductRepository)
	usecase := NewProductUsecase(productRepo, nil, nil)

	productRepo.On("GetByID", mock.Anything, int64(5)).
		Return(&domain.Product{
			ID: 5,
			Reviews: []domain.ProductReview{
				{UserID: 1, Author: "Ivan", Rating: 4, Text: "ok"},
			},
		}, nil).Once()
	productRepo.On("Update", mock.Anything, mock.AnythingOfType("*domain.Product")).
		Return(&domain.Product{ID: 5, Reviews: []domain.ProductReview{}}, nil).Once()

	product, err := usecase.DeleteUserReview(context.Background(), 1, 5)

	assert.NoError(t, err)
	assert.Equal(t, int64(5), product.ID)
}
