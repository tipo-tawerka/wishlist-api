package wishlist

import (
	"context"
	"errors"
	"io"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"github.com/tipo-tawerka/wishlist-api/internal/entity"
)

type wishlistRepositoryMock struct {
	mock.Mock
}

func (m *wishlistRepositoryMock) GetByUserId(ctx context.Context, userId uuid.UUID) ([]entity.WishList, error) {
	args := m.Called(ctx, userId)
	w, _ := args.Get(0).([]entity.WishList)
	return w, args.Error(1)
}

func (m *wishlistRepositoryMock) GetByWishId(ctx context.Context, wishId uuid.UUID) (entity.WishList, error) {
	args := m.Called(ctx, wishId)
	w, _ := args.Get(0).(entity.WishList)
	return w, args.Error(1)
}

func (m *wishlistRepositoryMock) Save(ctx context.Context, wishlist entity.WishList) (entity.WishList, error) {
	args := m.Called(ctx, wishlist)
	w, _ := args.Get(0).(entity.WishList)
	return w, args.Error(1)
}

func (m *wishlistRepositoryMock) Update(ctx context.Context, wishlist entity.WishList) (entity.WishList, error) {
	args := m.Called(ctx, wishlist)
	w, _ := args.Get(0).(entity.WishList)
	return w, args.Error(1)
}

func (m *wishlistRepositoryMock) Delete(ctx context.Context, userId, wishlistId uuid.UUID) error {
	args := m.Called(ctx, userId, wishlistId)
	return args.Error(0)
}

func newWishlistUseCaseForTest(repo *wishlistRepositoryMock) *UseCase {
	logger := zerolog.New(io.Discard)
	return NewUseCase(repo, &logger)
}

func TestUseCaseGetAllWishList(t *testing.T) {
	ctx := context.Background()
	userID := uuid.New()

	t.Run("repository error", func(t *testing.T) {
		repo := new(wishlistRepositoryMock)
		u := newWishlistUseCaseForTest(repo)

		errDB := errors.New("db failed")
		repo.On("GetByUserId", ctx, userID).Return([]entity.WishList(nil), errDB).Once()

		res, err := u.GetAllWishList(ctx, userID)
		require.Error(t, err)
		assert.ErrorIs(t, err, errDB)
		assert.Nil(t, res)
		repo.AssertExpectations(t)
	})

	t.Run("success", func(t *testing.T) {
		repo := new(wishlistRepositoryMock)
		u := newWishlistUseCaseForTest(repo)

		expected := []entity.WishList{{ID: uuid.New(), UserID: userID, Title: "Birthday", EventDate: time.Now()}}
		repo.On("GetByUserId", ctx, userID).Return(expected, nil).Once()

		res, err := u.GetAllWishList(ctx, userID)
		require.NoError(t, err)
		assert.Equal(t, expected, res)
		repo.AssertExpectations(t)
	})
}

func TestUseCaseCreateWishList(t *testing.T) {
	ctx := context.Background()
	userID := uuid.New()
	input := entity.WishList{UserID: userID, Title: "Birthday", EventDate: time.Now()}

	t.Run("repository error", func(t *testing.T) {
		repo := new(wishlistRepositoryMock)
		u := newWishlistUseCaseForTest(repo)

		errDB := errors.New("db failed")
		repo.On("Save", ctx, input).Return(entity.WishList{}, errDB).Once()

		res, err := u.CreateWishList(ctx, input)
		require.Error(t, err)
		assert.ErrorIs(t, err, errDB)
		assert.Equal(t, entity.WishList{}, res)
		repo.AssertExpectations(t)
	})

	t.Run("success", func(t *testing.T) {
		repo := new(wishlistRepositoryMock)
		u := newWishlistUseCaseForTest(repo)

		expected := input
		expected.ID = uuid.New()
		repo.On("Save", ctx, input).Return(expected, nil).Once()

		res, err := u.CreateWishList(ctx, input)
		require.NoError(t, err)
		assert.Equal(t, expected, res)
		repo.AssertExpectations(t)
	})
}

func TestUseCaseGetWishList(t *testing.T) {
	ctx := context.Background()
	userID := uuid.New()
	wishID := uuid.New()

	t.Run("not found", func(t *testing.T) {
		repo := new(wishlistRepositoryMock)
		u := newWishlistUseCaseForTest(repo)

		repo.On("GetByWishId", ctx, wishID).Return(entity.WishList{}, entity.ErrWishListNotFound).Once()

		res, err := u.GetWishList(ctx, userID, wishID)
		require.Error(t, err)
		assert.ErrorIs(t, err, entity.ErrWishListNotFound)
		assert.Equal(t, entity.WishList{}, res)
		repo.AssertExpectations(t)
	})

	t.Run("repository error", func(t *testing.T) {
		repo := new(wishlistRepositoryMock)
		u := newWishlistUseCaseForTest(repo)

		errDB := errors.New("db failed")
		repo.On("GetByWishId", ctx, wishID).Return(entity.WishList{}, errDB).Once()

		res, err := u.GetWishList(ctx, userID, wishID)
		require.Error(t, err)
		assert.ErrorIs(t, err, errDB)
		assert.Equal(t, entity.WishList{}, res)
		repo.AssertExpectations(t)
	})

	t.Run("wishlist belongs to another user", func(t *testing.T) {
		repo := new(wishlistRepositoryMock)
		u := newWishlistUseCaseForTest(repo)

		w := entity.WishList{ID: wishID, UserID: uuid.New(), Title: "Birthday"}
		repo.On("GetByWishId", ctx, wishID).Return(w, nil).Once()

		res, err := u.GetWishList(ctx, userID, wishID)
		require.Error(t, err)
		assert.ErrorIs(t, err, entity.ErrWishListNotFound)
		assert.Equal(t, entity.WishList{}, res)
		repo.AssertExpectations(t)
	})

	t.Run("success", func(t *testing.T) {
		repo := new(wishlistRepositoryMock)
		u := newWishlistUseCaseForTest(repo)

		expected := entity.WishList{ID: wishID, UserID: userID, Title: "Birthday"}
		repo.On("GetByWishId", ctx, wishID).Return(expected, nil).Once()

		res, err := u.GetWishList(ctx, userID, wishID)
		require.NoError(t, err)
		assert.Equal(t, expected, res)
		repo.AssertExpectations(t)
	})
}

func TestUseCaseUpdateWishList(t *testing.T) {
	ctx := context.Background()
	input := entity.WishList{ID: uuid.New(), UserID: uuid.New(), Title: "Birthday"}

	t.Run("not found", func(t *testing.T) {
		repo := new(wishlistRepositoryMock)
		u := newWishlistUseCaseForTest(repo)

		repo.On("Update", ctx, input).Return(entity.WishList{}, entity.ErrWishListNotFound).Once()

		res, err := u.UpdateWishList(ctx, input)
		require.Error(t, err)
		assert.ErrorIs(t, err, entity.ErrWishListNotFound)
		assert.Equal(t, entity.WishList{}, res)
		repo.AssertExpectations(t)
	})

	t.Run("repository error", func(t *testing.T) {
		repo := new(wishlistRepositoryMock)
		u := newWishlistUseCaseForTest(repo)

		errDB := errors.New("db failed")
		repo.On("Update", ctx, input).Return(entity.WishList{}, errDB).Once()

		res, err := u.UpdateWishList(ctx, input)
		require.Error(t, err)
		assert.ErrorIs(t, err, errDB)
		assert.Equal(t, entity.WishList{}, res)
		repo.AssertExpectations(t)
	})

	t.Run("success", func(t *testing.T) {
		repo := new(wishlistRepositoryMock)
		u := newWishlistUseCaseForTest(repo)

		expected := input
		expected.Title = "Updated"
		repo.On("Update", ctx, input).Return(expected, nil).Once()

		res, err := u.UpdateWishList(ctx, input)
		require.NoError(t, err)
		assert.Equal(t, expected, res)
		repo.AssertExpectations(t)
	})
}

func TestUseCaseDeleteWishList(t *testing.T) {
	ctx := context.Background()
	userID := uuid.New()
	wishID := uuid.New()

	t.Run("not found is idempotent", func(t *testing.T) {
		repo := new(wishlistRepositoryMock)
		u := newWishlistUseCaseForTest(repo)

		repo.On("Delete", ctx, userID, wishID).Return(entity.ErrWishListNotFound).Once()

		err := u.DeleteWishList(ctx, userID, wishID)
		require.NoError(t, err)
		repo.AssertExpectations(t)
	})

	t.Run("repository error", func(t *testing.T) {
		repo := new(wishlistRepositoryMock)
		u := newWishlistUseCaseForTest(repo)

		errDB := errors.New("db failed")
		repo.On("Delete", ctx, userID, wishID).Return(errDB).Once()

		err := u.DeleteWishList(ctx, userID, wishID)
		require.Error(t, err)
		assert.ErrorIs(t, err, errDB)
		repo.AssertExpectations(t)
	})

	t.Run("success", func(t *testing.T) {
		repo := new(wishlistRepositoryMock)
		u := newWishlistUseCaseForTest(repo)

		repo.On("Delete", ctx, userID, wishID).Return(nil).Once()

		err := u.DeleteWishList(ctx, userID, wishID)
		require.NoError(t, err)
		repo.AssertExpectations(t)
	})
}
