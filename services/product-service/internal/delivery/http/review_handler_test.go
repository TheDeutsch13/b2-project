package http

import (
	"bytes"
	stdhttp "net/http"
	"net/http/httptest"
	"testing"

	"github.com/TheDeutsch13/b2-project/services/product-service/internal/domain"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestProductHandler_UpsertReview_Success(t *testing.T) {
	productRepo := new(mockProductRepository)
	orderReviewRepo := new(mockOrderReviewRepository)
	router := setupTestRouterFull(routerTestDeps{
		productRepo:     productRepo,
		orderReviewRepo: orderReviewRepo,
	})

	orderReviewRepo.On("UserHasReceivedProduct", mock.Anything, int64(1), int64(10)).
		Return(true, nil).Once()
	productRepo.On("GetByID", mock.Anything, int64(10)).
		Return(&domain.Product{ID: 10, Title: "Mouse", Reviews: []domain.ProductReview{}}, nil).Once()
	productRepo.On("Update", mock.Anything, mock.AnythingOfType("*domain.Product")).
		Return(&domain.Product{ID: 10, Title: "Mouse", RatingCount: 1}, nil).Once()

	body := []byte(`{"author":"Ivan","rating":5,"text":"Great"}`)
	req := httptest.NewRequest(stdhttp.MethodPut, "/api/products/10/reviews", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", bearerToken(1, "user@example.com", "user"))
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	assert.Equal(t, stdhttp.StatusOK, rec.Code)
}

func TestProductHandler_UpsertReview_NotAllowed(t *testing.T) {
	productRepo := new(mockProductRepository)
	orderReviewRepo := new(mockOrderReviewRepository)
	router := setupTestRouterFull(routerTestDeps{
		productRepo:     productRepo,
		orderReviewRepo: orderReviewRepo,
	})

	orderReviewRepo.On("UserHasReceivedProduct", mock.Anything, int64(1), int64(10)).
		Return(false, nil).Once()

	body := []byte(`{"author":"Ivan","rating":5,"text":"Great"}`)
	req := httptest.NewRequest(stdhttp.MethodPut, "/api/products/10/reviews", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", bearerToken(1, "user@example.com", "user"))
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	assert.Equal(t, stdhttp.StatusForbidden, rec.Code)
}

func TestProductHandler_ListMyReviews_Success(t *testing.T) {
	reviewRepo := new(mockProductReviewRepository)
	router := setupTestRouterFull(routerTestDeps{reviewRepo: reviewRepo})

	reviewRepo.On("ListReviewsByUserID", mock.Anything, int64(1)).
		Return([]domain.UserProductReview{{
			ProductID:    5,
			ProductTitle: "Keyboard",
			Review:       domain.ProductReview{Author: "Ivan", Rating: 4, Text: "ok"},
		}}, nil).Once()

	req := httptest.NewRequest(stdhttp.MethodGet, "/api/reviews/my", nil)
	req.Header.Set("Authorization", bearerToken(1, "user@example.com", "user"))
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	assert.Equal(t, stdhttp.StatusOK, rec.Code)
	assert.Contains(t, rec.Body.String(), "Keyboard")
}

func TestProductHandler_ListAdminReviews_Success(t *testing.T) {
	reviewRepo := new(mockProductReviewRepository)
	router := setupTestRouterFull(routerTestDeps{reviewRepo: reviewRepo})

	reviewRepo.On("ListAllReviews", mock.Anything, domain.ReviewListFilter{Rating: 5}).
		Return([]domain.AdminProductReview{{
			ProductID:    1,
			ProductTitle: "GPU",
			Rating:       5,
			Text:         "Super",
		}}, nil).Once()

	req := httptest.NewRequest(stdhttp.MethodGet, "/api/reviews?rating=5", nil)
	req.Header.Set("Authorization", bearerToken(1, "admin@example.com", "admin"))
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	assert.Equal(t, stdhttp.StatusOK, rec.Code)
	assert.Contains(t, rec.Body.String(), "GPU")
}

func TestProductHandler_DeleteReview_Success(t *testing.T) {
	productRepo := new(mockProductRepository)
	router := setupTestRouterFull(routerTestDeps{productRepo: productRepo})

	productRepo.On("GetByID", mock.Anything, int64(7)).
		Return(&domain.Product{
			ID: 7,
			Reviews: []domain.ProductReview{
				{UserID: 1, Author: "Ivan", Rating: 3, Text: "ok"},
			},
		}, nil).Once()
	productRepo.On("Update", mock.Anything, mock.AnythingOfType("*domain.Product")).
		Return(&domain.Product{ID: 7, Reviews: []domain.ProductReview{}}, nil).Once()

	req := httptest.NewRequest(stdhttp.MethodDelete, "/api/products/7/reviews", nil)
	req.Header.Set("Authorization", bearerToken(1, "user@example.com", "user"))
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	assert.Equal(t, stdhttp.StatusOK, rec.Code)
}
