package items

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

type itemsRepositoryMock struct {
	mock.Mock
}

func (m *itemsRepositoryMock) GetAll(ctx context.Context, userID, wishListID uuid.UUID) ([]entity.Items, error) {
	args := m.Called(ctx, userID, wishListID)
	items, _ := args.Get(0).([]entity.Items)
	return items, args.Error(1)
}

func (m *itemsRepositoryMock) GetByID(ctx context.Context, userID, itemID uuid.UUID) (entity.Items, error) {
	args := m.Called(ctx, userID, itemID)
	item, _ := args.Get(0).(entity.Items)
	return item, args.Error(1)
}

func (m *itemsRepositoryMock) Create(ctx context.Context, userID uuid.UUID, item entity.Items) (entity.Items, error) {
	args := m.Called(ctx, userID, item)
	created, _ := args.Get(0).(entity.Items)
	return created, args.Error(1)
}

func (m *itemsRepositoryMock) Update(ctx context.Context, userID uuid.UUID, item entity.Items) (entity.Items, error) {
	args := m.Called(ctx, userID, item)
	updated, _ := args.Get(0).(entity.Items)
	return updated, args.Error(1)
}

func (m *itemsRepositoryMock) Delete(ctx context.Context, userID, wishlistID, itemID uuid.UUID) error {
	args := m.Called(ctx, userID, wishlistID, itemID)
	return args.Error(0)
}

type wishListRepositoryMock struct {
	mock.Mock
}

func (m *wishListRepositoryMock) GetWishList(ctx context.Context, userID, wishID uuid.UUID) (entity.WishList, error) {
	args := m.Called(ctx, userID, wishID)
	w, _ := args.Get(0).(entity.WishList)
	return w, args.Error(1)
}

func newItemsUseCaseForTest(itemRepo *itemsRepositoryMock, wishRepo *wishListRepositoryMock) *UseCase {
	logger := zerolog.New(io.Discard)
	return NewUseCase(itemRepo, wishRepo, &logger)
}

func TestUseCaseGetItems(t *testing.T) {
	ctx := context.Background()
	userID := uuid.New()
	wishlistID := uuid.New()

	t.Run("wishlist not found", func(t *testing.T) {
		itemRepo := new(itemsRepositoryMock)
		wishRepo := new(wishListRepositoryMock)
		u := newItemsUseCaseForTest(itemRepo, wishRepo)

		wishRepo.On("GetWishList", ctx, userID, wishlistID).Return(entity.WishList{}, entity.ErrWishListNotFound).Once()

		res, err := u.GetItems(ctx, userID, wishlistID)
		require.Error(t, err)
		assert.ErrorIs(t, err, entity.ErrWishListNotFound)
		assert.Nil(t, res)
		wishRepo.AssertExpectations(t)
		itemRepo.AssertNotCalled(t, "GetAll", mock.Anything, mock.Anything, mock.Anything)
	})

	t.Run("wishlist repository error", func(t *testing.T) {
		itemRepo := new(itemsRepositoryMock)
		wishRepo := new(wishListRepositoryMock)
		u := newItemsUseCaseForTest(itemRepo, wishRepo)

		errDB := errors.New("db failed")
		wishRepo.On("GetWishList", ctx, userID, wishlistID).Return(entity.WishList{}, errDB).Once()

		res, err := u.GetItems(ctx, userID, wishlistID)
		require.Error(t, err)
		assert.ErrorIs(t, err, errDB)
		assert.Nil(t, res)
		wishRepo.AssertExpectations(t)
		itemRepo.AssertNotCalled(t, "GetAll", mock.Anything, mock.Anything, mock.Anything)
	})

	t.Run("items repository error", func(t *testing.T) {
		itemRepo := new(itemsRepositoryMock)
		wishRepo := new(wishListRepositoryMock)
		u := newItemsUseCaseForTest(itemRepo, wishRepo)

		errDB := errors.New("db failed")
		wishRepo.On("GetWishList", ctx, userID, wishlistID).Return(entity.WishList{ID: wishlistID, UserID: userID}, nil).Once()
		itemRepo.On("GetAll", ctx, userID, wishlistID).Return([]entity.Items(nil), errDB).Once()

		res, err := u.GetItems(ctx, userID, wishlistID)
		require.Error(t, err)
		assert.ErrorIs(t, err, errDB)
		assert.Nil(t, res)
		wishRepo.AssertExpectations(t)
		itemRepo.AssertExpectations(t)
	})

	t.Run("success", func(t *testing.T) {
		itemRepo := new(itemsRepositoryMock)
		wishRepo := new(wishListRepositoryMock)
		u := newItemsUseCaseForTest(itemRepo, wishRepo)

		expected := []entity.Items{{ID: uuid.New(), WishListID: wishlistID, Title: "Book", Priority: 5}}
		wishRepo.On("GetWishList", ctx, userID, wishlistID).Return(entity.WishList{ID: wishlistID, UserID: userID}, nil).Once()
		itemRepo.On("GetAll", ctx, userID, wishlistID).Return(expected, nil).Once()

		res, err := u.GetItems(ctx, userID, wishlistID)
		require.NoError(t, err)
		assert.Equal(t, expected, res)
		wishRepo.AssertExpectations(t)
		itemRepo.AssertExpectations(t)
	})
}

func TestUseCaseCreateItem(t *testing.T) {
	ctx := context.Background()
	userID := uuid.New()
	input := entity.Items{WishListID: uuid.New(), Title: "Book", Priority: 5}

	t.Run("wishlist not found", func(t *testing.T) {
		itemRepo := new(itemsRepositoryMock)
		wishRepo := new(wishListRepositoryMock)
		u := newItemsUseCaseForTest(itemRepo, wishRepo)

		itemRepo.On("Create", ctx, userID, input).Return(entity.Items{}, entity.ErrWishListNotFound).Once()

		res, err := u.CreateItem(ctx, userID, input)
		require.Error(t, err)
		assert.ErrorIs(t, err, entity.ErrWishListNotFound)
		assert.Equal(t, entity.Items{}, res)
		itemRepo.AssertExpectations(t)
	})

	t.Run("repository error", func(t *testing.T) {
		itemRepo := new(itemsRepositoryMock)
		wishRepo := new(wishListRepositoryMock)
		u := newItemsUseCaseForTest(itemRepo, wishRepo)

		errDB := errors.New("db failed")
		itemRepo.On("Create", ctx, userID, input).Return(entity.Items{}, errDB).Once()

		res, err := u.CreateItem(ctx, userID, input)
		require.Error(t, err)
		assert.ErrorIs(t, err, errDB)
		assert.Equal(t, entity.Items{}, res)
		itemRepo.AssertExpectations(t)
	})

	t.Run("success", func(t *testing.T) {
		itemRepo := new(itemsRepositoryMock)
		wishRepo := new(wishListRepositoryMock)
		u := newItemsUseCaseForTest(itemRepo, wishRepo)

		expected := input
		expected.ID = uuid.New()
		itemRepo.On("Create", ctx, userID, input).Return(expected, nil).Once()

		res, err := u.CreateItem(ctx, userID, input)
		require.NoError(t, err)
		assert.Equal(t, expected, res)
		itemRepo.AssertExpectations(t)
	})
}

func TestUseCaseGetItem(t *testing.T) {
	ctx := context.Background()
	userID := uuid.New()
	wishlistID := uuid.New()
	itemID := uuid.New()

	t.Run("item not found", func(t *testing.T) {
		itemRepo := new(itemsRepositoryMock)
		wishRepo := new(wishListRepositoryMock)
		u := newItemsUseCaseForTest(itemRepo, wishRepo)

		itemRepo.On("GetByID", ctx, userID, itemID).Return(entity.Items{}, entity.ErrItemNotFound).Once()

		res, err := u.GetItem(ctx, userID, wishlistID, itemID)
		require.Error(t, err)
		assert.ErrorIs(t, err, entity.ErrItemNotFound)
		assert.Equal(t, entity.Items{}, res)
		itemRepo.AssertExpectations(t)
	})

	t.Run("repository error", func(t *testing.T) {
		itemRepo := new(itemsRepositoryMock)
		wishRepo := new(wishListRepositoryMock)
		u := newItemsUseCaseForTest(itemRepo, wishRepo)

		errDB := errors.New("db failed")
		itemRepo.On("GetByID", ctx, userID, itemID).Return(entity.Items{}, errDB).Once()

		res, err := u.GetItem(ctx, userID, wishlistID, itemID)
		require.Error(t, err)
		assert.ErrorIs(t, err, errDB)
		assert.Equal(t, entity.Items{}, res)
		itemRepo.AssertExpectations(t)
	})

	t.Run("item belongs to another wishlist", func(t *testing.T) {
		itemRepo := new(itemsRepositoryMock)
		wishRepo := new(wishListRepositoryMock)
		u := newItemsUseCaseForTest(itemRepo, wishRepo)

		itemRepo.On("GetByID", ctx, userID, itemID).Return(entity.Items{ID: itemID, WishListID: uuid.New()}, nil).Once()

		res, err := u.GetItem(ctx, userID, wishlistID, itemID)
		require.Error(t, err)
		assert.ErrorIs(t, err, entity.ErrItemNotFound)
		assert.Equal(t, entity.Items{}, res)
		itemRepo.AssertExpectations(t)
	})

	t.Run("success", func(t *testing.T) {
		itemRepo := new(itemsRepositoryMock)
		wishRepo := new(wishListRepositoryMock)
		u := newItemsUseCaseForTest(itemRepo, wishRepo)

		expected := entity.Items{ID: itemID, WishListID: wishlistID, Title: "Book"}
		itemRepo.On("GetByID", ctx, userID, itemID).Return(expected, nil).Once()

		res, err := u.GetItem(ctx, userID, wishlistID, itemID)
		require.NoError(t, err)
		assert.Equal(t, expected, res)
		itemRepo.AssertExpectations(t)
	})
}

func TestUseCaseUpdateItem(t *testing.T) {
	ctx := context.Background()
	userID := uuid.New()
	input := entity.Items{ID: uuid.New(), WishListID: uuid.New(), Title: "Book", Priority: 5}

	t.Run("item not found", func(t *testing.T) {
		itemRepo := new(itemsRepositoryMock)
		wishRepo := new(wishListRepositoryMock)
		u := newItemsUseCaseForTest(itemRepo, wishRepo)

		itemRepo.On("Update", ctx, userID, input).Return(entity.Items{}, entity.ErrItemNotFound).Once()

		res, err := u.UpdateItem(ctx, userID, input)
		require.Error(t, err)
		assert.ErrorIs(t, err, entity.ErrItemNotFound)
		assert.Equal(t, entity.Items{}, res)
		itemRepo.AssertExpectations(t)
	})

	t.Run("repository error", func(t *testing.T) {
		itemRepo := new(itemsRepositoryMock)
		wishRepo := new(wishListRepositoryMock)
		u := newItemsUseCaseForTest(itemRepo, wishRepo)

		errDB := errors.New("db failed")
		itemRepo.On("Update", ctx, userID, input).Return(entity.Items{}, errDB).Once()

		res, err := u.UpdateItem(ctx, userID, input)
		require.Error(t, err)
		assert.ErrorIs(t, err, errDB)
		assert.Equal(t, entity.Items{}, res)
		itemRepo.AssertExpectations(t)
	})

	t.Run("success", func(t *testing.T) {
		itemRepo := new(itemsRepositoryMock)
		wishRepo := new(wishListRepositoryMock)
		u := newItemsUseCaseForTest(itemRepo, wishRepo)

		expected := input
		expected.Title = "Updated"
		itemRepo.On("Update", ctx, userID, input).Return(expected, nil).Once()

		res, err := u.UpdateItem(ctx, userID, input)
		require.NoError(t, err)
		assert.Equal(t, expected, res)
		itemRepo.AssertExpectations(t)
	})
}

func TestUseCaseDeleteItem(t *testing.T) {
	ctx := context.Background()
	userID := uuid.New()
	wishlistID := uuid.New()
	itemID := uuid.New()

	t.Run("not found is idempotent", func(t *testing.T) {
		itemRepo := new(itemsRepositoryMock)
		wishRepo := new(wishListRepositoryMock)
		u := newItemsUseCaseForTest(itemRepo, wishRepo)

		itemRepo.On("Delete", ctx, userID, wishlistID, itemID).Return(entity.ErrItemNotFound).Once()

		err := u.DeleteItem(ctx, userID, wishlistID, itemID)
		require.NoError(t, err)
		itemRepo.AssertExpectations(t)
	})

	t.Run("repository error", func(t *testing.T) {
		itemRepo := new(itemsRepositoryMock)
		wishRepo := new(wishListRepositoryMock)
		u := newItemsUseCaseForTest(itemRepo, wishRepo)

		errDB := errors.New("db failed")
		itemRepo.On("Delete", ctx, userID, wishlistID, itemID).Return(errDB).Once()

		err := u.DeleteItem(ctx, userID, wishlistID, itemID)
		require.Error(t, err)
		assert.ErrorIs(t, err, errDB)
		itemRepo.AssertExpectations(t)
	})

	t.Run("success", func(t *testing.T) {
		itemRepo := new(itemsRepositoryMock)
		wishRepo := new(wishListRepositoryMock)
		u := newItemsUseCaseForTest(itemRepo, wishRepo)

		itemRepo.On("Delete", ctx, userID, wishlistID, itemID).Return(nil).Once()

		err := u.DeleteItem(ctx, userID, wishlistID, itemID)
		require.NoError(t, err)
		itemRepo.AssertExpectations(t)
	})
}
