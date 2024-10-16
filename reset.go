package main

import (
	"net/http"
)

func (cfg *apiConfig) handlerReset(w http.ResponseWriter, r *http.Request) {
	if cfg.platform != "dev" {
		respondWithError(w, http.StatusForbidden, "This endpoint is only available in development", nil)
		return
	}

	// Reset the users in the database
	err := cfg.db.DeleteAllUsers(r.Context())
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Could not reset users", nil)
		return
	}

	respondWithJSON(w, http.StatusOK, map[string]string{"message": "All users have been deleted"})
}
