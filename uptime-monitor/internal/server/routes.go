package server

import (
	"encoding/json"
	"net/http"
)

// RequestBody defines the JSON structure we expect from the user
type RequestBody struct {
	URL string `json:"url"`
}

// RegisterRoutes sets up our API endpoints.
// It accepts a write-only channel so the API can drop jobs into the pipeline.
func RegisterRoutes(jobs chan<- string) *http.ServeMux {
	mux := http.NewServeMux()

	mux.HandleFunc("/api/check", func(w http.ResponseWriter, r *http.Request) {
		var body RequestBody

		err := json.NewDecoder(r.Body).Decode(&body)
		if err != nil || body.URL == "" {
			http.Error(w, "Invalid request body. Please provide a JSON object with a 'url' field.", http.StatusBadRequest)
			return
		}

		jobs <- body.URL // Send the URL to the workers via the channel

		// Respond immediately to the client so they don't wait for the network check to finish
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusAccepted)
		json.NewEncoder(w).Encode(map[string]string{
			"message": "URL queued for checking successfully.",
			"url":     body.URL,
		})
	})

	return mux
}
