package wishlist

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/rs/zerolog"
	"github.com/tipo-tawerka/wishlist-api/internal/entity"
)

type WishlistRepository interface {
	GetByUserId(ctx context.Context, userId uuid.UUID) ([]entity.WishList, error)
	GetByWishId(ctx context.Context, wishId uuid.UUID) (entity.WishList, error)
	Save(ctx context.Context, wishlist entity.WishList) (entity.WishList, error)
	Update(ctx context.Context, wishlist entity.WishList) (entity.WishList, error)
	Delete(ctx context.Context, userId, wishlistId uuid.UUID) error
}

type UseCase struct {
	repo   WishlistRepository
	logger *zerolog.Logger
}

func NewUseCase(wishlistRepo WishlistRepository, logger *zerolog.Logger) *UseCase {
	return &UseCase{
		repo:   wishlistRepo,
		logger: logger,
	}
}

func (u *UseCase) GetAllWishList(ctx context.Context, userID uuid.UUID) ([]entity.WishList, error) {
	wishlist, err := u.repo.GetByUserId(ctx, userID)
	if err != nil {
		u.logger.Error().Err(err).Str("userID", userID.String()).Send()
		return nil, err
	}
	u.logger.Info().Str("userID", userID.String()).Msg("wishlists retrieved successfully")
	return wishlist, nil
}

func (u *UseCase) CreateWishList(ctx context.Context, wishlist entity.WishList) (entity.WishList, error) {
	wishlist, err := u.repo.Save(ctx, wishlist)
	if err != nil {
		u.logger.Error().Err(err).Str("userID", wishlist.UserID.String()).Send()
		return entity.WishList{}, err
	}
	u.logger.Info().Str("userID", wishlist.UserID.String()).Msg("wishlist created successfully")
	return wishlist, nil
}

func (u *UseCase) GetWishList(ctx context.Context, userID, wishID uuid.UUID) (entity.WishList, error) {
	wishlist, err := u.repo.GetByWishId(ctx, wishID)
	switch {
	case errors.Is(err, entity.ErrWishListNotFound):
		u.logger.Warn().Err(err).Str("wishID", wishID.String()).Send()
		return entity.WishList{}, entity.ErrWishListNotFound
	case err != nil:
		u.logger.Error().Err(err).Str("wishID", wishID.String()).Send()
		return entity.WishList{}, err
	}
	if wishlist.UserID != userID {
		u.logger.Warn().Str("userID", userID.String()).Str("wishID", wishID.String()).
			Msg("wishlist does not belong to user")
		return entity.WishList{}, entity.ErrWishListNotFound
	}
	u.logger.Info().Str("wishID", wishID.String()).Msg("wishlist retrieved successfully")
	return wishlist, nil
}

func (u *UseCase) UpdateWishList(ctx context.Context, wishlist entity.WishList) (entity.WishList, error) {
	wishlist, err := u.repo.Update(ctx, wishlist)
	switch {
	case errors.Is(err, entity.ErrWishListNotFound):
		u.logger.Warn().Err(err).Str("wishID", wishlist.ID.String()).Send()
		return entity.WishList{}, entity.ErrWishListNotFound
	case err != nil:
		u.logger.Error().Err(err).Str("wishID", wishlist.ID.String()).Send()
		return entity.WishList{}, err
	}
	u.logger.Info().Str("wishID", wishlist.ID.String()).Msg("wishlist updated successfully")
	return wishlist, nil
}

func (u *UseCase) DeleteWishList(ctx context.Context, userID, wishID uuid.UUID) error {
	err := u.repo.Delete(ctx, userID, wishID)
	switch {
	case errors.Is(err, entity.ErrWishListNotFound):
		u.logger.Warn().Err(err).Str("wishID", wishID.String()).Msg("wishlist does not exist")
		return nil
	case err != nil:
		u.logger.Error().Err(err).Str("wishID", wishID.String()).Send()
		return err
	}
	u.logger.Info().Str("wishID", wishID.String()).Msg("wishlist deleted successfully")
	return nil
}
