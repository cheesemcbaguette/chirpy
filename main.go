package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"sync/atomic"
)

// List of profane words
var profaneWords = []string{"kerfuffle", "sharbert", "fornax"}

type apiConfig struct {
	fileserverHits atomic.Int32
}

// Chirp represents the incoming request with a "body" field
type Chirp struct {
	Body string `json:"body"`
}

// CleanedChirp represents the outgoing response with a cleaned body
type CleanedChirp struct {
	CleanedBody string `json:"cleaned_body"`
}

// Helper function to respond with JSON
func respondWithJSON(w http.ResponseWriter, code int, payload interface{}) {
	response, _ := json.Marshal(payload)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	_, err := w.Write(response)
	if err != nil {
		fmt.Println(err)
	}
}

// Helper function to respond with an error message
func respondWithError(w http.ResponseWriter, code int, msg string) {
	respondWithJSON(w, code, map[string]string{"error": msg})
}

// Handler to validate and clean chirp
func validateChirpHandler(w http.ResponseWriter, r *http.Request) {
	var chirp Chirp

	// Parse the incoming request body
	if err := json.NewDecoder(r.Body).Decode(&chirp); err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	if len(chirp.Body) > 140 {
		respondWithError(w, http.StatusBadRequest, "Chirp is too long")
		return
	}

	// Clean the chirp by replacing profane words
	cleanedBody := cleanProfaneWords(chirp.Body)

	// Create the response struct with the cleaned body
	cleanedChirp := CleanedChirp{
		CleanedBody: cleanedBody,
	}

	// Return the cleaned chirp as JSON
	respondWithJSON(w, http.StatusOK, cleanedChirp)
}

// Separate function to replace profane words
func cleanProfaneWords(text string) string {
	words := strings.Split(text, " ")
	for i, word := range words {
		for _, badWord := range profaneWords {
			// Check if the word matches without punctuation
			lowerWord := strings.ToLower(word)
			if lowerWord == badWord {
				words[i] = "****"
				break
			}
		}
	}
	return strings.Join(words, " ")
}

func main() {
	// Create a new apiConfig to track state
	apiCfg := &apiConfig{}

	serveMux := http.NewServeMux()

	// Health check handler for /healthz
	serveMux.HandleFunc("GET /api/healthz", healthHandler)

	// Validate handler
	serveMux.HandleFunc("POST /api/validate_chirp", validateChirpHandler)

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
