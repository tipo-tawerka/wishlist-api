package handler

import (
	"errors"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/tipo-tawerka/wishlist-api/internal/controller/http/request"
	"github.com/tipo-tawerka/wishlist-api/internal/controller/http/response"
	"github.com/tipo-tawerka/wishlist-api/internal/entity"
)

func (h *Handler) GetItems(w http.ResponseWriter, r *http.Request) {
	userID, wishlistID, err := h.getUserWishListID(w, r)
	if err != nil {
		return
	}
	items, err := h.items.GetItems(r.Context(), userID, wishlistID)
	switch {
	case errors.Is(err, entity.ErrWishListNotFound):
		response.WriteError(w, http.StatusNotFound)
	case err != nil:
		response.WriteError(w, http.StatusInternalServerError)
	default:
		resp := make([]response.Items, len(items))
		for i, item := range items {
			resp[i] = h.toItemResponse(item)
		}
		response.WriteJSON(w, resp, http.StatusOK)
	}
}

func (h *Handler) CreateItem(w http.ResponseWriter, r *http.Request) {
	userID, wishlistID, err := h.getUserWishListID(w, r)
	if err != nil {
		return
	}
	var req request.SetItems
	if err := h.decodeAndValidate(r, &req); err != nil {
		response.WriteError(w, http.StatusBadRequest)
		return
	}
	item, err := h.items.CreateItem(r.Context(), userID, entity.Items{
		WishListID:  wishlistID,
		Title:       req.Title,
		Description: req.Description,
		ProductURL:  req.ProductUrl,
		Priority:    req.Priority,
	})
	switch {
	case errors.Is(err, entity.ErrWishListNotFound):
		response.WriteError(w, http.StatusNotFound)
	case err != nil:
		response.WriteError(w, http.StatusInternalServerError)
	default:
		response.WriteJSON(w, h.toItemResponse(item), http.StatusCreated)
	}
}

func (h *Handler) GetItem(w http.ResponseWriter, r *http.Request) {
	userID, wishlistID, err := h.getUserWishListID(w, r)
	if err != nil {
		return
	}
	itemID, err := uuid.Parse(r.PathValue("itemId"))
	if err != nil {
		response.WriteError(w, http.StatusNotFound)
		return
	}
	item, err := h.items.GetItem(r.Context(), userID, wishlistID, itemID)
	switch {
	case errors.Is(err, entity.ErrItemNotFound):
		response.WriteError(w, http.StatusNotFound)
	case err != nil:
		response.WriteError(w, http.StatusInternalServerError)
	default:
		response.WriteJSON(w, h.toItemResponse(item), http.StatusOK)
	}
}

func (h *Handler) UpdateItem(w http.ResponseWriter, r *http.Request) {
	userID, wishlistID, err := h.getUserWishListID(w, r)
	if err != nil {
		return
	}
	itemID, err := uuid.Parse(r.PathValue("itemId"))
	if err != nil {
		response.WriteError(w, http.StatusNotFound)
		return
	}
	var req request.UpdateItems
	if err := h.decodeAndValidate(r, &req); err != nil {
		response.WriteError(w, http.StatusBadRequest)
		return
	}
	current, err := h.items.GetItem(r.Context(), userID, wishlistID, itemID)
	switch {
	case errors.Is(err, entity.ErrItemNotFound):
		response.WriteError(w, http.StatusNotFound)
		return
	case err != nil:
		response.WriteError(w, http.StatusInternalServerError)
		return
	}

	if req.Title != nil {
		current.Title = *req.Title
	}
	if req.Description != nil {
		current.Description = *req.Description
	}
	if req.ProductUrl != nil {
		current.ProductURL = *req.ProductUrl
	}
	if req.Priority != nil {
		current.Priority = *req.Priority
	}

	item, err := h.items.UpdateItem(r.Context(), userID, current)
	switch {
	case errors.Is(err, entity.ErrItemNotFound):
		response.WriteError(w, http.StatusNotFound)
	case err != nil:
		response.WriteError(w, http.StatusInternalServerError)
	default:
		response.WriteJSON(w, h.toItemResponse(item), http.StatusOK)
	}
}

func (h *Handler) DeleteItem(w http.ResponseWriter, r *http.Request) {
	userID, wishlistID, err := h.getUserWishListID(w, r)
	if err != nil {
		return
	}
	itemID, err := uuid.Parse(r.PathValue("itemId"))
	if err != nil {
		response.WriteError(w, http.StatusNotFound)
		return
	}
	err = h.items.DeleteItem(r.Context(), userID, wishlistID, itemID)
	switch {
	case err != nil:
		response.WriteError(w, http.StatusInternalServerError)
	default:
		w.WriteHeader(http.StatusNoContent)
	}
}

func (h *Handler) toItemResponse(item entity.Items) response.Items {
	return response.Items{
		ID:          item.ID.String(),
		WishListID:  item.WishListID.String(),
		Title:       item.Title,
		Description: item.Description,
		ProductURL:  item.ProductURL,
		Priority:    item.Priority,
		Reserved:    item.Reserved,
		CreatedAt:   item.CreatedAt.Format(time.RFC3339),
		UpdatedAt:   item.UpdatedAt.Format(time.RFC3339),
	}
}

func (h *Handler) getUserWishListID(w http.ResponseWriter, r *http.Request) (uuid.UUID, uuid.UUID, error) {
	userID, err := request.GetUserID(r.Context())
	if err != nil {
		response.WriteError(w, http.StatusUnauthorized)
		return uuid.UUID{}, uuid.UUID{}, err
	}
	wishlistID, err := uuid.Parse(r.PathValue("wishlistId"))
	if err != nil {
		response.WriteError(w, http.StatusNotFound)
		return uuid.UUID{}, uuid.UUID{}, err
	}
	return userID, wishlistID, nil
}
