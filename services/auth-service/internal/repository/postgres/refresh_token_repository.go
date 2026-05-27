package postgres

import (
	"context"
	"time"

	"github.com/TheDeutsch13/b2-project/services/auth-service/internal/domain"
	"github.com/jackc/pgx/v5/pgxpool"
)

type RefreshTokenRepository struct {
	db *pgxpool.Pool
}

func NewRefreshTokenRepository(db *pgxpool.Pool) *RefreshTokenRepository {
	return &RefreshTokenRepository{db: db}
}

func (r *RefreshTokenRepository) Create(ctx context.Context, token *domain.RefreshToken) error {
	query := `
		INSERT INTO refresh_tokens (user_id, token, expires_at)
		VALUES ($1, $2, $3)
	`

	_, err := r.db.Exec(ctx, query, token.UserID, token.Token, token.ExpiresAt)

	return err
}

func (r *RefreshTokenRepository) GetByToken(ctx context.Context, token string) (*domain.RefreshToken, error) {
	query := `
		SELECT id, user_id, token, expires_at, created_at
		FROM refresh_tokens
		WHERE token = $1
	`

	var refreshToken domain.RefreshToken

	err := r.db.QueryRow(ctx, query, token).Scan(
		&refreshToken.ID,
		&refreshToken.UserID,
		&refreshToken.Token,
		&refreshToken.ExpiresAt,
		&refreshToken.CreatedAt,
	)
	if err != nil {
		return nil, err
	}

	return &refreshToken, nil
}

func (r *RefreshTokenRepository) DeleteByToken(ctx context.Context, token string) error {
	query := `DELETE FROM refresh_tokens WHERE token = $1`

	_, err := r.db.Exec(ctx, query, token)

	return err
}

func (r *RefreshTokenRepository) DeleteExpired(ctx context.Context, now time.Time) error {
	query := `DELETE FROM refresh_tokens WHERE expires_at < $1`

	_, err := r.db.Exec(ctx, query, now)

	return err
}
