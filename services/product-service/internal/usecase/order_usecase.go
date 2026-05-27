package usecase

import (
	"context"
	"errors"
	"strings"

	"github.com/TheDeutsch13/b2-project/services/product-service/internal/domain"
)

var (
	ErrOrderNotFound       = errors.New("order not found")
	ErrEmptyOrder          = errors.New("order must contain items")
	ErrInsufficientStock   = domain.ErrInsufficientStock
)

type OrderRepository interface {
	Create(ctx context.Context, order *domain.Order) (*domain.Order, error)
	ListByUserID(ctx context.Context, userID int64) ([]domain.Order, error)
	ListAll(ctx context.Context) ([]domain.Order, error)
	UpdateStatus(ctx context.Context, orderID int64, status domain.OrderStatus) (*domain.Order, error)
}

type OrderUsecase struct {
	orderRepo   OrderRepository
	productRepo ProductRepository
}

func NewOrderUsecase(orderRepo OrderRepository, productRepo ProductRepository) *OrderUsecase {
	return &OrderUsecase{
		orderRepo:   orderRepo,
		productRepo: productRepo,
	}
}

const (
	cdekDeliveryCost   = 305
	customDeliveryCost = 700
)

func normalizeDeliveryPayment(value string) string {
	switch strings.TrimSpace(value) {
	case "online":
		return "online"
	default:
		return "on_receipt"
	}
}

func deliveryCostForType(deliveryType domain.DeliveryType) (float64, bool) {
	switch deliveryType {
	case domain.DeliveryTypeCdek:
		return cdekDeliveryCost, true
	case domain.DeliveryTypeCustom:
		return customDeliveryCost, true
	default:
		return 0, false
	}
}

func (u *OrderUsecase) Create(ctx context.Context, input domain.CreateOrderInput) (*domain.Order, error) {
	if input.UserID == 0 ||
		strings.TrimSpace(input.ContactName) == "" ||
		strings.TrimSpace(input.ContactPhone) == "" ||
		strings.TrimSpace(input.ContactEmail) == "" ||
		strings.TrimSpace(input.DeliveryAddress) == "" ||
		strings.TrimSpace(input.PaymentMethod) == "" {
		return nil, ErrInvalidInput
	}

	deliveryCost, ok := deliveryCostForType(input.DeliveryType)
	if !ok {
		return nil, ErrInvalidInput
	}

	if input.DeliveryType == domain.DeliveryTypeCdek &&
		strings.TrimSpace(input.CdekPvzCode) == "" {
		return nil, ErrInvalidInput
	}

	if len(input.Items) == 0 {
		return nil, ErrEmptyOrder
	}

	order := &domain.Order{
		UserID:          input.UserID,
		Status:          domain.OrderStatusPending,
		ContactName:     strings.TrimSpace(input.ContactName),
		ContactPhone:    strings.TrimSpace(input.ContactPhone),
		ContactEmail:    strings.TrimSpace(input.ContactEmail),
		DeliveryAddress: strings.TrimSpace(input.DeliveryAddress),
		DeliveryType:    input.DeliveryType,
		DeliveryCost:    deliveryCost,
		DeliveryCity:    strings.TrimSpace(input.DeliveryCity),
		CdekPvzCode:     strings.TrimSpace(input.CdekPvzCode),
		DeliveryPayment: normalizeDeliveryPayment(input.DeliveryPayment),
		PaymentMethod:   strings.TrimSpace(input.PaymentMethod),
		Comment:         strings.TrimSpace(input.Comment),
		Items:           make([]domain.OrderItem, 0, len(input.Items)),
	}

	for _, itemInput := range input.Items {
		if itemInput.ProductID == 0 || itemInput.Quantity <= 0 {
			return nil, ErrInvalidInput
		}

		product, err := u.productRepo.GetByID(ctx, itemInput.ProductID)
		if err != nil {
			return nil, ErrInvalidInput
		}

		if product.Stock < itemInput.Quantity {
			return nil, ErrInsufficientStock
		}

		order.Items = append(order.Items, domain.OrderItem{
			ProductID: product.ID,
			Quantity:  itemInput.Quantity,
			Price:     product.Price,
			Title:     product.Title,
		})

		order.TotalAmount += product.Price * float64(itemInput.Quantity)
	}

	order.TotalAmount += deliveryCost

	created, err := u.orderRepo.Create(ctx, order)
	if err != nil {
		if errors.Is(err, domain.ErrInsufficientStock) {
			return nil, ErrInsufficientStock
		}
		return nil, err
	}

	return created, nil
}

func (u *OrderUsecase) ListMy(ctx context.Context, userID int64) ([]domain.Order, error) {
	return u.orderRepo.ListByUserID(ctx, userID)
}

func (u *OrderUsecase) ListAll(ctx context.Context) ([]domain.Order, error) {
	return u.orderRepo.ListAll(ctx)
}

func (u *OrderUsecase) UpdateStatus(ctx context.Context, orderID int64, status domain.OrderStatus) (*domain.Order, error) {
	switch status {
	case domain.OrderStatusPending,
		domain.OrderStatusConfirmed,
		domain.OrderStatusShipped,
		domain.OrderStatusDelivered,
		domain.OrderStatusReceived,
		domain.OrderStatusCancelled:
	default:
		return nil, ErrInvalidInput
	}

	order, err := u.orderRepo.UpdateStatus(ctx, orderID, status)
	if err != nil {
		if errors.Is(err, domain.ErrInsufficientStock) {
			return nil, ErrInsufficientStock
		}
		return nil, ErrOrderNotFound
	}

	return order, nil
}
