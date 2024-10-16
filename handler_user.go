package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"example.com/chirpy/internal/auth"
	"example.com/chirpy/internal/database"
	"github.com/google/uuid"
	"net/http"
	"time"
)

type User struct {
	ID           uuid.UUID `json:"id"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
	Email        string    `json:"email"`
	Token        string    `json:"token"`
	RefreshToken string    `json:"refresh_token"`
}

// UserCreationParams holds the request parameters for user creation.
type UserParams struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

func (cfg *apiConfig) handlerCreateUser(w http.ResponseWriter, r *http.Request) {
	var params UserParams
	err := json.NewDecoder(r.Body).Decode(&params)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Couldn't decode parameters", err)
		return
	}

	// Hash the password before saving it to the database
	hashedPassword, err := auth.HashPassword(params.Password)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Failed to hash password", err)
		return
	}

	// Insert the new user into the database
	user, err := cfg.db.CreateUser(context.Background(), database.CreateUserParams{
		Email:          params.Email,
		HashedPassword: hashedPassword,
	})
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Failed to create user", err)
		return
	}

	// Don't return the hashed password in the response
	respondWithJSON(w, http.StatusCreated, User{
		ID:        user.ID,
		CreatedAt: user.CreatedAt,
		UpdatedAt: user.UpdatedAt,
		Email:     user.Email,
	})
}

func (cfg *apiConfig) handlerLoginUser(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Password string `json:"password"`
		Email    string `json:"email"`
	}
	err := json.NewDecoder(r.Body).Decode(&body)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid request payload", err)
		return
	}

	// Fetch the user by email
	user, err := cfg.db.GetUserByEmail(context.Background(), body.Email)
	if err != nil || auth.CheckPasswordHash(body.Password, user.HashedPassword) != nil {
		respondWithError(w, 401, "Incorrect email or password", err)
		return
	}

	// Create access token
	token, err := auth.MakeJWT(user.ID, cfg.jwtSecret, time.Hour)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Could not generate token", err)
		return
	}

	// Create refresh token
	refreshToken, err := auth.MakeRefreshToken()
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Could not generate refresh token", err)
		return
	}

	// Store refresh token in the database
	expiresAt := time.Now().Add(60 * 24 * time.Hour) // 60 days
	err = cfg.db.CreateRefreshToken(context.Background(), database.CreateRefreshTokenParams{
		Token:     refreshToken,
		UserID:    user.ID,
		ExpiresAt: expiresAt,
		RevokedAt: sql.NullTime{},
	})

	// Password matched, return user data (without hashed password)
	respondWithJSON(w, http.StatusOK, User{
		ID:           user.ID,
		CreatedAt:    user.CreatedAt,
		UpdatedAt:    user.UpdatedAt,
		Email:        user.Email,
		Token:        token,
		RefreshToken: refreshToken,
	})
}

func (cfg *apiConfig) handlerRefresh(w http.ResponseWriter, r *http.Request) {
	// Extract the bearer token from the header
	refreshToken, err := auth.GetBearerToken(r.Header)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Invalid token", err)
		return
	}

	// Lookup the refresh token in the database
	storedToken, err := cfg.db.GetRefreshToken(context.Background(), refreshToken)
	if err != nil || storedToken.RevokedAt.Valid || storedToken.ExpiresAt.Before(time.Now()) {
		respondWithError(w, http.StatusUnauthorized, "Invalid or expired refresh token", err)
		return
	}

	// Generate a new access token
	token, err := auth.MakeJWT(storedToken.UserID, cfg.jwtSecret, time.Hour)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Could not generate new token", err)
		return
	}

	// Respond with the new access token
	response := map[string]string{
		"token": token,
	}

	respondWithJSON(w, http.StatusOK, response)
}

func (cfg *apiConfig) handlerRevoke(w http.ResponseWriter, r *http.Request) {
	// Extract the bearer token from the header
	refreshToken, err := auth.GetBearerToken(r.Header)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Invalid token", err)
		return
	}

	// Revoke the refresh token in the database
	err = cfg.db.RevokeRefreshToken(context.Background(), refreshToken)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Could not revoke token", err)
		return
	}

	// Respond with a 204 No Content status
	w.WriteHeader(http.StatusNoContent)
}
