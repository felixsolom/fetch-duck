package main

import (
	"database/sql"
	"embed"
	"fmt"
	"io/fs"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/felixsolom/fetch-duck/internal/config"
	"github.com/felixsolom/fetch-duck/internal/database"
	"github.com/felixsolom/fetch-duck/internal/s3service"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	_ "github.com/tursodatabase/libsql-client-go/libsql"
	"golang.org/x/oauth2"
)

//go:embed all:static
var staticFiles embed.FS

type apiConfig struct {
	DB           *database.Queries
	GoogleConfig *oauth2.Config
	App          config.AppConfig
	S3           *s3service.Service
}

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	s3Svc, err := s3service.New(cfg.AWS)
	if err != nil {
		log.Fatalf("Failed to create s3 service: %v", err)
	}
	log.Println("S3 services initialized successfully.")

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
		App:          cfg.App,
		S3:           s3Svc,
	}

	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{"https://*", "http://*"},
		AllowedMethods:   []string{"GET", "PUT", "POST", "DELETE", "HEAD", "OPTION"},
		AllowedHeaders:   []string{"*"},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: true,
		MaxAge:           300,
	}))

	apiRouter := chi.NewRouter()

	apiRouter.Get("/healthz", handlerReadiness)
	apiRouter.Post("/auth/verify-invite", apiCfg.handlerVerifyInvite)
	apiRouter.Get("/oauth/google/login", apiCfg.handlerOAuthGoogleLogin)
	apiRouter.Get("/oauth/google/callback", apiCfg.handlerOAuthGoogleCallback)
	apiRouter.Get("/auth/success", func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "/", http.StatusSeeOther)
	})

	apiRouter.Group(func(authedRouter chi.Router) {
		authedRouter.Use(apiCfg.authMiddleware)
		authedRouter.Get("/auth/status", apiCfg.handlerAuthStatus)
		authedRouter.Post("/auth/logout", apiCfg.handlerLogout)
		authedRouter.Get("/invoices/staged", apiCfg.handlerListStagedInvoices)
		authedRouter.Post("/invoices/{invoiceID}/approve", apiCfg.handlerApproveInvoice)
		authedRouter.Post("/invoices/{invoiceID}/reject", apiCfg.handlerRejectInvoice)
	})

	r.Mount("/api/v1", apiRouter)

	staticFS, err := fs.Sub(staticFiles, "static")
	if err != nil {
		log.Fatalf("Couldn't serve static files: %s", err)
	}
	r.Handle("/*", http.FileServer(http.FS(staticFS)))

	srv := &http.Server{
		Addr:              ":" + port,
		Handler:           r,
		ReadHeaderTimeout: time.Second * 10,
	}

	log.Printf("Serving on port: %s\n", port)
	log.Fatal(srv.ListenAndServe())
}
