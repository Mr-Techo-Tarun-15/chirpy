package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"sort"
	"time"

	"github.com/google/uuid"
)

type responseParams struct {
	ID        uuid.UUID `json:"id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	Body      string    `json:"body"`
	UserID    uuid.UUID `json:"user_id"`
}

func (cfg *apiConfig) metricsHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(
		w,
		"<html><body><h1>Welcome, Chirpy Admin</h1><p>Chirpy has been visited %d times!</p></body></html>",
		cfg.fileserverHits.Load(),
	)
}

func (cfg *apiConfig) getChirpsHandler(w http.ResponseWriter, r *http.Request) {
	sortMethod := "asc"
	sortMethod = r.URL.Query().Get("sort")

	chirps, err := cfg.databaseQueries.GetAllChirpsOrderedEarliest(r.Context())
	if err != nil {
		errorCase(w, http.StatusBadRequest, "couldn't retrieve Chirps from database", err.Error())
		return
	}

	if sortMethod == "desc" {
		sort.Slice(chirps, func(i int, j int) bool { return chirps[i].CreatedAt.After(chirps[j].CreatedAt) })
	}

	responseChirps := make([]responseParams, 0, len(chirps))

	authorID := r.URL.Query().Get("author_id")
	if authorID != "" {
		for _, chirp := range chirps {
			if chirp.UserID.UUID.String() == authorID {
				responseChirps = append(responseChirps, responseParams{
					ID:        chirp.ID,
					CreatedAt: chirp.CreatedAt,
					UpdatedAt: chirp.UpdatedAt,
					Body:      chirp.Body,
					UserID:    chirp.UserID.UUID,
				})
			}
		}
	} else {
		for _, chirp := range chirps {
			responseChirps = append(responseChirps, responseParams{
				ID:        chirp.ID,
				CreatedAt: chirp.CreatedAt,
				UpdatedAt: chirp.UpdatedAt,
				Body:      chirp.Body,
				UserID:    chirp.UserID.UUID,
			})
		}
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(responseChirps)
}

func (cfg *apiConfig) getChirpFromIDHandler(w http.ResponseWriter, r *http.Request) {
	chirpID := r.PathValue("chirpID")
	parsedChirpID, err := uuid.Parse(chirpID)
	if err != nil {
		errorCase(w, http.StatusBadRequest, "error parsing the id", err.Error())
		return
	}

	chirp, err := cfg.databaseQueries.GetChirpFromID(r.Context(), parsedChirpID)
	if err != nil {
		errorCase(w, 404, "Chirp was not found at this id", err.Error())
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(responseParams{
		ID:        chirp.ID,
		CreatedAt: chirp.CreatedAt,
		UpdatedAt: chirp.UpdatedAt,
		Body:      chirp.Body,
		UserID:    chirp.UserID.UUID,
	})
}
