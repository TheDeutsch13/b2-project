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

func TestOrderHandler_ListMy_Success(t *testing.T) {
	productRepo := new(mockProductRepository)
	orderRepo := new(mockOrderRepository)
	router := setupTestRouter(productRepo, orderRepo, nil)

	orderRepo.On("ListByUserID", mock.Anything, int64(1)).
		Return([]domain.Order{
			{ID: 10, UserID: 1, Status: domain.OrderStatusPending, TotalAmount: 500},
		}, nil).Once()

	req := httptest.NewRequest(stdhttp.MethodGet, "/api/orders/my", nil)
	req.Header.Set("Authorization", bearerToken(1, "user@example.com", "user"))
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	assert.Equal(t, stdhttp.StatusOK, rec.Code)
	assert.Contains(t, rec.Body.String(), `"id":10`)
}

func TestOrderHandler_ListAll_Admin(t *testing.T) {
	productRepo := new(mockProductRepository)
	orderRepo := new(mockOrderRepository)
	router := setupTestRouter(productRepo, orderRepo, nil)

	orderRepo.On("ListAll", mock.Anything).
		Return([]domain.Order{
			{ID: 1, Status: domain.OrderStatusConfirmed},
			{ID: 2, Status: domain.OrderStatusShipped},
		}, nil).Once()

	req := httptest.NewRequest(stdhttp.MethodGet, "/api/orders", nil)
	req.Header.Set("Authorization", bearerToken(1, "admin@example.com", "admin"))
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	assert.Equal(t, stdhttp.StatusOK, rec.Code)
}

func TestOrderHandler_CourierUpdateStatus(t *testing.T) {
	productRepo := new(mockProductRepository)
	orderRepo := new(mockOrderRepository)
	router := setupTestRouter(productRepo, orderRepo, nil)

	orderRepo.On("UpdateStatus", mock.Anything, int64(5), domain.OrderStatusDelivered).
		Return(&domain.Order{ID: 5, Status: domain.OrderStatusDelivered}, nil).Once()

	body := []byte(`{"status":"delivered"}`)
	req := httptest.NewRequest(
		stdhttp.MethodPatch,
		"/api/courier/orders/5/status",
		bytes.NewBuffer(body),
	)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", bearerToken(3, "courier@example.com", "courier"))
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	assert.Equal(t, stdhttp.StatusOK, rec.Code)
	assert.Contains(t, rec.Body.String(), "delivered")
}

func TestOrderHandler_Create_InvalidBody(t *testing.T) {
	router := setupTestRouter(new(mockProductRepository), nil, nil)

	req := httptest.NewRequest(stdhttp.MethodPost, "/api/orders", bytes.NewBuffer([]byte(`{}`)))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", bearerToken(1, "user@example.com", "user"))
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	assert.Equal(t, stdhttp.StatusBadRequest, rec.Code)
}

func TestOrderHandler_UpdateStatus_AdminBroadcast(t *testing.T) {
	productRepo := new(mockProductRepository)
	orderRepo := new(mockOrderRepository)
	router := setupTestRouter(productRepo, orderRepo, nil)

	orderRepo.On("UpdateStatus", mock.Anything, int64(2), domain.OrderStatusShipped).
		Return(&domain.Order{
			ID: 2, UserID: 5, Status: domain.OrderStatusShipped,
			Items: []domain.OrderItem{{ID: 1, ProductID: 1, Title: "GPU", Quantity: 1, Price: 100}},
		}, nil).Once()

	body := []byte(`{"status":"shipped"}`)
	req := httptest.NewRequest(stdhttp.MethodPatch, "/api/orders/2/status", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", bearerToken(1, "admin@example.com", "admin"))
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	assert.Equal(t, stdhttp.StatusOK, rec.Code)
	assert.Contains(t, rec.Body.String(), "shipped")
}

func TestOrderHandler_CourierForbiddenStatus(t *testing.T) {
	router := setupTestRouter(new(mockProductRepository), new(mockOrderRepository), nil)

	body := []byte(`{"status":"cancelled"}`)
	req := httptest.NewRequest(
		stdhttp.MethodPatch,
		"/api/courier/orders/1/status",
		bytes.NewBuffer(body),
	)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", bearerToken(3, "courier@example.com", "courier"))
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	assert.Equal(t, stdhttp.StatusForbidden, rec.Code)
}

func TestOrderHandler_Create_InsufficientStock(t *testing.T) {
	productRepo := new(mockProductRepository)
	orderRepo := new(mockOrderRepository)
	router := setupTestRouter(productRepo, orderRepo, nil)

	productRepo.On("GetByID", mock.Anything, int64(1)).
		Return(&domain.Product{ID: 1, Title: "GPU", Price: 100, Stock: 0}, nil).Once()

	body := []byte(`{
		"contact_name":"Ivan",
		"contact_phone":"+7999",
		"contact_email":"ivan@example.com",
		"delivery_address":"Moscow",
		"delivery_type":"custom",
		"payment_method":"card",
		"items":[{"product_id":1,"quantity":1}]
	}`)
	req := httptest.NewRequest(stdhttp.MethodPost, "/api/orders", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", bearerToken(1, "user@example.com", "user"))
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	assert.Equal(t, stdhttp.StatusConflict, rec.Code)
}

func TestOrderHandler_ListMy_Unauthorized(t *testing.T) {
	router := setupTestRouter(new(mockProductRepository), nil, nil)

	req := httptest.NewRequest(stdhttp.MethodGet, "/api/orders/my", nil)
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	assert.Equal(t, stdhttp.StatusUnauthorized, rec.Code)
}
