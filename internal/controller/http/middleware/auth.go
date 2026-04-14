package middleware

import (
	"net/http"
	"strings"

	"github.com/google/uuid"
	"github.com/tipo-tawerka/wishlist-api/internal/controller/http/request"
	"github.com/tipo-tawerka/wishlist-api/internal/controller/http/response"
)

type TokenManager interface {
	Parse(tokenString string) (uuid.UUID, error)
}

type AuthMiddleware struct {
	token TokenManager
}

func NewAuthMiddleware(token TokenManager) *AuthMiddleware {
	return &AuthMiddleware{token: token}
}

func (am *AuthMiddleware) Auth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		header := r.Header.Get("Authorization")
		if header == "" || !strings.HasPrefix(header, "Bearer ") {
			response.WriteError(w, http.StatusUnauthorized)
			return
		}
		id, err := am.token.Parse(header[len("Bearer "):])
		if err != nil {
			response.WriteError(w, http.StatusUnauthorized)
			return
		}
		next.ServeHTTP(w, r.WithContext(request.PutUserID(r.Context(), id)))
	})
}
