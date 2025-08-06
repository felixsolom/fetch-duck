package googleauth

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"

	"github.com/felixsolom/fetch-duck/internal/database"
	"golang.org/x/oauth2"
)

type GoogleUserInfo struct {
	ID    string `json:"id"`
	Email string `json:"email"`
}

func ExchangeCodeForToken(ctx context.Context, config *oauth2.Config, code string) (*oauth2.Token, error) {
	return config.Exchange(ctx, code)
}

func GetGoogleUserInfo(client *http.Client) (*GoogleUserInfo, error) {
	resp, err := client.Get("https://www.googleapis.com/oauth2/v2/userinfo")
	if err != nil {
		return nil, fmt.Errorf("failed to get user info: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read user info response body: %w", err)
	}

	var userInfo GoogleUserInfo
	if err = json.Unmarshal(body, &userInfo); err != nil {
		return nil, fmt.Errorf("failed to unmarshal user info: %w", err)
	}
	return &userInfo, nil
}

func StoreTokenInDB(ctx context.Context, db *database.Queries, user *database.User, token *oauth2.Token) error {
	now := time.Now().Unix()

	_, err := db.GetUser(ctx, user.Email)
	if err != nil {
		err = db.CreateUser(ctx, database.CreateUserParams{
			ID:        user.ID,
			Email:     user.Email,
			CreatedAt: now,
			UpdatedAt: now,
		})
		if err != nil {
			return fmt.Errorf("failed to create user: %w", err)
		}
		log.Printf("Created new user: %s", user.Email)
	}

	params := database.UpsertGoogleAuthParams{
		UserID:       user.ID,
		AccessToken:  token.AccessToken,
		RefreshToken: token.RefreshToken,
		TokenExpiry:  token.Expiry.Unix(),
		CreatedAt:    now,
		UpdatedAt:    now,
	}
	log.Printf("Upserting token for user: %s", user.Email)
	return db.UpsertGoogleAuth(ctx, params)
}

func GenerateOauthStateString(w http.ResponseWriter, r *http.Request) string {
	b := make([]byte, 16)
	rand.Read(b)
	state := base64.URLEncoding.EncodeToString(b)

	http.SetCookie(w, &http.Cookie{
		Name:     "oauthstate",
		Value:    state,
		Expires:  time.Now().Add(10 * time.Minute),
		HttpOnly: true,
	})
	return state
}
