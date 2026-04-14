package auth

import (
	"errors"
	"time"

	jwt "github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

type TokenManager struct {
	secret []byte
}

func NewTokenManager(secret string) (*TokenManager, error) {
	if secret == "" {
		return nil, errors.New("jwt secret must not be empty")
	}
	return &TokenManager{secret: []byte(secret)}, nil
}

const _ttl = time.Hour * 24 * 7 // one week

const (
	_claimUserID = "user_id"
	_claimExp    = "exp"
	_claimIat    = "iat"
)

func (tm *TokenManager) Generate(id uuid.UUID) (string, error) {
	claims := jwt.MapClaims{
		_claimUserID: id.String(),
		_claimIat:    time.Now().Unix(),
		_claimExp:    time.Now().Add(_ttl).Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(tm.secret)
}

func (tm *TokenManager) Parse(tokenString string) (uuid.UUID, error) {
	token, err := jwt.Parse(tokenString, func(t *jwt.Token) (any, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("unexpected signing method")
		}
		return tm.secret, nil
	})
	if err != nil {
		return uuid.UUID{}, err
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok || !token.Valid {
		return uuid.UUID{}, errors.New("invalid token")
	}

	idStr, ok := claims[_claimUserID].(string)
	if !ok {
		return uuid.UUID{}, errors.New("invalid user id")
	}
	id, err := uuid.Parse(idStr)
	if err != nil {
		return uuid.UUID{}, errors.New("id user is not a valid uuid")
	}
	return id, nil
}
