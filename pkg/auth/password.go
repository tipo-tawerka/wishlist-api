package auth

import (
	"crypto/sha256"
	"errors"

	"golang.org/x/crypto/bcrypt"
)

type PasswordHasher struct {
	pepper string
}

func NewPasswordHasher(pepper string) (*PasswordHasher, error) {
	if pepper == "" {
		return nil, errors.New("pepper must not be empty")
	}
	return &PasswordHasher{pepper: pepper}, nil
}

func (ph *PasswordHasher) Hash(password string) (string, error) {
	h := sha256.Sum256([]byte(ph.pepper + password))
	hashed, err := bcrypt.GenerateFromPassword(h[:], bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(hashed), nil
}

func (ph *PasswordHasher) Compare(hashedPassword string, password string) error {
	h := sha256.Sum256([]byte(ph.pepper + password))
	return bcrypt.CompareHashAndPassword([]byte(hashedPassword), h[:])
}
