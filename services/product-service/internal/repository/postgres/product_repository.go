package postgres

import (
	"context"
	"encoding/json"
	"errors"

	"github.com/TheDeutsch13/b2-project/services/product-service/internal/domain"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

const productSelectColumns = `
	p.id,
	p.category_id,
	COALESCE(c.name, '') AS category_name,
	p.title,
	p.description,
	p.price,
	p.brand,
	p.stock,
	p.images,
	p.specifications,
	p.variants,
	p.reviews,
	p.rating_avg,
	p.rating_count,
	p.created_at
`

type ProductRepository struct {
	db *pgxpool.Pool
}

func NewProductRepository(db *pgxpool.Pool) *ProductRepository {
	return &ProductRepository{db: db}
}

func (r *ProductRepository) scanProduct(row pgx.Row) (*domain.Product, error) {
	var product domain.Product
	var imagesRaw []byte
	var specificationsRaw []byte
	var variantsRaw []byte
	var reviewsRaw []byte

	err := row.Scan(
		&product.ID,
		&product.CategoryID,
		&product.CategoryName,
		&product.Title,
		&product.Description,
		&product.Price,
		&product.Brand,
		&product.Stock,
		&imagesRaw,
		&specificationsRaw,
		&variantsRaw,
		&reviewsRaw,
		&product.RatingAvg,
		&product.RatingCount,
		&product.CreatedAt,
	)
	if err != nil {
		return nil, err
	}

	product.Images, err = unmarshalStringSlice(imagesRaw)
	if err != nil {
		return nil, err
	}

	product.Specifications, err = unmarshalSpecifications(specificationsRaw)
	if err != nil {
		return nil, err
	}

	product.Variants, err = unmarshalStringSlice(variantsRaw)
	if err != nil {
		return nil, err
	}

	product.Reviews, err = unmarshalReviews(reviewsRaw)
	if err != nil {
		return nil, err
	}

	return &product, nil
}

func (r *ProductRepository) Create(ctx context.Context, product *domain.Product) (*domain.Product, error) {
	imagesJSON, err := marshalStringSlice(product.Images)
	if err != nil {
		return nil, err
	}

	specificationsJSON, err := marshalSpecifications(product.Specifications)
	if err != nil {
		return nil, err
	}

	variantsJSON, err := marshalStringSlice(product.Variants)
	if err != nil {
		return nil, err
	}

	reviewsJSON, err := marshalReviews(product.Reviews)
	if err != nil {
		return nil, err
	}

	query := `
		INSERT INTO products (
			title, description, price, category_id, brand, stock,
			images, specifications, variants, reviews, rating_avg, rating_count
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)
		RETURNING id
	`

	var id int64

	err = r.db.QueryRow(
		ctx,
		query,
		product.Title,
		product.Description,
		product.Price,
		product.CategoryID,
		product.Brand,
		product.Stock,
		string(imagesJSON),
		string(specificationsJSON),
		string(variantsJSON),
		string(reviewsJSON),
		product.RatingAvg,
		product.RatingCount,
	).Scan(&id)
	if err != nil {
		return nil, err
	}

	return r.GetByID(ctx, id)
}

func (r *ProductRepository) Update(ctx context.Context, product *domain.Product) (*domain.Product, error) {
	imagesJSON, err := marshalStringSlice(product.Images)
	if err != nil {
		return nil, err
	}

	specificationsJSON, err := marshalSpecifications(product.Specifications)
	if err != nil {
		return nil, err
	}

	variantsJSON, err := marshalStringSlice(product.Variants)
	if err != nil {
		return nil, err
	}

	reviewsJSON, err := marshalReviews(product.Reviews)
	if err != nil {
		return nil, err
	}

	query := `
		UPDATE products
		SET
			title = $2,
			description = $3,
			price = $4,
			category_id = $5,
			brand = $6,
			stock = $7,
			images = $8,
			specifications = $9,
			variants = $10,
			reviews = $11,
			rating_avg = $12,
			rating_count = $13,
			updated_at = NOW()
		WHERE id = $1
	`

	commandTag, err := r.db.Exec(
		ctx,
		query,
		product.ID,
		product.Title,
		product.Description,
		product.Price,
		product.CategoryID,
		product.Brand,
		product.Stock,
		string(imagesJSON),
		string(specificationsJSON),
		string(variantsJSON),
		string(reviewsJSON),
		product.RatingAvg,
		product.RatingCount,
	)
	if err != nil {
		return nil, err
	}

	if commandTag.RowsAffected() == 0 {
		return nil, domain.ErrProductNotFound
	}

	return r.GetByID(ctx, product.ID)
}

func (r *ProductRepository) ExistsDuplicate(
	ctx context.Context,
	title string,
	brand string,
	categoryID *int64,
	variantKey string,
	excludeID int64,
) (bool, error) {
	query := `
		SELECT EXISTS (
			SELECT 1
			FROM products p
			WHERE LOWER(TRIM(p.title)) = LOWER(TRIM($1))
				AND LOWER(TRIM(p.brand)) = LOWER(TRIM($2))
				AND p.category_id IS NOT DISTINCT FROM $3
				AND LOWER(TRIM(COALESCE(NULLIF(p.variants->>0, ''), 'стандарт'))) = LOWER(TRIM($4))
				AND ($5 = 0 OR p.id <> $5)
		)
	`

	var exists bool

	err := r.db.QueryRow(ctx, query, title, brand, categoryID, variantKey, excludeID).Scan(&exists)

	return exists, err
}

func (r *ProductRepository) Delete(ctx context.Context, id int64) error {
	var inOrders bool

	err := r.db.QueryRow(
		ctx,
		`SELECT EXISTS (SELECT 1 FROM order_items WHERE product_id = $1)`,
		id,
	).Scan(&inOrders)
	if err != nil {
		return err
	}

	if inOrders {
		return domain.ErrProductInOrders
	}

	commandTag, err := r.db.Exec(ctx, `DELETE FROM products WHERE id = $1`, id)
	if err != nil {
		return err
	}

	if commandTag.RowsAffected() == 0 {
		return domain.ErrProductNotFound
	}

	return nil
}

func (r *ProductRepository) List(ctx context.Context, categoryID *int64) ([]domain.Product, error) {
	query := `
		SELECT ` + productSelectColumns + `
		FROM products p
		LEFT JOIN categories c ON c.id = p.category_id
	`
	args := []interface{}{}

	if categoryID != nil {
		query += ` WHERE p.category_id = $1`
		args = append(args, *categoryID)
	}

	query += ` ORDER BY p.id DESC`

	rows, err := r.db.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	products := make([]domain.Product, 0)

	for rows.Next() {
		product, err := r.scanProduct(rows)
		if err != nil {
			return nil, err
		}

		products = append(products, *product)
	}

	return products, rows.Err()
}

func (r *ProductRepository) GetByID(ctx context.Context, id int64) (*domain.Product, error) {
	query := `
		SELECT ` + productSelectColumns + `
		FROM products p
		LEFT JOIN categories c ON c.id = p.category_id
		WHERE p.id = $1
	`

	product, err := r.scanProduct(r.db.QueryRow(ctx, query, id))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrProductNotFound
		}

		return nil, err
	}

	return product, nil
}

func (r *ProductRepository) ListReviewsByUserID(
	ctx context.Context,
	userID int64,
) ([]domain.UserProductReview, error) {
	query := `
		SELECT p.id, p.title, p.images, elem::text
		FROM products p
		CROSS JOIN LATERAL jsonb_array_elements(
			CASE
				WHEN jsonb_typeof(p.reviews) = 'array' THEN p.reviews
				ELSE '[]'::jsonb
			END
		) AS elem
		WHERE COALESCE((elem->>'user_id')::bigint, 0) = $1
		ORDER BY elem->>'created_at' DESC NULLS LAST, p.id DESC
	`

	rows, err := r.db.Query(ctx, query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	result := make([]domain.UserProductReview, 0)

	for rows.Next() {
		var productID int64
		var title string
		var imagesRaw []byte
		var reviewRaw []byte

		if err := rows.Scan(&productID, &title, &imagesRaw, &reviewRaw); err != nil {
			return nil, err
		}

		var review domain.ProductReview
		if err := json.Unmarshal(reviewRaw, &review); err != nil {
			return nil, err
		}

		images, err := unmarshalStringSlice(imagesRaw)
		if err != nil {
			return nil, err
		}

		image := ""
		if len(images) > 0 {
			image = images[0]
		}

		result = append(result, domain.UserProductReview{
			ProductID:    productID,
			ProductTitle: title,
			ProductImage: image,
			Review:       review,
		})
	}

	return result, rows.Err()
}

func (r *ProductRepository) ListAllReviews(
	ctx context.Context,
	filter domain.ReviewListFilter,
) ([]domain.AdminProductReview, error) {
	query := `
		SELECT
			p.id,
			p.title,
			COALESCE((elem->>'user_id')::bigint, 0),
			COALESCE(elem->>'author', ''),
			COALESCE((elem->>'rating')::int, 0),
			COALESCE(elem->>'text', ''),
			COALESCE(elem->>'created_at', '')
		FROM products p
		CROSS JOIN LATERAL jsonb_array_elements(
			CASE
				WHEN jsonb_typeof(p.reviews) = 'array' THEN p.reviews
				ELSE '[]'::jsonb
			END
		) AS elem
		WHERE COALESCE(elem->>'text', '') <> ''
		  AND COALESCE((elem->>'rating')::int, 0) BETWEEN 1 AND 5
		  AND ($1 = 0 OR (elem->>'rating')::int = $1)
		  AND ($2 = 0 OR p.id = $2)
		  AND (
			$3 = ''
			OR p.title ILIKE '%' || $3 || '%'
			OR COALESCE(elem->>'author', '') ILIKE '%' || $3 || '%'
			OR COALESCE(elem->>'text', '') ILIKE '%' || $3 || '%'
		  )
		ORDER BY elem->>'created_at' DESC NULLS LAST, p.id DESC
	`

	rows, err := r.db.Query(ctx, query, filter.Rating, filter.ProductID, filter.Query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	result := make([]domain.AdminProductReview, 0)

	for rows.Next() {
		var item domain.AdminProductReview
		if err := rows.Scan(
			&item.ProductID,
			&item.ProductTitle,
			&item.UserID,
			&item.Author,
			&item.Rating,
			&item.Text,
			&item.CreatedAt,
		); err != nil {
			return nil, err
		}

		result = append(result, item)
	}

	return result, rows.Err()
}
