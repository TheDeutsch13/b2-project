package postgres

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/zap"
)

func EnsureSupportSchema(ctx context.Context, db *pgxpool.Pool, logger *zap.Logger) error {
	statements := []string{
		`
		CREATE TABLE IF NOT EXISTS support_threads (
			id BIGSERIAL PRIMARY KEY,
			user_id BIGINT NOT NULL UNIQUE,
			user_email VARCHAR(255) NOT NULL DEFAULT '',
			status VARCHAR(20) NOT NULL DEFAULT 'open',
			created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
			updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
		)`,
		`
		CREATE TABLE IF NOT EXISTS support_messages (
			id BIGSERIAL PRIMARY KEY,
			thread_id BIGINT NOT NULL REFERENCES support_threads(id) ON DELETE CASCADE,
			sender_id BIGINT NOT NULL,
			sender_role VARCHAR(20) NOT NULL,
			sender_name VARCHAR(255) NOT NULL DEFAULT '',
			body TEXT NOT NULL,
			created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
		)`,
		`CREATE INDEX IF NOT EXISTS idx_support_messages_thread_id ON support_messages(thread_id)`,
		`CREATE INDEX IF NOT EXISTS idx_support_threads_updated_at ON support_threads(updated_at DESC)`,
	}

	for _, statement := range statements {
		if _, err := db.Exec(ctx, statement); err != nil {
			return err
		}
	}

	if logger != nil {
		logger.Info("support schema ensured")
	}

	return nil
}
