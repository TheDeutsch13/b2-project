package postgres

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/zap"
)

// EnsureProductSchema applies product table extensions idempotently.
func EnsureProductSchema(ctx context.Context, db *pgxpool.Pool, logger *zap.Logger) error {
	statements := []string{
		`ALTER TABLE products ADD COLUMN IF NOT EXISTS brand VARCHAR(255) NOT NULL DEFAULT ''`,
		`ALTER TABLE products ADD COLUMN IF NOT EXISTS stock INTEGER NOT NULL DEFAULT 0`,
		`ALTER TABLE products ADD COLUMN IF NOT EXISTS images JSONB NOT NULL DEFAULT '[]'::jsonb`,
		`ALTER TABLE products ADD COLUMN IF NOT EXISTS specifications JSONB NOT NULL DEFAULT '[]'::jsonb`,
		`ALTER TABLE products ADD COLUMN IF NOT EXISTS variants JSONB NOT NULL DEFAULT '[]'::jsonb`,
		`ALTER TABLE products ADD COLUMN IF NOT EXISTS reviews JSONB NOT NULL DEFAULT '[]'::jsonb`,
		`ALTER TABLE products ADD COLUMN IF NOT EXISTS rating_avg NUMERIC(3, 2) NOT NULL DEFAULT 0`,
		`ALTER TABLE products ADD COLUMN IF NOT EXISTS rating_count INTEGER NOT NULL DEFAULT 0`,
	}

	for _, statement := range statements {
		if _, err := db.Exec(ctx, statement); err != nil {
			return err
		}
	}

	if logger != nil {
		logger.Info("product schema ensured")
	}

	return nil
}

// EnsureOrdersSchema creates orders/order_items when migrations were not applied (local dev).
func EnsureOrdersSchema(ctx context.Context, db *pgxpool.Pool, logger *zap.Logger) error {
	statements := []string{
		`
		CREATE TABLE IF NOT EXISTS orders (
			id BIGSERIAL PRIMARY KEY,
			user_id BIGINT NOT NULL,
			status VARCHAR(50) NOT NULL DEFAULT 'pending',
			contact_name VARCHAR(255) NOT NULL,
			contact_phone VARCHAR(50) NOT NULL,
			contact_email VARCHAR(255) NOT NULL,
			delivery_address TEXT NOT NULL,
			payment_method VARCHAR(50) NOT NULL,
			comment TEXT,
			total_amount NUMERIC(10, 2) NOT NULL DEFAULT 0,
			created_at TIMESTAMP NOT NULL DEFAULT NOW(),
			updated_at TIMESTAMP NOT NULL DEFAULT NOW()
		)`,
		`
		CREATE TABLE IF NOT EXISTS order_items (
			id BIGSERIAL PRIMARY KEY,
			order_id BIGINT NOT NULL REFERENCES orders(id) ON DELETE CASCADE,
			product_id BIGINT NOT NULL REFERENCES products(id) ON DELETE RESTRICT,
			quantity INT NOT NULL CHECK (quantity > 0),
			price NUMERIC(10, 2) NOT NULL CHECK (price >= 0)
		)`,
		`CREATE INDEX IF NOT EXISTS idx_orders_user_id ON orders(user_id)`,
		`CREATE INDEX IF NOT EXISTS idx_order_items_order_id ON order_items(order_id)`,
	}

	for _, statement := range statements {
		if _, err := db.Exec(ctx, statement); err != nil {
			return err
		}
	}

	if err := EnsureOrderDeliveryColumns(ctx, db, logger); err != nil {
		return err
	}

	if logger != nil {
		logger.Info("orders schema ensured")
	}

	return nil
}

// EnsureOrderDeliveryColumns adds delivery metadata columns to orders.
func EnsureOrderDeliveryColumns(ctx context.Context, db *pgxpool.Pool, logger *zap.Logger) error {
	statements := []string{
		`ALTER TABLE orders ADD COLUMN IF NOT EXISTS delivery_type VARCHAR(32) NOT NULL DEFAULT 'custom'`,
		`ALTER TABLE orders ADD COLUMN IF NOT EXISTS delivery_cost NUMERIC(10, 2) NOT NULL DEFAULT 0`,
		`ALTER TABLE orders ADD COLUMN IF NOT EXISTS delivery_city VARCHAR(255) NOT NULL DEFAULT ''`,
		`ALTER TABLE orders ADD COLUMN IF NOT EXISTS cdek_pvz_code VARCHAR(64) NOT NULL DEFAULT ''`,
		`ALTER TABLE orders ADD COLUMN IF NOT EXISTS delivery_payment VARCHAR(32) NOT NULL DEFAULT 'on_receipt'`,
	}

	for _, statement := range statements {
		if _, err := db.Exec(ctx, statement); err != nil {
			return err
		}
	}

	if logger != nil {
		logger.Info("order delivery columns ensured")
	}

	return nil
}

var fixedCategoryNames = []string{
	"Мыши",
	"Коврики",
	"Клавиатуры",
	"Аксессуары",
}

// EnsureFixedCategories removes stray categories and guarantees the shop's 4 categories exist.
func EnsureFixedCategories(ctx context.Context, db *pgxpool.Pool, logger *zap.Logger) error {
	statements := []string{
		`INSERT INTO categories (name) VALUES ('Аксессуары') ON CONFLICT (name) DO NOTHING`,
		`INSERT INTO categories (name) VALUES ('Клавиатуры') ON CONFLICT (name) DO NOTHING`,
		`
		UPDATE products SET category_id = (SELECT id FROM categories WHERE name = 'Аксессуары' LIMIT 1)
		WHERE category_id IN (
			SELECT id FROM categories WHERE name IN ('Глайды/грипсы', 'Рукава')
		)`,
		`
		UPDATE products SET category_id = NULL
		WHERE category_id IN (
			SELECT id FROM categories
			WHERE name NOT IN ('Мыши', 'Коврики', 'Клавиатуры', 'Аксессуары')
		)`,
		`DELETE FROM categories WHERE name NOT IN ('Мыши', 'Коврики', 'Клавиатуры', 'Аксессуары')`,
	}

	for _, statement := range statements {
		if _, err := db.Exec(ctx, statement); err != nil {
			return err
		}
	}

	for _, name := range fixedCategoryNames {
		if _, err := db.Exec(
			ctx,
			`INSERT INTO categories (name) VALUES ($1) ON CONFLICT (name) DO NOTHING`,
			name,
		); err != nil {
			return err
		}
	}

	if logger != nil {
		logger.Info("fixed categories ensured", zap.Strings("categories", fixedCategoryNames))
	}

	return nil
}
