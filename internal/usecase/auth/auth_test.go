package auth

import (
	"context"
	"errors"
	"io"
	"testing"

	"github.com/google/uuid"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"github.com/tipo-tawerka/wishlist-api/internal/entity"
)

type userRepositoryMock struct {
	mock.Mock
}

func (m *userRepositoryMock) Save(ctx context.Context, email, passwordHash string) (entity.User, error) {
	args := m.Called(ctx, email, passwordHash)
	user, _ := args.Get(0).(entity.User)
	return user, args.Error(1)
}

func (m *userRepositoryMock) GetByEmail(ctx context.Context, email string) (entity.User, error) {
	args := m.Called(ctx, email)
	user, _ := args.Get(0).(entity.User)
	return user, args.Error(1)
}

type tokenManagerMock struct {
	mock.Mock
}

func (m *tokenManagerMock) Generate(id uuid.UUID) (string, error) {
	args := m.Called(id)
	return args.String(0), args.Error(1)
}

func (m *tokenManagerMock) Parse(tokenString string) (uuid.UUID, error) {
	args := m.Called(tokenString)
	id, _ := args.Get(0).(uuid.UUID)
	return id, args.Error(1)
}

type passwordManagerMock struct {
	mock.Mock
}

func (m *passwordManagerMock) Hash(password string) (string, error) {
	args := m.Called(password)
	return args.String(0), args.Error(1)
}

func (m *passwordManagerMock) Compare(hashedPassword string, password string) error {
	args := m.Called(hashedPassword, password)
	return args.Error(0)
}

func newAuthUseCaseForTest(
	repo *userRepositoryMock,
	token *tokenManagerMock,
	pass *passwordManagerMock,
) *UseCase {
	logger := zerolog.New(io.Discard)
	return NewUseCase(Options{
		TokenManager:   token,
		PasswordHasher: pass,
		UserRepo:       repo,
		Logger:         &logger,
	})
}

func TestUseCaseRegister(t *testing.T) {
	ctx := context.Background()

	t.Run("hash error", func(t *testing.T) {
		repo := new(userRepositoryMock)
		token := new(tokenManagerMock)
		pass := new(passwordManagerMock)
		u := newAuthUseCaseForTest(repo, token, pass)

		errHash := errors.New("hash failed")
		pass.On("Hash", "password").Return("", errHash).Once()

		user, err := u.Register(ctx, "user@example.com", "password")
		require.Error(t, err)
		assert.ErrorIs(t, err, errHash)
		assert.Equal(t, entity.User{}, user)
		repo.AssertNotCalled(t, "Save", mock.Anything, mock.Anything, mock.Anything)
		pass.AssertExpectations(t)
	})

	t.Run("already exists", func(t *testing.T) {
		repo := new(userRepositoryMock)
		token := new(tokenManagerMock)
		pass := new(passwordManagerMock)
		u := newAuthUseCaseForTest(repo, token, pass)

		pass.On("Hash", "password").Return("hashed", nil).Once()
		repo.On("Save", ctx, "user@example.com", "hashed").Return(entity.User{}, entity.ErrUserAlreadyExists).Once()

		user, err := u.Register(ctx, "user@example.com", "password")
		require.Error(t, err)
		assert.ErrorIs(t, err, entity.ErrUserAlreadyExists)
		assert.Equal(t, entity.User{}, user)
		pass.AssertExpectations(t)
		repo.AssertExpectations(t)
	})

	t.Run("repository error", func(t *testing.T) {
		repo := new(userRepositoryMock)
		token := new(tokenManagerMock)
		pass := new(passwordManagerMock)
		u := newAuthUseCaseForTest(repo, token, pass)

		errDB := errors.New("db failed")
		pass.On("Hash", "password").Return("hashed", nil).Once()
		repo.On("Save", ctx, "user@example.com", "hashed").Return(entity.User{}, errDB).Once()

		user, err := u.Register(ctx, "user@example.com", "password")
		require.Error(t, err)
		assert.ErrorIs(t, err, errDB)
		assert.Equal(t, entity.User{}, user)
		pass.AssertExpectations(t)
		repo.AssertExpectations(t)
	})

	t.Run("success", func(t *testing.T) {
		repo := new(userRepositoryMock)
		token := new(tokenManagerMock)
		pass := new(passwordManagerMock)
		u := newAuthUseCaseForTest(repo, token, pass)

		expected := entity.User{ID: uuid.New(), Email: "user@example.com"}
		pass.On("Hash", "password").Return("hashed", nil).Once()
		repo.On("Save", ctx, "user@example.com", "hashed").Return(expected, nil).Once()

		user, err := u.Register(ctx, "user@example.com", "password")
		require.NoError(t, err)
		assert.Equal(t, expected, user)
		pass.AssertExpectations(t)
		repo.AssertExpectations(t)
	})
}

func TestUseCaseLogin(t *testing.T) {
	ctx := context.Background()

	t.Run("invalid credentials from repository", func(t *testing.T) {
		repo := new(userRepositoryMock)
		token := new(tokenManagerMock)
		pass := new(passwordManagerMock)
		u := newAuthUseCaseForTest(repo, token, pass)

		repo.On("GetByEmail", ctx, "user@example.com").Return(entity.User{}, entity.ErrUserInvalidCredentials).Once()

		tok, err := u.Login(ctx, "user@example.com", "password")
		require.Error(t, err)
		assert.ErrorIs(t, err, entity.ErrUserInvalidCredentials)
		assert.Equal(t, "", tok)
		repo.AssertExpectations(t)
		pass.AssertNotCalled(t, "Compare", mock.Anything, mock.Anything)
		token.AssertNotCalled(t, "Generate", mock.Anything)
	})

	t.Run("repository error", func(t *testing.T) {
		repo := new(userRepositoryMock)
		token := new(tokenManagerMock)
		pass := new(passwordManagerMock)
		u := newAuthUseCaseForTest(repo, token, pass)

		errDB := errors.New("db failed")
		repo.On("GetByEmail", ctx, "user@example.com").Return(entity.User{}, errDB).Once()

		tok, err := u.Login(ctx, "user@example.com", "password")
		require.Error(t, err)
		assert.ErrorIs(t, err, errDB)
		assert.Equal(t, "", tok)
		repo.AssertExpectations(t)
		pass.AssertNotCalled(t, "Compare", mock.Anything, mock.Anything)
		token.AssertNotCalled(t, "Generate", mock.Anything)
	})

	t.Run("password mismatch", func(t *testing.T) {
		repo := new(userRepositoryMock)
		token := new(tokenManagerMock)
		pass := new(passwordManagerMock)
		u := newAuthUseCaseForTest(repo, token, pass)

		user := entity.User{ID: uuid.New(), Email: "user@example.com", Password: "hashed"}
		errCompare := errors.New("bad password")
		repo.On("GetByEmail", ctx, "user@example.com").Return(user, nil).Once()
		pass.On("Compare", "hashed", "password").Return(errCompare).Once()

		tok, err := u.Login(ctx, "user@example.com", "password")
		require.Error(t, err)
		assert.ErrorIs(t, err, entity.ErrUserInvalidCredentials)
		assert.Equal(t, "", tok)
		repo.AssertExpectations(t)
		pass.AssertExpectations(t)
		token.AssertNotCalled(t, "Generate", mock.Anything)
	})

	t.Run("token generation error", func(t *testing.T) {
		repo := new(userRepositoryMock)
		token := new(tokenManagerMock)
		pass := new(passwordManagerMock)
		u := newAuthUseCaseForTest(repo, token, pass)

		user := entity.User{ID: uuid.New(), Email: "user@example.com", Password: "hashed"}
		errToken := errors.New("token failed")
		repo.On("GetByEmail", ctx, "user@example.com").Return(user, nil).Once()
		pass.On("Compare", "hashed", "password").Return(nil).Once()
		token.On("Generate", user.ID).Return("", errToken).Once()

		tok, err := u.Login(ctx, "user@example.com", "password")
		require.Error(t, err)
		assert.ErrorIs(t, err, errToken)
		assert.Equal(t, "", tok)
		repo.AssertExpectations(t)
		pass.AssertExpectations(t)
		token.AssertExpectations(t)
	})

	t.Run("success", func(t *testing.T) {
		repo := new(userRepositoryMock)
		token := new(tokenManagerMock)
		pass := new(passwordManagerMock)
		u := newAuthUseCaseForTest(repo, token, pass)

		user := entity.User{ID: uuid.New(), Email: "user@example.com", Password: "hashed"}
		repo.On("GetByEmail", ctx, "user@example.com").Return(user, nil).Once()
		pass.On("Compare", "hashed", "password").Return(nil).Once()
		token.On("Generate", user.ID).Return("jwt-token", nil).Once()

		tok, err := u.Login(ctx, "user@example.com", "password")
		require.NoError(t, err)
		assert.Equal(t, "jwt-token", tok)
		repo.AssertExpectations(t)
		pass.AssertExpectations(t)
		token.AssertExpectations(t)
	})
}
