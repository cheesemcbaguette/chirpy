package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"example.com/chirpy/internal/auth"
	"github.com/google/uuid"
	"net/http"
	"os"
)

func (cfg *apiConfig) handlerWebhooks(w http.ResponseWriter, r *http.Request) {
	type PolkaWebhookRequest struct {
		Event string `json:"event"`
		Data  struct {
			UserID uuid.UUID `json:"user_id"`
		} `json:"data"`
	}

	var req PolkaWebhookRequest

	// Extract the API key from the header
	apiKey, err := auth.GetAPIKey(r.Header)
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Compare the extracted key with the one from the environment
	if apiKey != os.Getenv("POLKA_KEY") {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Bad Request", http.StatusBadRequest)
		return
	}

	// Only handle user.upgraded event
	if req.Event != "user.upgraded" {
		w.WriteHeader(http.StatusNoContent)
		return
	}

	// Get user by ID
	_, err = cfg.db.GetUserByID(context.Background(), req.Data.UserID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			respondWithError(w, http.StatusNotFound, "User not found", err)
		} else {
			respondWithError(w, http.StatusInternalServerError, "Failed to retrieve user", err)
		}
		return
	}

	// Upgrade user to Chirpy Red
	err = cfg.db.UpgradeUserToChirpRed(context.Background(), req.Data.UserID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) { // Handle case where user doesn't exist
			respondWithError(w, http.StatusNotFound, "User not found", err)
			return
		}
		respondWithError(w, http.StatusInternalServerError, "Failed to upgrade chirp to red", err)
		return
	}

	// Respond with 204 No Content on success
	w.WriteHeader(http.StatusNoContent)
}
