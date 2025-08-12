package googleauth

import (
	"context"
	"crypto/rand"
	"database/sql"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"

	"github.com/felixsolom/fetch-duck/internal/database"
	"github.com/google/uuid"
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

func StoreTokenInDB(ctx context.Context, db *database.Queries, userInfo *GoogleUserInfo, token *oauth2.Token) (string, error) {
	now := time.Now().Unix()

	existingUser, err := db.GetUser(ctx, userInfo.Email)
	if err != nil {
		if err == sql.ErrNoRows {
			newUserId := uuid.New().String()
			log.Printf("User %s not found. Creating new user with id %s", userInfo.Email, newUserId)
			err = db.CreateUser(ctx, database.CreateUserParams{
				ID:        newUserId,
				Email:     userInfo.Email,
				CreatedAt: now,
				UpdatedAt: now,
			})
			if err != nil {
				return "", fmt.Errorf("failed to create user: %w", err)
			}
			existingUser, err = db.GetUser(ctx, userInfo.Email)
			if err != nil {
				return "", fmt.Errorf("failed to get user after creation: %w", err)
			}
		} else {
			return "", fmt.Errorf("failed to get user: %w", err)
		}
	}

	params := database.UpsertGoogleAuthParams{
		UserID:       existingUser.ID,
		AccessToken:  token.AccessToken,
		RefreshToken: token.RefreshToken,
		TokenExpiry:  token.Expiry.Unix(),
		CreatedAt:    now,
		UpdatedAt:    now,
	}

	err = db.UpsertGoogleAuth(ctx, params)
	if err != nil {
		return "", fmt.Errorf("failed to upsert token: %w", err)
	}
	return existingUser.ID, nil
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
