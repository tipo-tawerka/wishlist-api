package public

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

type publicWishListRepositoryMock struct {
	mock.Mock
}

func (m *publicWishListRepositoryMock) GetByPublicToken(ctx context.Context, token uuid.UUID) (entity.WishList, error) {
	args := m.Called(ctx, token)
	w, _ := args.Get(0).(entity.WishList)
	return w, args.Error(1)
}

type publicItemsRepositoryMock struct {
	mock.Mock
}

func (m *publicItemsRepositoryMock) ListByWishlist(ctx context.Context, wishListId uuid.UUID) ([]entity.Items, error) {
	args := m.Called(ctx, wishListId)
	items, _ := args.Get(0).([]entity.Items)
	return items, args.Error(1)
}

func (m *publicItemsRepositoryMock) Reserve(ctx context.Context, wishlistId, itemId uuid.UUID) (entity.Items, error) {
	args := m.Called(ctx, wishlistId, itemId)
	item, _ := args.Get(0).(entity.Items)
	return item, args.Error(1)
}

func newPublicUseCaseForTest(wishlistRepo *publicWishListRepositoryMock, itemRepo *publicItemsRepositoryMock) *UseCase {
	logger := zerolog.New(io.Discard)
	return NewUseCase(wishlistRepo, itemRepo, &logger)
}

func TestUseCaseGetWishList(t *testing.T) {
	ctx := context.Background()
	token := uuid.New()
	wishlistID := uuid.New()

	t.Run("wishlist not found", func(t *testing.T) {
		wishRepo := new(publicWishListRepositoryMock)
		itemRepo := new(publicItemsRepositoryMock)
		u := newPublicUseCaseForTest(wishRepo, itemRepo)

		wishRepo.On("GetByPublicToken", ctx, token).Return(entity.WishList{}, entity.ErrWishListNotFound).Once()

		w, items, err := u.GetWishList(ctx, token)
		require.Error(t, err)
		assert.ErrorIs(t, err, entity.ErrWishListNotFound)
		assert.Equal(t, entity.WishList{}, w)
		assert.Nil(t, items)
		wishRepo.AssertExpectations(t)
		itemRepo.AssertNotCalled(t, "ListByWishlist", mock.Anything, mock.Anything)
	})

	t.Run("wishlist repository error", func(t *testing.T) {
		wishRepo := new(publicWishListRepositoryMock)
		itemRepo := new(publicItemsRepositoryMock)
		u := newPublicUseCaseForTest(wishRepo, itemRepo)

		errDB := errors.New("db failed")
		wishRepo.On("GetByPublicToken", ctx, token).Return(entity.WishList{}, errDB).Once()

		w, items, err := u.GetWishList(ctx, token)
		require.Error(t, err)
		assert.ErrorIs(t, err, errDB)
		assert.Equal(t, entity.WishList{}, w)
		assert.Nil(t, items)
		wishRepo.AssertExpectations(t)
		itemRepo.AssertNotCalled(t, "ListByWishlist", mock.Anything, mock.Anything)
	})

	t.Run("items repository error", func(t *testing.T) {
		wishRepo := new(publicWishListRepositoryMock)
		itemRepo := new(publicItemsRepositoryMock)
		u := newPublicUseCaseForTest(wishRepo, itemRepo)

		errDB := errors.New("db failed")
		wishlist := entity.WishList{ID: wishlistID}
		wishRepo.On("GetByPublicToken", ctx, token).Return(wishlist, nil).Once()
		itemRepo.On("ListByWishlist", ctx, wishlistID).Return([]entity.Items(nil), errDB).Once()

		w, items, err := u.GetWishList(ctx, token)
		require.Error(t, err)
		assert.ErrorIs(t, err, errDB)
		assert.Equal(t, entity.WishList{}, w)
		assert.Nil(t, items)
		wishRepo.AssertExpectations(t)
		itemRepo.AssertExpectations(t)
	})

	t.Run("success", func(t *testing.T) {
		wishRepo := new(publicWishListRepositoryMock)
		itemRepo := new(publicItemsRepositoryMock)
		u := newPublicUseCaseForTest(wishRepo, itemRepo)

		wishlist := entity.WishList{ID: wishlistID, Title: "Birthday"}
		expectedItems := []entity.Items{{ID: uuid.New(), WishListID: wishlistID, Title: "Book"}}
		wishRepo.On("GetByPublicToken", ctx, token).Return(wishlist, nil).Once()
		itemRepo.On("ListByWishlist", ctx, wishlistID).Return(expectedItems, nil).Once()

		w, items, err := u.GetWishList(ctx, token)
		require.NoError(t, err)
		assert.Equal(t, wishlist, w)
		assert.Equal(t, expectedItems, items)
		wishRepo.AssertExpectations(t)
		itemRepo.AssertExpectations(t)
	})
}

func TestUseCaseReserveItem(t *testing.T) {
	ctx := context.Background()
	token := uuid.New()
	wishlistID := uuid.New()
	itemID := uuid.New()

	t.Run("wishlist not found", func(t *testing.T) {
		wishRepo := new(publicWishListRepositoryMock)
		itemRepo := new(publicItemsRepositoryMock)
		u := newPublicUseCaseForTest(wishRepo, itemRepo)

		wishRepo.On("GetByPublicToken", ctx, token).Return(entity.WishList{}, entity.ErrWishListNotFound).Once()

		item, err := u.ReserveItem(ctx, token, itemID)
		require.Error(t, err)
		assert.ErrorIs(t, err, entity.ErrWishListNotFound)
		assert.Equal(t, entity.Items{}, item)
		wishRepo.AssertExpectations(t)
		itemRepo.AssertNotCalled(t, "Reserve", mock.Anything, mock.Anything, mock.Anything)
	})

	t.Run("wishlist repository error", func(t *testing.T) {
		wishRepo := new(publicWishListRepositoryMock)
		itemRepo := new(publicItemsRepositoryMock)
		u := newPublicUseCaseForTest(wishRepo, itemRepo)

		errDB := errors.New("db failed")
		wishRepo.On("GetByPublicToken", ctx, token).Return(entity.WishList{}, errDB).Once()

		item, err := u.ReserveItem(ctx, token, itemID)
		require.Error(t, err)
		assert.ErrorIs(t, err, errDB)
		assert.Equal(t, entity.Items{}, item)
		wishRepo.AssertExpectations(t)
		itemRepo.AssertNotCalled(t, "Reserve", mock.Anything, mock.Anything, mock.Anything)
	})

	t.Run("item not found", func(t *testing.T) {
		wishRepo := new(publicWishListRepositoryMock)
		itemRepo := new(publicItemsRepositoryMock)
		u := newPublicUseCaseForTest(wishRepo, itemRepo)

		wishRepo.On("GetByPublicToken", ctx, token).Return(entity.WishList{ID: wishlistID}, nil).Once()
		itemRepo.On("Reserve", ctx, wishlistID, itemID).Return(entity.Items{}, entity.ErrItemNotFound).Once()

		item, err := u.ReserveItem(ctx, token, itemID)
		require.Error(t, err)
		assert.ErrorIs(t, err, entity.ErrItemNotFound)
		assert.Equal(t, entity.Items{}, item)
		wishRepo.AssertExpectations(t)
		itemRepo.AssertExpectations(t)
	})

	t.Run("item already booked", func(t *testing.T) {
		wishRepo := new(publicWishListRepositoryMock)
		itemRepo := new(publicItemsRepositoryMock)
		u := newPublicUseCaseForTest(wishRepo, itemRepo)

		wishRepo.On("GetByPublicToken", ctx, token).Return(entity.WishList{ID: wishlistID}, nil).Once()
		itemRepo.On("Reserve", ctx, wishlistID, itemID).Return(entity.Items{}, entity.ErrItemAlreadyBooked).Once()

		item, err := u.ReserveItem(ctx, token, itemID)
		require.Error(t, err)
		assert.ErrorIs(t, err, entity.ErrItemAlreadyBooked)
		assert.Equal(t, entity.Items{}, item)
		wishRepo.AssertExpectations(t)
		itemRepo.AssertExpectations(t)
	})

	t.Run("items repository error", func(t *testing.T) {
		wishRepo := new(publicWishListRepositoryMock)
		itemRepo := new(publicItemsRepositoryMock)
		u := newPublicUseCaseForTest(wishRepo, itemRepo)

		errDB := errors.New("db failed")
		wishRepo.On("GetByPublicToken", ctx, token).Return(entity.WishList{ID: wishlistID}, nil).Once()
		itemRepo.On("Reserve", ctx, wishlistID, itemID).Return(entity.Items{}, errDB).Once()

		item, err := u.ReserveItem(ctx, token, itemID)
		require.Error(t, err)
		assert.ErrorIs(t, err, errDB)
		assert.Equal(t, entity.Items{}, item)
		wishRepo.AssertExpectations(t)
		itemRepo.AssertExpectations(t)
	})

	t.Run("success", func(t *testing.T) {
		wishRepo := new(publicWishListRepositoryMock)
		itemRepo := new(publicItemsRepositoryMock)
		u := newPublicUseCaseForTest(wishRepo, itemRepo)

		expected := entity.Items{ID: itemID, WishListID: wishlistID, Reserved: true}
		wishRepo.On("GetByPublicToken", ctx, token).Return(entity.WishList{ID: wishlistID}, nil).Once()
		itemRepo.On("Reserve", ctx, wishlistID, itemID).Return(expected, nil).Once()

		item, err := u.ReserveItem(ctx, token, itemID)
		require.NoError(t, err)
		assert.Equal(t, expected, item)
		wishRepo.AssertExpectations(t)
		itemRepo.AssertExpectations(t)
	})
}
