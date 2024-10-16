package main

import (
	"encoding/json"
	"fmt"
	"log"
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
	serveMux.HandleFunc("GET /api/healthz", healthHandler)

	// Validate handler
	serveMux.HandleFunc("POST /api/validate_chirp", validateHandler)

	// Serve root files directory at /app/
	fileServer := http.FileServer(http.Dir("./"))
	serveMux.Handle("/app/", apiCfg.middlewareMetricsInc(http.StripPrefix("/app/", fileServer)))

	// Metrics handler to show the number of hits
	serveMux.HandleFunc("GET /admin/metrics", apiCfg.metricsHandler)

	// Reset handler to reset the hit counter
	serveMux.HandleFunc("POST /admin/reset", apiCfg.resetHandler)

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

// Custom handler for /validate_chirp endpoint
func validateHandler(w http.ResponseWriter, r *http.Request) {
	// Set Content-Type header
	w.Header().Set("Content-Type", "application/json")

	type parameters struct {
		// these tags indicate how the keys in the JSON should be mapped to the struct fields
		// the struct fields must be exported (start with a capital letter) if you want them parsed
		Body string `json:"body"`
	}
	decoder := json.NewDecoder(r.Body)
	params := parameters{}
	err := decoder.Decode(&params)
	if err != nil {
		// an error will be thrown if the JSON is invalid or has the wrong types
		// any missing fields will simply have their values in the struct set to their zero value
		log.Printf("Error decoding parameters: %s", err)

		w.WriteHeader(http.StatusInternalServerError)
		_, err = w.Write([]byte("{\n  \"error\": \"Something went wrong\"\n}"))

		if err != nil {
			fmt.Println(err)
		}
	}

	if len(params.Body) > 140 {
		log.Println("Chirp is too long")

		w.WriteHeader(http.StatusBadRequest)
		_, err = w.Write([]byte("{\n  \"error\": \"Chirp is too long\"\n}"))

		if err != nil {
			fmt.Println(err)
		}
	} else {
		// Write 200 OK status
		w.WriteHeader(http.StatusOK)

		// Write the body message "OK"
		_, err = w.Write([]byte("{\n  \"valid\":true\n}"))
		if err != nil {
			fmt.Println(err)
		}
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
	w.Header().Set("Content-Type", "text/html; charset=utf-8")

	// Get the current hit count
	hits := cfg.fileserverHits.Load()

	// Write the hit count in plain text
	_, err := w.Write([]byte(fmt.Sprintf("<html>\n  <body>\n    <h1>Welcome, Chirpy Admin</h1>\n    <p>Chirpy has been visited %d times!</p>\n  </body>\n</html>", hits)))
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
