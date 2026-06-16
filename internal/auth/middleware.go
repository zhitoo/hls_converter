package auth

import (
	"context"
	"net/http"
	"strings"
)

type contextKey string

const userKey contextKey = "authenticated_user"

func Middleware(repo Repository) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			token := bearerToken(r)
			if token == "" {
				http.Error(w, `{"error":"missing authorization token"}`, http.StatusUnauthorized)
				return
			}

			user, err := repo.FindByToken(token)
			if err != nil {
				http.Error(w, `{"error":"invalid token"}`, http.StatusUnauthorized)
				return
			}

			next.ServeHTTP(w, r.WithContext(context.WithValue(r.Context(), userKey, user)))
		})
	}
}

func bearerToken(r *http.Request) string {
	h := r.Header.Get("Authorization")
	if strings.HasPrefix(h, "Bearer ") {
		return strings.TrimPrefix(h, "Bearer ")
	}
	return ""
}
