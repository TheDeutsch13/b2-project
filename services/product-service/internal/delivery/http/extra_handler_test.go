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

func TestCdekHandler_ListPoints_Demo(t *testing.T) {
	router := setupTestRouter(new(mockProductRepository), nil, nil)

	req := httptest.NewRequest(stdhttp.MethodGet, "/api/cdek/points?city=Саратов", nil)
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	assert.Equal(t, stdhttp.StatusOK, rec.Code)
	assert.Contains(t, rec.Body.String(), "SAR115")
}

func TestCdekHandler_ListPoints_MissingCity(t *testing.T) {
	router := setupTestRouter(new(mockProductRepository), nil, nil)

	req := httptest.NewRequest(stdhttp.MethodGet, "/api/cdek/points", nil)
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	assert.Equal(t, stdhttp.StatusBadRequest, rec.Code)
}

func TestProductHandler_Update_AdminSuccess(t *testing.T) {
	repo := new(mockProductRepository)
	router := setupTestRouter(repo, nil, nil)

	repo.On("GetByID", mock.Anything, int64(3)).
		Return(&domain.Product{ID: 3, Title: "Old", Price: 50}, nil).Once()
	repo.On(
		"ExistsDuplicate",
		mock.Anything,
		"New GPU",
		"",
		(*int64)(nil),
		"стандарт",
		int64(3),
	).Return(false, nil).Once()
	repo.On("Update", mock.Anything, mock.AnythingOfType("*domain.Product")).
		Return(&domain.Product{ID: 3, Title: "New GPU", Price: 120}, nil).Once()

	body := []byte(`{"title":"New GPU","description":"Updated","price":120}`)
	req := httptest.NewRequest(stdhttp.MethodPut, "/api/products/3", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", bearerToken(1, "admin@example.com", "admin"))
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	assert.Equal(t, stdhttp.StatusOK, rec.Code)
	assert.Contains(t, rec.Body.String(), "New GPU")
}

func TestProductHandler_ListAdminReviews_InvalidRating(t *testing.T) {
	router := setupTestRouterFull(routerTestDeps{reviewRepo: new(mockProductReviewRepository)})

	req := httptest.NewRequest(stdhttp.MethodGet, "/api/reviews?rating=9", nil)
	req.Header.Set("Authorization", bearerToken(1, "admin@example.com", "admin"))
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	assert.Equal(t, stdhttp.StatusBadRequest, rec.Code)
}

func TestProductHandler_Get_NotFound(t *testing.T) {
	repo := new(mockProductRepository)
	router := setupTestRouter(repo, nil, nil)

	repo.On("GetByID", mock.Anything, int64(99)).
		Return(nil, domain.ErrProductNotFound).Once()

	req := httptest.NewRequest(stdhttp.MethodGet, "/api/products/99", nil)
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	assert.Equal(t, stdhttp.StatusNotFound, rec.Code)
}

func TestProductHandler_Get_InvalidID(t *testing.T) {
	router := setupTestRouter(new(mockProductRepository), nil, nil)

	req := httptest.NewRequest(stdhttp.MethodGet, "/api/products/abc", nil)
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	assert.Equal(t, stdhttp.StatusBadRequest, rec.Code)
}

func TestProductHandler_Create_Duplicate(t *testing.T) {
	repo := new(mockProductRepository)
	router := setupTestRouter(repo, nil, nil)

	repo.On(
		"ExistsDuplicate",
		mock.Anything,
		"Dup",
		"",
		(*int64)(nil),
		"стандарт",
		int64(0),
	).Return(true, nil).Once()

	body := []byte(`{"title":"Dup","description":"Desc","price":10}`)
	req := httptest.NewRequest(stdhttp.MethodPost, "/api/products", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", bearerToken(1, "admin@example.com", "admin"))
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	assert.Equal(t, stdhttp.StatusBadRequest, rec.Code)
	assert.Contains(t, rec.Body.String(), "уже существует")
}

func TestProductHandler_Delete_InOrders(t *testing.T) {
	repo := new(mockProductRepository)
	router := setupTestRouter(repo, nil, nil)

	repo.On("Delete", mock.Anything, int64(4)).
		Return(domain.ErrProductInOrders).Once()

	req := httptest.NewRequest(stdhttp.MethodDelete, "/api/products/4", nil)
	req.Header.Set("Authorization", bearerToken(1, "admin@example.com", "admin"))
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	assert.Equal(t, stdhttp.StatusConflict, rec.Code)
}

func TestProductHandler_List_InvalidCategoryID(t *testing.T) {
	router := setupTestRouter(new(mockProductRepository), nil, nil)

	req := httptest.NewRequest(stdhttp.MethodGet, "/api/products?category_id=bad", nil)
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	assert.Equal(t, stdhttp.StatusBadRequest, rec.Code)
}

func TestProductHandler_List_WithCategory(t *testing.T) {
	repo := new(mockProductRepository)
	categoryID := int64(2)
	router := setupTestRouter(repo, nil, nil)

	repo.On("List", mock.Anything, &categoryID).
		Return([]domain.Product{{ID: 1, Title: "Mouse", CategoryID: &categoryID}}, nil).Once()

	req := httptest.NewRequest(stdhttp.MethodGet, "/api/products?category_id=2", nil)
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	assert.Equal(t, stdhttp.StatusOK, rec.Code)
}

func TestCdekHandler_ListPoints_DefaultCity(t *testing.T) {
	router := setupTestRouter(new(mockProductRepository), nil, nil)

	req := httptest.NewRequest(stdhttp.MethodGet, "/api/cdek/points?city=Moscow", nil)
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	assert.Equal(t, stdhttp.StatusOK, rec.Code)
}

func TestProductReview_Delete_NotFound(t *testing.T) {
	productRepo := new(mockProductRepository)
	router := setupTestRouterFull(routerTestDeps{productRepo: productRepo})

	productRepo.On("GetByID", mock.Anything, int64(8)).
		Return(&domain.Product{ID: 8, Reviews: []domain.ProductReview{}}, nil).Once()

	req := httptest.NewRequest(stdhttp.MethodDelete, "/api/products/8/reviews", nil)
	req.Header.Set("Authorization", bearerToken(1, "user@example.com", "user"))
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	assert.Equal(t, stdhttp.StatusNotFound, rec.Code)
}
