package auth

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/rs/zerolog"
	"github.com/tipo-tawerka/wishlist-api/internal/entity"
)

type UserRepository interface {
	Save(ctx context.Context, email, passwordHash string) (entity.User, error)
	GetByEmail(ctx context.Context, email string) (entity.User, error)
}

type TokenManager interface {
	Generate(id uuid.UUID) (string, error)
	Parse(tokenString string) (uuid.UUID, error)
}

type PasswordManager interface {
	Hash(password string) (string, error)
	Compare(hashedPassword string, password string) error
}

type Options struct {
	TokenManager   TokenManager
	PasswordHasher PasswordManager
	UserRepo       UserRepository
	Logger         *zerolog.Logger
}

type UseCase struct {
	token  TokenManager
	passwd PasswordManager
	repo   UserRepository
	logger *zerolog.Logger
}

func NewUseCase(opts Options) *UseCase {
	return &UseCase{
		token:  opts.TokenManager,
		passwd: opts.PasswordHasher,
		repo:   opts.UserRepo,
		logger: opts.Logger,
	}
}

func (u *UseCase) Register(ctx context.Context, email, password string) (entity.User, error) {
	hash, err := u.passwd.Hash(password)
	if err != nil {
		u.logger.Error().Err(err).Str("email", email).Send()
		return entity.User{}, err
	}
	user, err := u.repo.Save(ctx, email, hash)
	switch {
	case errors.Is(err, entity.ErrUserAlreadyExists):
		u.logger.Warn().Err(err).Str("email", email).Send()
		return entity.User{}, entity.ErrUserAlreadyExists
	case err != nil:
		u.logger.Error().Err(err).Str("email", email).Send()
		return entity.User{}, err
	default:
		u.logger.Info().Str("email", email).Msg("user registered successfully")
		return user, nil
	}
}

func (u *UseCase) Login(ctx context.Context, email, password string) (string, error) {
	user, err := u.repo.GetByEmail(ctx, email)
	switch {
	case errors.Is(err, entity.ErrUserInvalidCredentials):
		u.logger.Warn().Err(err).Str("email", email).Send()
		return "", entity.ErrUserInvalidCredentials
	case err != nil:
		u.logger.Error().Err(err).Str("email", email).Send()
		return "", err
	}
	if err := u.passwd.Compare(user.Password, password); err != nil {
		u.logger.Warn().Err(err).Str("email", email).Send()
		return "", entity.ErrUserInvalidCredentials
	}
	token, err := u.token.Generate(user.ID)
	if err != nil {
		u.logger.Error().Err(err).Str("email", email).Send()
		return "", err
	}
	u.logger.Info().Str("email", email).Msg("user logged successfully")
	return token, nil
}
