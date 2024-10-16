package main

import (
	"fmt"
	"net/http"
	"sync/atomic"
)

type apiConfig struct {
	fileserverHits atomic.Int32
}

func main() {
	// Create a new apiConfig to track state
	apiCfg := &apiConfig{}

	serveMux := http.NewServeMux()

	// Health check handler for /healthz
	serveMux.HandleFunc("GET /healthz", healthHandler)

	// Serve root files directory at /app/
	fileServer := http.FileServer(http.Dir("./"))
	serveMux.Handle("/app/", apiCfg.middlewareMetricsInc(http.StripPrefix("/app/", fileServer)))

	// Metrics handler to show the number of hits
	serveMux.HandleFunc("GET /metrics", apiCfg.metricsHandler)

	// Reset handler to reset the hit counter
	serveMux.HandleFunc("POST /reset", apiCfg.resetHandler)

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

// Middleware to count hits to the fileserver
func (cfg *apiConfig) middlewareMetricsInc(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Increment the counter
		cfg.fileserverHits.Add(1)
		// Call the next handler
		next.ServeHTTP(w, r)
	})
}

// Handler for /metrics to show the number of hits
func (cfg *apiConfig) metricsHandler(w http.ResponseWriter, r *http.Request) {
	// Set Content-Type header
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")

	// Get the current hit count
	hits := cfg.fileserverHits.Load()

	// Write the hit count in plain text
	_, err := w.Write([]byte(fmt.Sprintf("Hits: %d\n", hits)))
	if err != nil {
		fmt.Println(err)
	}
}

// Handler for /reset to reset the hit counter
func (cfg *apiConfig) resetHandler(w http.ResponseWriter, r *http.Request) {
	// Reset the hit counter to 0
	cfg.fileserverHits.Store(0)

	// Set Content-Type header
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")

	// Confirm the reset
	_, err := w.Write([]byte("Hits reset to 0\n"))
	if err != nil {
		fmt.Println(err)
	}
}
