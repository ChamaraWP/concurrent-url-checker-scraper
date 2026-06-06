package server

import (
	"bufio"
	"encoding/json"
	"net/http"
	"os"
)

// RequestBody defines the JSON structure we expect from the user
type RequestBody struct {
	URL string `json:"url"`
}

// LogEntry defines the JSON structure we return to the browser user
type LogEntry struct {
	Message string `json:"log"`
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

	// WIRING POINT: Endpoint 2: Read Log History (GET)
	mux.HandleFunc("/api/logs", func(w http.ResponseWriter, r *http.Request) {
		// 1. Open the file strictly in Read-Only mode so it doesn't collide with the writer
		file, err := os.Open("urls.txt")
		if err != nil {
			// If file doesn't exist yet, return an empty array instead of crashing
			w.Header().Set("Content-Type", "application/json")
			w.Write([]byte("[]"))
			return
		}
		defer file.Close() // Safely close the file read gate when function completes

		var logs []LogEntry

		// 2. Efficiently read the file line-by-line using a Scanner
		// (Prevents loading a huge 1GB file into RAM all at once)
		scanner := bufio.NewScanner(file)
		for scanner.Scan() {
			logs = append(logs, LogEntry{Message: scanner.Text()})
		}

		if err := scanner.Err(); err != nil {
			http.Error(w, "failed to read log file", http.StatusInternalServerError)
			return
		}

		// 3. Serialize our logs slice into a clean JSON layout and send it out
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(logs)

	})

	return mux
}
