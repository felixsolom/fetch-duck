package main

import (
	"context"
	"log"
	"net/http"

	"github.com/felixsolom/fetch-duck/internal/database"
	"github.com/felixsolom/fetch-duck/internal/googleauth"
	"github.com/google/uuid"
	"golang.org/x/oauth2"
)

type apiConfig struct {
	DB           *database.Queries
	GoogleConfig *oauth2.Config
}

func (cfg *apiConfig) handlerOAuthGoogleLogin(w http.ResponseWriter, r *http.Request) {
	state := googleauth.GenerateOauthStateString(w, r)
	url := cfg.GoogleConfig.AuthCodeURL(state)
	http.Redirect(w, r, url, http.StatusTemporaryRedirect)
}

func (cfg *apiConfig) handlerOAuthGoogleCallback(w http.ResponseWriter, r *http.Request) {
	oauthState, _ := r.Cookie("oauthstate")

	// the state from the cookie vs the state from the query parameters
	if r.FormValue("state") != oauthState.Value {
		log.Println("Invalid oauth google state")
		http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
		return
	}

	token, err := googleauth.ExchangeCodeForToken(context.Background(), cfg.GoogleConfig, r.FormValue("code"))
	if err != nil {
		log.Printf("Failed to exchange token %v\n", err)
		http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
		return
	}

	// getting http client from oauth.Token
	client := cfg.GoogleConfig.Client(context.Background(), token)
	userInfo, err := googleauth.GetGoogleUserInfo(client)
	if err != nil {
		log.Printf("Failed to get user info %v\n", err)
		http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
		return
	}

	user := &database.User{
		ID:    uuid.New().String(),
		Email: userInfo.Email,
	}

	err = googleauth.StoreTokenInDB(context.Background(), cfg.DB, user, token)
	if err != nil {
		log.Printf("failed to store token %v\n", err)
		http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
		return
	}
	log.Printf("Successfully authenticated user: %s", userInfo.Email)
	respondWithJSON(w, http.StatusOK, userInfo)
}
