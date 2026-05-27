package http

import (
	"errors"
	stdhttp "net/http"
	"strconv"
	"time"

	"github.com/TheDeutsch13/b2-common/httperr"
	commonmiddleware "github.com/TheDeutsch13/b2-common/middleware"
	"github.com/TheDeutsch13/b2-project/services/product-service/internal/delivery/ws"
	"github.com/TheDeutsch13/b2-project/services/product-service/internal/domain"
	"github.com/TheDeutsch13/b2-project/services/product-service/internal/usecase"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

type OrderHandler struct {
	orderUsecase *usecase.OrderUsecase
	hub          *ws.Hub
	logger       *zap.Logger
}

func NewOrderHandler(orderUsecase *usecase.OrderUsecase, hub *ws.Hub, logger *zap.Logger) *OrderHandler {
	return &OrderHandler{
		orderUsecase: orderUsecase,
		hub:          hub,
		logger:       logger,
	}
}

type orderItemRequest struct {
	ProductID int64 `json:"product_id"`
	Quantity  int   `json:"quantity"`
}

type createOrderRequest struct {
	ContactName     string             `json:"contact_name"`
	ContactPhone    string             `json:"contact_phone"`
	ContactEmail    string             `json:"contact_email"`
	DeliveryAddress string             `json:"delivery_address"`
	DeliveryType    string             `json:"delivery_type"`
	DeliveryCity    string             `json:"delivery_city"`
	CdekPvzCode     string             `json:"cdek_pvz_code"`
	DeliveryPayment string             `json:"delivery_payment"`
	PaymentMethod   string             `json:"payment_method"`
	Comment         string             `json:"comment"`
	Items           []orderItemRequest `json:"items"`
}

type updateOrderStatusRequest struct {
	Status string `json:"status"`
}

type orderItemResponse struct {
	ID        int64   `json:"id"`
	ProductID int64   `json:"product_id"`
	Title     string  `json:"title"`
	Quantity  int     `json:"quantity"`
	Price     float64 `json:"price"`
}

type orderResponse struct {
	ID              int64               `json:"id"`
	UserID          int64               `json:"user_id"`
	Status          string              `json:"status"`
	ContactName     string              `json:"contact_name"`
	ContactPhone    string              `json:"contact_phone"`
	ContactEmail    string              `json:"contact_email"`
	DeliveryAddress string              `json:"delivery_address"`
	DeliveryType    string              `json:"delivery_type"`
	DeliveryCost    float64             `json:"delivery_cost"`
	DeliveryCity    string              `json:"delivery_city"`
	CdekPvzCode     string              `json:"cdek_pvz_code"`
	DeliveryPayment string              `json:"delivery_payment"`
	PaymentMethod   string              `json:"payment_method"`
	Comment         string              `json:"comment"`
	TotalAmount     float64             `json:"total_amount"`
	CreatedAt       time.Time           `json:"created_at"`
	Items           []orderItemResponse `json:"items"`
}

func toOrderResponse(order domain.Order) orderResponse {
	items := make([]orderItemResponse, 0, len(order.Items))
	for _, item := range order.Items {
		items = append(items, orderItemResponse{
			ID:        item.ID,
			ProductID: item.ProductID,
			Title:     item.Title,
			Quantity:  item.Quantity,
			Price:     item.Price,
		})
	}

	return orderResponse{
		ID:              order.ID,
		UserID:          order.UserID,
		Status:          string(order.Status),
		ContactName:     order.ContactName,
		ContactPhone:    order.ContactPhone,
		ContactEmail:    order.ContactEmail,
		DeliveryAddress: order.DeliveryAddress,
		DeliveryType:    string(order.DeliveryType),
		DeliveryCost:    order.DeliveryCost,
		DeliveryCity:    order.DeliveryCity,
		CdekPvzCode:     order.CdekPvzCode,
		DeliveryPayment: order.DeliveryPayment,
		PaymentMethod:   order.PaymentMethod,
		Comment:         order.Comment,
		TotalAmount:     order.TotalAmount,
		CreatedAt:       order.CreatedAt,
		Items:           items,
	}
}

// CreateOrder godoc
// @Summary Create order
// @Tags orders
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body createOrderRequest true "Create order request"
// @Success 201 {object} orderResponse
// @Router /api/orders [post]
func (h *OrderHandler) Create(ctx *gin.Context) {
	userID, ok := commonmiddleware.GetUserID(ctx)
	if !ok {
		httperr.Unauthorized(ctx, "unauthorized")
		return
	}

	var req createOrderRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		httperr.BadRequest(ctx, "invalid request body")
		return
	}

	items := make([]domain.CreateOrderItemInput, 0, len(req.Items))
	for _, item := range req.Items {
		items = append(items, domain.CreateOrderItemInput{
			ProductID: item.ProductID,
			Quantity:  item.Quantity,
		})
	}

	order, err := h.orderUsecase.Create(ctx.Request.Context(), domain.CreateOrderInput{
		UserID:          userID,
		ContactName:     req.ContactName,
		ContactPhone:    req.ContactPhone,
		ContactEmail:    req.ContactEmail,
		DeliveryAddress: req.DeliveryAddress,
		DeliveryType:    domain.DeliveryType(req.DeliveryType),
		DeliveryCity:    req.DeliveryCity,
		CdekPvzCode:     req.CdekPvzCode,
		DeliveryPayment: req.DeliveryPayment,
		PaymentMethod:   req.PaymentMethod,
		Comment:         req.Comment,
		Items:           items,
	})
	if err != nil {
		if errors.Is(err, usecase.ErrInvalidInput) || errors.Is(err, usecase.ErrEmptyOrder) {
			httperr.BadRequest(ctx, err.Error())
			return
		}
		if errors.Is(err, usecase.ErrInsufficientStock) {
			httperr.Conflict(ctx, "insufficient stock")
			return
		}

		h.logger.Error("failed to create order", zap.Error(err))
		httperr.Internal(ctx, "internal server error")
		return
	}

	h.hub.Broadcast(ws.Notification{
		Type:    "order_created",
		Message: "Новый заказ оформлен",
		OrderID: order.ID,
		UserID:  order.UserID,
	})

	ctx.JSON(stdhttp.StatusCreated, toOrderResponse(*order))
}

// ListMyOrders godoc
// @Summary List my orders
// @Tags orders
// @Produce json
// @Security BearerAuth
// @Success 200 {array} orderResponse
// @Router /api/orders/my [get]
func (h *OrderHandler) ListMy(ctx *gin.Context) {
	userID, ok := commonmiddleware.GetUserID(ctx)
	if !ok {
		httperr.Unauthorized(ctx, "unauthorized")
		return
	}

	orders, err := h.orderUsecase.ListMy(ctx.Request.Context(), userID)
	if err != nil {
		h.logger.Error("failed to list orders", zap.Error(err))
		httperr.Internal(ctx, "internal server error")
		return
	}

	response := make([]orderResponse, 0, len(orders))
	for _, order := range orders {
		response = append(response, toOrderResponse(order))
	}

	ctx.JSON(stdhttp.StatusOK, response)
}

// ListAllOrders godoc
// @Summary List all orders
// @Tags orders
// @Produce json
// @Security BearerAuth
// @Success 200 {array} orderResponse
// @Router /api/orders [get]
func (h *OrderHandler) ListAll(ctx *gin.Context) {
	orders, err := h.orderUsecase.ListAll(ctx.Request.Context())
	if err != nil {
		h.logger.Error("failed to list all orders", zap.Error(err))
		httperr.Internal(ctx, "internal server error")
		return
	}

	response := make([]orderResponse, 0, len(orders))
	for _, order := range orders {
		response = append(response, toOrderResponse(order))
	}

	ctx.JSON(stdhttp.StatusOK, response)
}

// UpdateOrderStatus godoc
// @Summary Update order status
// @Tags orders
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "Order ID"
// @Param request body updateOrderStatusRequest true "Status update"
// @Success 200 {object} orderResponse
// @Router /api/orders/{id}/status [patch]
func (h *OrderHandler) UpdateStatus(ctx *gin.Context) {
	orderID, err := strconv.ParseInt(ctx.Param("id"), 10, 64)
	if err != nil {
		httperr.BadRequest(ctx, "invalid order id")
		return
	}

	var req updateOrderStatusRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		httperr.BadRequest(ctx, "invalid request body")
		return
	}

	status := domain.OrderStatus(req.Status)
	if role, ok := commonmiddleware.GetRole(ctx); ok && role == "courier" {
		if !isCourierAllowedStatus(status) {
			httperr.Forbidden(ctx, "courier cannot set this status")
			return
		}
	}

	order, err := h.orderUsecase.UpdateStatus(
		ctx.Request.Context(),
		orderID,
		status,
	)
	if err != nil {
		if errors.Is(err, usecase.ErrInvalidInput) {
			httperr.BadRequest(ctx, "invalid status")
			return
		}
		if errors.Is(err, usecase.ErrInsufficientStock) {
			httperr.Conflict(ctx, "insufficient stock")
			return
		}

		httperr.BadRequest(ctx, "order not found")
		return
	}

	h.hub.Broadcast(ws.Notification{
		Type:    "order_status_updated",
		Message: "Статус заказа изменён на " + req.Status,
		OrderID: order.ID,
		UserID:  order.UserID,
	})

	ctx.JSON(stdhttp.StatusOK, toOrderResponse(*order))
}

func isCourierAllowedStatus(status domain.OrderStatus) bool {
	switch status {
	case domain.OrderStatusConfirmed,
		domain.OrderStatusShipped,
		domain.OrderStatusDelivered:
		return true
	default:
		return false
	}
}
