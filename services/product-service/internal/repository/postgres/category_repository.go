package postgres

import (
	"context"

	"github.com/TheDeutsch13/b2-project/services/product-service/internal/domain"
	"github.com/jackc/pgx/v5/pgxpool"
)

type CategoryRepository struct {
	db *pgxpool.Pool
}

func NewCategoryRepository(db *pgxpool.Pool) *CategoryRepository {
	return &CategoryRepository{db: db}
}

func (r *CategoryRepository) List(ctx context.Context) ([]domain.Category, error) {
	query := `
		SELECT id, name, created_at
		FROM categories
		WHERE name IN ('Мыши', 'Коврики', 'Клавиатуры', 'Аксессуары')
		ORDER BY CASE name
			WHEN 'Мыши' THEN 1
			WHEN 'Коврики' THEN 2
			WHEN 'Клавиатуры' THEN 3
			WHEN 'Аксессуары' THEN 4
			ELSE 5
		END
	`

	rows, err := r.db.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	categories := make([]domain.Category, 0)

	for rows.Next() {
		var category domain.Category

		if err := rows.Scan(&category.ID, &category.Name, &category.CreatedAt); err != nil {
			return nil, err
		}

		categories = append(categories, category)
	}

	return categories, rows.Err()
}

func (r *CategoryRepository) Create(ctx context.Context, name string) (*domain.Category, error) {
	query := `
		INSERT INTO categories (name)
		VALUES ($1)
		RETURNING id, name, created_at
	`

	var category domain.Category

	err := r.db.QueryRow(ctx, query, name).Scan(&category.ID, &category.Name, &category.CreatedAt)
	if err != nil {
		return nil, mapPgError(err)
	}

	return &category, nil
}
