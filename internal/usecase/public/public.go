package public

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/rs/zerolog"
	"github.com/tipo-tawerka/wishlist-api/internal/entity"
)

type WishListRepository interface {
	GetByPublicToken(ctx context.Context, token uuid.UUID) (entity.WishList, error)
}

type ItemsRepository interface {
	ListByWishlist(ctx context.Context, wishListId uuid.UUID) ([]entity.Items, error)
	Reserve(ctx context.Context, wishlistId, itemId uuid.UUID) (entity.Items, error)
}

type UseCase struct {
	wishlistRepo WishListRepository
	itemRepo     ItemsRepository
	logger       *zerolog.Logger
}

func NewUseCase(wishlistRepo WishListRepository, itemRepo ItemsRepository, logger *zerolog.Logger) *UseCase {
	return &UseCase{
		wishlistRepo: wishlistRepo,
		itemRepo:     itemRepo,
		logger:       logger,
	}
}

func (u *UseCase) GetWishList(ctx context.Context, token uuid.UUID) (entity.WishList, []entity.Items, error) {
	wishlist, err := u.wishlistRepo.GetByPublicToken(ctx, token)
	switch {
	case errors.Is(err, entity.ErrWishListNotFound):
		u.logger.Warn().Err(err).Str("token", token.String()).Send()
		return entity.WishList{}, nil, entity.ErrWishListNotFound
	case err != nil:
		u.logger.Error().Err(err).Str("token", token.String()).Send()
		return entity.WishList{}, nil, err
	}
	items, err := u.itemRepo.ListByWishlist(ctx, wishlist.ID)
	if err != nil {
		u.logger.Error().Err(err).Str("wishlistID", wishlist.ID.String()).Send()
		return entity.WishList{}, nil, err
	}
	u.logger.Info().Str("token", token.String()).Msg("public wishlist retrieved successfully")
	return wishlist, items, nil
}

func (u *UseCase) ReserveItem(ctx context.Context, token, itemID uuid.UUID) (entity.Items, error) {
	wishlist, err := u.wishlistRepo.GetByPublicToken(ctx, token)
	switch {
	case errors.Is(err, entity.ErrWishListNotFound):
		u.logger.Warn().Err(err).Str("token", token.String()).Send()
		return entity.Items{}, entity.ErrWishListNotFound
	case err != nil:
		u.logger.Error().Err(err).Str("token", token.String()).Send()
		return entity.Items{}, err
	}
	reserved, err := u.itemRepo.Reserve(ctx, wishlist.ID, itemID)
	switch {
	case errors.Is(err, entity.ErrItemNotFound):
		u.logger.Warn().Err(err).Str("itemID", itemID.String()).Send()
		return entity.Items{}, entity.ErrItemNotFound
	case errors.Is(err, entity.ErrItemAlreadyBooked):
		u.logger.Warn().Err(err).Str("itemID", itemID.String()).Send()
		return entity.Items{}, entity.ErrItemAlreadyBooked
	case err != nil:
		u.logger.Error().Err(err).Str("itemID", itemID.String()).Send()
		return entity.Items{}, err
	}
	u.logger.Info().Str("itemID", itemID.String()).Msg("item reserved successfully")
	return reserved, nil
}
