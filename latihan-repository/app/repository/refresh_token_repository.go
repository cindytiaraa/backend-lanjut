package repository

import (
	"context"
	"errors"

	"latihan-repository/app/model"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

var ErrRefreshTokenNotFound = errors.New("refresh token not found")

type RefreshTokenRepository interface {
	Create(ctx context.Context, token model.RefreshToken) error
	FindActive(ctx context.Context, tokenHash string) (model.RefreshToken, error)
	Revoke(ctx context.Context, id int64) error
}

type refreshTokenPostgresRepository struct {
	pool *pgxpool.Pool
}

func NewRefreshTokenRepository(pool *pgxpool.Pool) RefreshTokenRepository {
	return &refreshTokenPostgresRepository{pool: pool}
}

func (r *refreshTokenPostgresRepository) Create(
	ctx context.Context,
	token model.RefreshToken,
) error {
	_, err := r.pool.Exec(ctx, `
		INSERT INTO refresh_tokens (user_id, token_hash, expires_at)
		VALUES ($1, $2, $3)
	`, token.UserID, token.TokenHash, token.ExpiresAt)

	return err
}

func (r *refreshTokenPostgresRepository) FindActive(
	ctx context.Context,
	tokenHash string,
) (model.RefreshToken, error) {
	var token model.RefreshToken

	err := r.pool.QueryRow(ctx, `
		SELECT id, user_id, token_hash, expires_at, revoked_at, created_at
		FROM refresh_tokens
		WHERE token_hash = $1
		  AND revoked_at IS NULL
		  AND expires_at > NOW()
	`, tokenHash).Scan(
		&token.ID,
		&token.UserID,
		&token.TokenHash,
		&token.ExpiresAt,
		&token.RevokedAt,
		&token.CreatedAt,
	)

	if errors.Is(err, pgx.ErrNoRows) {
		return model.RefreshToken{}, ErrRefreshTokenNotFound
	}

	if err != nil {
		return model.RefreshToken{}, err
	}

	return token, nil
}

func (r *refreshTokenPostgresRepository) Revoke(
	ctx context.Context,
	id int64,
) error {
	_, err := r.pool.Exec(ctx, `
		UPDATE refresh_tokens
		SET revoked_at = NOW()
		WHERE id = $1
		  AND revoked_at IS NULL
	`, id)

	return err
}