package config

import (
	"os"

	"github.com/joho/godotenv"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
)

type GoogleConfig struct {
	ClientID     string
	ClientSecret string
	RedirectURL  string
	Scopes       []string
}

type Config struct {
	Google GoogleConfig
}

func Load() (*Config, error) {
	err := godotenv.Load()
	if err != nil {
		return nil, err
	}
	cfg := &Config{
		Google: GoogleConfig{
			ClientID:     os.Getenv("GOOGLE_CLIENT_ID"),
			ClientSecret: os.Getenv("GOOGLE_CLIENT_SECRET"),
			RedirectURL:  "http://localhost:8080",
			Scopes: []string{
				"https://www.googleapis.com/auth/gmail.readonly",
			},
		},
	}
	return cfg, nil
}

func (g *GoogleConfig) ToOAuth2Confg() *oauth2.Config {
	return &oauth2.Config{
		ClientID:     g.ClientID,
		ClientSecret: g.ClientSecret,
		RedirectURL:  g.RedirectURL,
		Scopes:       g.Scopes,
		Endpoint:     google.Endpoint,
	}
}
