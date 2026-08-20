package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"slices"
	"strings"
	"sync/atomic"
	"time"

	"github.com/google/uuid"

	"github.com/Drag0neUsz/Chirpy/internal/database"
	"github.com/lib/pq"
)

type apiConfig struct {
	fileserverHits atomic.Int32
	dbQueries      *database.Queries
	platform       string
}

func (cfg *apiConfig) middlewareMetricsInc(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cfg.fileserverHits.Add(1)
		next.ServeHTTP(w, r)
	})
}

func (cfg *apiConfig) handleMetrics(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "text/html")
	w.Write([]byte(fmt.Sprintf("<html><body><h1>Welcome, Chirpy Admin</h1><p>Chirpy has been visited %d times!</p></body></html>", cfg.fileserverHits.Load())))
}

func (cfg *apiConfig) handleReset(w http.ResponseWriter, r *http.Request) {
	if cfg.platform == "dev" {
		cfg.fileserverHits.Store(0)
		cfg.dbQueries.ResetUsers(r.Context())
		respondWithJSON(w, http.StatusOK, map[string]string{"status": "success"})
	} else {
		respondWithError(w, http.StatusForbidden, "Forbidden")
	}
}

func handleHealthz(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "text/plain")
	w.Write([]byte("OK\n"))
}

type Chirp struct {
	Body      string    `json:"body"`
	UserID    uuid.UUID `json:"user_id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	ID        uuid.UUID `json:"id"`
}

func (cfg *apiConfig) handleGetChirps(w http.ResponseWriter, r *http.Request) {
	chirps, err := cfg.dbQueries.GetChirps(r.Context())
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Could not get chirps")
		return
	}
	chirpsJSON := []Chirp{}
	for _, chirp := range chirps {
		chirpsJSON = append(chirpsJSON, Chirp{
			ID:        chirp.ID,
			Body:      chirp.Body,
			UserID:    chirp.UserID,
			CreatedAt: chirp.CreatedAt,
			UpdatedAt: chirp.UpdatedAt,
		})
	}
	respondWithJSON(w, http.StatusOK, chirpsJSON)
}

func (cfg *apiConfig) handleGetChirp(w http.ResponseWriter, r *http.Request) {
	chirpID := r.PathValue("chirpID")
	if chirpID == "" {
		respondWithError(w, http.StatusBadRequest, "Chirp ID is required")
		return
	}
	chirpIDUUID, err := uuid.Parse(chirpID)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid Chirp ID")
		return
	}
	chirp, err := cfg.dbQueries.GetChirp(r.Context(), chirpIDUUID)
	if err != nil {
		respondWithError(w, http.StatusNotFound, "Chirp not found")
		return
	}
	respondWithJSON(w, http.StatusOK, Chirp{
		ID:        chirp.ID,
		Body:      chirp.Body,
		UserID:    chirp.UserID,
		CreatedAt: chirp.CreatedAt,
		UpdatedAt: chirp.UpdatedAt,
	})
}

func (cfg *apiConfig) handleCreateChirp(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Could not read request body")
		return
	}
	defer r.Body.Close()

	chirp := Chirp{}
	err = json.Unmarshal(body, &chirp)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid JSON body")
		return
	}

	if len(chirp.Body) > 140 {
		respondWithError(w, http.StatusBadRequest, "Chirp is too long")
		return
	}
	dbChirp, err := cfg.dbQueries.CreateChirp(r.Context(), database.CreateChirpParams{
		Body:   chirp.Body,
		UserID: chirp.UserID,
	})
	chirp.clean()
	respondWithJSON(w, http.StatusCreated, Chirp{
		ID:        dbChirp.ID,
		Body:      dbChirp.Body,
		UserID:    dbChirp.UserID,
		CreatedAt: dbChirp.CreatedAt,
		UpdatedAt: dbChirp.UpdatedAt,
	})

}

func (c *Chirp) clean() {
	words := strings.Fields(c.Body)
	profanities := []string{"kerfuffle", "sharbert", "fornax"}
	for i, word := range words {
		if slices.Contains(profanities, strings.ToLower(word)) {
			words[i] = "****"
		}
	}
	c.Body = strings.Join(words, " ")
}

func respondWithError(w http.ResponseWriter, code int, msg string) {
	w.WriteHeader(code)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"error": msg})
}

func respondWithJSON(w http.ResponseWriter, code int, payload interface{}) {
	w.WriteHeader(code)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(payload)
}

type User struct {
	ID        uuid.UUID `json:"id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	Email     string    `json:"email"`
}

func (cfg *apiConfig) handleCreateUser(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Could not read request body")
		return
	}
	defer r.Body.Close()

	var e struct {
		Email string `json:"email"`
	}
	err = json.Unmarshal(body, &e)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid JSON body")
		return
	}
	if e.Email == "" {
		respondWithError(w, http.StatusBadRequest, "Email is required")
		return
	}

	user, err := cfg.dbQueries.CreateUser(r.Context(), e.Email)
	if err != nil {
		if pqErr, ok := err.(*pq.Error); ok && pqErr.Code == "23505" {
			respondWithError(w, http.StatusConflict, "User already exists")
		} else {
			respondWithError(w, http.StatusInternalServerError, "Could not create user")
		}
		return
	}

	respondWithJSON(w, http.StatusCreated, User{
		ID:        user.ID,
		CreatedAt: user.CreatedAt,
		UpdatedAt: user.UpdatedAt,
		Email:     user.Email,
	})
}
