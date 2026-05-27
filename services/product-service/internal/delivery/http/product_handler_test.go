package http

import (
	"bytes"
	"context"
	"encoding/json"
	stdhttp "net/http"
	"net/http/httptest"
	"testing"
	"time"

	commonjwt "github.com/TheDeutsch13/b2-common/jwt"
	"github.com/TheDeutsch13/b2-project/services/product-service/internal/delivery/ws"
	"github.com/TheDeutsch13/b2-project/services/product-service/internal/domain"
	"github.com/TheDeutsch13/b2-project/services/product-service/internal/usecase"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"go.uber.org/zap"
)

type mockProductRepository struct {
	mock.Mock
}

func (m *mockProductRepository) Create(ctx context.Context, product *domain.Product) (*domain.Product, error) {
	args := m.Called(ctx, product)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Product), args.Error(1)
}

func (m *mockProductRepository) List(ctx context.Context, categoryID *int64) ([]domain.Product, error) {
	args := m.Called(ctx, categoryID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]domain.Product), args.Error(1)
}

func (m *mockProductRepository) GetByID(ctx context.Context, id int64) (*domain.Product, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Product), args.Error(1)
}

func (m *mockProductRepository) Update(ctx context.Context, product *domain.Product) (*domain.Product, error) {
	args := m.Called(ctx, product)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Product), args.Error(1)
}

func (m *mockProductRepository) Delete(ctx context.Context, id int64) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *mockProductRepository) ExistsDuplicate(
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

func setupTestRouter(
	productRepo *mockProductRepository,
	orderRepo *mockOrderRepository,
	categoryRepo *mockCategoryRepository,
) *gin.Engine {
	gin.SetMode(gin.TestMode)

	if orderRepo == nil {
		orderRepo = new(mockOrderRepository)
	}

	if categoryRepo == nil {
		categoryRepo = new(mockCategoryRepository)
	}

	logger := zap.NewNop()
	jwtManager := commonjwt.NewManager("test-secret", time.Hour)
	hub := ws.NewHub()

	productUsecase := usecase.NewProductUsecase(productRepo, nil, nil)
	categoryUsecase := usecase.NewCategoryUsecase(categoryRepo)
	orderUsecase := usecase.NewOrderUsecase(orderRepo, productRepo)

	productHandler := NewProductHandler(productUsecase, logger)
	categoryHandler := NewCategoryHandler(categoryUsecase, logger)
	orderHandler := NewOrderHandler(orderUsecase, hub, logger)
	cdekHandler := NewCdekHandler("", "", logger)
	uploadHandler := NewUploadHandler("uploads-test", logger)
	supportUsecase := usecase.NewSupportUsecase(nil)
	supportHandler := NewSupportHandler(supportUsecase, hub, logger)
	wsHandler := ws.NewHandler(hub, jwtManager)

	return NewRouter(
		logger,
		productHandler,
		categoryHandler,
		orderHandler,
		cdekHandler,
		uploadHandler,
		supportHandler,
		wsHandler,
		jwtManager,
		"uploads-test",
	)
}

type mockCategoryRepository struct{ mock.Mock }

func (m *mockCategoryRepository) List(ctx context.Context) ([]domain.Category, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]domain.Category), args.Error(1)
}

func (m *mockCategoryRepository) Create(ctx context.Context, name string) (*domain.Category, error) {
	args := m.Called(ctx, name)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Category), args.Error(1)
}

type mockOrderRepository struct{ mock.Mock }

func (m *mockOrderRepository) Create(ctx context.Context, order *domain.Order) (*domain.Order, error) {
	args := m.Called(ctx, order)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Order), args.Error(1)
}

func (m *mockOrderRepository) ListByUserID(ctx context.Context, userID int64) ([]domain.Order, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]domain.Order), args.Error(1)
}

func (m *mockOrderRepository) ListAll(ctx context.Context) ([]domain.Order, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]domain.Order), args.Error(1)
}

func (m *mockOrderRepository) UpdateStatus(ctx context.Context, orderID int64, status domain.OrderStatus) (*domain.Order, error) {
	args := m.Called(ctx, orderID, status)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Order), args.Error(1)
}

func TestProductHandler_List_Success(t *testing.T) {
	repo := new(mockProductRepository)
	router := setupTestRouter(repo, nil, nil)

	repo.On("List", mock.Anything, (*int64)(nil)).
		Return([]domain.Product{{ID: 1, Title: "Product 1"}}, nil).Once()

	req := httptest.NewRequest(stdhttp.MethodGet, "/api/products", nil)
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	assert.Equal(t, stdhttp.StatusOK, rec.Code)
}

func TestProductHandler_Create_Unauthorized(t *testing.T) {
	repo := new(mockProductRepository)
	router := setupTestRouter(repo, nil, nil)

	body := []byte(`{"title":"Test","description":"Desc","price":10}`)
	req := httptest.NewRequest(stdhttp.MethodPost, "/api/products", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	assert.Equal(t, stdhttp.StatusUnauthorized, rec.Code)
}

func TestProductHandler_Create_AdminSuccess(t *testing.T) {
	repo := new(mockProductRepository)
	router := setupTestRouter(repo, nil, nil)

	jwtManager := commonjwt.NewManager("test-secret", time.Hour)
	token, _ := jwtManager.Generate(1, "admin@example.com", "admin")

	repo.On(
		"ExistsDuplicate",
		mock.Anything,
		"GPU",
		"",
		(*int64)(nil),
		"стандарт",
		int64(0),
	).Return(false, nil).Once()
	repo.On("Create", mock.Anything, mock.AnythingOfType("*domain.Product")).
		Return(&domain.Product{ID: 1, Title: "GPU", Price: 99.99}, nil).Once()

	body := []byte(`{"title":"GPU","description":"Desc","price":99.99}`)
	req := httptest.NewRequest(stdhttp.MethodPost, "/api/products", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	assert.Equal(t, stdhttp.StatusCreated, rec.Code)

	var response map[string]interface{}
	assert.NoError(t, json.Unmarshal(rec.Body.Bytes(), &response))
	assert.Equal(t, "GPU", response["title"])
}

func TestProductHandler_Delete_AdminSuccess(t *testing.T) {
	repo := new(mockProductRepository)
	router := setupTestRouter(repo, nil, nil)

	jwtManager := commonjwt.NewManager("test-secret", time.Hour)
	token, _ := jwtManager.Generate(1, "admin@example.com", "admin")

	repo.On("Delete", mock.Anything, int64(1)).Return(nil).Once()

	req := httptest.NewRequest(stdhttp.MethodDelete, "/api/products/1", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	assert.Equal(t, stdhttp.StatusNoContent, rec.Code)
}

func TestCategoryHandler_List_Success(t *testing.T) {
	productRepo := new(mockProductRepository)
	categoryRepo := new(mockCategoryRepository)
	router := setupTestRouter(productRepo, nil, categoryRepo)

	categoryRepo.On("List", mock.Anything).
		Return([]domain.Category{{ID: 1, Name: "GPU"}}, nil).Once()

	req := httptest.NewRequest(stdhttp.MethodGet, "/api/categories", nil)
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	assert.Equal(t, stdhttp.StatusOK, rec.Code)
}

func TestOrderHandler_Create_Success(t *testing.T) {
	productRepo := new(mockProductRepository)
	orderRepo := new(mockOrderRepository)
	router := setupTestRouter(productRepo, orderRepo, nil)

	jwtManager := commonjwt.NewManager("test-secret", time.Hour)
	token, _ := jwtManager.Generate(1, "user@example.com", "user")

	productRepo.On("GetByID", mock.Anything, int64(1)).
		Return(&domain.Product{ID: 1, Title: "GPU", Price: 100}, nil).Once()
	orderRepo.On("Create", mock.Anything, mock.AnythingOfType("*domain.Order")).
		Return(&domain.Order{ID: 1, Status: domain.OrderStatusPending, TotalAmount: 100}, nil).Once()

	body := []byte(`{
		"contact_name":"Ivan",
		"contact_phone":"+7999",
		"contact_email":"ivan@example.com",
		"delivery_address":"Moscow",
		"payment_method":"card",
		"items":[{"product_id":1,"quantity":1}]
	}`)

	req := httptest.NewRequest(stdhttp.MethodPost, "/api/orders", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	assert.Equal(t, stdhttp.StatusCreated, rec.Code)
}

func TestOrderHandler_UpdateStatus_Admin(t *testing.T) {
	productRepo := new(mockProductRepository)
	orderRepo := new(mockOrderRepository)
	router := setupTestRouter(productRepo, orderRepo, nil)

	jwtManager := commonjwt.NewManager("test-secret", time.Hour)
	token, _ := jwtManager.Generate(1, "admin@example.com", "admin")

	orderRepo.On("UpdateStatus", mock.Anything, int64(1), domain.OrderStatusConfirmed).
		Return(&domain.Order{ID: 1, Status: domain.OrderStatusConfirmed}, nil).Once()

	body := []byte(`{"status":"confirmed"}`)
	req := httptest.NewRequest(stdhttp.MethodPatch, "/api/orders/1/status", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	assert.Equal(t, stdhttp.StatusOK, rec.Code)
}

func TestProductHandler_Health(t *testing.T) {
	repo := new(mockProductRepository)
	router := setupTestRouter(repo, nil, nil)

	req := httptest.NewRequest(stdhttp.MethodGet, "/health", nil)
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	assert.Equal(t, stdhttp.StatusOK, rec.Code)
	assert.Contains(t, rec.Body.String(), "product-service")
}
