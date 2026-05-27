package domain

import (
	"strings"
	"time"
)

type Category struct {
	ID        int64
	Name      string
	CreatedAt time.Time
}

type ProductSpecification struct {
	Label string `json:"label"`
	Value string `json:"value"`
}

type ProductReview struct {
	UserID    int64  `json:"user_id,omitempty"`
	Author    string `json:"author"`
	Rating    int    `json:"rating"`
	Text      string `json:"text"`
	CreatedAt string `json:"created_at,omitempty"`
}

// UserProductReview — отзыв пользователя с данными товара.
type UserProductReview struct {
	ProductID    int64
	ProductTitle string
	ProductImage string
	Review       ProductReview
}

// AdminProductReview — отзыв покупателя для админ-панели.
type AdminProductReview struct {
	ProductID    int64
	ProductTitle string
	UserID       int64
	Author       string
	Rating       int
	Text         string
	CreatedAt    string
}

// ReviewListFilter — фильтры списка отзывов.
type ReviewListFilter struct {
	Rating    int
	ProductID int64
	Query     string
}

type Product struct {
	ID             int64
	CategoryID     *int64
	CategoryName   string
	Title          string
	Description    string
	Price          float64
	Brand          string
	Stock          int
	Images         []string
	Specifications []ProductSpecification
	Variants       []string
	Reviews        []ProductReview
	RatingAvg      float64
	RatingCount    int
	CreatedAt      time.Time
}

func PrimaryVariant(variants []string) string {
	for _, item := range variants {
		value := strings.TrimSpace(item)
		if value != "" {
			return strings.ToLower(value)
		}
	}

	return "стандарт"
}

func CalcRatingFromReviews(reviews []ProductReview) (avg float64, count int) {
	count = len(reviews)
	if count == 0 {
		return 0, 0
	}

	sum := 0
	valid := 0

	for _, review := range reviews {
		rating := review.Rating
		if rating < 1 || rating > 5 {
			continue
		}

		sum += rating
		valid++
	}

	if valid == 0 {
		return 0, 0
	}

	return float64(sum) / float64(valid), valid
}

type OrderStatus string

const (
	OrderStatusPending   OrderStatus = "pending"
	OrderStatusConfirmed OrderStatus = "confirmed"
	OrderStatusShipped   OrderStatus = "shipped"
	OrderStatusDelivered OrderStatus = "delivered"
	OrderStatusReceived  OrderStatus = "received"
	OrderStatusCancelled OrderStatus = "cancelled"
)

type DeliveryType string

const (
	DeliveryTypeCdek   DeliveryType = "cdek"
	DeliveryTypeCustom DeliveryType = "custom"
)

type Order struct {
	ID              int64
	UserID          int64
	Status          OrderStatus
	ContactName     string
	ContactPhone    string
	ContactEmail    string
	DeliveryAddress string
	DeliveryType    DeliveryType
	DeliveryCost    float64
	DeliveryCity    string
	CdekPvzCode     string
	DeliveryPayment string
	PaymentMethod   string
	Comment         string
	TotalAmount     float64
	CreatedAt       time.Time
	Items           []OrderItem
}

type OrderItem struct {
	ID        int64
	OrderID   int64
	ProductID int64
	Quantity  int
	Price     float64
	Title     string
}

type CreateOrderInput struct {
	UserID          int64
	ContactName     string
	ContactPhone    string
	ContactEmail    string
	DeliveryAddress string
	DeliveryType    DeliveryType
	DeliveryCost    float64
	DeliveryCity    string
	CdekPvzCode     string
	DeliveryPayment string
	PaymentMethod   string
	Comment         string
	Items           []CreateOrderItemInput
}

type CreateOrderItemInput struct {
	ProductID int64
	Quantity  int
}
