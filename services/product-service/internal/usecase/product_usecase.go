package usecase

import (
	"context"
	"errors"
	"strings"

	"github.com/TheDeutsch13/b2-project/services/product-service/internal/domain"
)

var (
	ErrInvalidInput          = errors.New("invalid input")
	ErrCategoryAlreadyExists = errors.New("category already exists")
	ErrProductDuplicate      = domain.ErrProductDuplicate
	ErrProductInOrders       = domain.ErrProductInOrders
)

type ProductRepository interface {
	Create(ctx context.Context, product *domain.Product) (*domain.Product, error)
	Update(ctx context.Context, product *domain.Product) (*domain.Product, error)
	Delete(ctx context.Context, id int64) error
	List(ctx context.Context, categoryID *int64) ([]domain.Product, error)
	GetByID(ctx context.Context, id int64) (*domain.Product, error)
	ExistsDuplicate(
		ctx context.Context,
		title string,
		brand string,
		categoryID *int64,
		variantKey string,
		excludeID int64,
	) (bool, error)
}

type ProductInput struct {
	Title          string
	Description    string
	Price          float64
	CategoryID     *int64
	Brand          string
	Stock          int
	Images         []string
	Specifications []domain.ProductSpecification
	Variants       []string
	Reviews        []domain.ProductReview
}

type ProductUsecase struct {
	productRepo ProductRepository
	orderRepo   OrderReviewRepository
	reviewRepo  ProductReviewRepository
}

func NewProductUsecase(
	productRepo ProductRepository,
	orderRepo OrderReviewRepository,
	reviewRepo ProductReviewRepository,
) *ProductUsecase {
	return &ProductUsecase{
		productRepo: productRepo,
		orderRepo:   orderRepo,
		reviewRepo:  reviewRepo,
	}
}

func (u *ProductUsecase) Create(ctx context.Context, input ProductInput) (*domain.Product, error) {
	product, err := u.buildProduct(input)
	if err != nil {
		return nil, err
	}

	if err := u.ensureUniqueProduct(ctx, product, 0); err != nil {
		return nil, err
	}

	return u.productRepo.Create(ctx, product)
}

func (u *ProductUsecase) Update(ctx context.Context, id int64, input ProductInput) (*domain.Product, error) {
	product, err := u.buildProduct(input)
	if err != nil {
		return nil, err
	}

	product.ID = id

	if err := u.ensureUniqueProduct(ctx, product, id); err != nil {
		return nil, err
	}

	updatedProduct, err := u.productRepo.Update(ctx, product)
	if err != nil {
		if errors.Is(err, domain.ErrProductNotFound) {
			return nil, domain.ErrProductNotFound
		}

		return nil, err
	}

	return updatedProduct, nil
}

func (u *ProductUsecase) List(ctx context.Context, categoryID *int64) ([]domain.Product, error) {
	return u.productRepo.List(ctx, categoryID)
}

func (u *ProductUsecase) Delete(ctx context.Context, id int64) error {
	err := u.productRepo.Delete(ctx, id)
	if err != nil {
		if errors.Is(err, domain.ErrProductNotFound) {
			return domain.ErrProductNotFound
		}

		return err
	}

	return nil
}

func (u *ProductUsecase) GetByID(ctx context.Context, id int64) (*domain.Product, error) {
	product, err := u.productRepo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, domain.ErrProductNotFound) {
			return nil, domain.ErrProductNotFound
		}

		return nil, err
	}

	return product, nil
}

func (u *ProductUsecase) buildProduct(input ProductInput) (*domain.Product, error) {
	title := strings.TrimSpace(input.Title)

	if title == "" || input.Price < 0 || input.Stock < 0 {
		return nil, ErrInvalidInput
	}

	images := input.Images
	if images == nil {
		images = []string{}
	}

	specifications := input.Specifications
	if specifications == nil {
		specifications = []domain.ProductSpecification{}
	}

	variants := input.Variants
	if variants == nil {
		variants = []string{}
	}

	reviews := input.Reviews
	if reviews == nil {
		reviews = []domain.ProductReview{}
	}

	ratingAvg, ratingCount := domain.CalcRatingFromReviews(reviews)

	return &domain.Product{
		Title:          title,
		Description:    strings.TrimSpace(input.Description),
		Price:          input.Price,
		CategoryID:     input.CategoryID,
		Brand:          strings.TrimSpace(input.Brand),
		Stock:          input.Stock,
		Images:         images,
		Specifications: specifications,
		Variants:       variants,
		Reviews:        reviews,
		RatingAvg:      ratingAvg,
		RatingCount:    ratingCount,
	}, nil
}

func (u *ProductUsecase) ensureUniqueProduct(
	ctx context.Context,
	product *domain.Product,
	excludeID int64,
) error {
	exists, err := u.productRepo.ExistsDuplicate(
		ctx,
		product.Title,
		product.Brand,
		product.CategoryID,
		domain.PrimaryVariant(product.Variants),
		excludeID,
	)
	if err != nil {
		return err
	}

	if exists {
		return ErrProductDuplicate
	}

	return nil
}
