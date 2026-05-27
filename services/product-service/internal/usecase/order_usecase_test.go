package usecase

import (
	"context"
	"testing"

	"github.com/TheDeutsch13/b2-project/services/product-service/internal/domain"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type MockOrderRepository struct {
	mock.Mock
}

func (m *MockOrderRepository) Create(ctx context.Context, order *domain.Order) (*domain.Order, error) {
	args := m.Called(ctx, order)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Order), args.Error(1)
}

func (m *MockOrderRepository) ListByUserID(ctx context.Context, userID int64) ([]domain.Order, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]domain.Order), args.Error(1)
}

func (m *MockOrderRepository) ListAll(ctx context.Context) ([]domain.Order, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]domain.Order), args.Error(1)
}

func (m *MockOrderRepository) UpdateStatus(ctx context.Context, orderID int64, status domain.OrderStatus) (*domain.Order, error) {
	args := m.Called(ctx, orderID, status)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Order), args.Error(1)
}

func TestOrderUsecase_Create_Success(t *testing.T) {
	orderRepo := new(MockOrderRepository)
	productRepo := new(MockProductRepository)
	orderUsecase := NewOrderUsecase(orderRepo, productRepo)

	productRepo.On("GetByID", mock.Anything, int64(1)).
		Return(&domain.Product{ID: 1, Title: "GPU", Price: 100, Stock: 5}, nil).Once()
	orderRepo.On("Create", mock.Anything, mock.AnythingOfType("*domain.Order")).
		Return(&domain.Order{ID: 1, Status: domain.OrderStatusPending, TotalAmount: 100}, nil).Once()

	order, err := orderUsecase.Create(context.Background(), domain.CreateOrderInput{
		UserID:          1,
		ContactName:     "Ivan",
		ContactPhone:    "+79990001122",
		ContactEmail:    "ivan@example.com",
		DeliveryAddress: "Moscow",
		DeliveryType:    domain.DeliveryTypeCustom,
		PaymentMethod:   "card",
		Items:           []domain.CreateOrderItemInput{{ProductID: 1, Quantity: 1}},
	})

	assert.NoError(t, err)
	assert.Equal(t, int64(1), order.ID)
}

func TestOrderUsecase_Create_EmptyItems(t *testing.T) {
	orderUsecase := NewOrderUsecase(new(MockOrderRepository), new(MockProductRepository))

	order, err := orderUsecase.Create(context.Background(), domain.CreateOrderInput{
		UserID:          1,
		ContactName:     "Ivan",
		ContactPhone:    "+79990001122",
		ContactEmail:    "ivan@example.com",
		DeliveryAddress: "Moscow",
		DeliveryType:    domain.DeliveryTypeCustom,
		PaymentMethod:   "card",
		Items:           []domain.CreateOrderItemInput{},
	})

	assert.Nil(t, order)
	assert.ErrorIs(t, err, ErrEmptyOrder)
}

func TestOrderUsecase_UpdateStatus_Success(t *testing.T) {
	orderRepo := new(MockOrderRepository)
	orderUsecase := NewOrderUsecase(orderRepo, new(MockProductRepository))

	orderRepo.On("UpdateStatus", mock.Anything, int64(1), domain.OrderStatusConfirmed).
		Return(&domain.Order{ID: 1, Status: domain.OrderStatusConfirmed}, nil).Once()

	order, err := orderUsecase.UpdateStatus(context.Background(), 1, domain.OrderStatusConfirmed)

	assert.NoError(t, err)
	assert.Equal(t, domain.OrderStatusConfirmed, order.Status)
}

func TestOrderUsecase_Create_CdekRequiresPvz(t *testing.T) {
	orderUsecase := NewOrderUsecase(new(MockOrderRepository), new(MockProductRepository))

	order, err := orderUsecase.Create(context.Background(), domain.CreateOrderInput{
		UserID:          1,
		ContactName:     "Ivan",
		ContactPhone:    "+79990001122",
		ContactEmail:    "ivan@example.com",
		DeliveryAddress: "Moscow",
		DeliveryType:    domain.DeliveryTypeCdek,
		PaymentMethod:   "card",
		Items:           []domain.CreateOrderItemInput{{ProductID: 1, Quantity: 1}},
	})

	assert.Nil(t, order)
	assert.ErrorIs(t, err, ErrInvalidInput)
}

func TestOrderUsecase_Create_InsufficientStock(t *testing.T) {
	orderRepo := new(MockOrderRepository)
	productRepo := new(MockProductRepository)
	orderUsecase := NewOrderUsecase(orderRepo, productRepo)

	productRepo.On("GetByID", mock.Anything, int64(1)).
		Return(&domain.Product{ID: 1, Title: "GPU", Price: 100, Stock: 0}, nil).Once()

	order, err := orderUsecase.Create(context.Background(), domain.CreateOrderInput{
		UserID:          1,
		ContactName:     "Ivan",
		ContactPhone:    "+79990001122",
		ContactEmail:    "ivan@example.com",
		DeliveryAddress: "Moscow",
		DeliveryType:    domain.DeliveryTypeCustom,
		PaymentMethod:   "card",
		Items:           []domain.CreateOrderItemInput{{ProductID: 1, Quantity: 1}},
	})

	assert.Nil(t, order)
	assert.ErrorIs(t, err, ErrInsufficientStock)
}

func TestOrderUsecase_UpdateStatus_InvalidStatus(t *testing.T) {
	orderUsecase := NewOrderUsecase(new(MockOrderRepository), new(MockProductRepository))

	order, err := orderUsecase.UpdateStatus(context.Background(), 1, domain.OrderStatus("invalid"))

	assert.Nil(t, order)
	assert.ErrorIs(t, err, ErrInvalidInput)
}

func TestOrderUsecase_Create_CdekSuccess(t *testing.T) {
	orderRepo := new(MockOrderRepository)
	productRepo := new(MockProductRepository)
	orderUsecase := NewOrderUsecase(orderRepo, productRepo)

	productRepo.On("GetByID", mock.Anything, int64(1)).
		Return(&domain.Product{ID: 1, Title: "GPU", Price: 100, Stock: 5}, nil).Once()
	orderRepo.On("Create", mock.Anything, mock.AnythingOfType("*domain.Order")).
		Return(&domain.Order{ID: 2, Status: domain.OrderStatusPending, TotalAmount: 405}, nil).Once()

	order, err := orderUsecase.Create(context.Background(), domain.CreateOrderInput{
		UserID:          1,
		ContactName:     "Ivan",
		ContactPhone:    "+79990001122",
		ContactEmail:    "ivan@example.com",
		DeliveryAddress: "PVZ",
		DeliveryType:    domain.DeliveryTypeCdek,
		CdekPvzCode:     "SAR115",
		PaymentMethod:   "card",
		Items:           []domain.CreateOrderItemInput{{ProductID: 1, Quantity: 1}},
	})

	assert.NoError(t, err)
	assert.Equal(t, int64(2), order.ID)
}

func TestOrderUsecase_ListMy(t *testing.T) {
	orderRepo := new(MockOrderRepository)
	orderUsecase := NewOrderUsecase(orderRepo, new(MockProductRepository))

	orderRepo.On("ListByUserID", mock.Anything, int64(3)).
		Return([]domain.Order{{ID: 7, UserID: 3}}, nil).Once()

	orders, err := orderUsecase.ListMy(context.Background(), 3)

	assert.NoError(t, err)
	assert.Len(t, orders, 1)
}

func TestOrderUsecase_ListAll(t *testing.T) {
	orderRepo := new(MockOrderRepository)
	orderUsecase := NewOrderUsecase(orderRepo, new(MockProductRepository))

	orderRepo.On("ListAll", mock.Anything).
		Return([]domain.Order{{ID: 1}, {ID: 2}}, nil).Once()

	orders, err := orderUsecase.ListAll(context.Background())

	assert.NoError(t, err)
	assert.Len(t, orders, 2)
}

func TestOrderUsecase_Create_InvalidDeliveryType(t *testing.T) {
	orderUsecase := NewOrderUsecase(new(MockOrderRepository), new(MockProductRepository))

	order, err := orderUsecase.Create(context.Background(), domain.CreateOrderInput{
		UserID:          1,
		ContactName:     "Ivan",
		ContactPhone:    "+79990001122",
		ContactEmail:    "ivan@example.com",
		DeliveryAddress: "Moscow",
		DeliveryType:    domain.DeliveryType("unknown"),
		PaymentMethod:   "card",
		Items:           []domain.CreateOrderItemInput{{ProductID: 1, Quantity: 1}},
	})

	assert.Nil(t, order)
	assert.ErrorIs(t, err, ErrInvalidInput)
}
