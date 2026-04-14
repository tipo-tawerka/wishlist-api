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

func (h *Handler) GetAllWishList(w http.ResponseWriter, r *http.Request) {
	userID, err := request.GetUserID(r.Context())
	if err != nil {
		response.WriteError(w, http.StatusUnauthorized)
		return
	}
	wishLists, err := h.wishList.GetAllWishList(r.Context(), userID)
	if err != nil {
		response.WriteError(w, http.StatusInternalServerError)
		return
	}
	resp := make([]response.WishList, len(wishLists))
	for i, wl := range wishLists {
		resp[i] = h.toWishListResponse(wl)
	}
	response.WriteJSON(w, resp, http.StatusOK)
}

func (h *Handler) CreateWishList(w http.ResponseWriter, r *http.Request) {
	userID, err := request.GetUserID(r.Context())
	if err != nil {
		response.WriteError(w, http.StatusUnauthorized)
		return
	}
	var req request.SetWishlist
	if err := h.decodeAndValidate(r, &req); err != nil {
		response.WriteError(w, http.StatusBadRequest)
		return
	}
	eventDate, err := time.Parse(time.DateOnly, req.EventDate)
	if err != nil {
		response.WriteError(w, http.StatusBadRequest)
		return
	}
	token := uuid.New()
	wl, err := h.wishList.CreateWishList(r.Context(), entity.WishList{
		UserID:      userID,
		Title:       req.Title,
		Description: req.Description,
		EventDate:   eventDate,
		PublicToken: &token,
	})
	if err != nil {
		response.WriteError(w, http.StatusInternalServerError)
		return
	}
	response.WriteJSON(w, h.toWishListResponse(wl), http.StatusCreated)
}

func (h *Handler) GetWishList(w http.ResponseWriter, r *http.Request) {
	userID, wishlistID, err := h.getUserWishListID(w, r)
	if err != nil {
		return
	}
	wl, err := h.wishList.GetWishList(r.Context(), userID, wishlistID)
	switch {
	case errors.Is(err, entity.ErrWishListNotFound):
		response.WriteError(w, http.StatusNotFound)
	case err != nil:
		response.WriteError(w, http.StatusInternalServerError)
	default:
		items, err := h.items.GetItems(r.Context(), userID, wishlistID)
		switch {
		case errors.Is(err, entity.ErrWishListNotFound):
			response.WriteError(w, http.StatusNotFound)
		case err != nil:
			response.WriteError(w, http.StatusInternalServerError)
		default:
			respItems := make([]response.Items, len(items))
			for i, item := range items {
				respItems[i] = h.toItemResponse(item)
			}
			response.WriteJSON(w, response.WishListWithItems{
				WishList: h.toWishListResponse(wl),
				Items:    respItems,
			}, http.StatusOK)
		}
	}
}

func (h *Handler) UpdateWishList(w http.ResponseWriter, r *http.Request) {
	userID, wishlistID, err := h.getUserWishListID(w, r)
	if err != nil {
		return
	}
	var req request.UpdateWishlist
	if err := h.decodeAndValidate(r, &req); err != nil {
		response.WriteError(w, http.StatusBadRequest)
		return
	}
	current, err := h.wishList.GetWishList(r.Context(), userID, wishlistID)
	switch {
	case errors.Is(err, entity.ErrWishListNotFound):
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
	if req.EventDate != nil {
		eventDate, err := time.Parse(time.DateOnly, *req.EventDate)
		if err != nil {
			response.WriteError(w, http.StatusBadRequest)
			return
		}
		current.EventDate = eventDate
	}

	wl, err := h.wishList.UpdateWishList(r.Context(), current)
	switch {
	case errors.Is(err, entity.ErrWishListNotFound):
		response.WriteError(w, http.StatusNotFound)
	case err != nil:
		response.WriteError(w, http.StatusInternalServerError)
	default:
		response.WriteJSON(w, h.toWishListResponse(wl), http.StatusOK)
	}
}

func (h *Handler) DeleteWishList(w http.ResponseWriter, r *http.Request) {
	userID, wishlistID, err := h.getUserWishListID(w, r)
	if err != nil {
		return
	}
	err = h.wishList.DeleteWishList(r.Context(), userID, wishlistID)
	switch {
	case err != nil:
		response.WriteError(w, http.StatusInternalServerError)
	default:
		w.WriteHeader(http.StatusNoContent)
	}
}

func (h *Handler) toWishListResponse(wl entity.WishList) response.WishList {
	var publicToken *string
	if wl.PublicToken != nil {
		token := wl.PublicToken.String()
		publicToken = &token
	}
	return response.WishList{
		ID:          wl.ID.String(),
		Title:       wl.Title,
		Description: wl.Description,
		EventDate:   wl.EventDate.Format(time.DateOnly),
		PublicToken: publicToken,
		CreatedAt:   wl.CreatedAt.Format(time.RFC3339),
		UpdatedAt:   wl.UpdatedAt.Format(time.RFC3339),
	}
}
