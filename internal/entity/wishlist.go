package entity

import (
	"errors"
	"time"

	"github.com/google/uuid"
)

type WishList struct {
	ID          uuid.UUID
	UserID      uuid.UUID
	Title       string
	Description string
	EventDate   time.Time
	PublicToken *uuid.UUID
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

var (
	ErrWishListNotFound      = errors.New("wishlist not found")
	ErrWishListAlreadyExists = errors.New("wishlist already exists")
)
