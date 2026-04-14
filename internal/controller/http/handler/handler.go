package handler

import (
	"encoding/json"
	"net/http"

	"github.com/go-playground/validator/v10"
	"github.com/tipo-tawerka/wishlist-api/internal/usecase/auth"
	"github.com/tipo-tawerka/wishlist-api/internal/usecase/items"
	"github.com/tipo-tawerka/wishlist-api/internal/usecase/public"
	"github.com/tipo-tawerka/wishlist-api/internal/usecase/wishlist"
)

type Handler struct {
	auth      *auth.UseCase
	wishList  *wishlist.UseCase
	items     *items.UseCase
	public    *public.UseCase
	validator *validator.Validate
}

type Options struct {
	Auth      *auth.UseCase
	WishList  *wishlist.UseCase
	Items     *items.UseCase
	Public    *public.UseCase
	Validator *validator.Validate
}

func NewHandler(opts Options) *Handler {
	return &Handler{
		auth:      opts.Auth,
		wishList:  opts.WishList,
		items:     opts.Items,
		public:    opts.Public,
		validator: opts.Validator,
	}
}

func (h *Handler) decodeAndValidate(r *http.Request, dst any) error {
	if err := json.NewDecoder(r.Body).Decode(dst); err != nil {
		return err
	}
	return h.validator.Struct(dst)
}
