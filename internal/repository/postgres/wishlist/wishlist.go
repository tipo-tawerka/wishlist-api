package wishlist

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/tipo-tawerka/wishlist-api/internal/entity"
)

type WishlistRepository struct {
	pool *pgxpool.Pool
}

func NewWishlistRepository(pool *pgxpool.Pool) *WishlistRepository {
	return &WishlistRepository{pool: pool}
}

func (r *WishlistRepository) GetByUserId(ctx context.Context, userID uuid.UUID) ([]entity.WishList, error) {
	const query = `
		SELECT id, user_id, title, description, event_date, public_token, created_at, updated_at
		FROM wishlists
		WHERE user_id = $1
		ORDER BY updated_at DESC`

	rows, err := r.pool.Query(ctx, query, userID)
	if err != nil {
		return nil, fmt.Errorf("WishlistRepository.GetByUserId: %w", err)
	}
	defer rows.Close()

	wishlists := make([]entity.WishList, 0)
	for rows.Next() {
		var w entity.WishList
		if w, err = r.scanWishlist(rows); err != nil {
			return nil, fmt.Errorf("WishlistRepository.GetByUserId: %w", err)
		}
		wishlists = append(wishlists, w)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("WishlistRepository.GetByUserId: %w", err)
	}
	return wishlists, nil
}

func (r *WishlistRepository) GetByWishId(ctx context.Context, wishID uuid.UUID) (entity.WishList, error) {
	const query = `
		SELECT id, user_id, title, description, event_date, public_token, created_at, updated_at
		FROM wishlists
		WHERE id = $1`

	w, err := r.scanWishlist(r.pool.QueryRow(ctx, query, wishID))
	switch {
	case errors.Is(err, pgx.ErrNoRows):
		return entity.WishList{}, entity.ErrWishListNotFound
	case err != nil:
		return entity.WishList{}, fmt.Errorf("WishlistRepository.GetByWishId: %w", err)
	}
	return w, nil
}

func (r *WishlistRepository) GetByPublicToken(ctx context.Context, token uuid.UUID) (entity.WishList, error) {
	const query = `
		SELECT id, user_id, title, description, event_date, public_token, created_at, updated_at
		FROM wishlists
		WHERE public_token = $1`

	w, err := r.scanWishlist(r.pool.QueryRow(ctx, query, token))
	switch {
	case errors.Is(err, pgx.ErrNoRows):
		return entity.WishList{}, entity.ErrWishListNotFound
	case err != nil:
		return entity.WishList{}, fmt.Errorf("WishlistRepository.GetByPublicToken: %w", err)
	}
	return w, nil
}

func (r *WishlistRepository) Save(ctx context.Context, wishlist entity.WishList) (entity.WishList, error) {
	const query = `
		INSERT INTO wishlists (user_id, title, description, event_date, public_token)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id, user_id, title, description, event_date, public_token, created_at, updated_at`

	w, err := r.scanWishlist(r.pool.QueryRow(
		ctx, query,
		wishlist.UserID, wishlist.Title, wishlist.Description, wishlist.EventDate, wishlist.PublicToken,
	))
	if err != nil {
		return entity.WishList{}, fmt.Errorf("WishlistRepository.Save: %w", err)
	}
	return w, nil
}

func (r *WishlistRepository) Update(ctx context.Context, wishlist entity.WishList) (entity.WishList, error) {
	const query = `
		UPDATE wishlists
		SET title = $1, description = $2, event_date = $3, public_token = $4, updated_at = now()
		WHERE id = $5 AND user_id = $6
		RETURNING id, user_id, title, description, event_date, public_token, created_at, updated_at`

	w, err := r.scanWishlist(r.pool.QueryRow(
		ctx, query,
		wishlist.Title, wishlist.Description, wishlist.EventDate, wishlist.PublicToken,
		wishlist.ID, wishlist.UserID,
	))
	switch {
	case errors.Is(err, pgx.ErrNoRows):
		return entity.WishList{}, entity.ErrWishListNotFound
	case err != nil:
		return entity.WishList{}, fmt.Errorf("WishlistRepository.Update: %w", err)
	}
	return w, nil
}

// GetWishList returns a wishlist only if it belongs to userID — used by items usecase.
func (r *WishlistRepository) GetWishList(ctx context.Context, userID, wishID uuid.UUID) (entity.WishList, error) {
	const query = `
		SELECT id, user_id, title, description, event_date, public_token, created_at, updated_at
		FROM wishlists
		WHERE id = $1 AND user_id = $2`

	w, err := r.scanWishlist(r.pool.QueryRow(ctx, query, wishID, userID))
	switch {
	case errors.Is(err, pgx.ErrNoRows):
		return entity.WishList{}, entity.ErrWishListNotFound
	case err != nil:
		return entity.WishList{}, fmt.Errorf("WishlistRepository.GetWishList: %w", err)
	}
	return w, nil
}

func (r *WishlistRepository) Delete(ctx context.Context, userID, wishlistID uuid.UUID) error {
	const query = `DELETE FROM wishlists WHERE id = $1 AND user_id = $2`

	tag, err := r.pool.Exec(ctx, query, wishlistID, userID)
	if err != nil {
		return fmt.Errorf("WishlistRepository.Delete: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return entity.ErrWishListNotFound
	}
	return nil
}

func (r *WishlistRepository) scanWishlist(row pgx.Row) (entity.WishList, error) {
	var w entity.WishList
	err := row.Scan(&w.ID, &w.UserID, &w.Title, &w.Description,
		&w.EventDate, &w.PublicToken, &w.CreatedAt, &w.UpdatedAt)
	return w, err
}
