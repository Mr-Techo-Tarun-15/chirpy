package main

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/Mr-Techo-Tarun-15/chirpy/internal/auth"
	"github.com/Mr-Techo-Tarun-15/chirpy/internal/database"
	"github.com/google/uuid"
)

func (cfg *apiConfig) updateUsersHandler(w http.ResponseWriter, r *http.Request) {
	type inputParams struct {
		Password string `json:"password"`
		Email    string `json:"email"`
	}

	type responseParams struct {
		ID             uuid.UUID `json:"id"`
		CreatedAt      time.Time `json:"created_at"`
		UpdatedAt      time.Time `json:"updated_at"`
		Email          string    `json:"email"`
		HashedPassword string    `json:"hashed_password"`
		IsChirpyRed    bool      `json:"is_chirpy_red"`
	}

	var params inputParams
	if err := json.NewDecoder(r.Body).Decode(&params); err != nil {
		errorCase(w, http.StatusBadRequest, "Something went wrong", err.Error())
		return
	}

	token, err := auth.GetBearerToken((*r).Header)
	if err != nil {
		errorCase(w, 401, "couldn't get token", err.Error())
		return
	}

	userIDFromToken, err := auth.ValidateJWT(token, cfg.secretForJWT)
	if err != nil {
		errorCase(w, 401, "couldn't validate token", err.Error())
		return
	}

	hashedPassword, err := auth.HashPassword(params.Password)
	if err != nil {
		errorCase(w, http.StatusBadRequest, "couldn't hash password", err.Error())
		return
	}

	err = cfg.databaseQueries.UpdateUserEmailAndPassword(
		r.Context(),
		database.UpdateUserEmailAndPasswordParams{
			Email:          params.Email,
			HashedPassword: hashedPassword,
			ID:             userIDFromToken,
			UpdatedAt:      time.Now(),
		},
	)
	if err != nil {
		errorCase(w, http.StatusBadRequest, "couldn't update user's email and password", err.Error())
		return
	}

	user, err := cfg.databaseQueries.GetUserFromEmail(r.Context(), params.Email)
	if err != nil {
		errorCase(w, 401, "couldn't get user with new email", err.Error())
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(responseParams{
		ID:             user.ID,
		CreatedAt:      user.CreatedAt,
		UpdatedAt:      user.UpdatedAt,
		Email:          user.Email,
		HashedPassword: user.HashedPassword,
	})
}
