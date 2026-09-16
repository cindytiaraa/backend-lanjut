package repository

import (
	"context"
	"errors"

	"latihan-repository/app/model"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

var ErrUserNotFound = errors.New("user not found")

type UserRepository interface {
	Create(ctx context.Context, user model.User) (model.User, error)
	FindByUsername(ctx context.Context, username string) (model.User, error)
	FindByID(ctx context.Context, id int) (model.User, error)
}

type userPostgresRepository struct {
	pool *pgxpool.Pool
}

func NewUserRepository(pool *pgxpool.Pool) UserRepository {
	return &userPostgresRepository{
		pool: pool,
	}
}

func (r *userPostgresRepository) Create(
	ctx context.Context,
	user model.User,
) (model.User, error) {
	var created model.User

	err := r.pool.QueryRow(
		ctx,
		`
		INSERT INTO users (
			username,
			email,
			password,
			role,
			is_active
		)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id, username, email, password, role, is_active, created_at
		`,
		user.Username,
		user.Email,
		user.Password,
		user.Role,
		user.IsActive,
	).Scan(
		&created.ID,
		&created.Username,
		&created.Email,
		&created.Password,
		&created.Role,
		&created.IsActive,
		&created.CreatedAt,
	)

	if err != nil {
		return model.User{}, err
	}

	return created, nil
}

func (r *userPostgresRepository) FindByUsername(
	ctx context.Context,
	username string,
) (model.User, error) {
	var user model.User

	err := r.pool.QueryRow(
		ctx,
		`
		SELECT
			id,
			username,
			email,
			password,
			role,
			is_active,
			created_at
		FROM users
		WHERE LOWER(username) = LOWER($1)
		`,
		username,
	).Scan(
		&user.ID,
		&user.Username,
		&user.Email,
		&user.Password,
		&user.Role,
		&user.IsActive,
		&user.CreatedAt,
	)

	if errors.Is(err, pgx.ErrNoRows) {
		return model.User{}, ErrUserNotFound
	}

	if err != nil {
		return model.User{}, err
	}

	return user, nil
}

func (r *userPostgresRepository) FindByID(
	ctx context.Context,
	id int,
) (model.User, error) {
	var user model.User

	err := r.pool.QueryRow(
		ctx,
		`
		SELECT
			id,
			username,
			email,
			password,
			role,
			is_active,
			created_at
		FROM users
		WHERE id = $1
		`,
		id,
	).Scan(
		&user.ID,
		&user.Username,
		&user.Email,
		&user.Password,
		&user.Role,
		&user.IsActive,
		&user.CreatedAt,
	)

	if errors.Is(err, pgx.ErrNoRows) {
		return model.User{}, ErrUserNotFound
	}

	if err != nil {
		return model.User{}, err
	}

	return user, nil
}
