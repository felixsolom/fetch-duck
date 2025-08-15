package main

import (
	"log"
	"net/http"
	"time"
)

func (cfg *apiConfig) handlerAuthStatus(w http.ResponseWriter, r *http.Request) {
	user, ok := getUserFromContext(r)
	if !ok {
		respondWithError(w, http.StatusInternalServerError, "Could not get user from context", nil)
		return
	}
	respondWithJSON(w, http.StatusOK, user)
}

func (cfg *apiConfig) handlerLogout(w http.ResponseWriter, r *http.Request) {
	cookie, err := r.Cookie("session_token")
	if err != nil {
		respondWithJSON(w, http.StatusOK, map[string]string{"status": "logged out"})
		return
	}
	sessionToken := cookie.Value

	err = cfg.DB.DeleteSession(r.Context(), sessionToken)
	if err != nil {
		log.Printf("Failed to delete session from DB: %v", err)
	}

	http.SetCookie(w, &http.Cookie{
		Name:     "session_token",
		Value:    "",
		Expires:  time.Unix(0, 0),
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
	})
	respondWithJSON(w, http.StatusOK, map[string]string{"status": "logged out"})
}
