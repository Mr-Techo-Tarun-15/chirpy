package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/Mr-Techo-Tarun-15/chirpy/internal/auth"
	"github.com/Mr-Techo-Tarun-15/chirpy/internal/database"
	"github.com/google/uuid"
)

func (cfg *apiConfig) resetHandler(w http.ResponseWriter, r *http.Request) {
	if cfg.currentPlatform != "dev" {
		fmt.Println("you do not have permission to use this command")
		fmt.Println("403 forbidden")
	}

	cfg.fileserverHits.Swap(0)
	err := cfg.databaseQueries.ClearUsers(r.Context())
	if err != nil {
		errorCase(w, http.StatusBadRequest, "couldn't clear the database", err.Error())
		return
	}
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusOK)
}

func (cfg *apiConfig) createUsersHandler(w http.ResponseWriter, r *http.Request) {
	type inputParams struct {
		Email    string `json:"email"`
		Password string `json:"password"`
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

	if strings.TrimSpace(params.Email) == "" {
		errorCase(w, http.StatusBadRequest, "email must not be empty", "empty emails are not usable")
		return
	}

	hashedPassword, err := auth.HashPassword(params.Password)
	if err != nil {
		errorCase(w, http.StatusBadRequest, "couldn't create password", err.Error())
		return
	}

	user, err := cfg.databaseQueries.CreateUser(
		r.Context(),
		database.CreateUserParams{
			ID:        uuid.New(),
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
			Email:     params.Email,
		},
	)
	if err != nil {
		errorCase(w, http.StatusBadRequest, "couln't create user", err.Error())
		return
	}

	err = cfg.databaseQueries.StorePassword(
		r.Context(),
		database.StorePasswordParams{
			Email:          params.Email,
			HashedPassword: hashedPassword,
		},
	)
	if err != nil {
		errorCase(w, http.StatusBadRequest, "couldn't store password", err.Error())
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(responseParams{
		ID:        user.ID,
		CreatedAt: user.CreatedAt,
		UpdatedAt: user.UpdatedAt,
		Email:     user.Email,
	})
}

func (cfg *apiConfig) postChirpsHandler(w http.ResponseWriter, r *http.Request) {
	type inputParams struct {
		Body string `json:"body"`
	}

	type responseParams struct {
		ID        uuid.UUID `json:"id"`
		CreatedAt time.Time `json:"created_at"`
		UpdatedAt time.Time `json:"updated_at"`
		Body      string    `json:"body"`
		UserID    uuid.UUID `json:"user_id"`
		Token     string    `json:"token"`
	}

	var params inputParams
	if err := json.NewDecoder(r.Body).Decode(&params); err != nil {
		errorCase(w, http.StatusBadRequest, "Something went wrong", err.Error())
		return
	}

	if strings.TrimSpace(params.Body) == "" {
		errorCase(w, http.StatusBadRequest, "body must not be empty", "empty body is not usable")
		return
	}

	if len(params.Body) > 140 {
		errorCase(w, http.StatusBadRequest, "Chirp is too long", "maximum processing length is 140 characters")
		return
	}

	token, err := auth.GetBearerToken((*r).Header)
	if err != nil {
		errorCase(w, http.StatusBadRequest, "couldn't get token", err.Error())
		return
	}

	userIDFromToken, err := auth.ValidateJWT(token, cfg.secretForJWT)
	if err != nil {
		errorCase(w, 401, "couldn't validate token", err.Error())
		return
	}

	splitParamsBody := strings.Split(params.Body, " ")
	splitFilterdParamsBody := []string{}
	var filterdParamsBody string = ""
	for _, word := range splitParamsBody {
		lowerWord := strings.ToLower(word)
		if lowerWord == "kerfuffle" || lowerWord == "sharbert" || lowerWord == "fornax" {
			splitFilterdParamsBody = append(splitFilterdParamsBody, "****")
		} else {
			splitFilterdParamsBody = append(splitFilterdParamsBody, word)
		}
	}
	filterdParamsBody = strings.Join(splitFilterdParamsBody, " ")

	storedChirp, err := cfg.databaseQueries.NewChirp(
		r.Context(),
		database.NewChirpParams{
			ID:        uuid.New(),
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
			Body:      filterdParamsBody,
			UserID:    uuid.NullUUID{UUID: userIDFromToken, Valid: true},
		},
	)
	if err != nil {
		errorCase(w, 401, "couldn't make a new Chirp", err.Error())
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(responseParams{
		ID:        storedChirp.ID,
		CreatedAt: storedChirp.CreatedAt,
		UpdatedAt: storedChirp.UpdatedAt,
		Body:      storedChirp.Body,
		UserID:    userIDFromToken,
		Token:     token,
	})
}

func (cfg *apiConfig) loginHandler(w http.ResponseWriter, r *http.Request) {
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
		AccessToken    string    `json:"token"`
		RefreshToken   string    `json:"refresh_token"`
	}

	var params inputParams
	if err := json.NewDecoder(r.Body).Decode(&params); err != nil {
		errorCase(w, http.StatusBadRequest, "Something went wrong", err.Error())
		return
	}

	userFromEmail, err := cfg.databaseQueries.GetUserFromEmail(r.Context(), params.Email)
	if err != nil {
		errorCase(w, 401, "Incorrect email or password", err.Error())
		return
	}

	pswrdMatches, err := auth.CheckPasswordHash(params.Password, userFromEmail.HashedPassword)
	if err != nil {
		errorCase(w, 401, "Incorrect email or password", err.Error())
		return
	}

	if !pswrdMatches {
		errorCase(w, 401, "Incorrect email or password", "access denied")
		return
	}

	newJWT, err := auth.MakeJWT(userFromEmail.ID, cfg.secretForJWT, time.Duration(time.Hour))
	if err != nil {
		errorCase(w, http.StatusBadRequest, "couldn't create a new JWT for users", err.Error())
		return
	}

	newRefreshToken := auth.MakeRefreshToken()
	err = cfg.databaseQueries.AddRefreshToken(
		r.Context(),
		database.AddRefreshTokenParams{
			Token:     newRefreshToken,
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
			UserID:    uuid.NullUUID{UUID: userFromEmail.ID, Valid: true},
			ExpiresAt: time.Now().AddDate(0, 0, 60),
		},
	)
	if err != nil {
		errorCase(w, http.StatusBadRequest, "cannot add new refresh token to database", err.Error())
		return
	}

	if pswrdMatches {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(responseParams{
			ID:             userFromEmail.ID,
			CreatedAt:      userFromEmail.CreatedAt,
			UpdatedAt:      userFromEmail.UpdatedAt,
			Email:          userFromEmail.Email,
			HashedPassword: userFromEmail.HashedPassword,
			IsChirpyRed:    userFromEmail.IsChirpyRed,
			AccessToken:    newJWT,
			RefreshToken:   newRefreshToken,
		})
	}
}

func (cfg *apiConfig) refreshHandler(w http.ResponseWriter, r *http.Request) {
	type responseParams struct {
		Token string `json:"token"`
	}

	refreshBearerToken, err := auth.GetBearerToken((*r).Header)
	if err != nil {
		errorCase(w, http.StatusBadRequest, "couldn't get refresh token from header", err.Error())
		return
	}

	refreshToken, err := cfg.databaseQueries.GetRefreshTokenDataFromToken(r.Context(), refreshBearerToken)
	if err != nil {
		errorCase(w, 401, "couldn't get refresh token data from database", err.Error())
		return
	}

	if !refreshToken.UserID.Valid || refreshToken.UserID.UUID == uuid.Nil {
		errorCase(w, 401, "refresh token is invalid", "please log in using your password")
		return
	}

	if time.Now().After(refreshToken.ExpiresAt) {
		errorCase(w, 401, "refresh token expired", "please log in using your password")
		return
	}

	if refreshToken.RevokedAt.Valid {
		errorCase(w, 401, "refresh token has been revoked", "please log in using your password")
		return
	}

	newAccessToken, err := auth.MakeJWT(refreshToken.UserID.UUID, cfg.secretForJWT, time.Hour)
	if err != nil {
		errorCase(w, http.StatusInternalServerError, "couldn't create a new access token", err.Error())
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(responseParams{Token: newAccessToken})
}

func (cfg *apiConfig) revokeHandler(w http.ResponseWriter, r *http.Request) {
	token, err := auth.GetBearerToken((*r).Header)
	if err != nil {
		errorCase(w, http.StatusBadRequest, "couldn't get refresh token from header", err.Error())
		return
	}

	refreshToken, err := cfg.databaseQueries.GetRefreshTokenDataFromToken(r.Context(), token)
	if err != nil {
		errorCase(w, 401, "couldn't get refresh token data from database", err.Error())
		return
	}

	err = cfg.databaseQueries.UpdateRefreshTokenRevoke(
		r.Context(),
		database.UpdateRefreshTokenRevokeParams{
			RevokedAt: sql.NullTime{Time: time.Now(), Valid: true},
			Token:     refreshToken.Token,
		},
	)
	if err != nil {
		errorCase(w, 401, "couldn't update revoked_at", err.Error())
	}
	cfg.databaseQueries.UpdateRefreshTokenUpdate(
		r.Context(),
		database.UpdateRefreshTokenUpdateParams{
			UpdatedAt: time.Now(),
			Token:     refreshToken.Token,
		},
	)
	if err != nil {
		errorCase(w, 401, "couldn't update updated_at", err.Error())
	}

	w.WriteHeader(204)
}

func (cfg *apiConfig) webhooksHandler(w http.ResponseWriter, r *http.Request) {
	type inputParams struct {
		Event string `json:"event"`
		Data  struct {
			UserID uuid.UUID `json:"user_id"`
		} `json:"data"`
	}

	APIKey, err := auth.GetAPIKey(r.Header)
	if err != nil {
		errorCase(w, 401, "couldn't get API Key", err.Error())
		return
	}

	if APIKey != cfg.polkaKey {
		errorCase(w, 401, "you do not have permission", "")
		return
	}

	var params inputParams
	if err := json.NewDecoder(r.Body).Decode(&params); err != nil {
		errorCase(w, 404, "Something went wrong", err.Error())
		return
	}

	if params.Event != "user.upgraded" {
		errorCase(w, 204, "no events other than 'user.upgraded'", "'event' should be 'user.upgraded'")
		return
	}

	if params.Event == "user.upgraded" {
		err := cfg.databaseQueries.UpgradeUserToChirpyRedFromID(r.Context(), params.Data.UserID)
		if err != nil {
			errorCase(w, 404, "user couldn't be found with the provided ID", err.Error())
		}

		w.WriteHeader(204)
	}

}
