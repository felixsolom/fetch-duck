package main

import (
	"context"
	"net/http"
	"time"

	"github.com/felixsolom/fetch-duck/internal/database"
)

type contextKey string

const userContextKey = contextKey("user")

func (cfg *apiConfig) authMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cookie, err := r.Cookie("session_token")
		if err != nil {
			respondWithError(w, http.StatusUnauthorized, "Not authenticated", err)
			return
		}
		sessionToken := cookie.Value

		user, err := cfg.DB.GetUserBySessionToken(r.Context(), database.GetUserBySessionTokenParams{
			Token:  sessionToken,
			Expiry: time.Now().Unix(),
		})
		if err != nil {
			respondWithError(w, http.StatusUnauthorized, "Invalid session", err)
			return
		}
		ctx := context.WithValue(r.Context(), userContextKey, user)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func getUserFromContext(r *http.Request) (database.User, bool) {
	user, ok := r.Context().Value(userContextKey).(database.User)
	return user, ok
}
