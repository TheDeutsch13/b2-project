package usecase

import (
	"context"
	"strings"
	"time"

	"github.com/TheDeutsch13/b2-project/services/product-service/internal/domain"
)

type OrderReviewRepository interface {
	UserHasReceivedProduct(ctx context.Context, userID, productID int64) (bool, error)
}

type ProductReviewRepository interface {
	ListReviewsByUserID(ctx context.Context, userID int64) ([]domain.UserProductReview, error)
	ListAllReviews(ctx context.Context, filter domain.ReviewListFilter) ([]domain.AdminProductReview, error)
}

func (u *ProductUsecase) UpsertUserReview(
	ctx context.Context,
	userID int64,
	productID int64,
	author string,
	rating int,
	text string,
) (*domain.Product, error) {
	author = strings.TrimSpace(author)
	text = strings.TrimSpace(text)

	if userID == 0 || productID == 0 || author == "" || rating < 1 || rating > 5 || text == "" {
		return nil, ErrInvalidInput
	}

	if u.orderRepo == nil {
		return nil, domain.ErrReviewNotAllowed
	}

	allowed, err := u.orderRepo.UserHasReceivedProduct(ctx, userID, productID)
	if err != nil || !allowed {
		return nil, domain.ErrReviewNotAllowed
	}

	product, err := u.productRepo.GetByID(ctx, productID)
	if err != nil {
		return nil, err
	}

	now := time.Now().UTC().Format(time.RFC3339)
	reviews := product.Reviews
	if reviews == nil {
		reviews = []domain.ProductReview{}
	}

	found := false

	for index, item := range reviews {
		if item.UserID == userID {
			createdAt := item.CreatedAt
			if createdAt == "" {
				createdAt = now
			}

			reviews[index] = domain.ProductReview{
				UserID:    userID,
				Author:    author,
				Rating:    rating,
				Text:      text,
				CreatedAt: createdAt,
			}
			found = true

			break
		}
	}

	if !found {
		reviews = append(reviews, domain.ProductReview{
			UserID:    userID,
			Author:    author,
			Rating:    rating,
			Text:      text,
			CreatedAt: now,
		})
	}

	product.Reviews = reviews
	product.RatingAvg, product.RatingCount = domain.CalcRatingFromReviews(reviews)

	return u.productRepo.Update(ctx, product)
}

func (u *ProductUsecase) DeleteUserReview(
	ctx context.Context,
	userID int64,
	productID int64,
) (*domain.Product, error) {
	if userID == 0 || productID == 0 {
		return nil, ErrInvalidInput
	}

	product, err := u.productRepo.GetByID(ctx, productID)
	if err != nil {
		return nil, err
	}

	reviews := make([]domain.ProductReview, 0, len(product.Reviews))
	removed := false

	for _, item := range product.Reviews {
		if item.UserID == userID {
			removed = true
			continue
		}

		reviews = append(reviews, item)
	}

	if !removed {
		return nil, domain.ErrReviewNotFound
	}

	product.Reviews = reviews
	product.RatingAvg, product.RatingCount = domain.CalcRatingFromReviews(reviews)

	return u.productRepo.Update(ctx, product)
}

func (u *ProductUsecase) ListUserReviews(
	ctx context.Context,
	userID int64,
) ([]domain.UserProductReview, error) {
	if userID == 0 {
		return nil, ErrInvalidInput
	}

	if u.reviewRepo == nil {
		return []domain.UserProductReview{}, nil
	}

	return u.reviewRepo.ListReviewsByUserID(ctx, userID)
}

func (u *ProductUsecase) ListAllReviews(
	ctx context.Context,
	filter domain.ReviewListFilter,
) ([]domain.AdminProductReview, error) {
	if u.reviewRepo == nil {
		return []domain.AdminProductReview{}, nil
	}

	filter.Query = strings.TrimSpace(filter.Query)

	return u.reviewRepo.ListAllReviews(ctx, filter)
}
