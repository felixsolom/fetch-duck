package main

import (
	"context"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/felixsolom/fetch-duck/internal/database"
	"github.com/go-chi/chi/v5"
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

	err := cfg.DB.UpdateStagedInvoiceStatus(r.Context(), database.UpdateStagedInvoiceStatusParams{
		ID:        invoiceID,
		UserID:    user.ID,
		Status:    "approved",
		UpdatedAt: time.Now().Unix(),
	})
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Failed to approve invoice", err)
		return
	}
	respondWithJSON(w, http.StatusOK, map[string]string{"status": "approved"})
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
