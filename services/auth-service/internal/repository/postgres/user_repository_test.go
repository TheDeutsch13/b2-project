//go:build integration

package postgres

import (
	"context"
	"os"
	"testing"

	"github.com/TheDeutsch13/b2-project/services/auth-service/internal/domain"
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

	_, err = db.Exec(ctx, `
		ALTER TABLE users ADD COLUMN IF NOT EXISTS role VARCHAR(20) NOT NULL DEFAULT 'user';
	`)
	require.NoError(t, err)

	_, err = db.Exec(ctx, "TRUNCATE TABLE refresh_tokens, users RESTART IDENTITY CASCADE")
	require.NoError(t, err)

	t.Cleanup(func() {
		_, _ = db.Exec(ctx, "TRUNCATE TABLE refresh_tokens, users RESTART IDENTITY CASCADE")
		db.Close()
	})

	return db
}

func TestNewUserRepository(t *testing.T) {
	db := setupTestDB(t)

	repo := NewUserRepository(db)

	assert.NotNil(t, repo)
}

func TestUserRepository_ExistsByEmail(t *testing.T) {
	db := setupTestDB(t)
	repo := NewUserRepository(db)

	exists, err := repo.ExistsByEmail(context.Background(), "test@example.com")

	assert.NoError(t, err)
	assert.False(t, exists)
}

func TestUserRepository_Create(t *testing.T) {
	db := setupTestDB(t)
	repo := NewUserRepository(db)

	user := &domain.User{
		Email:        "test@example.com",
		PasswordHash: "hashed-password",
		Role:         domain.RoleUser,
	}

	createdUser, err := repo.Create(context.Background(), user)

	assert.NoError(t, err)
	assert.NotNil(t, createdUser)
	assert.Equal(t, int64(1), createdUser.ID)
	assert.Equal(t, "test@example.com", createdUser.Email)
	assert.Equal(t, "hashed-password", createdUser.PasswordHash)
	assert.False(t, createdUser.CreatedAt.IsZero())
}

func TestUserRepository_ExistsByEmail_AfterCreate(t *testing.T) {
	db := setupTestDB(t)
	repo := NewUserRepository(db)

	_, err := repo.Create(context.Background(), &domain.User{
		Email:        "test@example.com",
		PasswordHash: "hashed-password",
		Role:         domain.RoleUser,
	})
	require.NoError(t, err)

	exists, err := repo.ExistsByEmail(context.Background(), "test@example.com")

	assert.NoError(t, err)
	assert.True(t, exists)
}
