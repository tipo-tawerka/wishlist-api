package entity

import (
	"errors"
	"time"

	"github.com/google/uuid"
)

type Items struct {
	ID          uuid.UUID
	WishListID  uuid.UUID
	Title       string
	Description string
	ProductURL  string
	Priority    int
	Reserved    bool
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

var (
	ErrItemNotFound      = errors.New("item not found")
	ErrItemAlreadyBooked = errors.New("item is already booked")
)
