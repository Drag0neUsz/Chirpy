package main

import (
	"database/sql"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/Drag0neUsz/Chirpy/internal/database"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
)

func main() {
	godotenv.Load()
	dbURL := os.Getenv("DB_URL")
	if dbURL == "" {
		log.Fatal("DB_URL is not set")
	}

	db, err := sql.Open("postgres", dbURL)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	config := apiConfig{
		dbQueries: database.New(db),
		platform:  strings.ToLower(os.Getenv("PLATFORM")),
	}
	mux := http.NewServeMux()

	port := "8080"

	s := http.Server{
		Addr:    ":" + port,
		Handler: mux,
	}

	mux.Handle("/app/", config.middlewareMetricsInc(http.StripPrefix("/app/", http.FileServer(http.Dir("./")))))

	mux.HandleFunc("GET /api/healthz", handleHealthz)

	//Admin
	mux.HandleFunc("GET /admin/metrics", config.handleMetrics)
	mux.HandleFunc("POST /admin/reset", config.handleReset)

	//API
	mux.HandleFunc("POST /api/users", config.handleCreateUser)
	mux.HandleFunc("GET /api/chirps", config.handleGetChirps)
	mux.HandleFunc("GET /api/chirps/{chirpID}", config.handleGetChirp)
	mux.HandleFunc("POST /api/chirps", config.handleCreateChirp)
	log.Fatal(s.ListenAndServe())
}
