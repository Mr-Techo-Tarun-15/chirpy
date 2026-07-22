package main

import (
	"net/http"

	"github.com/Mr-Techo-Tarun-15/chirpy/internal/auth"
	"github.com/google/uuid"
)

func (cfg *apiConfig) deleteChirpFromIDHandler(w http.ResponseWriter, r *http.Request) {
	stringChirpID := r.PathValue("chirpID")
	properChirpID, err := uuid.Parse(stringChirpID)
	if err != nil {
		errorCase(w, 404, "cannot parse given ID", err.Error())
	}
	chirp, err := cfg.databaseQueries.GetChirpFromID(r.Context(), properChirpID)
	if err != nil {
		errorCase(w, 404, "provided id does not match any Chirps", err.Error())
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

	if userIDFromToken != chirp.UserID.UUID {
		errorCase(w, 403, "You are not authorized to delete this Chirp", "your ID doesn't match the ID of the user who made this Chirp")
		return
	}

	err = cfg.databaseQueries.DeleteChripWithID(r.Context(), chirp.ID)
	if err != nil {
		errorCase(w, 404, "Chirp couldn't be deleted with the provided ID", err.Error())
		return
	}

	w.WriteHeader(204)
}
