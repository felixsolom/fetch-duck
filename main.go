package main

import (
	"database/sql"
	"embed"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/felixsolom/fetch-duck/internal/config"
	"github.com/felixsolom/fetch-duck/internal/database"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	_ "github.com/tursodatabase/libsql-client-go/libsql"
)

//go:embed all:static
var staticFiles embed.FS

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	port := os.Getenv("PORT")
	if port == "" {
		log.Fatal("PORT environment variable is not set")
	}

	db, err := sql.Open("libsql", cfg.DB.URL)
	if err != nil {
		log.Fatalf("Failed to open database connection: %v", err)
	}
	defer db.Close()

	dbQueries := database.New(db)
	fmt.Println("Configuration loaded and database connection established")
	fmt.Println("Google Client ID:", cfg.Google.ClientID)

	apiCfg := &apiConfig{
		DB:           dbQueries,
		GoogleConfig: cfg.Google.ToOAuth2Confg(),
	}

	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{"https://*", "http://*"},
		AllowedMethods:   []string{"GET", "PUT", "POST", "DELETE", "HEAD", "OPTION"},
		AllowedHeaders:   []string{"User-Agent", "Content-Type", "Accept", "Accept-Encoding", "Accept-Language", "Cache-Control", "Connection", "DNT", "Host", "Origin", "Pragma", "Referer"},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: true,
		MaxAge:           300,
	}))

	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		f, err := staticFiles.Open("static/index.html")
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		defer f.Close()
		if _, err := io.Copy(w, f); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	})

	v1Router := chi.NewRouter()

	v1Router.Get("/healthz", handlerReadiness)
	v1Router.Get("/oauth/google/login", apiCfg.handlerOAuthGoogleLogin)
	v1Router.Get("/oauth/google/callback", apiCfg.handlerOAuthGoogleCallback)
	v1Router.Get("/auth/success", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.Write([]byte(`<h1>Login Successful!</h1><p><a href="/v1/invoices/staged">Click here to view your staged invoices.</a></p>`))
	})

	authedRouter := chi.NewRouter()
	authedRouter.Use(apiCfg.authMiddleware)

	authedRouter.Get("/invoices/staged", apiCfg.handlerListStagedInvoices)
	authedRouter.Post("/invoices/{invoiceID}/approve", apiCfg.handlerApproveInvoice)
	authedRouter.Post("/invoices/{invoiceID}/reject", apiCfg.handlerRejectInvoice)

	v1Router.Mount("/", authedRouter)
	r.Mount("/v1", v1Router)
	srv := &http.Server{
		Addr:              ":" + port,
		Handler:           r,
		ReadHeaderTimeout: time.Second * 10,
	}

	log.Printf("Serving on port: %s\n", port)
	log.Fatal(srv.ListenAndServe())
}
