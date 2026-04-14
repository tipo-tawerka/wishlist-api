package http

import (
	"time"

	"github.com/go-chi/chi/v5"
	chiMiddleware "github.com/go-chi/chi/v5/middleware"
	"github.com/tipo-tawerka/wishlist-api/internal/controller/http/handler"
	"github.com/tipo-tawerka/wishlist-api/internal/controller/http/middleware"
)

func NewRouter(h *handler.Handler, am *middleware.AuthMiddleware) chi.Router {
	r := chi.NewRouter()

	r.Use(chiMiddleware.Timeout(5 * time.Second))
	r.Use(chiMiddleware.Recoverer)
	r.Use(chiMiddleware.Logger)

	r.Route("/api/v1", func(r chi.Router) {
		r.Group(func(r chi.Router) {
			r.Route("/auth", func(r chi.Router) {
				r.Post("/register", h.Registration)
				r.Post("/login", h.Login)
			})

			r.Route("/public", func(r chi.Router) {
				r.Get("/wishlists/{token}", h.GetPublicWishList)
				r.Post("/wishlists/{token}/items/{itemId}/reserve", h.ReserveItem)
			})
		})

		r.Group(func(r chi.Router) {
			r.Use(am.Auth)

			r.Route("/wishlists", func(r chi.Router) {
				r.Get("/", h.GetAllWishList)
				r.Post("/", h.CreateWishList)

				r.Route("/{wishlistId}", func(r chi.Router) {
					r.Get("/", h.GetWishList)
					r.Put("/", h.UpdateWishList)
					r.Delete("/", h.DeleteWishList)

					r.Route("/items", func(r chi.Router) {
						r.Get("/", h.GetItems)
						r.Post("/", h.CreateItem)

						r.Route("/{itemId}", func(r chi.Router) {
							r.Get("/", h.GetItem)
							r.Put("/", h.UpdateItem)
							r.Delete("/", h.DeleteItem)
						})
					})
				})
			})
		})
	})

	return r
}
