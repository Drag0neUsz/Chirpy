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

	mux.HandleFunc("GET /healthz", handleHealthz)

	mux.HandleFunc("GET /metrics", config.handleMetrics)
	mux.HandleFunc("POST /reset", config.handleReset)

	log.Fatal(s.ListenAndServe())
}
