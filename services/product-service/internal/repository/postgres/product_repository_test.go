//go:build integration

package postgres

import (
	"context"
	"os"
	"testing"

	"github.com/TheDeutsch13/b2-project/services/product-service/internal/domain"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupTestDB(t *testing.T) *pgxpool.Pool {
	t.Helper()

	databaseURL := os.Getenv("TEST_DATABASE_URL")
	if databaseURL == "" {
		databaseURL = "postgres://b2user:b2password@localhost:5432/b2db?sslmode=disable"
	}

	ctx := context.Background()

	db, err := pgxpool.New(ctx, databaseURL)
	require.NoError(t, err)

	err = db.Ping(ctx)
	require.NoError(t, err)

	_, err = db.Exec(ctx, "TRUNCATE TABLE products RESTART IDENTITY CASCADE")
	require.NoError(t, err)

	t.Cleanup(func() {
		_, _ = db.Exec(ctx, "TRUNCATE TABLE products RESTART IDENTITY CASCADE")
		db.Close()
	})

	return db
}

func TestNewProductRepository(t *testing.T) {
	db := setupTestDB(t)

	repo := NewProductRepository(db)

	assert.NotNil(t, repo)
}

func TestProductRepository_Create(t *testing.T) {
	db := setupTestDB(t)
	repo := NewProductRepository(db)

	product := &domain.Product{
		Title:       "Test Product",
		Description: "Test Description",
		Price:       99.99,
	}

	createdProduct, err := repo.Create(context.Background(), product)

	assert.NoError(t, err)
	assert.NotNil(t, createdProduct)
	assert.Equal(t, int64(1), createdProduct.ID)
	assert.Equal(t, "Test Product", createdProduct.Title)
	assert.Equal(t, "Test Description", createdProduct.Description)
	assert.Equal(t, 99.99, createdProduct.Price)
	assert.False(t, createdProduct.CreatedAt.IsZero())
}

func TestProductRepository_List(t *testing.T) {
	db := setupTestDB(t)
	repo := NewProductRepository(db)

	firstProduct := &domain.Product{
		Title:       "First Product",
		Description: "First Description",
		Price:       100,
	}

	secondProduct := &domain.Product{
		Title:       "Second Product",
		Description: "Second Description",
		Price:       200,
	}

	_, err := repo.Create(context.Background(), firstProduct)
	require.NoError(t, err)

	_, err = repo.Create(context.Background(), secondProduct)
	require.NoError(t, err)

	products, err := repo.List(context.Background(), nil)

	assert.NoError(t, err)
	assert.Len(t, products, 2)

	assert.Equal(t, "Second Product", products[0].Title)
	assert.Equal(t, "First Product", products[1].Title)
}
