package handler

import (
	"errors"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/tipo-tawerka/wishlist-api/internal/controller/http/response"
	"github.com/tipo-tawerka/wishlist-api/internal/entity"
)

func (h *Handler) GetPublicWishList(w http.ResponseWriter, r *http.Request) {
	token, err := uuid.Parse(r.PathValue("token"))
	if err != nil {
		response.WriteError(w, http.StatusNotFound)
		return
	}
	wl, items, err := h.public.GetWishList(r.Context(), token)
	switch {
	case errors.Is(err, entity.ErrWishListNotFound):
		response.WriteError(w, http.StatusNotFound)
	case err != nil:
		response.WriteError(w, http.StatusInternalServerError)
	default:
		resp := response.PublicWishList{
			Title:       wl.Title,
			Description: wl.Description,
			EventDate:   wl.EventDate.Format(time.DateOnly),
			Items:       make([]response.PublicItems, len(items)),
		}
		for i, item := range items {
			resp.Items[i] = h.toPublicItemResponse(item)
		}
		response.WriteJSON(w, resp, http.StatusOK)
	}
}

func (h *Handler) ReserveItem(w http.ResponseWriter, r *http.Request) {
	token, err := uuid.Parse(r.PathValue("token"))
	if err != nil {
		response.WriteError(w, http.StatusNotFound)
		return
	}
	itemID, err := uuid.Parse(r.PathValue("itemId"))
	if err != nil {
		response.WriteError(w, http.StatusNotFound)
		return
	}
	item, err := h.public.ReserveItem(r.Context(), token, itemID)
	switch {
	case errors.Is(err, entity.ErrWishListNotFound), errors.Is(err, entity.ErrItemNotFound):
		response.WriteError(w, http.StatusNotFound)
	case errors.Is(err, entity.ErrItemAlreadyBooked):
		response.WriteError(w, http.StatusConflict)
	case err != nil:
		response.WriteError(w, http.StatusInternalServerError)
	default:
		response.WriteJSON(w, h.toPublicItemResponse(item), http.StatusOK)
	}
}

func (h *Handler) toPublicItemResponse(item entity.Items) response.PublicItems {
	return response.PublicItems{
		ID:          item.ID.String(),
		Title:       item.Title,
		Description: item.Description,
		ProductURL:  item.ProductURL,
		Priority:    item.Priority,
		Reserved:    item.Reserved,
	}
}
