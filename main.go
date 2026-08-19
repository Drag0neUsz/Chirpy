package main

import (
	"log"
	"net/http"
)

func main() {
	config := apiConfig{}
	mux := http.NewServeMux()

	port := "8080"

	s := http.Server{
		Addr:    ":" + port,
		Handler: mux,
	}

	mux.Handle("/app/", config.middlewareMetricsInc(http.StripPrefix("/app/", http.FileServer(http.Dir("./")))))

	mux.HandleFunc("GET /api/healthz", handleHealthz)
	mux.HandleFunc("POST /api/validate_chirp", handleValidateChirp)

	mux.HandleFunc("GET /admin/metrics", config.handleMetrics)
	mux.HandleFunc("POST /admin/reset", config.handleReset)

	log.Fatal(s.ListenAndServe())
}
