package main

import (
	"fmt"
	"log"

	"github.com/felixsolom/fetch-duck/internal/config"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}
	fmt.Println("Configuration loaded successfully")
	fmt.Println("Google Client ID:", cfg.Google.ClientID)

	oauth2config := cfg.Google.ToOAuth2Confg()
	fmt.Println("OAuth2 endpoint URL:", oauth2config.Endpoint.AuthURL)
}
