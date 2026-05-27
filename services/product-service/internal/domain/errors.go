package domain

import "errors"

var ErrProductNotFound = errors.New("product not found")
var ErrProductDuplicate = errors.New("product duplicate")
var ErrProductInOrders = errors.New("product in orders")
var ErrReviewNotAllowed = errors.New("review not allowed")
var ErrReviewNotFound = errors.New("review not found")
var ErrInsufficientStock = errors.New("insufficient stock")
