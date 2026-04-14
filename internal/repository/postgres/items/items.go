package items

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/tipo-tawerka/wishlist-api/internal/entity"
)

type ItemsRepository struct {
	pool *pgxpool.Pool
}

func NewItemsRepository(pool *pgxpool.Pool) *ItemsRepository {
	return &ItemsRepository{pool: pool}
}

func (r *ItemsRepository) GetAll(ctx context.Context, userID, wishListID uuid.UUID) ([]entity.Items, error) {
	const query = `
		SELECT i.id, i.wishlist_id, i.title, i.description, i.product_url, i.priority, i.reserved, i.created_at, i.updated_at
		FROM items i
		JOIN wishlists w ON w.id = i.wishlist_id
		WHERE i.wishlist_id = $1 AND w.user_id = $2
		ORDER BY i.priority DESC`

	rows, err := r.pool.Query(ctx, query, wishListID, userID)
	if err != nil {
		return nil, fmt.Errorf("ItemsRepository.GetAll: %w", err)
	}
	defer rows.Close()

	items := make([]entity.Items, 0, 8)
	for rows.Next() {
		var i entity.Items
		if i, err = r.scanItem(rows); err != nil {
			return nil, fmt.Errorf("ItemsRepository.GetAll: %w", err)
		}
		items = append(items, i)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("ItemsRepository.GetAll: %w", err)
	}
	return items, nil
}

func (r *ItemsRepository) ListByWishlist(ctx context.Context, wishListID uuid.UUID) ([]entity.Items, error) {
	const query = `
		SELECT id, wishlist_id, title, description, product_url, priority, reserved, created_at, updated_at
		FROM items
		WHERE wishlist_id = $1
		ORDER BY priority DESC`

	rows, err := r.pool.Query(ctx, query, wishListID)
	if err != nil {
		return nil, fmt.Errorf("ItemsRepository.ListByWishlist: %w", err)
	}
	defer rows.Close()

	items := make([]entity.Items, 0)
	for rows.Next() {
		var i entity.Items
		if err := rows.Scan(&i.ID, &i.WishListID, &i.Title, &i.Description, &i.ProductURL, &i.Priority, &i.Reserved, &i.CreatedAt, &i.UpdatedAt); err != nil {
			return nil, fmt.Errorf("ItemsRepository.ListByWishlist: %w", err)
		}
		items = append(items, i)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("ItemsRepository.ListByWishlist: %w", err)
	}
	return items, nil
}

func (r *ItemsRepository) GetByID(ctx context.Context, userID, itemID uuid.UUID) (entity.Items, error) {
	const query = `
		SELECT i.id, i.wishlist_id, i.title, i.description, i.product_url, i.priority, i.reserved, i.created_at, i.updated_at
		FROM items i
		JOIN wishlists w ON w.id = i.wishlist_id
		WHERE i.id = $1 AND w.user_id = $2`

	item, err := r.scanItem(r.pool.QueryRow(ctx, query, itemID, userID))
	switch {
	case errors.Is(err, pgx.ErrNoRows):
		return entity.Items{}, entity.ErrItemNotFound
	case err != nil:
		return entity.Items{}, fmt.Errorf("ItemsRepository.GetByID: %w", err)
	}
	return item, nil
}

func (r *ItemsRepository) Create(ctx context.Context, userID uuid.UUID, item entity.Items) (entity.Items, error) {
	const query = `
		INSERT INTO items (wishlist_id, title, description, product_url, priority)
		SELECT $1, $2, $3, $4, $5
		FROM wishlists
		WHERE id = $1 AND user_id = $6
		RETURNING id, wishlist_id, title, description, product_url, priority, reserved, created_at, updated_at`

	created, err := r.scanItem(r.pool.QueryRow(ctx, query,
		item.WishListID, item.Title, item.Description, item.ProductURL, item.Priority, userID,
	))
	switch {
	case errors.Is(err, pgx.ErrNoRows):
		return entity.Items{}, entity.ErrWishListNotFound
	case err != nil:
		return entity.Items{}, fmt.Errorf("ItemsRepository.Create: %w", err)
	}
	return created, nil
}

func (r *ItemsRepository) Update(ctx context.Context, userID uuid.UUID, item entity.Items) (entity.Items, error) {
	const query = `
		UPDATE items i
		SET title = $1, description = $2, product_url = $3, priority = $4, updated_at = now()
		FROM wishlists w
		WHERE i.id = $5 AND i.wishlist_id = w.id AND w.user_id = $6
		RETURNING i.id, i.wishlist_id, i.title, i.description, i.product_url, i.priority, i.reserved, i.created_at, i.updated_at`

	updated, err := r.scanItem(r.pool.QueryRow(ctx, query,
		item.Title, item.Description, item.ProductURL, item.Priority,
		item.ID, userID,
	))
	switch {
	case errors.Is(err, pgx.ErrNoRows):
		return entity.Items{}, entity.ErrItemNotFound
	case err != nil:
		return entity.Items{}, fmt.Errorf("ItemsRepository.Update: %w", err)
	}
	return updated, nil
}

func (r *ItemsRepository) Delete(ctx context.Context, userID, wishlistID, itemID uuid.UUID) error {
	const query = `
		DELETE FROM items i
		USING wishlists w
		WHERE i.id = $1 AND i.wishlist_id = $2 AND i.wishlist_id = w.id AND w.user_id = $3`

	tag, err := r.pool.Exec(ctx, query, itemID, wishlistID, userID)
	if err != nil {
		return fmt.Errorf("ItemsRepository.Delete: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return entity.ErrItemNotFound
	}
	return nil
}

func (r *ItemsRepository) Reserve(ctx context.Context, wishlistID, itemID uuid.UUID) (entity.Items, error) {
	const query = `
		UPDATE items
		SET reserved = true
		WHERE id = $1 AND wishlist_id = $2 AND reserved = false
		RETURNING id, wishlist_id, title, description, product_url, priority, reserved, created_at, updated_at`

	item, err := r.scanItem(r.pool.QueryRow(ctx, query, itemID, wishlistID))
	if err == nil {
		return item, nil
	}
	if !errors.Is(err, pgx.ErrNoRows) {
		return entity.Items{}, fmt.Errorf("ItemsRepository.Reserve: %w", err)
	}

	const checkQuery = `SELECT id FROM items WHERE id = $1 AND wishlist_id = $2`
	var id uuid.UUID
	checkErr := r.pool.QueryRow(ctx, checkQuery, itemID, wishlistID).Scan(&id)
	switch {
	case errors.Is(checkErr, pgx.ErrNoRows):
		return entity.Items{}, entity.ErrItemNotFound
	case checkErr != nil:
		return entity.Items{}, fmt.Errorf("ItemsRepository.Reserve: %w", checkErr)
	default:
		return entity.Items{}, entity.ErrItemAlreadyBooked
	}
}

func (r *ItemsRepository) scanItem(row pgx.Row) (entity.Items, error) {
	var i entity.Items
	err := row.Scan(&i.ID, &i.WishListID, &i.Title, &i.Description, &i.ProductURL,
		&i.Priority, &i.Reserved, &i.CreatedAt, &i.UpdatedAt)
	return i, err
}
