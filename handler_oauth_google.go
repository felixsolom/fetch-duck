package main

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"log"
	"net/http"
	"time"

	"github.com/felixsolom/fetch-duck/internal/database"
	"github.com/felixsolom/fetch-duck/internal/gmailservice"
	"github.com/felixsolom/fetch-duck/internal/googleauth"
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
		respondWithError(w, http.StatusInternalServerError, "Failed to get user info", err)
		return
	}

	userID, err := googleauth.StoreTokenInDB(context.Background(), cfg.DB, userInfo, token)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Failed to store token", err)
		return
	}

	log.Printf("Successfully authenticated user: %s", userInfo.Email)

	sessionTokenBytes := make([]byte, 32)
	rand.Read(sessionTokenBytes)
	sessionToken := hex.EncodeToString(sessionTokenBytes)

	expiry := time.Now().Add(24 * time.Hour)

	_, err = cfg.DB.CreateSession(context.Background(), database.CreateSessionParams{
		Token:     sessionToken,
		UserID:    userID,
		Expiry:    expiry.Unix(),
		CreatedAt: time.Now().Unix(),
	})
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Failed to create session", err)
		return
	}

	http.SetCookie(w, &http.Cookie{
		Name:     "session_token",
		Value:    sessionToken,
		Expires:  expiry,
		HttpOnly: true,
		Path:     "/",
		SameSite: http.SameSiteLaxMode,
	})

	go func() {
		log.Println("Authentication successfull. Starting background scan and stage process...")
		gmailService, err := gmailservice.New(client)
		if err != nil {
			log.Printf("Error creating gmail service: %v", err)
			return
		}
		err = gmailService.ScanAndStageInvoices(context.Background(), cfg.DB, userID)
		if err != nil {
			log.Printf("Error scanning and staging invoices: %v", err)
			return
		}
	}()

	http.Redirect(w, r, "/v1/invoices/staged", http.StatusSeeOther)
}
