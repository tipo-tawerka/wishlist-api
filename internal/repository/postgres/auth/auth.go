package auth

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/tipo-tawerka/wishlist-api/internal/entity"
)

type UserRepository struct {
	pool *pgxpool.Pool
}

func NewUserRepository(pool *pgxpool.Pool) *UserRepository {
	return &UserRepository{pool: pool}
}

func (r *UserRepository) Save(ctx context.Context, email, passwordHash string) (entity.User, error) {
	const query = `
		INSERT INTO users (email, password_hash)
		VALUES ($1, $2)
		RETURNING id, email`

	var user entity.User
	err := r.pool.QueryRow(ctx, query, email, passwordHash).Scan(&user.ID, &user.Email)

	var pgErr *pgconn.PgError
	switch {
	case errors.As(err, &pgErr) && pgErr.Code == pgerrcode.UniqueViolation:
		return entity.User{}, entity.ErrUserAlreadyExists
	case err != nil:
		return entity.User{}, fmt.Errorf("UserRepository.Save: %w", err)
	}
	return user, nil
}

func (r *UserRepository) GetByEmail(ctx context.Context, email string) (entity.User, error) {
	const query = `SELECT id, email, password_hash FROM users WHERE email = $1`

	var user entity.User
	err := r.pool.QueryRow(ctx, query, email).Scan(&user.ID, &user.Email, &user.Password)

	switch {
	case errors.Is(err, pgx.ErrNoRows):
		return entity.User{}, entity.ErrUserInvalidCredentials
	case err != nil:
		return entity.User{}, fmt.Errorf("UserRepository.GetByEmail: %w", err)
	}
	return user, nil
}
