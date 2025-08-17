package main

import (
	"encoding/json"
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

type invitePayload struct {
	Code string `json:"code"`
}

func (cfg *apiConfig) handlerVerifyInvite(w http.ResponseWriter, r *http.Request) {
	var payload invitePayload
	err := json.NewDecoder(r.Body).Decode(&payload)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid request payload", err)
		return
	}

	if payload.Code == cfg.App.InviteCode {
		http.SetCookie(w, &http.Cookie{
			Name:     "pre_auth_token",
			Value:    "valid",
			Expires:  time.Now().Add(15 * time.Minute),
			HttpOnly: true,
			Path:     "/",
			SameSite: http.SameSiteLaxMode,
		})
		respondWithJSON(w, http.StatusOK, map[string]string{"status": "ok"})
	} else {
		respondWithError(w, http.StatusUnauthorized, "Invalid invite code", nil)
	}
}
