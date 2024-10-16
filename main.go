package main

import (
	"fmt"
	"net/http"
)

func main() {
	serveMux := http.NewServeMux()

	// Health check handler for /healthz
	serveMux.HandleFunc("/healthz", healthHandler)

	// Serve root files directory at /app/
	fileServer := http.FileServer(http.Dir("./"))
	serveMux.Handle("/app/", http.StripPrefix("/app/", fileServer))

	server := &http.Server{
		Addr:    ":8080",  // Bind to localhost:8080
		Handler: serveMux, // Use the ServeMux as the handler
	}

	// Start the server
	err := server.ListenAndServe()
	if err != nil {
		fmt.Println(err)
	}
}

// Custom handler for /healthz endpoint
func healthHandler(w http.ResponseWriter, r *http.Request) {
	// Set Content-Type header
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")

	// Write 200 OK status
	w.WriteHeader(http.StatusOK)

	// Write the body message "OK"
	_, err := w.Write([]byte("OK"))
	if err != nil {
		fmt.Println(err)
	}
}
