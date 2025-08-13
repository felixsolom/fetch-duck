package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"time"

	"github.com/felixsolom/fetch-duck/internal/database"
	"github.com/felixsolom/fetch-duck/internal/gmailservice"
	"github.com/go-chi/chi/v5"
	"golang.org/x/oauth2"
)

func (cfg *apiConfig) handlerListStagedInvoices(w http.ResponseWriter, r *http.Request) {
	user, ok := getUserFromContext(r)
	if !ok {
		respondWithError(w, http.StatusInternalServerError, "Could not get user from context", nil)
		return
	}
	log.Printf("Fetching staged invoices for user: %s", user.Email)

	limitStr := r.URL.Query().Get("limit")
	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit <= 0 {
		limit = 25
	}

	offsetStr := r.URL.Query().Get("offset")
	offset, err := strconv.Atoi(offsetStr)
	if err != nil || offset < 0 {
		offset = 0
	}

	now := time.Now()
	twoMonthAgo := now.AddDate(0, -2, 0)

	params := database.ListStagedInvoicesByUserParams{
		UserID:       user.ID,
		ReceivedAt:   twoMonthAgo.Unix(),
		ReceivedAt_2: now.Unix(),
		Limit:        int64(limit),
		Offset:       int64(offset),
	}

	invoices, err := cfg.DB.ListStagedInvoicesByUser(context.Background(), params)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Failed to list staged invoices", err)
		return
	}
	respondWithJSON(w, http.StatusOK, invoices)
}

func (cfg *apiConfig) handlerApproveInvoice(w http.ResponseWriter, r *http.Request) {
	invoiceID := chi.URLParam(r, "invoiceID")
	user, ok := getUserFromContext(r)
	if !ok {
		respondWithError(w, http.StatusInternalServerError, "Could not get user from context", nil)
		return
	}
	log.Printf("User %s is approving invoice %s", user.Email, invoiceID)

	//aws server logic for later

	//staged invoice details for db
	stagedInvoice, err := cfg.DB.GetStagedInvoice(r.Context(), database.GetStagedInvoiceParams{
		ID:     invoiceID,
		UserID: user.ID,
	})
	if err != nil {
		respondWithError(w, http.StatusNotFound, "Staged invoice not found", err)
		return
	}

	//google auth token from google_auths table
	dbAuth, err := cfg.DB.GetGoogleAuthByUserID(r.Context(), user.ID)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Couldn't get auth token for user", err)
		return
	}

	//auth client now
	token := &oauth2.Token{
		AccessToken:  dbAuth.AccessToken,
		RefreshToken: dbAuth.RefreshToken,
		Expiry:       time.Unix(dbAuth.TokenExpiry, 1000),
		TokenType:    "Bearer",
	}
	client := cfg.GoogleConfig.Client(context.Background(), token)
	gmailService, err := gmailservice.New(client)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Failed to create gmail service", err)
		return
	}

	//now the attachment
	attachmentData, filename, err := gmailService.GetFirstAttachment(stagedInvoice.GmailMessageID)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Failed to get attachment from gmail", err)
		return
	}

	tempDir := "temp_invoices"
	if err := os.MkdirAll(tempDir, 0755); err != nil {
		respondWithError(w, http.StatusInternalServerError, "Failed to create temp directory", err)
		return
	}

	savePath := filepath.Join(tempDir, fmt.Sprintf("approved-%s-%s", stagedInvoice.ID, filename))
	err = os.WriteFile(savePath, attachmentData, 0644)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Failed to save file", err)
		return
	}
	log.Printf("Successfully saved approved invoice to %s", savePath)

	err = cfg.DB.UpdateStagedInvoiceStatus(r.Context(), database.UpdateStagedInvoiceStatusParams{
		ID:        invoiceID,
		UserID:    user.ID,
		Status:    "approved",
		UpdatedAt: time.Now().Unix(),
	})
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Failed to approve invoice", err)
		return
	}
	respondWithJSON(w, http.StatusOK, map[string]string{
		"status":    "approved",
		"filename":  filename,
		"save_path": savePath,
	})
}

func (cfg *apiConfig) handlerRejectInvoice(w http.ResponseWriter, r *http.Request) {
	invoiceID := chi.URLParam(r, "invoiceID")
	user, ok := getUserFromContext(r)
	if !ok {
		respondWithError(w, http.StatusInternalServerError, "Could not get user from context", nil)
		return
	}
	log.Printf("User %s is rejecting invoice %s", user.Email, invoiceID)

	err := cfg.DB.UpdateStagedInvoiceStatus(r.Context(), database.UpdateStagedInvoiceStatusParams{
		ID:        invoiceID,
		UserID:    user.ID,
		Status:    "rejected",
		UpdatedAt: time.Now().Unix(),
	})
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Failed to reject invoice", err)
		return
	}
	respondWithJSON(w, http.StatusOK, map[string]string{"status": "rejected"})
}
