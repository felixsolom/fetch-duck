package config

import (
	"log"
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

type DBConfig struct {
	URL string
}

type AppConfig struct {
	InviteCode string
}

type AWSConfig struct {
	AccessKeyID     string
	SecretAccessKey string
	Region          string
	BucketName      string
}

type AccountingConfig struct {
	APIKey    string
	APISecret string
	BaseURL   string
}

type Config struct {
	Google     GoogleConfig
	DB         DBConfig
	App        AppConfig
	AWS        AWSConfig
	Accounting AccountingConfig
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
			RedirectURL:  "http://localhost:8080/api/v1/oauth/google/callback",
			Scopes: []string{
				"https://www.googleapis.com/auth/gmail.readonly",
				"https://www.googleapis.com/auth/userinfo.email",
			},
		},
		DB: DBConfig{
			URL: os.Getenv("DATABASE_URL"),
		},
		App: AppConfig{
			InviteCode: os.Getenv("APP_SECRET_INVITE_CODE"),
		},
		AWS: AWSConfig{
			AccessKeyID:     os.Getenv("AWS_ACCESS_KEY_ID"),
			SecretAccessKey: os.Getenv("AWS_SECRET_ACCESS_KEY"),
			Region:          os.Getenv("S3_REGION"),
			BucketName:      os.Getenv("S3_BUCKET"),
		},
		Accounting: AccountingConfig{
			APIKey:    os.Getenv("GREEN_INVOICE_API_KEY"),
			APISecret: os.Getenv("GREEN_INVOICE_API_SECRET"),
			BaseURL:   os.Getenv("GREEN_INVOICE_BASE_URL"),
		},
	}

	if cfg.App.InviteCode == "" {
		log.Fatal("CRITICAL: APP_SECRET_INVITE_CODE environment variable is not set")
	}

	if cfg.AWS.AccessKeyID == "" || cfg.AWS.SecretAccessKey == "" || cfg.AWS.BucketName == "" {
		log.Fatal("CRITICAL: AWS configuration (bucket name, keys) is not fully set")
	}

	if cfg.Accounting.APIKey == "" || cfg.Accounting.APISecret == "" || cfg.Accounting.BaseURL == "" {
		log.Fatal("CRITICAL: Accounting credentials (API key, secret, base url) are not fully set")
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
