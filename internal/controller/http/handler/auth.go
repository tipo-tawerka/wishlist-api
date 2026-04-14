package handler

import (
	"errors"
	"net/http"

	"github.com/tipo-tawerka/wishlist-api/internal/controller/http/request"
	"github.com/tipo-tawerka/wishlist-api/internal/controller/http/response"
	"github.com/tipo-tawerka/wishlist-api/internal/entity"
)

func (h *Handler) Registration(w http.ResponseWriter, r *http.Request) {
	var req request.Register
	if err := h.decodeAndValidate(r, &req); err != nil {
		response.WriteError(w, http.StatusBadRequest)
		return
	}
	_, err := h.auth.Register(r.Context(), req.Email, req.Password)
	switch {
	case errors.Is(err, entity.ErrUserAlreadyExists):
		response.WriteError(w, http.StatusConflict)
	case err != nil:
		response.WriteError(w, http.StatusInternalServerError)
	default:
		response.WriteJSON(w, nil, http.StatusCreated)
	}
}

func (h *Handler) Login(w http.ResponseWriter, r *http.Request) {
	var req request.Login
	if err := h.decodeAndValidate(r, &req); err != nil {
		response.WriteError(w, http.StatusBadRequest)
		return
	}
	token, err := h.auth.Login(r.Context(), req.Email, req.Password)
	switch {
	case errors.Is(err, entity.ErrUserInvalidCredentials):
		response.WriteError(w, http.StatusUnauthorized)
	case err != nil:
		response.WriteError(w, http.StatusInternalServerError)
	default:
		response.WriteJSON(w, response.Login{Token: token}, http.StatusOK)
	}
}
