package entity

import (
	"errors"

	"github.com/google/uuid"
)

type User struct {
	ID       uuid.UUID
	Email    string
	Password string
}

var (
	ErrUserAlreadyExists      = errors.New("user with this email already exists")
	ErrUserInvalidCredentials = errors.New("invalid email or password")
)
