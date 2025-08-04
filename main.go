package main

import (
	"database/sql"
	"fmt"
	"log"

	"github.com/felixsolom/fetch-duck/internal/config"
	"github.com/felixsolom/fetch-duck/internal/database"
	_ "github.com/tursodatabase/libsql-client-go/libsql"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	db, err := sql.Open("libsql", cfg.DB.URL)
	if err != nil {
		log.Fatalf("Failed to open database connection: %v", err)
	}
	defer db.Close()

	dbQueries := database.New(db)
	fmt.Println("Configuration loaded and database connection established")
	fmt.Println("Google Client ID:", cfg.Google.ClientID)

	oauth2config := cfg.Google.ToOAuth2Confg()
	fmt.Println("OAuth2 endpoint URL:", oauth2config.Endpoint.AuthURL)
	_ = dbQueries
}
