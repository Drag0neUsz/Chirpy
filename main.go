package main

import (
	"fmt"
	"log"
	"net/http"
)

func main() {
	mux := http.NewServeMux()

	port := "8080"

	s := http.Server{
		Addr:    ":" + port,
		Handler: mux,
	}

	mux.Handle("/app/", http.StripPrefix("/app/", http.FileServer(http.Dir("./"))))

	mux.HandleFunc("/healthz", handleHealthz)
	mux.HandleFunc("/test", func(w http.ResponseWriter, r *http.Request) { fmt.Println("Hello, World!") })
	log.Fatal(s.ListenAndServe())
}
