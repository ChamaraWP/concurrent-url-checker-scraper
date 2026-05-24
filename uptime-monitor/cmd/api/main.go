package main

import (
	"fmt"
	"log"
	"net/http"
	"sync"

	"concurrent-url-checker-scraper/uptime-monitor/internal/monitor"
	"concurrent-url-checker-scraper/uptime-monitor/internal/server"
)

func main() {
	// 1. Create channels for our system pipeline
	jobs := make(chan string, 100)
	results := make(chan string, 100)

	var wg sync.WaitGroup
	numberOfWorkers := 3

	for w := 1; w <= numberOfWorkers; w++ {
		wg.Add(1)
		go monitor.StartWorker(w, jobs, results, &wg) // Capitalized means Public/Exported
	}

	// 3. Start a background manager to print results out of the channel
	go func() {
		for res := range results {
			log.Println("[MONITOR SYSTEM]", res)
		}
	}()

	// 4. Initialize our HTTP server routes, passing it the 'jobs' channel
	// so the API endpoints can feed the workers dynamically.
	mux := server.RegisterRoutes(jobs)

	fmt.Println("🚀 Uptime Monitor Service starting on port :8080...")

	if err := http.ListenAndServe(":8080", mux); err != nil {
		log.Fatalf("Server failed to start: %v", err)
	}
}
