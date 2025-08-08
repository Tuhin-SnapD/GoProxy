//go:build testserver
// +build testserver

package main

import (
	"fmt"
	"log"
	"net/http"
	"time"
)

func main() {
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/plain")
		w.WriteHeader(http.StatusOK)
		fmt.Fprintf(w, "Hello from backend server!\n")
		fmt.Fprintf(w, "Time: %s\n", time.Now().Format(time.RFC3339))
		fmt.Fprintf(w, "Request: %s %s\n", r.Method, r.URL.Path)
		fmt.Fprintf(w, "User-Agent: %s\n", r.UserAgent())
	})

	http.HandleFunc("/api/data", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		fmt.Fprintf(w, `{"message": "API data", "timestamp": "%s", "path": "%s"}`, 
			time.Now().Format(time.RFC3339), r.URL.Path)
	})

	http.HandleFunc("/slow", func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(2 * time.Second) // Simulate slow response
		w.Header().Set("Content-Type", "text/plain")
		w.WriteHeader(http.StatusOK)
		fmt.Fprintf(w, "Slow response after 2 seconds\n")
	})

	fmt.Println("Starting test backend server on :8081")
	fmt.Println("Available endpoints:")
	fmt.Println("  - GET / (basic response)")
	fmt.Println("  - GET /api/data (JSON response)")
	fmt.Println("  - GET /slow (slow response for testing)")
	
	log.Fatal(http.ListenAndServe(":8081", nil))
} 