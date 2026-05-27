package postgres

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/TheDeutsch13/b2-project/services/product-service/internal/domain"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type SupportRepository struct {
	db *pgxpool.Pool
}

func NewSupportRepository(db *pgxpool.Pool) *SupportRepository {
	return &SupportRepository{db: db}
}

func scanThread(row pgx.Row) (*domain.SupportThread, error) {
	var thread domain.SupportThread

	err := row.Scan(
		&thread.ID,
		&thread.UserID,
		&thread.UserEmail,
		&thread.Status,
		&thread.CreatedAt,
		&thread.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}

	return &thread, nil
}

func scanMessage(row pgx.Row) (*domain.SupportMessage, error) {
	var message domain.SupportMessage

	err := row.Scan(
		&message.ID,
		&message.ThreadID,
		&message.SenderID,
		&message.SenderRole,
		&message.SenderName,
		&message.Body,
		&message.CreatedAt,
	)
	if err != nil {
		return nil, err
	}

	return &message, nil
}

func (r *SupportRepository) GetOrCreateThread(
	ctx context.Context,
	userID int64,
	userEmail string,
) (*domain.SupportThread, error) {
	thread, err := r.GetThreadByUserID(ctx, userID)
	if err == nil {
		return thread, nil
	}

	if !errors.Is(err, pgx.ErrNoRows) {
		return nil, err
	}

	query := `
		INSERT INTO support_threads (user_id, user_email)
		VALUES ($1, $2)
		RETURNING id, user_id, user_email, status, created_at, updated_at
	`

	return scanThread(r.db.QueryRow(ctx, query, userID, strings.TrimSpace(userEmail)))
}

func (r *SupportRepository) GetThreadByUserID(
	ctx context.Context,
	userID int64,
) (*domain.SupportThread, error) {
	query := `
		SELECT id, user_id, user_email, status, created_at, updated_at
		FROM support_threads
		WHERE user_id = $1
	`

	return scanThread(r.db.QueryRow(ctx, query, userID))
}

func (r *SupportRepository) GetThreadByID(
	ctx context.Context,
	threadID int64,
) (*domain.SupportThread, error) {
	query := `
		SELECT id, user_id, user_email, status, created_at, updated_at
		FROM support_threads
		WHERE id = $1
	`

	thread, err := scanThread(r.db.QueryRow(ctx, query, threadID))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, err
		}
		return nil, err
	}

	return thread, nil
}

func (r *SupportRepository) ListThreads(
	ctx context.Context,
	openOnly bool,
) ([]domain.SupportThreadListItem, error) {
	whereClause := ""
	if openOnly {
		whereClause = `WHERE t.status = 'open'`
	}

	query := `
		SELECT
			t.id, t.user_id, t.user_email, t.status, t.created_at, t.updated_at,
			COALESCE((
				SELECT m.body
				FROM support_messages m
				WHERE m.thread_id = t.id
				ORDER BY m.created_at DESC
				LIMIT 1
			), '') AS last_body,
			(
				SELECT m.created_at
				FROM support_messages m
				WHERE m.thread_id = t.id
				ORDER BY m.created_at DESC
				LIMIT 1
			) AS last_at,
			(
				SELECT COUNT(*)::int
				FROM support_messages m
				WHERE m.thread_id = t.id
			) AS message_count
		FROM support_threads t
		` + whereClause + `
		ORDER BY t.updated_at DESC
	`

	rows, err := r.db.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := make([]domain.SupportThreadListItem, 0)

	for rows.Next() {
		var item domain.SupportThreadListItem
		var lastAt *time.Time

		if err := rows.Scan(
			&item.ID,
			&item.UserID,
			&item.UserEmail,
			&item.Status,
			&item.CreatedAt,
			&item.UpdatedAt,
			&item.LastMessageBody,
			&lastAt,
			&item.MessageCount,
		); err != nil {
			return nil, err
		}

		item.LastMessageAt = lastAt
		items = append(items, item)
	}

	return items, rows.Err()
}

func (r *SupportRepository) ListMessages(
	ctx context.Context,
	threadID int64,
) ([]domain.SupportMessage, error) {
	query := `
		SELECT id, thread_id, sender_id, sender_role, sender_name, body, created_at
		FROM support_messages
		WHERE thread_id = $1
		ORDER BY created_at ASC
	`

	rows, err := r.db.Query(ctx, query, threadID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	messages := make([]domain.SupportMessage, 0)

	for rows.Next() {
		message, err := scanMessage(rows)
		if err != nil {
			return nil, err
		}

		messages = append(messages, *message)
	}

	return messages, rows.Err()
}

func (r *SupportRepository) CreateMessage(
	ctx context.Context,
	threadID int64,
	senderID int64,
	senderRole string,
	senderName string,
	body string,
) (*domain.SupportMessage, error) {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(ctx)

	messageQuery := `
		INSERT INTO support_messages (thread_id, sender_id, sender_role, sender_name, body)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id, thread_id, sender_id, sender_role, sender_name, body, created_at
	`

	message, err := scanMessage(
		tx.QueryRow(
			ctx,
			messageQuery,
			threadID,
			senderID,
			senderRole,
			strings.TrimSpace(senderName),
			strings.TrimSpace(body),
		),
	)
	if err != nil {
		return nil, err
	}

	_, err = tx.Exec(
		ctx,
		`UPDATE support_threads SET updated_at = NOW(), status = 'open' WHERE id = $1`,
		threadID,
	)
	if err != nil {
		return nil, err
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, err
	}

	return message, nil
}

func (r *SupportRepository) UpdateThreadStatus(
	ctx context.Context,
	threadID int64,
	status string,
) (*domain.SupportThread, error) {
	query := `
		UPDATE support_threads
		SET status = $1, updated_at = NOW()
		WHERE id = $2
		RETURNING id, user_id, user_email, status, created_at, updated_at
	`

	thread, err := scanThread(r.db.QueryRow(ctx, query, status, threadID))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, err
		}
		return nil, err
	}

	return thread, nil
}

func (r *SupportRepository) DeleteThread(ctx context.Context, threadID int64) error {
	result, err := r.db.Exec(ctx, `DELETE FROM support_threads WHERE id = $1`, threadID)
	if err != nil {
		return err
	}

	if result.RowsAffected() == 0 {
		return pgx.ErrNoRows
	}

	return nil
}
