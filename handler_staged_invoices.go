package main

import (
	"context"
	"net/http"
)

func (cfg *apiConfig) handlerListStagedInvoices(w http.ResponseWriter, r *http.Request) {
	users, err := cfg.DB.GetUsers(context.Background())
	if err != nil || len(users) == 0 {
		respondWithError(w, http.StatusNotFound, "No users found in the system", err)
		return
	}

}
