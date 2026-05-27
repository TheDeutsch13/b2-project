package postgres

import (
	"context"
	"errors"

	"github.com/TheDeutsch13/b2-project/services/auth-service/internal/domain"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

const userSelectColumns = `
	id, email, password_hash, role,
	first_name, last_name, nickname, birth_date, gender, phone, avatar_url,
	created_at, updated_at
`

type UserRepository struct {
	db *pgxpool.Pool
}

func NewUserRepository(db *pgxpool.Pool) *UserRepository {
	return &UserRepository{db: db}
}

func scanUser(row pgx.Row) (*domain.User, error) {
	var user domain.User

	err := row.Scan(
		&user.ID,
		&user.Email,
		&user.PasswordHash,
		&user.Role,
		&user.FirstName,
		&user.LastName,
		&user.Nickname,
		&user.BirthDate,
		&user.Gender,
		&user.Phone,
		&user.AvatarURL,
		&user.CreatedAt,
		&user.UpdatedAt,
	)

	if err != nil {
		return nil, err
	}

	return &user, nil
}

func (r *UserRepository) ExistsByEmail(ctx context.Context, email string) (bool, error) {
	query := `SELECT EXISTS(SELECT 1 FROM users WHERE email = $1)`

	var exists bool
	err := r.db.QueryRow(ctx, query, email).Scan(&exists)

	return exists, err
}

func (r *UserRepository) Create(ctx context.Context, user *domain.User) (*domain.User, error) {
	query := `
		INSERT INTO users (email, password_hash, role)
		VALUES ($1, $2, $3)
		RETURNING ` + userSelectColumns + `
	`

	return scanUser(r.db.QueryRow(ctx, query, user.Email, user.PasswordHash, user.Role))
}

func (r *UserRepository) GetByEmail(ctx context.Context, email string) (*domain.User, error) {
	query := `SELECT ` + userSelectColumns + ` FROM users WHERE email = $1`

	user, err := scanUser(r.db.QueryRow(ctx, query, email))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrUserNotFound
		}

		return nil, err
	}

	return user, nil
}

func (r *UserRepository) GetByID(ctx context.Context, id int64) (*domain.User, error) {
	query := `SELECT ` + userSelectColumns + ` FROM users WHERE id = $1`

	user, err := scanUser(r.db.QueryRow(ctx, query, id))
	if err != nil {
		return nil, err
	}

	return user, nil
}

func (r *UserRepository) UpdateProfile(
	ctx context.Context,
	userID int64,
	input domain.UserProfileInput,
) (*domain.User, error) {
	query := `
		UPDATE users
		SET
			first_name = $1,
			last_name = $2,
			nickname = $3,
			birth_date = $4,
			gender = $5,
			phone = $6,
			updated_at = NOW()
		WHERE id = $7
		RETURNING ` + userSelectColumns

	return scanUser(
		r.db.QueryRow(
			ctx,
			query,
			input.FirstName,
			input.LastName,
			input.Nickname,
			input.BirthDate,
			input.Gender,
			input.Phone,
			userID,
		),
	)
}

func (r *UserRepository) UpdateAvatar(
	ctx context.Context,
	userID int64,
	avatarURL string,
) (*domain.User, error) {
	query := `
		UPDATE users
		SET avatar_url = $1, updated_at = NOW()
		WHERE id = $2
		RETURNING ` + userSelectColumns

	return scanUser(r.db.QueryRow(ctx, query, avatarURL, userID))
}

func (r *UserRepository) UpdateRole(
	ctx context.Context,
	userID int64,
	role string,
) (*domain.User, error) {
	query := `
		UPDATE users
		SET role = $1, updated_at = NOW()
		WHERE id = $2
		RETURNING ` + userSelectColumns

	user, err := scanUser(r.db.QueryRow(ctx, query, role, userID))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrUserNotFound
		}

		return nil, err
	}

	return user, nil
}

func (r *UserRepository) List(ctx context.Context) ([]domain.User, error) {
	query := `SELECT ` + userSelectColumns + ` FROM users ORDER BY id DESC`

	rows, err := r.db.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	users := make([]domain.User, 0)

	for rows.Next() {
		user, err := scanUser(rows)
		if err != nil {
			return nil, err
		}

		users = append(users, *user)
	}

	return users, rows.Err()
}

func (r *UserRepository) ListPublicByIDs(ctx context.Context, ids []int64) ([]domain.User, error) {
	if len(ids) == 0 {
		return []domain.User{}, nil
	}

	query := `
		SELECT id, first_name, last_name, nickname, avatar_url
		FROM users
		WHERE id = ANY($1)
	`

	rows, err := r.db.Query(ctx, query, ids)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	users := make([]domain.User, 0)

	for rows.Next() {
		var user domain.User

		if err := rows.Scan(
			&user.ID,
			&user.FirstName,
			&user.LastName,
			&user.Nickname,
			&user.AvatarURL,
		); err != nil {
			return nil, err
		}

		users = append(users, user)
	}

	return users, rows.Err()
}
