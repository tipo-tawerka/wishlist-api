package items

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/rs/zerolog"
	"github.com/tipo-tawerka/wishlist-api/internal/entity"
)

type ItemsRepository interface {
	GetAll(ctx context.Context, userID, wishListID uuid.UUID) ([]entity.Items, error)
	GetByID(ctx context.Context, userID, itemID uuid.UUID) (entity.Items, error)
	Create(ctx context.Context, userID uuid.UUID, item entity.Items) (entity.Items, error)
	Update(ctx context.Context, userID uuid.UUID, item entity.Items) (entity.Items, error)
	Delete(ctx context.Context, userID, wishlistID, itemID uuid.UUID) error
}

type WishListRepository interface {
	GetWishList(ctx context.Context, userID, wishID uuid.UUID) (entity.WishList, error)
}

type UseCase struct {
	itemRepo ItemsRepository
	wishRepo WishListRepository
	logger   *zerolog.Logger
}

func NewUseCase(itemRepo ItemsRepository, wishRepo WishListRepository, logger *zerolog.Logger) *UseCase {
	return &UseCase{
		itemRepo: itemRepo,
		wishRepo: wishRepo,
		logger:   logger,
	}
}

func (u *UseCase) GetItems(ctx context.Context, userID, wishlistID uuid.UUID) ([]entity.Items, error) {
	wishlist, err := u.wishRepo.GetWishList(ctx, userID, wishlistID)
	switch {
	case errors.Is(err, entity.ErrWishListNotFound):
		u.logger.Warn().Err(err).Str("wishlistID", wishlistID.String()).Send()
		return nil, entity.ErrWishListNotFound
	case err != nil:
		u.logger.Error().Err(err).Str("wishlistID", wishlistID.String()).Send()
		return nil, err
	}
	u.logger.Info().Str("wishlistID", wishlist.ID.String()).Msg("wishlist found, retrieving items")
	items, err := u.itemRepo.GetAll(ctx, userID, wishlistID)
	if err != nil {
		u.logger.Error().Err(err).Str("wishlistID", wishlistID.String()).Send()
		return nil, err
	}
	u.logger.Info().Str("wishlistID", wishlistID.String()).Msg("items retrieved successfully")
	return items, nil
}

func (u *UseCase) CreateItem(ctx context.Context, userID uuid.UUID, item entity.Items) (entity.Items, error) {
	created, err := u.itemRepo.Create(ctx, userID, item)
	switch {
	case errors.Is(err, entity.ErrWishListNotFound):
		u.logger.Warn().Err(err).Str("wishlistID", item.WishListID.String()).Send()
		return entity.Items{}, entity.ErrWishListNotFound
	case err != nil:
		u.logger.Error().Err(err).Str("wishlistID", item.WishListID.String()).Send()
		return entity.Items{}, err
	}
	u.logger.Info().Str("itemID", created.ID.String()).Msg("item created successfully")
	return created, nil
}

func (u *UseCase) GetItem(ctx context.Context, userID, wishlistID, itemID uuid.UUID) (entity.Items, error) {
	item, err := u.itemRepo.GetByID(ctx, userID, itemID)
	switch {
	case errors.Is(err, entity.ErrItemNotFound):
		u.logger.Warn().Err(err).Str("itemID", itemID.String()).Send()
		return entity.Items{}, entity.ErrItemNotFound
	case err != nil:
		u.logger.Error().Err(err).Str("itemID", itemID.String()).Send()
		return entity.Items{}, err
	}
	if item.WishListID != wishlistID {
		u.logger.Warn().Str("itemID", itemID.String()).Str("wishlistID", wishlistID.String()).
			Msg("item does not belong to wishlist")
		return entity.Items{}, entity.ErrItemNotFound
	}
	u.logger.Info().Str("itemID", itemID.String()).Msg("item retrieved successfully")
	return item, nil
}

func (u *UseCase) UpdateItem(ctx context.Context, userID uuid.UUID, item entity.Items) (entity.Items, error) {
	updated, err := u.itemRepo.Update(ctx, userID, item)
	switch {
	case errors.Is(err, entity.ErrItemNotFound):
		u.logger.Warn().Err(err).Str("itemID", item.ID.String()).Send()
		return entity.Items{}, entity.ErrItemNotFound
	case err != nil:
		u.logger.Error().Err(err).Str("itemID", item.ID.String()).Send()
		return entity.Items{}, err
	}
	u.logger.Info().Str("itemID", updated.ID.String()).Msg("item updated successfully")
	return updated, nil
}

func (u *UseCase) DeleteItem(ctx context.Context, userID, wishlistID, itemID uuid.UUID) error {
	err := u.itemRepo.Delete(ctx, userID, wishlistID, itemID)
	switch {
	case errors.Is(err, entity.ErrItemNotFound):
		u.logger.Warn().Err(err).Str("itemID", itemID.String()).Msg("item does not exist")
		return nil
	case err != nil:
		u.logger.Error().Err(err).Str("itemID", itemID.String()).Send()
		return err
	}
	u.logger.Info().Str("itemID", itemID.String()).Msg("item deleted successfully")
	return nil
}
