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

	"github.com/Drag0neUsz/Chirpy/internal/auth"
	"github.com/Drag0neUsz/Chirpy/internal/database"
	"github.com/lib/pq"
)

const accessTokenDuration = time.Hour

type apiConfig struct {
	fileserverHits atomic.Int32
	dbQueries      *database.Queries
	platform       string
	tokenSecret    string
	polkaKey       string
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
	order := strings.ToLower(r.URL.Query().Get("sort"))
	if order != "" {
		if order != "desc" && order != "asc" {
			respondWithError(w, http.StatusBadRequest, "Invalid sort order")
			return
		}
	}
	authorID := r.URL.Query().Get("author_id")
	var chirps []database.Chirp
	var err error
	if authorID != "" {
		authorIDUUID, err := uuid.Parse(authorID)
		if err != nil {
			respondWithError(w, http.StatusBadRequest, "Invalid author ID")
			return
		}
		chirps, err = cfg.dbQueries.GetChirpsByAuthorID(r.Context(), authorIDUUID)
	} else {
		chirps, err = cfg.dbQueries.GetChirps(r.Context())
	}
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
	if order == "desc" {
		slices.Reverse(chirpsJSON)
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
	token, err := auth.GetBearerToken(r.Header)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}
	userID, err := auth.ValidateJWT(token, cfg.tokenSecret)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}
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
		UserID: userID,
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

func (cfg *apiConfig) handleDeleteChirp(w http.ResponseWriter, r *http.Request) {
	chirpID := r.PathValue("chirpID")
	if chirpID == "" {
		respondWithError(w, http.StatusBadRequest, "Chirp ID is required")
		return
	}
	token, err := auth.GetBearerToken(r.Header)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}
	userID, err := auth.ValidateJWT(token, cfg.tokenSecret)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Unauthorized")
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
	if chirp.UserID != userID {
		respondWithError(w, http.StatusForbidden, "Forbidden")
		return
	}
	err = cfg.dbQueries.DeleteChirp(r.Context(), chirpIDUUID)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Could not delete chirp")
		return
	}
	w.WriteHeader(http.StatusNoContent)
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
	ID          uuid.UUID `json:"id"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
	Email       string    `json:"email"`
	IsChirpyRed bool      `json:"is_chirpy_red"`
}

type UserResponseWithToken struct {
	User
	Token        string `json:"token"`
	RefreshToken string `json:"refresh_token"`
}

func (cfg *apiConfig) handleRegisterUser(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Could not read request body")
		return
	}
	defer r.Body.Close()

	var e struct {
		Email    string `json:"email"`
		Password string `json:"password"`
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
	if e.Password == "" {
		respondWithError(w, http.StatusBadRequest, "Password is required")
		return
	}
	hashedPassword, err := auth.HashPassword(e.Password)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Could not hash password")
		return
	}
	user, err := cfg.dbQueries.CreateUser(r.Context(), database.CreateUserParams{
		Email:          e.Email,
		HashedPassword: hashedPassword,
	})
	if err != nil {
		if pqErr, ok := err.(*pq.Error); ok && pqErr.Code == "23505" {
			respondWithError(w, http.StatusConflict, "User already exists")
		} else {
			respondWithError(w, http.StatusInternalServerError, "Could not create user")
		}
		return
	}
	token, err := auth.MakeJWT(user.ID, cfg.tokenSecret, accessTokenDuration)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Could not make JWT")
		return
	}
	refreshToken, err := auth.MakeRefreshToken()
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Could not make refresh token")
		return
	}
	_, err = cfg.dbQueries.CreateRefreshToken(r.Context(), database.CreateRefreshTokenParams{
		Token:  refreshToken,
		UserID: user.ID,
	})
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Could not create refresh token")
		return
	}
	respondWithJSON(w, http.StatusCreated, UserResponseWithToken{
		User: User{
			ID:          user.ID,
			CreatedAt:   user.CreatedAt,
			UpdatedAt:   user.UpdatedAt,
			Email:       user.Email,
			IsChirpyRed: user.IsChirpyRed,
		},
		Token:        token,
		RefreshToken: refreshToken,
	})
}

func (cfg *apiConfig) handleLoginUser(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Could not read request body")
		return
	}
	defer r.Body.Close()

	var e struct {
		Email    string `json:"email"`
		Password string `json:"password"`
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
	if e.Password == "" {
		respondWithError(w, http.StatusBadRequest, "Password is required")
		return
	}
	user, err := cfg.dbQueries.GetUser(r.Context(), e.Email)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Invalid email or password")
		return
	}
	valid, err := auth.CheckPasswordHash(e.Password, user.HashedPassword)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Could not check password")
		return
	}
	if !valid {
		respondWithError(w, http.StatusUnauthorized, "Invalid email or password")
		return
	}
	token, err := auth.MakeJWT(user.ID, cfg.tokenSecret, accessTokenDuration)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Could not make JWT")
		return
	}
	refreshToken, err := auth.MakeRefreshToken()
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Could not make refresh token")
		return
	}
	_, err = cfg.dbQueries.CreateRefreshToken(r.Context(), database.CreateRefreshTokenParams{
		Token:  refreshToken,
		UserID: user.ID,
	})
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Could not create refresh token")
		return
	}
	respondWithJSON(w, http.StatusOK, UserResponseWithToken{
		User: User{
			ID:          user.ID,
			CreatedAt:   user.CreatedAt,
			UpdatedAt:   user.UpdatedAt,
			Email:       user.Email,
			IsChirpyRed: user.IsChirpyRed,
		},
		Token:        token,
		RefreshToken: refreshToken,
	})
}

func (cfg *apiConfig) handleUpdateUser(w http.ResponseWriter, r *http.Request) {

	body, err := io.ReadAll(r.Body)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Could not read request body")
		return
	}
	defer r.Body.Close()
	token, err := auth.GetBearerToken(r.Header)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}
	var e struct {
		Email    string `json:"email"`
		Password string `json:"password"`
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
	if e.Password == "" {
		respondWithError(w, http.StatusBadRequest, "Password is required")
		return
	}
	userID, err := auth.ValidateJWT(token, cfg.tokenSecret)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}
	hashedPassword, err := auth.HashPassword(e.Password)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Could not hash password")
		return
	}
	user, err := cfg.dbQueries.UpdateUser(r.Context(), database.UpdateUserParams{
		ID:             userID,
		Email:          e.Email,
		HashedPassword: hashedPassword,
	})
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Could not update user")
		return
	}
	respondWithJSON(w, http.StatusOK, User{
		ID:          user.ID,
		Email:       user.Email,
		CreatedAt:   user.CreatedAt,
		UpdatedAt:   user.UpdatedAt,
		IsChirpyRed: user.IsChirpyRed,
	})
}

type polkaWebhook struct {
	Event string `json:"event"`
	Data  struct {
		UserID string `json:"user_id"`
	} `json:"data"`
}

func (cfg *apiConfig) handlePolkaWebhook(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Could not read request body")
		return
	}
	defer r.Body.Close()
	apiKey, err := auth.GetAPIKey(r.Header)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}
	if apiKey != cfg.polkaKey {
		respondWithError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}
	var p polkaWebhook
	err = json.Unmarshal(body, &p)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid JSON body")
		return
	}
	if p.Event != "user.upgraded" {
		w.WriteHeader(http.StatusNoContent)
		return
	}
	userID, err := uuid.Parse(p.Data.UserID)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid user ID")
		return
	}
	_, err = cfg.dbQueries.UpdateUserIsChirpyRed(r.Context(), database.UpdateUserIsChirpyRedParams{
		ID:          userID,
		IsChirpyRed: true,
	})
	if err != nil {
		if pqErr, ok := err.(*pq.Error); ok && pqErr.Code == "23505" {
			respondWithError(w, http.StatusNotFound, "User not found")
		} else {
			respondWithError(w, http.StatusInternalServerError, "Could not update user")
		}
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (cfg *apiConfig) handleRefresh(w http.ResponseWriter, r *http.Request) {
	refreshToken, err := auth.GetBearerToken(r.Header)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}
	dataRow, err := cfg.dbQueries.GetUserIDFromRefreshToken(r.Context(), refreshToken)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}
	if dataRow.RevokedAt.Valid {
		respondWithError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}
	if dataRow.ExpiresAt.Before(time.Now()) {
		respondWithError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}
	accessToken, err := auth.MakeJWT(dataRow.UserID, cfg.tokenSecret, accessTokenDuration)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}
	respondWithJSON(w, http.StatusOK, map[string]string{"token": accessToken})
}

func (cfg *apiConfig) handleRevoke(w http.ResponseWriter, r *http.Request) {
	refreshToken, err := auth.GetBearerToken(r.Header)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}
	err = cfg.dbQueries.RevokeRefreshToken(r.Context(), refreshToken)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Could not revoke refresh token")
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
